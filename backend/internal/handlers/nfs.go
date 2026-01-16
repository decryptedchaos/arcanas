/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package handlers

import (
	"arcanas/internal/models"
	"arcanas/internal/utils"
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

func GetNFSExports(w http.ResponseWriter, r *http.Request) {
	exports, err := getRealNFSExports()
	if err != nil {
		fmt.Printf("getRealNFSExports error: %v\n", err)
		exports = []models.NFSExport{}
	}

	// Ensure we always return a valid JSON array
	if exports == nil {
		exports = []models.NFSExport{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(exports); err != nil {
		fmt.Printf("Failed to encode exports: %v\n", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func getRealNFSExports() ([]models.NFSExport, error) {
	var exports []models.NFSExport

	// Read /etc/exports file
	cmd := exec.Command("cat", "/etc/exports")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to read /etc/exports: %v\n", err)
		return nil, fmt.Errorf("failed to read /etc/exports: %v", err)
	}

	fmt.Printf("Raw /etc/exports content: %s\n", string(output))

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fmt.Printf("Processing line: %s\n", line)

		// Parse export line
		parts := strings.Fields(line)
		if len(parts) < 2 {
			fmt.Printf("Skipping line - insufficient parts: %v\n", parts)
			continue
		}

		exportPath := parts[0]
		var clients []models.NFSClient

		for i := 1; i < len(parts); i++ {
			clientSpec := parts[i]

			// Extract network and options
			network := clientSpec
			options := "rw,sync"
			access := "read-write"

			// Check if this contains options in parentheses
			if strings.Contains(clientSpec, "(") && strings.Contains(clientSpec, ")") {
				start := strings.Index(clientSpec, "(")
				end := strings.Index(clientSpec, ")")
				if start != -1 && end != -1 && end > start {
					network = clientSpec[:start]
					options = clientSpec[start+1 : end]

					// Determine access based on options
					if strings.Contains(options, "rw") {
						access = "read-write"
					} else {
						access = "read-only"
					}
				}
			}

			clients = append(clients, models.NFSClient{
				Network: network,
				Options: options,
				Access:  access,
			})
		}

		// Get filesystem info for this path
		filesystem, size, used := getFilesystemInfo(exportPath)

		// Get active connections
		connections := getActiveConnections(exportPath)

		exports = append(exports, models.NFSExport{
			ID:                len(exports) + 1,
			Path:              exportPath,
			Clients:           clients,
			Filesystem:        filesystem,
			Size:              size,
			Used:              used,
			ActiveConnections: connections,
			Created:           time.Now(), // Would need to get from filesystem stats
			LastModified:      time.Now(),
		})
	}

	fmt.Printf("Final exports count: %d\n", len(exports))
	return exports, nil
}

func getFilesystemInfo(path string) (string, string, string) {
	// Get filesystem info using df command
	cmd := exec.Command("df", "-h", path)
	output, err := cmd.Output()
	if err != nil {
		return "unknown", "unknown", "unknown"
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return "unknown", "unknown", "unknown"
	}

	// Parse df output (skip header line)
	fields := strings.Fields(lines[1])
	if len(fields) < 6 {
		return "unknown", "unknown", "unknown"
	}

	filesystem := fields[0]
	size := fields[1]
	used := fields[2]

	return filesystem, size, used
}

func getActiveConnections(exportPath string) int {
	// Get NFS connections using showmount command
	cmd := exec.Command("showmount", "-a")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, exportPath) {
			count++
		}
	}

	return count
}

// getFilesystemType detects the filesystem type of a given path
func getFilesystemType(path string) string {
	cmd := exec.Command("findmnt", "-n", "-o", "FSTYPE", path)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// needsFsid determines if a filesystem requires fsid= option for NFS export
func needsFsid(fsType string) bool {
	switch fsType {
	case "fuse", "fuse.mergerfs", "fuse.sshfs", "fuse.rclone",
		"btrfs", "zfs", "overlay", "aufs":
		return true
	default:
		return false
	}
}

// getNextFsid generates the next available fsid number
func getNextFsid() int {
	cmd := exec.Command("exportfs", "-v")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	maxFsid := -1
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "fsid=") {
			// Extract fsid value
			idx := strings.Index(line, "fsid=")
			if idx != -1 {
				remaining := line[idx+5:]
				var value string
				for i, c := range remaining {
					if c >= '0' && c <= '9' {
						value += string(c)
					} else {
						break
					}
					if i > 4 { // Limit parsing
						break
					}
				}
				if value != "" {
					var fsid int
					fmt.Sscanf(value, "%d", &fsid)
					if fsid > maxFsid {
						maxFsid = fsid
					}
				}
			}
		}
	}
	return maxFsid + 1
}

func CreateNFSExport(w http.ResponseWriter, r *http.Request) {
	var export models.NFSExport
	if err := json.NewDecoder(r.Body).Decode(&export); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if export.Path == "" || len(export.Clients) == 0 {
		http.Error(w, "Path and at least one client are required", http.StatusBadRequest)
		return
	}

	// Check if path needs fsid option (FUSE, btrfs, zfs, etc.)
	fsType := getFilesystemType(export.Path)
	needsID := needsFsid(fsType)
	nextFsid := 0
	if needsID {
		nextFsid = getNextFsid()
		fmt.Printf("Export path %s is on %s filesystem, adding fsid=%d\n", export.Path, fsType, nextFsid)
	}

	// Build export line for /etc/exports
	exportLine := export.Path
	for i, client := range export.Clients {
		options := client.Options
		// Add fsid option if needed and not already present
		if needsID && !strings.Contains(options, "fsid=") {
			if options == "" {
				options = fmt.Sprintf("fsid=%d", nextFsid)
			} else {
				options = fmt.Sprintf("%s,fsid=%d", options, nextFsid)
			}
		}
		exportLine += " " + client.Network + "(" + options + ")"
		_ = i // Use i to avoid unused variable warning
	}

	// Read existing exports file
	cmd := exec.Command("cat", "/etc/exports")
	output, err := cmd.Output()
	var existingContent string
	if err == nil {
		existingContent = string(output)
	}

	// Append new export line
	newContent := existingContent
	if newContent != "" && !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	newContent += exportLine + "\n"

	// Write to /etc/exports file using sudo wrapper
	if err := utils.SudoWriteFile("/etc/exports", newContent); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write to /etc/exports: %v", err), http.StatusInternalServerError)
		return
	}

	// Reload NFS service
	reloadCmd := exec.Command("sudo", "systemctl", "reload", "nfs-server")
	if err := reloadCmd.Run(); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to reload NFS service: %v\n", err)
	}

	// Set response data
	export.ID = 999 // Would get from actual system
	export.Created = time.Now()
	export.LastModified = time.Now()

	// Get real filesystem info
	filesystem, size, used := getFilesystemInfo(export.Path)
	export.Filesystem = filesystem
	export.Size = size
	export.Used = used
	export.ActiveConnections = 0

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(export); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func UpdateNFSExport(w http.ResponseWriter, r *http.Request) {
	// Extract path from query parameter since we don't have ID mapping
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path parameter is required", http.StatusBadRequest)
		return
	}

	var export models.NFSExport
	if err := json.NewDecoder(r.Body).Decode(&export); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(export.Clients) == 0 {
		http.Error(w, "At least one client is required", http.StatusBadRequest)
		return
	}

	// Check if path needs fsid option (FUSE, btrfs, zfs, etc.)
	fsType := getFilesystemType(export.Path)
	needsID := needsFsid(fsType)
	nextFsid := 0
	if needsID {
		nextFsid = getNextFsid()
		fmt.Printf("Export path %s is on %s filesystem, adding fsid=%d\n", export.Path, fsType, nextFsid)
	}

	// Build export line for /etc/exports
	exportLine := export.Path
	for _, client := range export.Clients {
		options := client.Options
		// Add fsid option if needed and not already present
		if needsID && !strings.Contains(options, "fsid=") {
			if options == "" {
				options = fmt.Sprintf("fsid=%d", nextFsid)
			} else {
				options = fmt.Sprintf("%s,fsid=%d", options, nextFsid)
			}
		}
		exportLine += " " + client.Network + "(" + options + ")"
	}

	// Read existing exports file
	cmd := exec.Command("cat", "/etc/exports")
	output, err := cmd.Output()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read /etc/exports: %v", err), http.StatusInternalServerError)
		return
	}

	// Find and replace the export line
	lines := strings.Split(string(output), "\n")
	newLines := []string{}
	replaced := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			if trimmed != "" {
				newLines = append(newLines, line)
			}
			continue
		}

		// Check if this is the export line we want to replace
		if strings.HasPrefix(trimmed, path+" ") || strings.HasPrefix(trimmed, path+"\t") {
			newLines = append(newLines, exportLine)
			replaced = true
		} else {
			newLines = append(newLines, line)
		}
	}

	// If we didn't find and replace the line, append it
	if !replaced {
		newLines = append(newLines, exportLine)
	}

	// Write back to file using sudo wrapper
	if err := utils.SudoWriteFile("/etc/exports", strings.Join(newLines, "\n")+"\n"); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write to /etc/exports: %v", err), http.StatusInternalServerError)
		return
	}

	// Reload NFS service
	reloadCmd := exec.Command("sudo", "systemctl", "reload", "nfs-server")
	if err := reloadCmd.Run(); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to reload NFS service: %v\n", err)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "updated", "path": path}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func DeleteNFSExport(w http.ResponseWriter, r *http.Request) {
	// Extract path from query parameter since we don't have ID mapping
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path parameter is required", http.StatusBadRequest)
		return
	}

	// Read existing exports file
	cmd := exec.Command("cat", "/etc/exports")
	output, err := cmd.Output()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read /etc/exports: %v", err), http.StatusInternalServerError)
		return
	}

	// Filter out the line to delete
	lines := strings.Split(string(output), "\n")
	var newLines []string
	for _, line := range lines {
		// Skip lines that start with the path (but be careful with partial matches)
		if strings.HasPrefix(strings.TrimSpace(line), path) {
			continue
		}
		if strings.TrimSpace(line) != "" {
			newLines = append(newLines, line)
		}
	}

	// Write back to file using sudo wrapper
	if err := utils.SudoWriteFile("/etc/exports", strings.Join(newLines, "\n")+"\n"); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write to /etc/exports: %v", err), http.StatusInternalServerError)
		return
	}

	// Reload NFS service
	reloadCmd := exec.Command("sudo", "systemctl", "reload", "nfs-server")
	if err := reloadCmd.Run(); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to reload NFS service: %v\n", err)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "path": path}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func GetNFSExportStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "active",
		"clients":     5,
		"last_reload": time.Now().Add(-1 * time.Hour),
	}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func ReloadNFSConfig(w http.ResponseWriter, r *http.Request) {
	// In production, actually reload NFS configuration
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":    "reloaded",
		"timestamp": time.Now().Format(time.RFC3339),
	}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}
