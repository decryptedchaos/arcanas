/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"arcanas/internal/models"
	"arcanas/internal/system"
	"arcanas/internal/utils"
)

func GetSCSITargets(w http.ResponseWriter, r *http.Request) {
	// Initialize to empty slice to avoid null JSON
	targets := []models.SCSITarget{}

	// Use targetcli to get targets
	output, err := utils.SudoCombinedOutput("sh", "-c", "echo '/iscsi ls' | sudo targetcli")
	if err != nil {
		// targetcli might not be installed or no targets configured
		// Return empty array instead of error
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(targets)
		return
	}

	lines := strings.Split(string(output), "\n")
	targetID := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "o- ") && strings.Contains(line, "iqn.") {
			// Extract IQN from targetcli output
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				iqn := parts[1]
				target, err := getTargetDetails(iqn, targetID)
				if err == nil {
					targets = append(targets, target)
					targetID++
				}
				// If getting details fails, just skip this target
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(targets); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func getLIOTargets() ([]models.SCSITarget, error) {
	// Initialize to empty slice to avoid null JSON
	targets := make([]models.SCSITarget, 0)

	// Use targetcli to get targets
	output, err := utils.SudoCombinedOutput("sh", "-c", "echo '/iscsi ls' | sudo targetcli")
	if err != nil {
		// targetcli might not be installed or no targets configured
		// Return empty array instead of error
		return targets, nil
	}

	lines := strings.Split(string(output), "\n")
	targetID := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "o- ") && strings.Contains(line, "iqn.") {
			// Extract IQN from targetcli output
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				iqn := parts[1]
				target, err := getTargetDetails(iqn, targetID)
				if err == nil {
					targets = append(targets, target)
					targetID++
				}
				// If getting details fails, just skip this target
			}
		}
	}

	return targets, nil
}

func getTargetDetails(iqn string, targetID int) (models.SCSITarget, error) {
	target := models.SCSITarget{
		ID:           targetID,
		Name:         iqn,
		Status:       "inactive", // Default status
		InitiatorIPs: []string{},  // Initialize to empty slice to prevent null in JSON
		ACLs:         []models.ACL{},
		LUNs:         []models.LUN{},
	}

	// Get target status using correct targetcli path syntax
	output, err := utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1 info' | sudo targetcli", iqn))
	if err == nil {
		if strings.Contains(string(output), "Attribute: enabled") && strings.Contains(string(output), "1") {
			target.Status = "active"
		}
	}

	// Get LUNs and sessions
	if luns, err := getTargetLUNs(iqn); err == nil {
		target.LUNs = luns
		target.LUNCount = len(luns)
		if len(luns) > 0 {
			// Set BackingStore from the first LUN's backing file path
			// Extract device path from backstore path (e.g., "/backstores/block/md0" -> "/dev/md0")
			if luns[0].BackingFile != "" {
				// Get the actual device from the backstore
				backstoreDevice := getDeviceFromBackstore(luns[0].BackingFile)
				if backstoreDevice != "" {
					target.BackingStore = backstoreDevice
				} else {
					// Fallback to showing the backstore path itself
					target.BackingStore = luns[0].BackingFile
				}
			}

			// Calculate total size from LUNs
			var totalSize int64
			for _, lun := range luns {
				totalSize += lun.Size
			}
			target.Size = formatSize(totalSize)
		}
	} // If error getting LUNs, just leave LUNCount as 0 and Size as empty string

	// Get sessions
	if sessions, err := getTargetSessions(iqn); err == nil {
		target.Sessions = len(sessions)
		var initiatorIPs []string
		for _, session := range sessions {
			initiatorIPs = append(initiatorIPs, session.IP)
		}
		target.InitiatorIPs = initiatorIPs
	} // If error getting sessions, just leave Sessions as 0 and InitiatorIPs as empty slice

	// Get ACLs
	if acls, err := getTargetACLs(iqn); err == nil {
		for i := range acls {
			acls[i].TargetID = targetID
		}
		target.ACLs = acls
	} // If error getting ACLs, just leave ACLs as empty slice

	target.Created = time.Now()
	target.LastAccess = time.Now()

	return target, nil
}

// getDeviceFromBackstore extracts the underlying device path from a backstore path
// e.g., "/backstores/block/bs_iqn_2024_01_com_nas_target_123" -> "/dev/md0"
//      "/backstores/block/md0" -> "/dev/md0"
func getDeviceFromBackstore(backstorePath string) string {
	// First, check if the backstore name itself is a device path
	// Format: "/backstores/block/{name}" or "/backstores/fileio/{name}"
	if strings.HasPrefix(backstorePath, "/backstores/block/") {
		backstoreName := strings.TrimPrefix(backstorePath, "/backstores/block/")
		// If the backstore name looks like a device name (e.g., "md0", "sda1")
		// construct the full device path
		if strings.Contains(backstoreName, "md") || strings.Contains(backstoreName, "sd") ||
			strings.Contains(backstoreName, "nvme") || strings.Contains(backstoreName, "vd") {
			return "/dev/" + backstoreName
		}
	}

	// Query targetcli for backstore details to find the underlying device
	output, err := utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '%s info' | sudo targetcli", backstorePath))
	if err != nil {
		return ""
	}

	// Parse the output to find the device path
	// The output contains various attributes, look for the actual device
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for udev_path format (most reliable)
		if strings.Contains(line, "udev_path") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				devicePath := strings.TrimSpace(parts[1])
				// Remove any quotes or special characters
				devicePath = strings.Trim(devicePath, `"' `)
				if strings.HasPrefix(devicePath, "/dev/") {
					return devicePath
				}
			}
		}

		// Check for path attribute
		if strings.Contains(line, "path:") || strings.Contains(line, "Path:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				devicePath := strings.TrimSpace(parts[1])
				devicePath = strings.Trim(devicePath, `"' `)
				if strings.HasPrefix(devicePath, "/dev/") {
					return devicePath
				}
			}
		}

		// Look for any /dev/ path in the line
		if strings.Contains(line, "/dev/") {
			// Extract the device path - split by whitespace and check each part
			parts := strings.Fields(line)
			for _, part := range parts {
				part = strings.Trim(part, `,;)'"`)

				// Check if this part looks like a device path
				if strings.HasPrefix(part, "/dev/") {
					return part
				}
			}
		}
	}

	return ""
}

