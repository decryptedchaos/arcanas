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
	"os/exec"
	"strconv"
	"strings"
	"time"

	"arcanas/internal/models"
	"arcanas/internal/system"
	"arcanas/internal/utils"
)

type MountInfo struct {
	Mountpoint string
	Filesystem string
}

func getMountInfo(device string) *MountInfo {
	// Use lsblk to get mountpoint and filesystem info - more reliable than df
	// lsblk -J -o NAME,MOUNTPOINT,FSTYPE,PATH
	cmd := exec.Command("lsblk", "-J", "-o", "NAME,MOUNTPOINT,FSTYPE,PATH")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	type LsblkDevice struct {
		Name       string        `json:"name"`
		Path       string        `json:"path"`
		Mountpoint *string       `json:"mountpoint"`
		Fstype     *string       `json:"fstype"`
		Children   []LsblkDevice `json:"children"`
	}
	type LsblkOutput struct {
		Blockdevices []LsblkDevice `json:"blockdevices"`
	}

	var result LsblkOutput
	if err := json.Unmarshal(output, &result); err != nil {
		return nil
	}

	// Search for matching device and filesystem info
	var searchDevice func(d LsblkDevice) *MountInfo
	searchDevice = func(d LsblkDevice) *MountInfo {
		// Check if this is our device
		if d.Path == device || strings.Contains(d.Path, strings.TrimPrefix(device, "/dev/")) {
			// If this device has mountpoint/fstype, return it
			if d.Fstype != nil || d.Mountpoint != nil {
				fstype := "unknown"
				mountpoint := ""
				if d.Fstype != nil {
					fstype = *d.Fstype
				}
				if d.Mountpoint != nil {
					mountpoint = *d.Mountpoint
				}
				return &MountInfo{
					Mountpoint: mountpoint,
					Filesystem: fstype,
				}
			}
		}

		// Search children (partitions)
		for _, child := range d.Children {
			// For partitions, check if they belong to our device
			if strings.HasPrefix(child.Path, strings.TrimPrefix(device, "/dev/")) ||
				strings.HasPrefix(child.Path, device) {
				fstype := "unknown"
				mountpoint := ""
				if child.Fstype != nil {
					fstype = *child.Fstype
				}
				if child.Mountpoint != nil {
					mountpoint = *child.Mountpoint
				}
				if fstype != "unknown" || mountpoint != "" {
					return &MountInfo{
						Mountpoint: mountpoint,
						Filesystem: fstype,
					}
				}
			}
			if result := searchDevice(child); result != nil {
				return result
			}
		}

		return nil
	}

	for _, blockdev := range result.Blockdevices {
		if result := searchDevice(blockdev); result != nil {
			return result
		}
	}

	return nil
}

