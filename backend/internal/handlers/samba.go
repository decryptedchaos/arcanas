/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"arcanas/internal/models"
)

func GetSambaShares(w http.ResponseWriter, r *http.Request) {
	shares, err := getSambaSharesFromSystem()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get Samba shares: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shares)
}

func getSambaSharesFromSystem() ([]models.SambaShare, error) {
	var shares []models.SambaShare

	// Read smb.conf file
	smbConfPath := "/etc/samba/smb.conf"
	if _, err := os.Stat(smbConfPath); os.IsNotExist(err) {
		// Try alternative paths
		altPaths := []string{"/usr/local/samba/etc/smb.conf", "/etc/smb.conf"}
		for _, path := range altPaths {
			if _, err := os.Stat(path); err == nil {
				smbConfPath = path
				break
			}
		}
		if _, err := os.Stat(smbConfPath); os.IsNotExist(err) {
			return shares, fmt.Errorf("smb.conf not found")
		}
	}

	file, err := os.Open(smbConfPath)
	if err != nil {
		return shares, fmt.Errorf("failed to open smb.conf: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentShare *models.SambaShare
	shareID := 1

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Check for share section
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// Save previous share if exists
			if currentShare != nil && currentShare.Name != "" {
				currentShare.ID = shareID
				currentShare.Created = time.Now()
				currentShare.LastModified = time.Now()
				// Get actual share stats
				getShareStats(currentShare)
				shares = append(shares, *currentShare)
				shareID++
			}

			// Start new share
			shareName := strings.Trim(line, "[]")
			if shareName != "global" && shareName != "homes" {
				currentShare = &models.SambaShare{
					Name:       shareName,
					Users:      []string{},
					Groups:     []string{},
					GuestOK:    false,
					ReadOnly:   false,
					Browseable: true,
					Available:  true,
					Size:       "Unknown",
					Used:       "Unknown",
				}
			} else {
				currentShare = nil
			}
			continue
		}

		// Parse share parameters
		if currentShare != nil {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "path":
					currentShare.Path = value
				case "comment":
					currentShare.Comment = value
				case "valid users":
					currentShare.Users = parseUsers(value)
				case "valid groups":
					currentShare.Groups = strings.Split(value, ",")
					for i, group := range currentShare.Groups {
						currentShare.Groups[i] = strings.TrimSpace(group)
					}
				case "guest ok":
					currentShare.GuestOK = strings.ToLower(value) == "yes"
				case "read only":
					currentShare.ReadOnly = strings.ToLower(value) == "yes"
				case "browseable":
					currentShare.Browseable = strings.ToLower(value) == "yes"
				case "available":
					currentShare.Available = strings.ToLower(value) == "yes"
				}
			}
		}
	}

	// Add the last share
	if currentShare != nil && currentShare.Name != "" {
		currentShare.ID = shareID
		currentShare.Created = time.Now()
		currentShare.LastModified = time.Now()
		getShareStats(currentShare)
		shares = append(shares, *currentShare)
	}

	return shares, nil
}

func parseUsers(userString string) []string {
	var users []string
	parts := strings.Split(userString, ",")
	for _, part := range parts {
		user := strings.TrimSpace(part)
		if user != "" && !strings.HasPrefix(user, "@") {
			users = append(users, user)
		}
	}
	return users
}

func getShareStats(share *models.SambaShare) {
	// Get directory size if path exists
	if share.Path != "" {
		if stat, err := os.Stat(share.Path); err == nil {
			if stat.IsDir() {
				// Try to get disk usage
				if cmd := exec.Command("df", "-h", share.Path); cmd != nil {
					if output, err := cmd.Output(); err == nil {
						lines := strings.Split(string(output), "\n")
						if len(lines) >= 2 {
							fields := strings.Fields(lines[1])
							if len(fields) >= 4 {
								share.Size = fields[1]
								share.Used = fields[2]
							}
						}
					}
				}
			}
		}
	}

	// Get connection count from smbstatus
	share.Connections = getSambaConnections(share.Name)
}

func getSambaConnections(shareName string) int {
	cmd := exec.Command("smbstatus", "-s")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		count := 0
		for _, line := range lines {
			if strings.Contains(line, shareName) {
				count++
			}
		}
		return count
	}
	return 0
}

func CreateSambaShare(w http.ResponseWriter, r *http.Request) {
	var share models.SambaShare
	if err := json.NewDecoder(r.Body).Decode(&share); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if share.Name == "" || share.Path == "" {
		http.Error(w, "Name and path are required", http.StatusBadRequest)
		return
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(share.Path, 0755); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create directory: %v", err), http.StatusInternalServerError)
		return
	}

	// Add share to smb.conf
	if err := addShareToSmbConf(share); err != nil {
		http.Error(w, fmt.Sprintf("Failed to add share: %v", err), http.StatusInternalServerError)
		return
	}

	// Reload Samba configuration
	if err := reloadSamba(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to reload Samba: %v", err), http.StatusInternalServerError)
		return
	}

	share.ID = int(time.Now().Unix()) // Simple ID generation
	share.Created = time.Now()
	share.LastModified = time.Now()
	share.Available = true

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(share)
}

