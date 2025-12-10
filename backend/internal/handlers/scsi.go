/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"arcanas/internal/models"
)

func GetSCSITargets(w http.ResponseWriter, r *http.Request) {
	targets, err := getLIOTargets()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get LIO targets: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(targets)
}

func getLIOTargets() ([]models.SCSITarget, error) {
	var targets []models.SCSITarget

	// Use targetcli to get targets
	cmd := exec.Command("targetcli", "ls", "/iscsi")
	output, err := cmd.Output()
	if err != nil {
		return targets, fmt.Errorf("failed to run targetcli: %v", err)
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
			}
		}
	}

	return targets, nil
}

func getTargetDetails(iqn string, targetID int) (models.SCSITarget, error) {
	target := models.SCSITarget{
		ID:     targetID,
		Name:   iqn,
		Status: "inactive", // Default status
	}

	// Get target status
	if cmd := exec.Command("targetcli", "iscsi", iqn, "info"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			if strings.Contains(string(output), "Attribute: enabled") && strings.Contains(string(output), "1") {
				target.Status = "active"
			}
		}
	}

	// Get LUNs and sessions
	if luns, err := getTargetLUNs(iqn); err == nil {
		target.LUNCount = len(luns)
		if len(luns) > 0 {
			// Calculate total size from LUNs
			var totalSize int64
			for _, lun := range luns {
				totalSize += lun.Size
			}
			target.Size = formatSize(totalSize)
		}
	}

	// Get sessions
	if sessions, err := getTargetSessions(iqn); err == nil {
		target.Sessions = len(sessions)
		var initiatorIPs []string
		for _, session := range sessions {
			initiatorIPs = append(initiatorIPs, session.IP)
		}
		target.InitiatorIPs = initiatorIPs
	}

	target.Created = time.Now()
	target.LastAccess = time.Now()

	return target, nil
}

func getTargetLUNs(iqn string) ([]models.LUN, error) {
	var luns []models.LUN

	cmd := exec.Command("targetcli", "iscsi", iqn, "tpg1", "luns", "ls")
	output, err := cmd.Output()
	if err != nil {
		return luns, err
	}

	lines := strings.Split(string(output), "\n")
	lunID := 1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "lun") {
			// Parse LUN information
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				lunNum, _ := strconv.Atoi(strings.Trim(parts[1], "lun"))
				lun := models.LUN{
					ID:       lunID,
					TargetID: 0, // Will be set by caller
					LUN:      lunNum,
					Device:   parts[2],
				}
				luns = append(luns, lun)
				lunID++
			}
		}
	}

	return luns, nil
}

func getTargetSessions(iqn string) ([]models.Session, error) {
	var sessions []models.Session

	cmd := exec.Command("targetcli", "iscsi", iqn, "tpg1", "sessions", "ls")
	output, err := cmd.Output()
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

	target.ID = int(time.Now().Unix())
	target.Created = time.Now()
	target.Status = "active"

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(target)
}

func createLIOTarget(target models.SCSITarget) error {
	// Create iSCSI target
	cmd := exec.Command("targetcli", "iscsi", "create", target.Name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create iSCSI target: %v", err)
	}

	// Enable the target
	cmd = exec.Command("targetcli", "iscsi", target.Name, "tpg1", "attr", "set", "attribute=enable", "value=1")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable target: %v", err)
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
	json.NewEncoder(w).Encode(target)
}

func updateLIOTarget(target models.SCSITarget) error {
	// For LIO, most updates involve modifying attributes
	// Enable/disable target
	enabled := "0"
	if target.Status == "active" {
		enabled = "1"
	}

	cmd := exec.Command("targetcli", "iscsi", target.Name, "tpg1", "attr", "set", "attribute=enable", "value="+enabled)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update target status: %v", err)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "target": targetName})
}

func deleteLIOTarget(targetName string) error {
	// First, disable the target
	cmd := exec.Command("targetcli", "iscsi", targetName, "tpg1", "attr", "set", "attribute=enable", "value=0")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to disable target: %v", err)
	}

	// Delete the target
	cmd = exec.Command("targetcli", "iscsi", "delete", targetName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete target: %v", err)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "toggled", "target": targetName})
}

func toggleLIOTarget(targetName string) error {
	// Get current status
	cmd := exec.Command("targetcli", "iscsi", targetName, "info")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get target info: %v", err)
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

	// Set new status
	cmd = exec.Command("targetcli", "iscsi", targetName, "tpg1", "attr", "set", "attribute=enable", "value="+newStatus)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to toggle target: %v", err)
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
	json.NewEncoder(w).Encode(sessions)
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