// TODO: Rename this function - it returns disk info, not stats
func GetDiskStats(w http.ResponseWriter, r *http.Request) {
	// Get real storage stats
	storageStats, err := system.GetStorageStats()
	if err != nil {
		log.Printf("Error getting storage stats: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert storage stats to disk stats format
	var disks []models.DiskStats
	for _, disk := range storageStats.Disks {
		// Calculate usage percentage
		usage := float64(0)
		if disk.Size > 0 {
			usage = float64(disk.Used) / float64(disk.Size) * 100
		}

		// Get filesystem info - prefer storage stats fstype (includes linux_raid_member)
		filesystem := disk.Fstype
		if filesystem == "" {
			filesystem = "unknown"
		}
		mountpoint := ""

		// Try to get mountpoint from df (more reliable for mounted filesystems)
		if fsInfo := getMountInfo(disk.Device); fsInfo != nil {
			if fsInfo.Mountpoint != "" {
				mountpoint = fsInfo.Mountpoint
			}
			// Use storage stats fstype as primary source, fall back to mountinfo for actual mounted filesystems
			if fsInfo.Filesystem != "" && fsInfo.Filesystem != "unknown" && filesystem == "unknown" {
				filesystem = fsInfo.Filesystem
			}
		}

		disks = append(disks, models.DiskStats{
			Device:     disk.Device,
			Model:      disk.Model,
			Size:       disk.Size,
			Used:       disk.Used,
			Available:  disk.Size - disk.Used,
			Usage:      usage,
			Mountpoint: mountpoint,
			Filesystem: filesystem,
			ReadOnly:   false, // TODO: Implement read-only detection
			Smart: models.SmartInfo{
				Status:      disk.SmartStatus,
				Health:      disk.Health,
				Temperature: int(disk.Temperature),
				PassedTests: 0, // TODO: Implement SMART test count
				FailedTests: 0,
				LastTest:    time.Now(), // TODO: Implement actual test time
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(disks); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func GetSmartStatus(w http.ResponseWriter, r *http.Request) {
	disk := r.URL.Query().Get("disk")

	if disk == "" {
		http.Error(w, "Disk parameter is required", http.StatusBadRequest)
		return
	}

	smart, err := getSmartFullInfo(disk)
	if err != nil {
		log.Printf("Error getting SMART status for %s: %v", disk, err)
		http.Error(w, "Failed to get SMART status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(smart); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// GetAllSmartStatus returns SMART info for all disks
func GetAllSmartStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get list of all disks
	disks, err := getDiskList()
	if err != nil {
		log.Printf("Error getting disk list: %v", err)
		http.Error(w, "Failed to get disk list", http.StatusInternalServerError)
		return
	}

	// Get SMART info for each disk
	var results []models.SmartFullInfo
	for _, disk := range disks {
		smart, err := getSmartFullInfo(disk)
		if err != nil {
			log.Printf("Error getting SMART for %s: %v", disk, err)
			continue
		}
		results = append(results, smart)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// getSmartFullInfo gets complete SMART information for a disk
func getSmartFullInfo(disk string) (models.SmartFullInfo, error) {
	smart := models.SmartFullInfo{
		Device: disk,
	}

	// Get device info (model, serial, firmware)
	info, err := getSmartInfo(disk)
	if err == nil {
		smart.Model = info.model
		smart.Serial = info.serial
		smart.Firmware = info.firmware
	}

	// Get SMART health
	health, err := getSmartHealth(disk)
	if err == nil {
		smart.Health = health.percentage
		smart.Status = health.status
		smart.Enabled = health.enabled
	}

	// Get temperature
	temp, err := getDiskTemperature(disk)
	if err == nil && temp > 0 {
		smart.Temperature = temp
	}

	// Get power on hours and cycles
	powerInfo, err := getSmartPowerInfo(disk)
	if err == nil {
		smart.PowerOnHours = powerInfo.hours
		smart.PowerCycles = powerInfo.cycles
	}

	// Get SMART attributes
	attrs, err := getSmartAttributes(disk)
	if err == nil {
		smart.Attributes = attrs
	}

	// Get self-test log
	tests, err := getSmartSelfTests(disk)
	if err == nil {
		smart.SelfTests = tests
		smart.PassedTests = countPassedTests(tests)
		smart.FailedTests = countFailedTests(tests)
	}

	// Get error log
	errors, err := getSmartErrors(disk)
	if err == nil {
		smart.Errors = errors
	}

	return smart, nil
}

// getSmartInfo gets device information from smartctl -i
func getSmartInfo(disk string) (struct {
	model    string
	serial   string
	firmware string
}, error) {
	cmd := exec.Command("smartctl", "-i", disk)
	output, err := cmd.Output()
	if err != nil {
		return struct{ model, serial, firmware string }{}, err
	}

	info := struct {
		model    string
		serial   string
		firmware string
	}{}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Device Model:") {
			info.model = strings.TrimSpace(strings.TrimPrefix(line, "Device Model:"))
		} else if strings.HasPrefix(line, "Serial Number:") {
			info.serial = strings.TrimSpace(strings.TrimPrefix(line, "Serial Number:"))
		} else if strings.HasPrefix(line, "Firmware Version:") {
			info.firmware = strings.TrimSpace(strings.TrimPrefix(line, "Firmware Version:"))
		}
	}

	return info, nil
}

// getSmartHealth gets SMART health assessment from smartctl -H
func getSmartHealth(disk string) (struct {
	percentage int
	status     string
	enabled    bool
}, error) {
	cmd := exec.Command("smartctl", "-H", disk)
	output, err := cmd.Output()
	if err != nil {
		return struct{ percentage int; status string; enabled bool }{}, err
	}

	result := struct {
		percentage int
		status     string
		enabled    bool
	}{}

	// Parse overall health
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "SMART overall-health self-assessment test result") {
			if strings.Contains(line, "PASSED") {
				result.status = "healthy"
				result.percentage = 100
			} else if strings.Contains(line, "FAILED") {
				result.status = "failed"
				result.percentage = 0
			}
		}
	}

	// Check if SMART is enabled
	cmd = exec.Command("smartctl", "-c", disk)
	output, err = cmd.Output()
	if err == nil {
		lines = strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "SMART support is:") {
				result.enabled = strings.Contains(line, "Enabled")
			}
		}
	}

	return result, nil
}

// getSmartPowerInfo gets power on hours and cycles from smartctl
func getSmartPowerInfo(disk string) (struct {
	hours  int
	cycles int
}, error) {
	cmd := exec.Command("smartctl", "-A", disk)
	output, err := cmd.Output()
	if err != nil {
		return struct{ hours int; cycles int }{}, err
	}

	result := struct{ hours int; cycles int }{}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Power_On_Hours") {
			parts := strings.Fields(line)
			if len(parts) >= 10 {
				if val, err := strconv.ParseInt(strings.Trim(parts[9], "-"), 10, 64); err == nil {
					result.hours = int(val)
				}
			}
		} else if strings.Contains(line, "Power_Cycle_Count") {
			parts := strings.Fields(line)
			if len(parts) >= 10 {
				if val, err := strconv.ParseInt(strings.Trim(parts[9], "-"), 10, 64); err == nil {
					result.cycles = int(val)
				}
			}
		}
	}

	return result, nil
}

// getDiskTemperature gets the current disk temperature from smartctl
func getDiskTemperature(disk string) (int, error) {
	cmd := exec.Command("smartctl", "-A", disk)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// Look for temperature attribute (usually ID 194 or 190)
		if strings.Contains(line, "Temperature") || strings.Contains(line, "Airflow_Temperature") {
			parts := strings.Fields(line)
			if len(parts) >= 10 {
				// Try to parse the raw value (column 9) or normalized value (column 3)
				if val, err := strconv.ParseInt(strings.Trim(parts[9], "-"), 10, 64); err == nil && val > 0 {
					return int(val), nil
				}
				if val, err := strconv.ParseInt(parts[3], 10, 64); err == nil && val > 0 {
					return int(val), nil
				}
			}
		}
	}

	return 0, fmt.Errorf("temperature not found")
}

// getSmartAttributes parses SMART attributes from smartctl -A
func getSmartAttributes(disk string) ([]models.SmartAttribute, error) {
	cmd := exec.Command("smartctl", "-A", disk)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var attributes []models.SmartAttribute
	lines := strings.Split(string(output), "\n")

	inAttributesSection := false
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Start of attributes section
		if strings.HasPrefix(line, "ID# ATTRIBUTE_NAME") {
			inAttributesSection = true
			continue
		}

		// Vendor-specific attributes section starts after an empty line
		if inAttributesSection && line == "" {
			continue
		}

		if !inAttributesSection || line == "" || strings.HasPrefix(line, "===") {
			if strings.HasPrefix(line, "===") {
				break
			}
			continue
		}

		// Parse attribute line
		// Format: ID# ATTRIBUTE_NAME FLAGS VALUE WORST THRESH FAILING_RAW_VALUE
		parts := strings.Fields(line)
		if len(parts) >= 10 {
			attr := models.SmartAttribute{}

			// Parse ID
			if strings.Contains(parts[0], "#") {
				idStr := strings.TrimSuffix(parts[0], "#")
				if id, err := strconv.Atoi(idStr); err == nil {
					attr.ID = id
				}
			}

			attr.Name = parts[1]
			attr.Flag = parts[2]

			if value, err := strconv.Atoi(parts[3]); err == nil {
				attr.Value = value
			}
			if worst, err := strconv.Atoi(parts[4]); err == nil {
				attr.Worst = worst
			}
			if thresh, err := strconv.Atoi(parts[5]); err == nil {
				attr.Threshold = thresh
			}

			// Parse raw value (format: VALUE (raw))
			rawStr := parts[9]
			if strings.HasPrefix(rawStr, "-") {
				attr.RawValue, _ = strconv.ParseInt(strings.TrimPrefix(rawStr, "-"), 10, 64)
			} else {
				// Try to extract the raw value from parentheses
				startIdx := strings.LastIndex(rawStr, "(")
				endIdx := strings.LastIndex(rawStr, ")")
				if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
					rawVal := rawStr[startIdx+1 : endIdx]
					attr.RawValue, _ = strconv.ParseInt(rawVal, 10, 64)
				}
			}

			// Check if failed
			attr.Failed = strings.Contains(parts[6], "FAILING")

			attributes = append(attributes, attr)
		}
	}

	return attributes, nil
}