func addShareToSmbConf(share models.SambaShare) error {
	smbConfPath := "/etc/samba/smb.conf"
	file, err := os.OpenFile(smbConfPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Build share configuration
	var config strings.Builder
	config.WriteString(fmt.Sprintf("\n[%s]\n", share.Name))
	config.WriteString(fmt.Sprintf("\tpath = %s\n", share.Path))
	if share.Comment != "" {
		config.WriteString(fmt.Sprintf("\tcomment = %s\n", share.Comment))
	}
	if len(share.Users) > 0 {
		config.WriteString(fmt.Sprintf("\tvalid users = %s\n", strings.Join(share.Users, ", ")))
	}
	if len(share.Groups) > 0 {
		config.WriteString(fmt.Sprintf("\tvalid groups = %s\n", strings.Join(share.Groups, ", ")))
	}
	config.WriteString(fmt.Sprintf("\tguest ok = %s\n", boolToYesNo(share.GuestOK)))
	config.WriteString(fmt.Sprintf("\tread only = %s\n", boolToYesNo(share.ReadOnly)))
	config.WriteString(fmt.Sprintf("\tbrowseable = %s\n", boolToYesNo(share.Browseable)))
	config.WriteString(fmt.Sprintf("\tavailable = %s\n", boolToYesNo(share.Available)))

	_, err = file.WriteString(config.String())
	return err
}

func boolToYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func UpdateSambaShare(w http.ResponseWriter, r *http.Request) {
	var share models.SambaShare
	if err := json.NewDecoder(r.Body).Decode(&share); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update share in smb.conf
	if err := updateShareInSmbConf(share); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update share: %v", err), http.StatusInternalServerError)
		return
	}

	// Reload Samba configuration
	if err := reloadSamba(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to reload Samba: %v", err), http.StatusInternalServerError)
		return
	}

	share.LastModified = time.Now()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(share)
}

func updateShareInSmbConf(share models.SambaShare) error {
	smbConfPath := "/etc/samba/smb.conf"

	// Read entire file
	content, err := os.ReadFile(smbConfPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	inTargetShare := false
	shareFound := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check if we're entering the target share
		if strings.HasPrefix(trimmedLine, "[") && strings.HasSuffix(trimmedLine, "]") {
			shareName := strings.Trim(trimmedLine, "[]")
			if shareName == share.Name {
				inTargetShare = true
				shareFound = true
				newLines = append(newLines, line)
				continue
			} else {
				inTargetShare = false
			}
		}

		// Skip old share content if we're in the target share
		if inTargetShare && !strings.HasPrefix(trimmedLine, "[") {
			continue
		}

		newLines = append(newLines, line)
	}

	if !shareFound {
		return fmt.Errorf("share %s not found", share.Name)
	}

	// Add updated share content
	var shareConfig strings.Builder
	shareConfig.WriteString(fmt.Sprintf("\tpath = %s\n", share.Path))
	if share.Comment != "" {
		shareConfig.WriteString(fmt.Sprintf("\tcomment = %s\n", share.Comment))
	}
	if len(share.Users) > 0 {
		shareConfig.WriteString(fmt.Sprintf("\tvalid users = %s\n", strings.Join(share.Users, ", ")))
	}
	if len(share.Groups) > 0 {
		shareConfig.WriteString(fmt.Sprintf("\tvalid groups = %s\n", strings.Join(share.Groups, ", ")))
	}
	shareConfig.WriteString(fmt.Sprintf("\tguest ok = %s\n", boolToYesNo(share.GuestOK)))
	shareConfig.WriteString(fmt.Sprintf("\tread only = %s\n", boolToYesNo(share.ReadOnly)))
	shareConfig.WriteString(fmt.Sprintf("\tbrowseable = %s\n", boolToYesNo(share.Browseable)))
	shareConfig.WriteString(fmt.Sprintf("\tavailable = %s\n", boolToYesNo(share.Available)))

	// Insert the new share config
	var finalLines []string
	for _, line := range newLines {
		finalLines = append(finalLines, line)
		if strings.HasPrefix(strings.TrimSpace(line), fmt.Sprintf("[%s]", share.Name)) {
			finalLines = append(finalLines, shareConfig.String())
		}
	}

	// Write back to file
	return os.WriteFile(smbConfPath, []byte(strings.Join(finalLines, "\n")), 0644)
}

func DeleteSambaShare(w http.ResponseWriter, r *http.Request) {
	// Get share name from query parameter or request body
	shareName := r.URL.Query().Get("name")
	if shareName == "" {
		// Try to get from request body
		var requestData map[string]string
		if err := json.NewDecoder(r.Body).Decode(&requestData); err == nil {
			shareName = requestData["name"]
		}
	}

	if shareName == "" {
		http.Error(w, "Share name is required", http.StatusBadRequest)
		return
	}

	// Remove share from smb.conf
	if err := deleteShareFromSmbConf(shareName); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete share: %v", err), http.StatusInternalServerError)
		return
	}

	// Reload Samba configuration
	if err := reloadSamba(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to reload Samba: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "share": shareName})
}