func getTargetLUNs(iqn string) ([]models.LUN, error) {
	var luns []models.LUN

	output, err := utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1/luns ls' | sudo targetcli", iqn))
	if err != nil {
		return luns, err
	}

	lines := strings.Split(string(output), "\n")
	lunID := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Only match lines starting with "o- lun" followed by a number (e.g., "o- lun0")
		// This avoids matching the header line "o- luns ... [LUNs: 1]"
		if strings.HasPrefix(line, "o- lun") && len(line) > 6 && line[6] >= '0' && line[6] <= '9' {
			// Parse LUN information - format can vary:
			// "o- lun0 ...................................................... [/backstores/block/md0]"
			// "lun0" on its own line followed by path info
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				// Extract LUN number from "lun0"
				lunStr := strings.TrimPrefix(parts[0], "o- ")
				lunStr = strings.TrimPrefix(lunStr, "lun")
				lunNum, _ := strconv.Atoi(lunStr)

				// Extract backstore path - can be in different formats
				backstorePath := ""

				// Method 1: Look for bracketed path at the end
				if strings.Contains(line, "[") && strings.Contains(line, "]") {
					startIdx := strings.Index(line, "[")
					endIdx := strings.LastIndex(line, "]")
					if startIdx >= 0 && endIdx > startIdx {
						backstorePath = line[startIdx+1 : endIdx]
					}
				}

				// Method 2: Check if last field looks like a path
				if backstorePath == "" && len(parts) >= 2 {
					lastPart := parts[len(parts)-1]
					if strings.HasPrefix(lastPart, "/backstores/") {
						backstorePath = strings.Trim(lastPart, "[]")
					}
				}

				// Method 3: Check if any field contains /backstores/
				if backstorePath == "" {
					for _, part := range parts {
						if strings.Contains(part, "/backstores/") {
							backstorePath = strings.Trim(part, "[]")
							break
						}
					}
				}

				lun := models.LUN{
					ID:          lunID,
					TargetID:    0, // Will be set by caller
					LUN:         lunNum,
					BackingFile: backstorePath,
				}

				// Query the LUN individually to get size information
				// Use "info" command on the specific LUN path
				lunOutput, err := utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1/luns/lun%d info' | sudo targetcli", iqn, lunNum))
				if err == nil {
					// Parse size from info output - look for "size:" field
					for _, infoLine := range strings.Split(string(lunOutput), "\n") {
						infoLine = strings.TrimSpace(infoLine)
						if strings.Contains(infoLine, "size:") || strings.Contains(infoLine, "Size:") {
							// Extract size value (e.g., "size: 1073741824" or "size: 1G")
							sizeParts := strings.Fields(infoLine)
							if len(sizeParts) >= 2 {
								sizeStr := strings.Trim(sizeParts[1], ",")
								// Try to parse as bytes first
								if sizeBytes, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
									lun.Size = sizeBytes
								} else {
									// Try to parse human-readable format (e.g., "1G", "512M")
									lun.Size = parseHumanSize(sizeStr)
								}
							}
							break
						}
					}
				}

				luns = append(luns, lun)
				lunID++
			}
		}
	}

	return luns, nil
}

// parseHumanSize converts human-readable size strings (like "1G", "512M") to bytes
func parseHumanSize(sizeStr string) int64 {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	if sizeStr == "" {
		return 0
	}

	// Extract numeric prefix and unit suffix
	var numStr string
	var unit string
	for i, r := range sizeStr {
		if r >= '0' && r <= '9' || r == '.' {
			numStr += string(r)
		} else {
			unit = sizeStr[i:]
			break
		}
	}
	// If no unit found, the whole string is the number
	if unit == "" {
		unit = sizeStr[len(numStr):]
	}

	// Parse the number
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}

	// Convert based on unit
	multiplier := int64(1)
	switch unit {
	case "T", "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "G", "GB":
		multiplier = 1024 * 1024 * 1024
	case "M", "MB":
		multiplier = 1024 * 1024
	case "K", "KB":
		multiplier = 1024
	case "B", "":
		multiplier = 1
	}

	return int64(num * float64(multiplier))
}

func getTargetSessions(iqn string) ([]models.Session, error) {
	var sessions []models.Session

	output, err := utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1/sessions ls' | sudo targetcli", iqn))
	if err != nil {
		return sessions, err
	}

	lines := strings.Split(string(output), "\n")
	sessionID := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "iqn.") {
			// Parse session information
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				session := models.Session{
					ID:        sessionID,
					TargetID:  0, // Will be set by caller
					Initiator: parts[1],
					Connected: time.Now(),
				}
				sessions = append(sessions, session)
				sessionID++
			}
		}
	}

	return sessions, nil
}