// getSmartSelfTests gets self-test log from smartctl -l selftest
func getSmartSelfTests(disk string) ([]models.SmartTestEntry, error) {
	cmd := exec.Command("smartctl", "-l", "selftest", disk)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var tests []models.SmartTestEntry
	lines := strings.Split(string(output), "\n")

	inTestSection := false
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "SMART Self-test log structure") {
			inTestSection = true
			continue
		}

		if !inTestSection {
			continue
		}

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse test entry
		// Format: # Num Test_Description Status Remaining Lifetime(hours) LBA_of_first_error
		parts := strings.Fields(line)
		if len(parts) >= 6 {
			test := models.SmartTestEntry{}

			if testNum, err := strconv.Atoi(strings.TrimPrefix(parts[0], "#")); err == nil {
				test.TestNum = testNum
			}
			test.Type = strings.Join(parts[1:4], " ")
			test.Status = parts[4]

			// Handle case where remaining might be empty
			if len(parts) >= 6 {
				// Try to parse remaining percentage
				if strings.Contains(parts[5], "%") {
					test.Remaining, _ = strconv.Atoi(strings.TrimSuffix(parts[5], "%"))
				}
			}

			// Parse LBA if present
			for i := 6; i < len(parts); i++ {
				if parts[i] != "-" {
					if lba, err := strconv.ParseInt(parts[i], 10, 64); err == nil {
						test.LBA = lba
					}
				}
			}

			tests = append(tests, test)
		}
	}

	return tests, nil
}

