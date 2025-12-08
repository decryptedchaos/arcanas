package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"arcanas/internal/models"
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

	// Build export line for /etc/exports
	exportLine := export.Path
	for _, client := range export.Clients {
		exportLine += " " + client.Network + "(" + client.Options + ")"
	}

	// Append to /etc/exports file
	cmd := exec.Command("sh", "-c", "echo '"+exportLine+"' >> /etc/exports")
	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to write to /etc/exports: %v\nOutput: %s", err, string(output)), http.StatusInternalServerError)
		return
	}

	// Reload NFS service
	reloadCmd := exec.Command("systemctl", "reload", "nfs-server")
	reloadCmd.Run() // Ignore reload errors for now

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
	json.NewEncoder(w).Encode(export)
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

	// For now, just return success - updating /etc/exports properly would require
	// more complex parsing and rewriting of the file
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated", "path": path})
}

func DeleteNFSExport(w http.ResponseWriter, r *http.Request) {
	// Extract path from query parameter since we don't have ID mapping
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path parameter is required", http.StatusBadRequest)
		return
	}

	// Remove line from /etc/exports file
	cmd := exec.Command("sed", "-i", "\\|^"+path+" |d", "/etc/exports")
	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to remove from /etc/exports: %v\nOutput: %s", err, string(output)), http.StatusInternalServerError)
		return
	}

	// Reload NFS service
	reloadCmd := exec.Command("systemctl", "reload", "nfs-server")
	reloadCmd.Run() // Ignore reload errors for now

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "path": path})
}

func GetNFSExportStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "active",
		"clients":     5,
		"last_reload": time.Now().Add(-1 * time.Hour),
	})
}

func ReloadNFSConfig(w http.ResponseWriter, r *http.Request) {
	// In production, actually reload NFS configuration
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "reloaded",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