func getTargetACLs(iqn string) ([]models.ACL, error) {
	var acls []models.ACL

	output, err := utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1/acls ls' | sudo targetcli", iqn))
	if err != nil {
		return acls, err
	}

	lines := strings.Split(string(output), "\n")
	aclID := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "iqn.") {
			// Parse ACL information
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				acl := models.ACL{
					ID:           aclID,
					TargetID:     0, // Will be set by caller
					InitiatorIQN: parts[1],
					MappedLUNs:   "all", // Default to all LUNs mapped
				}
				acls = append(acls, acl)
				aclID++
			}
		}
	}

	return acls, nil
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func CreateSCSITarget(w http.ResponseWriter, r *http.Request) {
	var target models.SCSITarget
	if err := json.NewDecoder(r.Body).Decode(&target); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate IQN format
	if !strings.HasPrefix(target.Name, "iqn.") {
		http.Error(w, "Invalid IQN format", http.StatusBadRequest)
		return
	}

	// Create LIO target using targetcli
	if err := createLIOTarget(target); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create LIO target: %v", err), http.StatusInternalServerError)
		return
	}

	// Fetch the complete target details including LUNs
	createdTarget, err := getTargetDetails(target.Name, int(time.Now().Unix()))
	if err != nil {
		// Log error but still return basic target info
		fmt.Printf("Warning: failed to fetch target details after creation: %v\n", err)
		target.ID = int(time.Now().Unix())
		target.Created = time.Now()
		target.Status = "active"
	} else {
		createdTarget.ID = int(time.Now().Unix())
		// Preserve initiator_ips from the original request since getTargetDetails won't have them
		createdTarget.InitiatorIPs = target.InitiatorIPs
		target = createdTarget
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(target); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func createLIOTarget(target models.SCSITarget) error {
	// Create iSCSI target using correct targetcli path syntax
	output, err := utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi create %s' | sudo targetcli", target.Name))
	if err != nil {
		return fmt.Errorf("failed to create iSCSI target: %v, output: %s", err, string(output))
	}

	// Enable the target using correct targetcli path syntax
	output, err = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1 set attribute enable=1' | sudo targetcli", target.Name))
	if err != nil {
		return fmt.Errorf("failed to enable target: %v, output: %s", err, string(output))
	}

	// Create LUN with backing store if provided
	if target.BackingStore != "" {
		return createBackstoreAndLUNForTarget(target.Name, target.BackingStore, false)
	}

	return nil
}

// createBackstoreAndLUNForTarget creates a block backstore and attaches it as LUN to a target
func createBackstoreAndLUNForTarget(targetName, backingStore string, useFileio bool) error {
	// Generate a name for the backstore from the target IQN
	backstoreName := strings.ReplaceAll(targetName, ":", "_")
	backstoreName = strings.ReplaceAll(backstoreName, ".", "_")
	backstoreName = "bs_" + backstoreName

	var backstorePath string
	var output []byte
	var err error

	if useFileio {
		// Create a fileio backstore (file-based storage)
		filePath := "/var/lib/arcanas/iscsi/" + backstoreName + ".img"
		_ = utils.SudoRunCommand("mkdir", "-p", "/var/lib/arcanas/iscsi")

		output, err = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/backstores/fileio create %s %s 1G' | sudo targetcli", backstoreName, filePath))
		if err != nil {
			return fmt.Errorf("failed to create fileio backstore: %v, output: %s", err, string(output))
		}
		backstorePath = "/backstores/fileio/" + backstoreName
	} else {
		// Create a block backstore using the device
		output, err = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/backstores/block create %s %s' | sudo targetcli", backstoreName, backingStore))
		if err != nil {
			return fmt.Errorf("failed to create backstore: %v, output: %s", err, string(output))
		}
		backstorePath = "/backstores/block/" + backstoreName
	}

	// Create LUN 0 and link it to the backstore
	output, err = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1/luns create %s' | sudo targetcli", targetName, backstorePath))
	if err != nil {
		// Clean up backstore on failure
		backstoreType := "block"
		if useFileio {
			backstoreType = "fileio"
		}
		_, _ = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/backstores/%s delete %s' | sudo targetcli", backstoreType, backstoreName))
		return fmt.Errorf("failed to create LUN: %v, output: %s", err, string(output))
	}

	return nil
}

func UpdateSCSITarget(w http.ResponseWriter, r *http.Request) {
	var target models.SCSITarget
	if err := json.NewDecoder(r.Body).Decode(&target); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update LIO target
	if err := updateLIOTarget(target); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update LIO target: %v", err), http.StatusInternalServerError)
		return
	}

	target.LastAccess = time.Now()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(target); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func updateLIOTarget(target models.SCSITarget) error {
	// Enable/disable target using correct targetcli path syntax
	enabled := "0"
	if target.Status == "active" {
		enabled = "1"
	}

	output, err := utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1 set attribute enable=%s' | sudo targetcli", target.Name, enabled))
	if err != nil {
		return fmt.Errorf("failed to update target status: %v, output: %s", err, string(output))
	}

	// If a new backing store is specified, create a LUN for it
	if target.BackingStore != "" {
		// Check if there are existing LUNs
		luns, err := getTargetLUNs(target.Name)
		if err == nil && len(luns) > 0 {
			// Delete existing LUNs first (this is a simplification - for production,
			// you might want to support multiple LUNs or selective updates)
			for _, lun := range luns {
				_ = deleteLUN(target.Name, lun.LUN)
			}
		}
		// Create new LUN with the specified backing store
		return createBackstoreAndLUNForTarget(target.Name, target.BackingStore, false)
	}

	return nil
}

// deleteLUN deletes a specific LUN from a target
func deleteLUN(targetName string, lunNumber int) error {
	// First, get the backstore path for this LUN so we can clean it up
	output, err := utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1/luns ls' | sudo targetcli", targetName))
	if err != nil {
		return fmt.Errorf("failed to list LUNs: %v, output: %s", err, string(output))
	}

	// Parse output to find backstore path
	lines := strings.Split(string(output), "\n")
	var backstorePath string
	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf("lun%d", lunNumber)) {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				backstorePath = strings.TrimSuffix(parts[2], "/>") // Remove trailing /> if present
			}
			break
		}
	}

	// Delete the LUN
	output, err = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1/luns delete lun%d' | sudo targetcli", targetName, lunNumber))
	if err != nil {
		return fmt.Errorf("failed to delete LUN: %v, output: %s", err, string(output))
	}

	// Delete the backstore to free up storage
	if backstorePath != "" {
		var backstoreType, backstoreName string
		if strings.Contains(backstorePath, "/backstores/block/") {
			backstoreType = "block"
			backstoreName = strings.TrimPrefix(backstorePath, "/backstores/block/")
		} else if strings.Contains(backstorePath, "/backstores/fileio/") {
			backstoreType = "fileio"
			backstoreName = strings.TrimPrefix(backstorePath, "/backstores/fileio/")
		}

		if backstoreName != "" {
			_, _ = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/backstores/%s delete %s' | sudo targetcli", backstoreType, backstoreName))
		}
	}

	return nil
}