// getSmartErrors gets error log from smartctl -l error
func getSmartErrors(disk string) ([]models.SmartError, error) {
	cmd := exec.Command("smartctl", "-l", "error", disk)
	output, err := cmd.Output()
	if err != nil {
		// No errors is OK
		return []models.SmartError{}, nil
	}

	var errors []models.SmartError
	lines := strings.Split(string(output), "\n")

	inErrorSection := false
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "SMART Error Log structure") {
			inErrorSection = true
			continue
		}

		if !inErrorSection {
			continue
		}

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse error entry
		// Simplified parsing for common format
		parts := strings.Fields(line)
		if len(parts) >= 7 {
			smartErr := models.SmartError{}

			if errNum, parseErr := strconv.Atoi(strings.TrimPrefix(parts[0], "#")); parseErr == nil {
				smartErr.ErrorNum = errNum
			}
			smartErr.Type = parts[1]
			smartErr.State = parts[3]

			// Try to parse LBA if present
			for i := 4; i < len(parts); i++ {
				if parts[i] != "-" {
					if lba, parseErr := strconv.ParseInt(parts[i], 10, 64); parseErr == nil {
						smartErr.LBA = lba
					}
					break
				}
			}

			errors = append(errors, smartErr)
		}
	}

	return errors, nil
}

// countPassedTests counts passed self-tests
func countPassedTests(tests []models.SmartTestEntry) int {
	count := 0
	for _, test := range tests {
		if strings.Contains(strings.ToLower(test.Status), "completed") ||
		   strings.Contains(strings.ToLower(test.Status), "passed") {
			count++
		}
	}
	return count
}

// countFailedTests counts failed self-tests
func countFailedTests(tests []models.SmartTestEntry) int {
	count := 0
	for _, test := range tests {
		if strings.Contains(strings.ToLower(test.Status), "failed") ||
		   strings.Contains(strings.ToLower(test.Status), "error") {
			count++
		}
	}
	return count
}

// getDiskList returns a list of all disk devices
func getDiskList() ([]string, error) {
	cmd := exec.Command("lsblk", "-d", "-n", "-o", "NAME")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var disks []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "NAME") {
			continue
		}
		// Only include actual disks (sd*, nvme*, etc.)
		if strings.HasPrefix(line, "sd") || strings.HasPrefix(line, "nvme") {
			disks = append(disks, "/dev/"+line)
		}
	}

	return disks, nil
}