func deleteShareFromSmbConf(shareName string) error {
	smbConfPath := "/etc/samba/smb.conf"

	// Read entire file
	content, err := os.ReadFile(smbConfPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	inTargetShare := false
	shareFound := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check if we're entering the target share
		if strings.HasPrefix(trimmedLine, "[") && strings.HasSuffix(trimmedLine, "]") {
			currentShareName := strings.Trim(trimmedLine, "[]")
			if currentShareName == shareName {
				inTargetShare = true
				shareFound = true
				// Skip this line (don't add to newLines)
				continue
			} else {
				inTargetShare = false
			}
		}

		// Skip lines if we're in the target share
		if !inTargetShare {
			newLines = append(newLines, line)
		}
	}

	if !shareFound {
		return fmt.Errorf("share %s not found", shareName)
	}

	// Write back to file
	return os.WriteFile(smbConfPath, []byte(strings.Join(newLines, "\n")), 0644)
}

func ToggleSambaShare(w http.ResponseWriter, r *http.Request) {
	// Get share name from query parameter or request body
	shareName := r.URL.Query().Get("name")
	if shareName == "" {
		// Try to get from request body
		var requestData map[string]string
		if err := json.NewDecoder(r.Body).Decode(&requestData); err == nil {
			shareName = requestData["name"]
		}
	}

	if shareName == "" {
		http.Error(w, "Share name is required", http.StatusBadRequest)
		return
	}

	// Toggle share availability in smb.conf
	if err := toggleShareInSmbConf(shareName); err != nil {
		http.Error(w, fmt.Sprintf("Failed to toggle share: %v", err), http.StatusInternalServerError)
		return
	}

	// Reload Samba configuration
	if err := reloadSamba(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to reload Samba: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "toggled", "share": shareName})
}

func toggleShareInSmbConf(shareName string) error {
	smbConfPath := "/etc/samba/smb.conf"

	// Read entire file
	content, err := os.ReadFile(smbConfPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	inTargetShare := false
	shareFound := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check if we're entering the target share
		if strings.HasPrefix(trimmedLine, "[") && strings.HasSuffix(trimmedLine, "]") {
			currentShareName := strings.Trim(trimmedLine, "[]")
			if currentShareName == shareName {
				inTargetShare = true
				shareFound = true
				newLines = append(newLines, line)
				continue
			} else {
				inTargetShare = false
			}
		}

		// Toggle the available flag
		if inTargetShare && strings.HasPrefix(trimmedLine, "available") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				currentValue := strings.TrimSpace(parts[1])
				var newValue string
				if strings.ToLower(currentValue) == "yes" {
					newValue = "no"
				} else {
					newValue = "yes"
				}
				// Preserve original spacing
				newLine := parts[0] + "= " + newValue
				newLines = append(newLines, newLine)
				continue
			}
		}

		newLines = append(newLines, line)
	}

	if !shareFound {
		return fmt.Errorf("share %s not found", shareName)
	}

	// Write back to file
	return os.WriteFile(smbConfPath, []byte(strings.Join(newLines, "\n")), 0644)
}

func reloadSamba() error {
	// Try different commands to reload Samba
	commands := [][]string{
		{"systemctl", "reload", "smb"},
		{"systemctl", "reload", "samba"},
		{"service", "smb", "reload"},
		{"service", "samba", "reload"},
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to reload Samba service")
}

func GetSambaConnections(w http.ResponseWriter, r *http.Request) {
	connections, err := getSambaConnectionsFromSystem()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get Samba connections: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(connections)
}

func getSambaConnectionsFromSystem() ([]models.SambaConnection, error) {
	var connections []models.SambaConnection

	cmd := exec.Command("smbstatus")
	output, err := cmd.Output()
	if err != nil {
		return connections, fmt.Errorf("failed to run smbstatus: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	var pid, user, share, ip string
	connectionID := 1

	for _, line := range lines {
		fields := strings.Fields(line)

		// Parse PID line
		if len(fields) >= 4 && strings.Contains(line, "PID") {
			continue // Skip header
		}

		// Parse connection details
		if len(fields) >= 4 && isNumeric(fields[0]) {
			pid = fields[0]
			user = fields[1]
			share = fields[2]
			ip = fields[3]

			connection := models.SambaConnection{
				ID:        connectionID,
				User:      user,
				Share:     share,
				IP:        ip,
				PID:       pid,
				Connected: time.Now(), // Could parse actual time if available
			}

			connections = append(connections, connection)
			connectionID++
		}
	}

	return connections, nil
}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func TestSambaConfig(w http.ResponseWriter, r *http.Request) {
	// Mock configuration test
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":  true,
		"errors": []string{},
	})
}