func DeleteSCSITarget(w http.ResponseWriter, r *http.Request) {
	// Get target name from request body
	var requestData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	targetName := requestData["name"]
	if targetName == "" {
		http.Error(w, "Target name is required", http.StatusBadRequest)
		return
	}

	// Delete LIO target
	if err := deleteLIOTarget(targetName); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete LIO target: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "target": targetName}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func deleteLIOTarget(targetName string) error {
	// First, disable the target using correct targetcli path syntax
	output, err := utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1 set attribute enable=0' | sudo targetcli", targetName))
	if err != nil {
		return fmt.Errorf("failed to disable target: %v, output: %s", err, string(output))
	}

	// Delete the target using correct targetcli path syntax
	output, err = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi delete %s' | sudo targetcli", targetName))
	if err != nil {
		return fmt.Errorf("failed to delete target: %v, output: %s", err, string(output))
	}

	return nil
}

func ToggleSCSITarget(w http.ResponseWriter, r *http.Request) {
	// Get target name from request body
	var requestData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	targetName := requestData["name"]
	if targetName == "" {
		http.Error(w, "Target name is required", http.StatusBadRequest)
		return
	}

	// Toggle LIO target
	if err := toggleLIOTarget(targetName); err != nil {
		http.Error(w, fmt.Sprintf("Failed to toggle LIO target: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "toggled", "target": targetName}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func toggleLIOTarget(targetName string) error {
	// Get current status
	output, err := utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1 info' | sudo targetcli", targetName))
	if err != nil {
		return fmt.Errorf("failed to get target info: %v, output: %s", err, string(output))
	}

	// Determine current status and toggle
	currentStatus := "0"
	if strings.Contains(string(output), "Attribute: enabled") && strings.Contains(string(output), "1") {
		currentStatus = "1"
	}

	newStatus := "0"
	if currentStatus == "0" {
		newStatus = "1"
	}

	// Set new status using correct targetcli path syntax
	output, err = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1 set attribute enable=%s' | sudo targetcli", targetName, newStatus))
	if err != nil {
		return fmt.Errorf("failed to toggle target: %v, output: %s", err, string(output))
	}

	return nil
}

func GetSCSISessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := getLIOSessions()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get LIO sessions: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sessions); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func getLIOSessions() ([]models.Session, error) {
	var allSessions []models.Session

	// Get all targets first
	targets, err := getLIOTargets()
	if err != nil {
		return allSessions, err
	}

	// Get sessions for each target
	for _, target := range targets {
		sessions, err := getTargetSessions(target.Name)
		if err == nil {
			for i := range sessions {
				sessions[i].TargetID = target.ID
			}
			allSessions = append(allSessions, sessions...)
		}
	}

	return allSessions, nil
}

// GetAvailableBackingStores returns a list of block devices with their availability status for iSCSI backing stores
// This includes regular partitions and RAID arrays (md devices)
func GetAvailableBackingStores(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Use a map to track unique devices and prevent duplicates
	seenDevices := make(map[string]bool)
	backingStores := make([]models.BackingStore, 0)

	// Get all block devices using lsblk with size information
	// lsblk -J -o NAME,PATH,TYPE,FSTYPE,MOUNTPOINT,SIZE
	cmd := exec.Command("lsblk", "-J", "-o", "NAME,PATH,TYPE,FSTYPE,MOUNTPOINT,SIZE")
	output, err := cmd.Output()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list block devices: %v", err), http.StatusInternalServerError)
		return
	}

	type LsblkDevice struct {
		Name       string        `json:"name"`
		Path       string        `json:"path"`
		Type       string        `json:"type"`
		Fstype     *string       `json:"fstype"`
		Mountpoint *string       `json:"mountpoint"`
		Size       string        `json:"size"`
		Children   []LsblkDevice `json:"children"`
	}

	type LsblkOutput struct {
		Blockdevices []LsblkDevice `json:"blockdevices"`
	}

	var result LsblkOutput
	if err := json.Unmarshal(output, &result); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse block devices: %v", err), http.StatusInternalServerError)
		return
	}

	// Recursively process devices to find suitable backing stores
	var processDevice func(d LsblkDevice)
	processDevice = func(d LsblkDevice) {
		if d.Path == "" {
			for _, child := range d.Children {
				processDevice(child)
			}
			return
		}

		// Skip if we've already seen this device (handles RAID duplicates)
		if seenDevices[d.Path] {
			for _, child := range d.Children {
				processDevice(child)
			}
			return
		}

		// Check if this device is suitable for iSCSI backing store
		var isSuitable bool
		var reason string

		// RAID member devices - skip but process children
		isRAIDMember := d.Fstype != nil && *d.Fstype == "linux_raid_member"
		isLVMMember := d.Fstype != nil && *d.Fstype == "LVM2_member"
		if isRAIDMember || isLVMMember {
			for _, child := range d.Children {
				processDevice(child)
			}
			return
		}

		// Check device type
		switch {
		case d.Type == "md" || strings.HasPrefix(d.Type, "raid"):
			// RAID arrays are suitable
			isSuitable = true
		case d.Type == "lvm":
			// LVM volumes are suitable
			isSuitable = true
		case d.Type == "part" && d.Fstype != nil && *d.Fstype != "" && *d.Fstype != "swap":
			// Partitions with filesystems are suitable
			isSuitable = true
		default:
			// Other device types are not suitable
			isSuitable = false
			reason = "Unsupported device type"
		}

		// Build backing store info
		store := models.BackingStore{
			Path:       d.Path,
			Type:       d.Type,
			Size:       d.Size,
			Available:  isSuitable,
			Reason:     reason,
		}

		// Check mount status
		if d.Mountpoint != nil && *d.Mountpoint != "" {
			store.MountPoint = *d.Mountpoint
			if isSuitable {
				// Device is mounted but will be auto-unmounted when used as iSCSI backing store
				store.Reason = "Will be auto-unmounted from " + *d.Mountpoint
			}
		}

		// Add to result if suitable
		if isSuitable || store.MountPoint != "" {
			seenDevices[d.Path] = true
			backingStores = append(backingStores, store)
		}

		// Process children
		for _, child := range d.Children {
			processDevice(child)
		}
	}

	for _, device := range result.Blockdevices {
		processDevice(device)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(backingStores); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// CreateLUN adds a new LUN to an iSCSI target
func CreateLUN(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		TargetName   string `json:"target_name"`
		BackingStore string `json:"backing_store"`
		Size         string `json:"size"`        // Size for fileio backstores (e.g., "1G")
		UseFileio    bool   `json:"use_fileio"` // Use fileio instead of block
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if requestData.TargetName == "" {
		http.Error(w, "Target name is required", http.StatusBadRequest)
		return
	}

	// Generate a name for the backstore
	backstoreName := fmt.Sprintf("lun_%d", time.Now().UnixNano())
	backstoreName = strings.ReplaceAll(backstoreName, ":", "_")
	backstoreName = strings.ReplaceAll(backstoreName, ".", "_")

	var backstorePath string
	var output []byte
	var err error

	if requestData.UseFileio {
		// Create a fileio backstore (file-based storage)
		if requestData.Size == "" {
			requestData.Size = "1G" // Default size
		}
		if requestData.BackingStore == "" {
			// Default path for fileio backstores
			requestData.BackingStore = "/var/lib/arcanas/iscsi/" + backstoreName + ".img"
		}

		// Create the directory if it doesn't exist
		_ = utils.SudoRunCommand("mkdir", "-p", "/var/lib/arcanas/iscsi")

		// Create fileio backstore using correct targetcli path syntax
		output, err = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/backstores/fileio create %s %s %s' | sudo targetcli", backstoreName, requestData.BackingStore, requestData.Size))
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create fileio backstore: %v, output: %s", err, string(output)), http.StatusInternalServerError)
			return
		}
		backstorePath = "/backstores/fileio/" + backstoreName
	} else {
		// Create a block backstore using the device
		if requestData.BackingStore == "" {
			http.Error(w, "Backing store is required", http.StatusBadRequest)
			return
		}

		// Check if device is mounted and unmount it if so
		// This is necessary because iSCSI requires exclusive access to the block device
		findmntCmd := exec.Command("findmnt", "-n", "-o", "TARGET", "--source", requestData.BackingStore)
		mountOutput, _ := findmntCmd.Output()
		mountPoint := strings.TrimSpace(string(mountOutput))

		if mountPoint != "" {
			// Device is mounted, unmount it first
			fmt.Printf("Device %s is mounted at %s, unmounting for iSCSI use...\n", requestData.BackingStore, mountPoint)
			umountCmd := exec.Command("sudo", "umount", mountPoint)
			if err := umountCmd.Run(); err != nil {
				// Try lazy unmount if normal unmount fails
				fmt.Printf("Normal unmount failed, trying lazy unmount...\n")
				umountCmd = exec.Command("sudo", "umount", "-l", mountPoint)
				if err := umountCmd.Run(); err != nil {
					http.Error(w, fmt.Sprintf("Failed to unmount device %s from %s: %v (device is in use)", requestData.BackingStore, mountPoint, err), http.StatusInternalServerError)
					return
				}
			}
			fmt.Printf("Successfully unmounted %s from %s\n", requestData.BackingStore, mountPoint)
		}

		output, err = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/backstores/block create %s %s' | sudo targetcli", backstoreName, requestData.BackingStore))
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create block backstore: %v, output: %s", err, string(output)), http.StatusInternalServerError)
			return
		}
		backstorePath = "/backstores/block/" + backstoreName
	}

	// Create LUN and link it to the backstore using correct targetcli path syntax
	output, err = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1/luns create %s' | sudo targetcli", requestData.TargetName, backstorePath))
	if err != nil {
		// Try to clean up the backstore if LUN creation failed
		backstoreType := "fileio"
		if !requestData.UseFileio {
			backstoreType = "block"
		}
		_, _ = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/backstores/%s delete %s' | sudo targetcli", backstoreType, backstoreName))
		http.Error(w, fmt.Sprintf("Failed to create LUN: %v, output: %s", err, string(output)), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "created",
		"backing_store": backstoreName,
		"backing_path":  backstorePath,
		"target":        requestData.TargetName,
	})
}