// RunSmartTest runs a SMART self-test on a disk
func RunSmartTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Disk     string `json:"disk"`
		TestType string `json:"test_type"` // short, long, conveyance
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Disk == "" {
		http.Error(w, "Disk is required", http.StatusBadRequest)
		return
	}

	// Map test type to smartctl argument
	var testArg string
	switch req.TestType {
	case "short":
		testArg = "-t short"
	case "long":
		testArg = "-t long"
	case "conveyance":
		testArg = "-t conveyance"
	case "offline":
		testArg = "-t offline"
	default:
		http.Error(w, "Invalid test type", http.StatusBadRequest)
		return
	}

	// Run the test using SudoCommand
	utils.SudoCommand("smartctl", testArg, req.Disk)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "test_started",
		"disk":     req.Disk,
		"testType": req.TestType,
	})
}

// GetSmartAttributes returns detailed SMART attributes for a disk
func GetSmartAttributes(w http.ResponseWriter, r *http.Request) {
	disk := r.URL.Query().Get("disk")
	if disk == "" {
		http.Error(w, "Disk parameter is required", http.StatusBadRequest)
		return
	}

	attrs, err := getSmartAttributes(disk)
	if err != nil {
		log.Printf("Error getting SMART attributes for %s: %v", disk, err)
		http.Error(w, "Failed to get SMART attributes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(attrs); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetSmartErrors returns error log for a disk
func GetSmartErrors(w http.ResponseWriter, r *http.Request) {
	disk := r.URL.Query().Get("disk")
	if disk == "" {
		http.Error(w, "Disk parameter is required", http.StatusBadRequest)
		return
	}

	errors, err := getSmartErrors(disk)
	if err != nil {
		log.Printf("Error getting SMART errors for %s: %v", disk, err)
		http.Error(w, "Failed to get error log", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(errors); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// SetSmartSetting modifies SMART settings for a disk
func SetSmartSetting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Disk   string `json:"disk"`
		Setting string `json:"setting"` // enable, disable, offline_on, offline_off
		Value  string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Disk == "" {
		http.Error(w, "Disk is required", http.StatusBadRequest)
		return
	}

	var args []string
	switch req.Setting {
	case "enable":
		args = []string{"-s", "on", req.Disk}
	case "disable":
		args = []string{"-s", "off", req.Disk}
	case "offline_on":
		args = []string{"-o", "on", req.Disk}
	case "offline_off":
		args = []string{"-o", "off", req.Disk}
	default:
		http.Error(w, "Invalid setting", http.StatusBadRequest)
		return
	}

	// Apply SMART setting
	utils.SudoCommand("smartctl", args...)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "setting_applied",
		"disk":     req.Disk,
		"setting":  req.Setting,
		"value":    req.Value,
	})
}

func GetPartitions(w http.ResponseWriter, r *http.Request) {
	disk := r.URL.Query().Get("disk")
	if disk == "" {
		// If no disk specified, return all partitions
		log.Printf("No disk specified, returning all partitions")
	}

	// 1. Run lsblk JSON output
	// lsblk -J -b -o NAME,MOUNTPOINT,SIZE,FSTYPE,UUID,PATH <disk>
	args := []string{"-J", "-b", "-o", "NAME,MOUNTPOINT,SIZE,FSTYPE,UUID,PATH"}
	if disk != "" {
		args = append(args, disk)
	}

	cmd := exec.Command("lsblk", args...)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error running lsblk: %v", err)
		http.Error(w, "Failed to list partitions", http.StatusInternalServerError)
		return
	}

	// 2. Parse lsblk JSON
	type LsblkDevice struct {
		Name       string        `json:"name"`
		Path       string        `json:"path"`
		Size       interface{}   `json:"size"` // can be string or number? usually number with -b
		Mountpoint *string       `json:"mountpoint"`
		Fstype     *string       `json:"fstype"`
		Children   []LsblkDevice `json:"children"`
	}
	type LsblkOutput struct {
		Blockdevices []LsblkDevice `json:"blockdevices"`
	}

	var result LsblkOutput
	if err := json.Unmarshal(output, &result); err != nil {
		log.Printf("Error parsing lsblk json: %v", err)
		http.Error(w, "Failed to parse partitions", http.StatusInternalServerError)
		return
	}

	// 3. Flatten and Convert to API Model
	var partitions []models.Partition

	var processDevice func(d LsblkDevice)
	processDevice = func(d LsblkDevice) {
		// Check if it's a partition (children usually, or leaf node)
		// Or just return everything that has a path
		if d.Path != "" {
			part := models.Partition{
				Device:     d.Path,
				Filesystem: "",
			}

			if d.Fstype != nil {
				part.Filesystem = *d.Fstype
			}
			if d.Mountpoint != nil {
				part.Mountpoint = *d.Mountpoint

				// Get usage if mounted
				if size, usedResult := system.GetPathUsage(part.Mountpoint); size > 0 {
					part.Size = size
					part.Used = usedResult
					part.Available = size - usedResult
					if size > 0 {
						part.Usage = float64(usedResult) / float64(size) * 100.0
					}
				}
			} else {
				// Parse size from lsblk (bytes)
				// lsblk JSON size is number if -b used, or is it?
				// json decoder will handle float64 for numbers
				// Let's handle it safely
				switch v := d.Size.(type) {
				case float64:
					part.Size = int64(v)
				case string:
					// try parsing
					if s, err := strconv.ParseInt(v, 10, 64); err == nil {
						part.Size = s
					}
				}
			}

			// Only add if it looks like a partition or logical volume (not the disk itself if it has children)
			// But user might want to format the whole disk?
			// GetPartitions implies "parts".
			// Let's include it.
			partitions = append(partitions, part)
		}

		for _, child := range d.Children {
			processDevice(child)
		}
	}

	for _, d := range result.Blockdevices {
		// If disk was specified, lsblk returns just that disk object
		// If we are looking at the ROOT device, we might skip adding it to partitions list unless it IS a partition?
		// For now, process children. If no children, process self.
		if len(d.Children) > 0 {
			for _, child := range d.Children {
				processDevice(child)
			}
		} else {
			// Single device (e.g. partition passed directly or disk with no parts)
			processDevice(d)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(partitions); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// DeviceMount represents a mounted device that can be managed
type DeviceMount struct {
	Device     string `json:"device"`
	MountPoint string `json:"mount_point"`
	Filesystem string `json:"filesystem"`
	Used       int64  `json:"used"`
	Size       int64  `json:"size"`
	Available  int64  `json:"available"`
}

// GetDeviceMounts returns all arcanas-managed device mounts
// This includes both /srv/* pools (for new direct mount architecture) and /mnt/arcanas-disk-* (legacy)
func GetDeviceMounts(w http.ResponseWriter, r *http.Request) {
	mounts := []DeviceMount{}
	// Use a map to deduplicate by mount point
	seenMounts := make(map[string]bool)

	// Use lsblk to find mounted devices more reliably
	// lsblk -J -o NAME,PATH,MOUNTPOINT,FSTYPE,SIZE
	cmd := exec.Command("lsblk", "-J", "-o", "NAME,PATH,MOUNTPOINT,FSTYPE,SIZE")
	output, err := cmd.Output()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list mounts: %v", err), http.StatusInternalServerError)
		return
	}

	type LsblkDevice struct {
		Name       string        `json:"name"`
		Path       string        `json:"path"`
		Mountpoint *string       `json:"mountpoint"`
		Fstype     *string       `json:"fstype"`
		Size       string        `json:"size"`
		Children   []LsblkDevice `json:"children"`
	}
	type LsblkOutput struct {
		Blockdevices []LsblkDevice `json:"blockdevices"`
	}

	var result LsblkOutput
	if err := json.Unmarshal(output, &result); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse mount info: %v", err), http.StatusInternalServerError)
		return
	}

	// Recursive function to process devices
	var processDevice func(d LsblkDevice)
	processDevice = func(d LsblkDevice) {
		// Check if this device has an arcanas-disk- mount point (legacy)
		// OR a /srv/* mount point (new architecture)
		if d.Mountpoint != nil {
			mountPoint := *d.Mountpoint
			isArcanasMount := strings.HasPrefix(mountPoint, "/mnt/arcanas-disk-") ||
				strings.HasPrefix(mountPoint, "/srv/") && mountPoint != "/srv/"

			if isArcanasMount {
				// Skip if we've already seen this mount point
				if seenMounts[mountPoint] {
					return
				}
				seenMounts[mountPoint] = true

				// Get disk usage for this mount point
				var used, size int64

				// Try to get usage from df
				dfCmd := exec.Command("df", "-B1", "--output=size,used", mountPoint)
				dfOutput, _ := dfCmd.Output()
				if dfOutput != nil {
					lines := strings.Split(string(dfOutput), "\n")
					if len(lines) >= 2 {
						fields := strings.Fields(lines[1])
						if len(fields) >= 2 {
							size, _ = strconv.ParseInt(fields[0], 10, 64)
							used, _ = strconv.ParseInt(fields[1], 10, 64)
						}
					}
				}

				mounts = append(mounts, DeviceMount{
					Device:     d.Path,
					MountPoint: mountPoint,
					Filesystem: func() string { if d.Fstype != nil { return *d.Fstype }; return "" }(),
					Used:       used,
					Size:       size,
					Available:  size - used,
				})
			}
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
	json.NewEncoder(w).Encode(mounts)
}

// UnmountDeviceRequest contains the device path to unmount
type UnmountDeviceRequest struct {
	Device     string `json:"device"`
	MountPoint string `json:"mount_point"`
}

// UnmountDevice unmounts a device, freeing it for iSCSI or other uses
func UnmountDevice(w http.ResponseWriter, r *http.Request) {
	var req UnmountDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate mount point is an arcanas-managed mount
	// Supports both /srv/* pools (new architecture) and /mnt/arcanas-disk-* (legacy)
	isValidMount := strings.HasPrefix(req.MountPoint, "/mnt/arcanas-disk-") ||
		(strings.HasPrefix(req.MountPoint, "/srv/") && req.MountPoint != "/srv/")
	if !isValidMount {
		http.Error(w, "Only arcanas-managed device mounts can be unmounted via this API", http.StatusBadRequest)
		return
	}

	// Check if mount point exists and is mounted
	cmd := exec.Command("findmnt", "-n", req.MountPoint)
	if err := cmd.Run(); err != nil {
		http.Error(w, fmt.Sprintf("Mount point %s is not mounted", req.MountPoint), http.StatusBadRequest)
		return
	}

	// Try normal unmount first
	cmd = exec.Command("sudo", "umount", req.MountPoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If normal unmount fails, try lazy unmount
		cmd = exec.Command("sudo", "umount", "-l", req.MountPoint)
		output, err = cmd.CombinedOutput()
		if err != nil {
			log.Printf("Failed to unmount %s: %v, output: %s", req.MountPoint, err, string(output))
			http.Error(w, fmt.Sprintf("Failed to unmount: %v", err), http.StatusInternalServerError)
			return
		}
	}

	log.Printf("Unmounted %s - device %s is now available for iSCSI", req.MountPoint, req.Device)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Device unmounted successfully",
		"device":     req.Device,
		"mount_point": req.MountPoint,
	})
}

// MountDeviceRequest contains the device path and mount point to mount
type MountDeviceRequest struct {
	Device     string `json:"device"`
	MountPoint string `json:"mount_point"`
}

// MountDevice remounts a previously unmounted device
func MountDevice(w http.ResponseWriter, r *http.Request) {
	var req MountDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate mount point is an arcanas-managed mount
	// Supports both /srv/* pools (new architecture) and /mnt/arcanas-disk-* (legacy)
	isValidMount := strings.HasPrefix(req.MountPoint, "/mnt/arcanas-disk-") ||
		(strings.HasPrefix(req.MountPoint, "/srv/") && req.MountPoint != "/srv/")
	if !isValidMount {
		http.Error(w, "Only arcanas-managed device mounts can be mounted via this API", http.StatusBadRequest)
		return
	}

	// Check if already mounted
	cmd := exec.Command("findmnt", "-n", req.MountPoint)
	if err := cmd.Run(); err == nil {
		http.Error(w, fmt.Sprintf("Mount point %s is already mounted", req.MountPoint), http.StatusConflict)
		return
	}

	// Create mount point directory if it doesn't exist
	cmd = exec.Command("sudo", "mkdir", "-p", req.MountPoint)
	if err := cmd.Run(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create mount point: %v", err), http.StatusInternalServerError)
		return
	}

	// Mount the device
	cmd = exec.Command("sudo", "mount", req.Device, req.MountPoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to mount %s at %s: %v, output: %s", req.Device, req.MountPoint, err, string(output))
		http.Error(w, fmt.Sprintf("Failed to mount: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Mounted %s at %s", req.Device, req.MountPoint)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Device mounted successfully",
		"device":     req.Device,
		"mount_point": req.MountPoint,
	})
}