// DeleteLUN removes a LUN from an iSCSI target
func DeleteLUN(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		TargetName string `json:"target_name"`
		LUN        int    `json:"lun"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if requestData.TargetName == "" {
		http.Error(w, "Target name is required", http.StatusBadRequest)
		return
	}

	// Delete the LUN using targetcli
	// Note: We need to find which backstore this LUN is using first
	output, err := utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1/luns ls' | sudo targetcli", requestData.TargetName))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list LUNs: %v, output: %s", err, string(output)), http.StatusInternalServerError)
		return
	}

	// Parse the output to find the backstore for this LUN
	lines := strings.Split(string(output), "\n")
	var backstorePath string
	lunNum := requestData.LUN
	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf("lun%d", lunNum)) {
			// Extract the backstore path from the line
			// Format is typically: "o- lun0 ... [/backstores/block/md0]"
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				// Get the last field which contains the backstore path in brackets
				backstorePath = strings.TrimSuffix(strings.TrimPrefix(parts[len(parts)-1], "["), "]")
			}
			break
		}
	}

	if backstorePath == "" {
		http.Error(w, fmt.Sprintf("LUN %d not found", lunNum), http.StatusNotFound)
		return
	}

	// Delete the LUN using correct targetcli path syntax
	output, err = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1/luns delete lun%d' | sudo targetcli", requestData.TargetName, lunNum))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete LUN: %v, output: %s", err, string(output)), http.StatusInternalServerError)
		return
	}

	// Also delete the backstore to free up the storage
	// Determine backstore type and name from the path
	var backstoreType, backstoreName string
	if strings.Contains(backstorePath, "/backstores/block/") {
		backstoreType = "block"
		backstoreName = strings.TrimPrefix(backstorePath, "/backstores/block/")
	} else if strings.Contains(backstorePath, "/backstores/fileio/") {
		backstoreType = "fileio"
		backstoreName = strings.TrimPrefix(backstorePath, "/backstores/fileio/")
	} else {
		// Unknown backstore type, skip deletion
		backstoreType = ""
		backstoreName = ""
	}

	if backstoreName != "" {
		_, _ = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/backstores/%s delete %s' | sudo targetcli", backstoreType, backstoreName))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
		"target": requestData.TargetName,
		"lun":    fmt.Sprintf("%d", lunNum),
	})
}

// CreateACL adds an ACL (Access Control List) entry for an initiator
func CreateACL(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		TargetName   string `json:"target_name"`
		InitiatorIQN string `json:"initiator_iqn"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if requestData.TargetName == "" {
		http.Error(w, "Target name is required", http.StatusBadRequest)
		return
	}

	if requestData.InitiatorIQN == "" {
		http.Error(w, "Initiator IQN is required", http.StatusBadRequest)
		return
	}

	// Create the ACL using correct targetcli path syntax
	output, err := utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1/acls create %s' | sudo targetcli", requestData.TargetName, requestData.InitiatorIQN))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create ACL: %v, output: %s", err, string(output)), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":         "created",
		"target":        requestData.TargetName,
		"initiator_iqn": requestData.InitiatorIQN,
	})
}

// DeleteACL removes an ACL entry
func DeleteACL(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		TargetName   string `json:"target_name"`
		InitiatorIQN string `json:"initiator_iqn"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if requestData.TargetName == "" {
		http.Error(w, "Target name is required", http.StatusBadRequest)
		return
	}

	if requestData.InitiatorIQN == "" {
		http.Error(w, "Initiator IQN is required", http.StatusBadRequest)
		return
	}

	// Delete the ACL using correct targetcli path syntax
	output, err := utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1/acls delete %s' | sudo targetcli", requestData.TargetName, requestData.InitiatorIQN))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete ACL: %v, output: %s", err, string(output)), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":         "deleted",
		"target":        requestData.TargetName,
		"initiator_iqn": requestData.InitiatorIQN,
	})
}

// ============================================================================
// NEW iSCSI API (Single Target Model)
// ============================================================================

// GetISCSITarget returns the single iSCSI target with all its LUNs
func GetISCSITarget(w http.ResponseWriter, r *http.Request) {
	target := models.ISCSITarget{
		IQN:     "iqn.2024-01.com.nas:storage",
		Name:    "Arcanas NAS Storage",
		Status:  "active",
		LUNs:    []models.ISCSILUN{},
		Created: time.Now(),
	}

	// Get LUNs from targetcli
	output, err := utils.SudoCombinedOutput("sh", "-c", "echo '/iscsi/iqn.2024-01.com.nas:storage/tpg1/luns ls' | sudo targetcli")
	if err == nil {
		target.LUNs = parseLUNsFromTargetcli(string(output))
	}

	// Get session count
	sessions, _ := getTargetSessions(target.IQN)
	target.Sessions = len(sessions)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(target)
}

// GetISCSILUNs returns all LUNs for the target
func GetISCSILUNs(w http.ResponseWriter, r *http.Request) {
	// Ensure the iSCSI target exists and is configured with no authentication
	if err := system.EnsureISCSITargetConfigured(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to configure iSCSI target: %v", err), http.StatusInternalServerError)
		return
	}

	output, err := utils.SudoCombinedOutput("sh", "-c", "echo '/iscsi/iqn.2024-01.com.nas:storage/tpg1/luns ls' | sudo targetcli")
	if err != nil {
		http.Error(w, "Failed to get LUNs", http.StatusInternalServerError)
		return
	}

	luns := parseLUNsFromTargetcli(string(output))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(luns)
}

// CreateISCSILUN creates a new LUN with the specified backend
func CreateISCSILUN(w http.ResponseWriter, r *http.Request) {
	var req models.LUNCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if req.SizeGB <= 0 {
		http.Error(w, "Size must be greater than 0", http.StatusBadRequest)
		return
	}
	if req.BackendType == "" {
		http.Error(w, "Backend type is required (lvm, block, or fileio)", http.StatusBadRequest)
		return
	}

	// Validate backend-specific fields
	switch req.BackendType {
	case "lvm":
		if req.VolumeGroup == "" {
			http.Error(w, "Volume group is required for LVM backend", http.StatusBadRequest)
			return
		}
	case "block":
		if req.DevicePath == "" {
			http.Error(w, "Device path is required for block backend", http.StatusBadRequest)
			return
		}
	case "fileio":
		// FilePath is optional, will be auto-generated
	default:
		http.Error(w, "Invalid backend type (must be lvm, block, or fileio)", http.StatusBadRequest)
		return
	}

	// Call the LUN manager to create the LUN
	lun, err := system.CreateLUN(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create LUN: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(lun)
}

// DeleteISCSILUN removes a LUN
func DeleteISCSILUN(w http.ResponseWriter, r *http.Request) {
	// Extract LUN number from URL path
	// URL pattern: /api/iscsi/luns/{lun}
	path := strings.TrimPrefix(r.URL.Path, "/api/iscsi/luns/")
	path = strings.TrimSuffix(path, "/")
	lunNumStr := strings.TrimPrefix(path, "/")
	lunNum, err := strconv.Atoi(lunNumStr)
	if err != nil {
		http.Error(w, "Invalid LUN number", http.StatusBadRequest)
		return
	}

	// First, get the LUN details so we can properly clean up the backend
	output, err := utils.SudoCombinedOutput("sh", "-c", "echo '/iscsi/iqn.2024-01.com.nas:storage/tpg1/luns ls' | sudo targetcli")
	if err != nil {
		http.Error(w, "Failed to list LUNs", http.StatusInternalServerError)
		return
	}

	// Find the LUN to get its backend path
	var lunToDelete models.ISCSILUN
	luns := parseLUNsFromTargetcli(string(output))
	found := false
	for _, lun := range luns {
		if lun.LUN == lunNum {
			lunToDelete = lun
			found = true
			break
		}
	}

	if !found {
		http.Error(w, fmt.Sprintf("LUN %d not found", lunNum), http.StatusNotFound)
		return
	}

	// Delete the LUN using the system function
	if err := system.DeleteLUN(lunNum, lunToDelete); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete LUN: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "deleted",
		"lun":    lunNum,
	})
}

// GetLUNBackends returns available backend options for LUN creation
func GetLUNBackends(w http.ResponseWriter, r *http.Request) {
	backends, err := system.GetAvailableBackends()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get backend options: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(backends)
}

// parseLUNsFromTargetcli parses LUN information from targetcli output
func parseLUNsFromTargetcli(output string) []models.ISCSILUN {
	var luns []models.ISCSILUN
	lines := strings.Split(output, "\n")
	id := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Only match lines starting with "o- lun" followed by a number (e.g., "o- lun0")
		// This avoids matching the header line "o- luns ... [LUNs: 1]"
		if strings.HasPrefix(line, "o- lun") && len(line) > 6 && line[6] >= '0' && line[6] <= '9' {
			// Extract LUN number using regex (format: "o- lun0" or "  o- lun0")
			re := regexp.MustCompile(`lun(\d+)`)
			matches := re.FindStringSubmatch(line)
			lunNum := 0
			if len(matches) >= 2 {
				lunNum, _ = strconv.Atoi(matches[1])
			}

			// Extract backstore path and device path from brackets
			// Format: [block/bs_lun0 (/dev/R0/c1) (default_tg_pt_gp)]
			// We need to extract:
			// - backstorePath: /backstores/block/bs_lun0
			// - devicePath: /dev/R0/c1 (for LVM detection)
			backstorePath := ""
			devicePath := ""
			if strings.Contains(line, "[") && strings.Contains(line, "]") {
				startIdx := strings.Index(line, "[")
				endIdx := strings.LastIndex(line, "]")
				if startIdx >= 0 && endIdx > startIdx {
					content := line[startIdx+1 : endIdx]
					// Extract the backstore type and name (e.g., "block/bs_lun0")
					// The content format is typically: "type/name (/dev/...) (extra_info)"
					parts := strings.Fields(content)
					if len(parts) > 0 {
						backstoreTypeAndName := parts[0] // e.g., "block/bs_lun0"
						backstorePath = "/backstores/" + backstoreTypeAndName
						// Try to extract the device path (e.g., /dev/R0/c1)
						// The path may be wrapped in parentheses like "(/dev/R0/c1)"
						if len(parts) > 1 {
							// Remove leading/trailing parentheses from the path
							candidatePath := strings.Trim(parts[1], "()")
							if strings.HasPrefix(candidatePath, "/dev/") {
								devicePath = candidatePath
							}
						}
					}
				}
			}

			// Query the LUN size
			var sizeGB float64
			// First, try to get size from the device path (more accurate)
			if devicePath != "" {
				sizeOutput, err := utils.SudoCombinedOutput("blockdev", "--getsize64", devicePath)
				if err == nil {
					// Parse size in bytes and convert to GB
					if sizeBytes, err := strconv.ParseInt(strings.TrimSpace(string(sizeOutput)), 10, 64); err == nil {
						sizeGB = float64(sizeBytes) / (1024 * 1024 * 1024)
					}
				}
			}
			// Fallback: query backstore size from targetcli
			if sizeGB == 0 && backstorePath != "" {
				sizeOutput, err := utils.SudoCombinedOutput("targetcli", fmt.Sprintf("%s info", backstorePath))
				if err == nil {
					// Parse size from info output - look for "size:" field
					for _, infoLine := range strings.Split(string(sizeOutput), "\n") {
						infoLine = strings.TrimSpace(infoLine)
						if strings.Contains(infoLine, "size:") || strings.Contains(infoLine, "Size:") {
							// Extract size value (e.g., "size: 1073741824" or "size: 1G")
							sizeParts := strings.Fields(infoLine)
							if len(sizeParts) >= 2 {
								sizeStr := strings.Trim(sizeParts[1], ",")
								// Try to parse as bytes first
								if sizeBytes, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
									sizeGB = float64(sizeBytes) / (1024 * 1024 * 1024)
								} else {
									// Try to parse human-readable format (e.g., "1G", "512M")
									sizeGB = float64(parseHumanSize(sizeStr))
								}
							}
							break
						}
					}
				}
			}

			// Determine backend type from path and device
			backendType := "block"
			vgName := ""
			lvSizeBytes := int64(0)
			lvDataPercent := float64(0)
			lvUsedBytes := int64(0)
			if strings.Contains(backstorePath, "/fileio/") {
				backendType = "fileio"
			} else if strings.Contains(backstorePath, "/block/") {
				// Check if the device is an LVM LV (path format: /dev/vg_name/lv_name)
				if devicePath != "" && isLVMDevice(devicePath) {
					backendType = "lvm"
					// Extract VG name and get LV stats
					parts := strings.Split(devicePath, "/")
					if len(parts) >= 4 {
						vgName = parts[2]
						lvName := parts[3]
						// Get LV size in bytes
						lvSizeBytes = getLVSizeBytes(vgName, lvName)
						// Get LV data percent
						lvDataPercent = getLVDataPercent(vgName, lvName)
						// Calculate used bytes
						if lvSizeBytes > 0 && lvDataPercent > 0 {
							lvUsedBytes = int64(float64(lvSizeBytes) * lvDataPercent / 100)
						}
					}
				}
			}

			luns = append(luns, models.ISCSILUN{
				ID:            id,
				LUN:           lunNum,
				Name:          fmt.Sprintf("LUN %d", lunNum),
				SizeGB:        sizeGB,
				BackendType:   backendType,
				BackendPath:   backstorePath,
				LVPath:        devicePath, // Store device path for LV deletion
				Status:        "active",
				Created:       time.Now(),
				VolumeGroup:   vgName,
				LVSizeBytes:   lvSizeBytes,
				LVDataPercent: lvDataPercent,
				LVUsedBytes:   lvUsedBytes,
			})
			id++
		}
	}

	return luns
}

// isLVMDevice checks if a device path is an LVM Logical Volume
func isLVMDevice(devicePath string) bool {
	// LVM LVs have the format /dev/vg_name/lv_name
	// We can detect this by checking if the path has 3+ parts and
	// if the middle part is a valid volume group
	parts := strings.Split(devicePath, "/")
	if len(parts) < 4 || parts[0] != "" || parts[1] != "dev" {
		return false
	}
	// Check if it's actually an LV by trying to query it
	_, err := utils.SudoCombinedOutput("lvs", "--noheadings", "-o", "lv_attr", devicePath)
	return err == nil
}

// getLVSizeBytes returns the size of a logical volume in bytes using lvdisplay
func getLVSizeBytes(vgName, lvName string) int64 {
	output, err := utils.SudoCombinedOutput("lvdisplay", fmt.Sprintf("%s/%s", vgName, lvName))
	if err != nil {
		return 0
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "LV Size") {
			// Parse "LV Size                25.00 GiB"
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				sizeStr := parts[2]
				// Convert to bytes
				size, err := parseSizeToBytes(sizeStr)
				if err == nil {
					return size
				}
			}
		}
	}
	return 0
}

// parseSizeToBytes converts a size string like "25.00 GiB" to bytes
func parseSizeToBytes(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)

	// Extract the numeric part and unit
	parts := strings.Fields(sizeStr)
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid size format")
	}

	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}

	unit := strings.ToLower(parts[1])
	var multiplier float64

	switch {
	case strings.HasPrefix(unit, "pb"):
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024
	case strings.HasPrefix(unit, "tb"):
		multiplier = 1024 * 1024 * 1024 * 1024
	case strings.HasPrefix(unit, "gb"):
		multiplier = 1024 * 1024 * 1024
	case strings.HasPrefix(unit, "mb"):
		multiplier = 1024 * 1024
	case strings.HasPrefix(unit, "kb"):
		multiplier = 1024
	case strings.HasPrefix(unit, "b"):
		multiplier = 1
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}

	return int64(value * multiplier), nil
}

// getLVDataPercent returns the data percentage of a logical volume (0-100)
func getLVDataPercent(vgName, lvName string) float64 {
	output, err := utils.SudoCombinedOutput("lvs", "--noheadings", "--units", "b", "-o", "data_percent", fmt.Sprintf("%s/%s", vgName, lvName))
	if err != nil {
		return 0
	}

	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) < 1 {
		return 0
	}

	percent, _ := strconv.ParseFloat(parts[0], 64)
	return percent
}

// ============================================================================
// NEW iSCSI ACL API (Single Target Model with Per-Client LUN Access)
// ============================================================================

// GetISCSIACLs returns all ACL entries for the iSCSI target
func GetISCSIACLs(w http.ResponseWriter, r *http.Request) {
	acls, err := system.GetISCSIACLs()
	if err != nil {
		log.Printf("Error getting ACLs: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(acls); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// CreateISCSIACL creates a new ACL entry for an initiator IQN
func CreateISCSIACL(w http.ResponseWriter, r *http.Request) {
	var req models.ACLCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.InitiatorIQN == "" {
		http.Error(w, "Initiator IQN is required", http.StatusBadRequest)
		return
	}

	if err := system.CreateISCSIACL(req.InitiatorIQN, req.Name); err != nil {
		log.Printf("Error creating ACL: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message":      "ACL created successfully",
		"initiator_iqn": req.InitiatorIQN,
		"name":         req.Name,
	}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// DeleteISCSIACL removes an ACL entry
func DeleteISCSIACL(w http.ResponseWriter, r *http.Request) {
	// Extract IQN from URL path and URL-decode it
	path := strings.TrimPrefix(r.URL.Path, "/api/iscsi/acls/")
	iqn, err := url.QueryUnescape(strings.TrimSuffix(path, "/"))
	if err != nil {
		http.Error(w, "Invalid IQN format", http.StatusBadRequest)
		return
	}

	if iqn == "" {
		http.Error(w, "Initiator IQN is required", http.StatusBadRequest)
		return
	}

	if err := system.DeleteISCSIACL(iqn); err != nil {
		log.Printf("Error deleting ACL: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]string{
		"message":      "ACL deleted successfully",
		"initiator_iqn": iqn,
	}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// MapLUNToACL maps a LUN to an ACL (gives the initiator access to that LUN)
func MapLUNToACL(w http.ResponseWriter, r *http.Request) {
	// Extract IQN from URL path and URL-decode it
	path := strings.TrimPrefix(r.URL.Path, "/api/iscsi/acls/")
	path = strings.TrimSuffix(path, "/luns")
	iqn, err := url.QueryUnescape(path)
	if err != nil {
		http.Error(w, "Invalid IQN format", http.StatusBadRequest)
		return
	}

	if iqn == "" {
		http.Error(w, "Initiator IQN is required", http.StatusBadRequest)
		return
	}

	var req models.ACLMapLUNRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SourceLUN < 0 {
		http.Error(w, "Source LUN is required", http.StatusBadRequest)
		return
	}

	targetLUN := req.TargetLUN
	if targetLUN < 0 {
		targetLUN = req.SourceLUN
	}

	if err := system.MapLUNToACL(iqn, req.SourceLUN, targetLUN); err != nil {
		log.Printf("Error mapping LUN to ACL: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "LUN mapped successfully",
		"initiator_iqn": iqn,
		"source_lun":   req.SourceLUN,
		"target_lun":   targetLUN,
	}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// UnmapLUNFromACL removes a LUN mapping from an ACL
func UnmapLUNFromACL(w http.ResponseWriter, r *http.Request) {
	// Extract IQN and target LUN from URL path
	// Format: /api/iscsi/acls/{iqn}/luns/{target_lun}
	path := strings.TrimPrefix(r.URL.Path, "/api/iscsi/acls/")
	parts := strings.Split(path, "/luns/")
	if len(parts) < 2 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	iqn, err := url.QueryUnescape(parts[0])
	if err != nil {
		http.Error(w, "Invalid IQN format", http.StatusBadRequest)
		return
	}

	targetLUNStr := strings.TrimSuffix(parts[1], "/")

	targetLUN, err := strconv.Atoi(targetLUNStr)
	if err != nil {
		http.Error(w, "Invalid target LUN number", http.StatusBadRequest)
		return
	}

	if err := system.UnmapLUNFromACL(iqn, targetLUN); err != nil {
		log.Printf("Error unmapping LUN from ACL: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "LUN unmapped successfully",
		"initiator_iqn": iqn,
		"target_lun":   targetLUN,
	}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// GetACLsForLUN returns all ACLs that have a specific LUN mapped
func GetACLsForLUN(w http.ResponseWriter, r *http.Request) {
	// Extract LUN number from URL query parameter
	lunStr := r.URL.Query().Get("lun")
	if lunStr == "" {
		http.Error(w, "LUN parameter is required", http.StatusBadRequest)
		return
	}

	lunNum, err := strconv.Atoi(lunStr)
	if err != nil {
		http.Error(w, "Invalid LUN number", http.StatusBadRequest)
		return
	}

	acls, err := system.GetACLsForLUN(lunNum)
	if err != nil {
		log.Printf("Error getting ACLs for LUN: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(acls); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}
