/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package system

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"arcanas/internal/models"
	"arcanas/internal/utils"
)

func GetRAIDArrays() ([]models.RAIDArray, error) {
	var arrays []models.RAIDArray

	// Get RAID arrays from mdadm using sudo
	cmd := exec.Command("sudo", "mdadm", "--detail", "--scan")
	output, err := cmd.Output()
	if err != nil {
		// No RAID arrays found, return empty slice
		return arrays, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Parse mdadm output line
		// Example: ARRAY /dev/md0 metadata=1.2 name=arcanas:0 UUID=...
		fields := strings.Fields(line)
		if len(fields) < 4 || fields[0] != "ARRAY" {
			continue
		}

		device := fields[1]
		array, err := getRAIDDetails(device)
		if err != nil {
			continue
		}

		arrays = append(arrays, array)
	}

	return arrays, nil
}

func getRAIDDetails(device string) (models.RAIDArray, error) {
	var array models.RAIDArray

	// Get detailed RAID info using sudo
	cmd := exec.Command("sudo", "mdadm", "--detail", device)
	output, err := cmd.Output()
	if err != nil {
		return array, err
	}

	// Parse mdadm detail output
	// Example output:
	//   /dev/md0:
	//            Version : 1.2
	//      Creation Time : ...
	//         Raid Level : raid0
	//         Array Size : 1953511936 (1.82 TiB 2.00 TB)
	//       Used Dev Size : 976755968 (931.51 GiB 1.00 TB)
	//        Raid Devices : 2
	//       Total Devices : 2
	//             State : clean
	//      Active Devices : 2
	//     Working Devices : 2
	//      Failed Devices : 0
	//       Spare Devices : 0
	//              UUID : ...

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Name :") {
			// Extract name from "Name : arcanas:0  (local to host arcanas)"
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				nameParts := strings.Fields(strings.TrimSpace(parts[1]))
				if len(nameParts) > 0 {
					array.Name = nameParts[0]
				}
			}
		} else if strings.Contains(line, "Raid Level :") {
			// Extract raid level from "Raid Level : raid0"
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				array.Level = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Array Size :") {
			// Extract size from "Array Size : 1953511936 (1.82 TiB 2.00 TB)"
			// The number is in kilobytes
			re := regexp.MustCompile(`Array Size\s*:\s*(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 2 {
				if sizeKB, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
					array.Size = sizeKB * 1024 // Convert KB to bytes
				}
			}
		} else if strings.Contains(line, "State :") {
			// Extract state from "State : clean" or "State : active, clean"
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				rawState := strings.TrimSpace(parts[1])
				stateLower := strings.ToLower(rawState)
				// Normalize state: "clean" and "active" both mean healthy, use "clean" for consistency
				// RAID arrays can show "active", "clean", or "active, clean" - all mean healthy
				if strings.Contains(stateLower, "clean") || strings.Contains(stateLower, "active") {
					array.State = "clean"
				} else {
					// For other states like "degraded", "failed", use as-is
					array.State = strings.TrimSuffix(rawState, ",")
					array.State = strings.TrimSpace(array.State)
				}
			}
		} else if strings.Contains(line, "UUID :") {
			// Extract UUID
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				array.UUID = strings.TrimSpace(parts[1])
			}
		}
	}

	// Get device list from the detailed output
	array.Devices = parseRAIDDevicesFromDetail(string(output))

	// Get mount point and usage
	array.MountPoint, array.Used = getMountPointAndUsage(device)

	// Store the actual device path for operations like deletion
	array.Device = device

	array.CreatedAt = time.Now() // TODO: Get actual creation time
	array.Health = calculateRAIDHealth(array.State)

	return array, nil
}

func parseRAIDDevicesFromDetail(output string) []string {
	var devices []string
	lines := strings.Split(output, "\n")

	// Look for device lines like:
	//    0       8  0  0  active sync  /dev/sda
	//    1       8 16  1  active sync  /dev/sdb
	re := regexp.MustCompile(`/dev/\w+`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "/dev/") || re.MatchString(line) {
			matches := re.FindAllString(line, -1)
			for _, match := range matches {
				// Exclude the md device itself (e.g., /dev/md0, /dev/md1)
				// Only include physical member devices
				if !strings.HasPrefix(match, "/dev/md") {
					devices = append(devices, match)
				}
			}
		}
	}

	return devices
}

// findNextAvailableMDNumber scans for existing md devices and returns the next available number
func findNextAvailableMDNumber() (int, error) {
	// Check /proc/mdstat for active arrays
	data, err := os.ReadFile("/proc/mdstat")
	if err != nil {
		return 0, fmt.Errorf("failed to read /proc/mdstat: %w", err)
	}

	usedNumbers := make(map[int]bool)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		// Look for lines like "md0 : active" or "md1 : active"
		if strings.Contains(line, ": active") || strings.Contains(line, ": inactive") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				deviceName := strings.TrimSuffix(fields[0], ":")
				if strings.HasPrefix(deviceName, "md") {
					numStr := strings.TrimPrefix(deviceName, "md")
					if num, err := strconv.Atoi(numStr); err == nil {
						usedNumbers[num] = true
					}
				}
			}
		}
	}

	// Find the first available number
	for i := 0; i < 128; i++ {
		if !usedNumbers[i] {
			return i, nil
		}
	}

	return 0, fmt.Errorf("no available md device numbers")
}

func CreateRAIDArray(req models.RAIDCreateRequest) (string, error) {
	// Validate RAID level
	validLevels := map[string]bool{
		"raid0": true, "raid1": true, "raid5": true, "raid6": true, "raid10": true,
	}
	if !validLevels[req.Level] {
		return "", fmt.Errorf("invalid RAID level: %s", req.Level)
	}

	// Auto-generate name if not provided
	arrayName := req.Name
	if arrayName == "" {
		nextNum, err := findNextAvailableMDNumber()
		if err != nil {
			return "", fmt.Errorf("failed to find available md device number: %w", err)
		}
		arrayName = fmt.Sprintf("md%d", nextNum)
	}

	// Normalize device name (e.g., "md0" or "/dev/md0")
	deviceName := arrayName
	if !strings.HasPrefix(deviceName, "/dev/") {
		if !strings.HasPrefix(deviceName, "md") {
			deviceName = "md" + deviceName
		}
		deviceName = "/dev/" + deviceName
	}

	// Extract the numeric part for the --name parameter (e.g., "0" from "md0")
	nameNum := strings.TrimPrefix(arrayName, "md")
	if strings.HasPrefix(nameNum, "/") {
		parts := strings.Split(nameNum, "/")
		nameNum = strings.TrimPrefix(parts[len(parts)-1], "md")
	}

	// Build create command with --homehost to prevent md127 rename issue
	// Using "arcanas" as homehost ensures consistent naming across reboots
	args := []string{
		"--create", "--verbose", deviceName,
		"--homehost=arcanas",
		"--name=" + nameNum,
		"--level=" + strings.TrimPrefix(req.Level, "raid"),
		"--raid-devices=" + strconv.Itoa(len(req.Devices)),
	}
	args = append(args, req.Devices...)

	// Build full command with sudo and mdadm
	fullArgs := append([]string{"mdadm"}, args...)
	cmd := exec.Command("sudo", fullArgs...)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create RAID array: %w", err)
	}

	// Write array configuration to mdadm.conf to persist across reboots
	if err := writeMDAdmConf(deviceName); err != nil {
		// Log warning but don't fail - array was created successfully
		fmt.Printf("Warning: failed to write mdadm.conf: %v\n", err)
	}

	return arrayName, nil
}

func DeleteRAIDArray(name string) error {
	// Handle both "md0" and "0" formats
	device := name
	if !strings.HasPrefix(name, "/dev/") {
		if !strings.HasPrefix(name, "md") {
			device = "/dev/md" + name
		} else {
			device = "/dev/" + name
		}
	}

	// Get array details for UUID and member devices BEFORE stopping
	detailCmd := exec.Command("sudo", "mdadm", "--detail", device)
	detailOutput, err := detailCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get array details: %w", err)
	}

	// Extract UUID from detail output
	arrayUUID := extractUUIDFromDetail(string(detailOutput))

	// Extract member devices from full detail output
	memberDevices := parseRAIDDevicesFromDetail(string(detailOutput))

	// IMPORTANT: First, wipe any LVM PV metadata from the RAID device
	// This is the #1 cause of "Immutable 36" errors - stale LVM signatures
	fmt.Printf("Wiping any LVM metadata from %s...\n", device)
	pvremoveCmd := exec.Command("sudo", "pvremove", "-ff", "-y", device)
	pvremoveOutput, pvremoveErr := pvremoveCmd.CombinedOutput()
	if pvremoveErr == nil {
		fmt.Printf("Successfully removed LVM PV metadata from %s\n", device)
	} else {
		// pvremove might fail if there's no PV metadata, that's OK
		fmt.Printf("Note: pvremove on %s: %v (this is OK if no LVM metadata exists)\n", device, string(pvremoveOutput))
	}

	// Also check if any of the member devices have LVM metadata and wipe it
	for _, memberDev := range memberDevices {
		fmt.Printf("Checking for LVM metadata on member device %s...\n", memberDev)
		memberPVRemoveCmd := exec.Command("sudo", "pvremove", "-ff", "-y", memberDev)
		memberPVRemoveCmd.Run() // Ignore errors
	}

	// Check if this RAID array is being used as a storage pool and delete it first
	// Storage pools might be mounted at various locations
	mountPoint, _ := getMountPointAndUsage(device)
	if mountPoint != "" && mountPoint != "/" {
		fmt.Printf("RAID array is mounted at %s - attempting to unmount and remove storage pool first...\n", mountPoint)

		// First, try to find and delete the storage pool that uses this mount point
		pools, err := GetStoragePools()
		if err == nil {
			for _, pool := range pools {
				if pool.MountPoint == mountPoint {
					fmt.Printf("Found storage pool '%s' using this RAID array, deleting it first...\n", pool.Name)
					if poolErr := DeleteStoragePool(pool.Name); poolErr != nil {
						fmt.Printf("Warning: failed to delete storage pool: %v\n", poolErr)
						// Continue anyway - the pool might already be partially cleaned up
					} else {
						fmt.Printf("Successfully deleted storage pool '%s'\n", pool.Name)
					}
					break
				}
			}
		}

		// Now try to unmount the array
		fmt.Printf("Unmounting array from %s...\n", mountPoint)
		umountCmd := exec.Command("sudo", "umount", mountPoint)
		if err := umountCmd.Run(); err != nil {
			// Try lazy unmount if normal unmount fails
			fmt.Printf("Normal unmount failed, trying lazy unmount...\n")
			umountCmd = exec.Command("sudo", "umount", "-l", mountPoint)
			if err := umountCmd.Run(); err != nil {
				return fmt.Errorf("failed to unmount array %s: %w (array is in use)", device, err)
			}
		}
		fmt.Printf("Successfully unmounted %s\n", mountPoint)
	}

	// Stop the array using sudo
	output, err := utils.SudoCombinedOutput("mdadm", "--stop", device)
	if err != nil {
		return fmt.Errorf("failed to stop array %s: %w, output: %s", device, err, string(output))
	}

	// Zero the superblock on all member devices to complete removal
	for _, dev := range memberDevices {
		fmt.Printf("Zeroing superblock on %s...\n", dev)
		_, zeroErr := utils.SudoCombinedOutput("mdadm", "--zero-superblock", dev)
		if zeroErr != nil {
			fmt.Printf("Warning: failed to zero superblock on %s: %v\n", dev, zeroErr)
		}
	}

	// Wipe all filesystem and RAID signatures from member devices
	// This is critical to ensure disks are fully freed and reusable
	for _, dev := range memberDevices {
		fmt.Printf("Wiping all signatures from %s...\n", dev)
		wipeOutput, wipeErr := utils.SudoCombinedOutput("wipefs", "-a", dev)
		if wipeErr != nil {
			fmt.Printf("Warning: failed to wipe signatures from %s: %v, output: %s\n", dev, wipeErr, string(wipeOutput))
		} else {
			fmt.Printf("Successfully wiped all signatures from %s\n", dev)
		}
	}

	// Remove array from mdadm.conf
	if arrayUUID != "" {
		if err := removeMDAdmConfEntry(device, arrayUUID); err != nil {
			fmt.Printf("Warning: failed to remove array from mdadm.conf: %v\n", err)
		}
	}

	return nil
}

func AddDiskToRAID(arrayName, device string) error {
	cmd := exec.Command("sudo", "mdadm", "--add", "/dev/"+arrayName, device)
	return cmd.Run()
}

// WipeRAIDSuperblock removes mdadm superblock from a device
// This is useful for cleaning up orphaned RAID metadata from devices
// that were previously part of an array but are no longer active
func WipeRAIDSuperblock(device string) error {
	// Normalize device path
	if !strings.HasPrefix(device, "/dev/") {
		device = "/dev/" + device
	}

	// First, check if device has RAID metadata
	examineCmd := exec.Command("sudo", "mdadm", "--examine", device)
	examineOutput, examineErr := examineCmd.CombinedOutput()
	if examineErr != nil {
		// Device might not have RAID metadata at all
		return fmt.Errorf("device %s does not appear to have RAID metadata: %w", device, examineErr)
	}

	// Check if the array is still active
	examineStr := string(examineOutput)
	if strings.Contains(examineStr, "Array State :") || strings.Contains(examineStr, "State :") {
		// Device might be part of an active array
		return fmt.Errorf("device %s appears to be part of an active RAID array, please delete the array first", device)
	}

	fmt.Printf("Wiping RAID superblock from %s...\n", device)

	// Zero the superblock
	zeroOutput, zeroErr := utils.SudoCombinedOutput("mdadm", "--zero-superblock", device)
	if zeroErr != nil {
		return fmt.Errorf("failed to zero superblock on %s: %w, output: %s", device, zeroErr, string(zeroOutput))
	}

	fmt.Printf("Successfully wiped superblock from %s\n", device)

	// Also wipe any other filesystem/signature remnants for a clean slate
	fmt.Printf("Wiping all signatures from %s...\n", device)
	wipeOutput, wipeErr := utils.SudoCombinedOutput("wipefs", "-a", device)
	if wipeErr != nil {
		fmt.Printf("Warning: failed to wipe all signatures from %s: %v, output: %s\n", device, wipeErr, string(wipeOutput))
		// Don't fail on wipefs errors - the superblock was removed successfully
	} else {
		fmt.Printf("Successfully wiped all signatures from %s\n", device)
	}

	return nil
}

// ExamineRAIDDevice returns RAID metadata information for a device
// This helps identify orphaned RAID superblocks
func ExamineRAIDDevice(device string) (map[string]string, error) {
	// Normalize device path
	if !strings.HasPrefix(device, "/dev/") {
		device = "/dev/" + device
	}

	result := make(map[string]string)

	cmd := exec.Command("sudo", "mdadm", "--examine", device)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Device might not have RAID metadata
		return result, fmt.Errorf("device %s does not have RAID metadata", device)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Parse key-value pairs like "Array UUID : xxx"
		if strings.Contains(line, " : ") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(strings.ReplaceAll(parts[0], " ", "_"))
				value := strings.TrimSpace(parts[1])
				result[key] = value
			}
		}
	}

	return result, nil
}

func parseRAIDDevices(output string) []string {
	var devices []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "devices=") {
			devicesStr := strings.Split(line, "devices=")[1]
			devices = strings.Split(devicesStr, ",")
			break
		}
	}
	return devices
}

func calculateRAIDHealth(state string) int {
	switch state {
	case "clean", "active":
		return 100
	case "degraded":
		return 50
	case "failed", "removing":
		return 0
	default:
		return 75
	}
}

func parseSize(sizeStr string) (int64, error) {
	// Parse size like "1.2T" or "500G"
	sizeStr = strings.TrimSpace(sizeStr)
	if len(sizeStr) < 2 {
		return 0, fmt.Errorf("invalid size format")
	}

	numStr := sizeStr[:len(sizeStr)-1]
	unit := strings.ToUpper(sizeStr[len(sizeStr)-1:])

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, err
	}

	var multiplier int64 = 1
	switch unit {
	case "K":
		multiplier = 1024
	case "M":
		multiplier = 1024 * 1024
	case "G":
		multiplier = 1024 * 1024 * 1024
	case "T":
		multiplier = 1024 * 1024 * 1024 * 1024
	}

	return int64(num * float64(multiplier)), nil
}

func getMountPointAndUsage(device string) (string, int64) {
	// First, try to find mount point using lsblk (more reliable)
	cmd := exec.Command("lsblk", "-no", "MOUNTPOINT", device)
	output, err := cmd.Output()
	if err == nil {
		mountPoint := strings.TrimSpace(string(output))
		if mountPoint != "" && mountPoint != "/" {
			// Device is mounted, get usage
			cmd := exec.Command("df", "-B1", "--output=used", mountPoint)
			dfOutput, dfErr := cmd.Output()
			if dfErr == nil {
				lines := strings.Split(string(dfOutput), "\n")
				if len(lines) >= 2 {
					fields := strings.Fields(lines[1])
					if len(fields) >= 1 {
						used, _ := strconv.ParseInt(fields[0], 10, 64)
						return mountPoint, used
					}
				}
			}
		}
	}

	// Fallback: try df directly on device
	cmd = exec.Command("df", "-B1", "--output=target,used", device)
	output, err = cmd.Output()
	if err != nil {
		return "", 0
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return "", 0
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 2 {
		return "", 0
	}

	mountPoint := fields[0]
	used, _ := strconv.ParseInt(fields[1], 10, 64)

	return mountPoint, used
}

// writeMDAdmConf writes the RAID array configuration to /etc/mdadm/mdadm.conf
// This ensures the array is properly assembled with the correct name on reboot
func writeMDAdmConf(device string) error {
	// Get array details to extract UUID and metadata
	cmd := exec.Command("sudo", "mdadm", "--detail", "--brief", device)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get array details: %w", err)
	}

	// Parse the ARRAY line from output
	// Format: ARRAY /dev/md0 metadata=1.2 UUID=xxx name=arcanas:0
	lines := strings.Split(string(output), "\n")
	var arrayLine string
	for _, line := range lines {
		if strings.HasPrefix(line, "ARRAY ") {
			arrayLine = line
			break
		}
	}

	if arrayLine == "" {
		return fmt.Errorf("no ARRAY line found in mdadm output")
	}

	// Ensure the ARRAY line has all necessary fields
	// Add explicit device mapping to prevent md127 rename
	// Format: ARRAY /dev/md0 metadata=1.2 UUID=xxx name=arcanas:0 devices=/dev/sda,/dev/sdb
	confLine := arrayLine

	// Read existing conf to check for duplicates
	confPath := "/etc/mdadm/mdadm.conf"
	var existingLines []string
	if data, err := os.ReadFile(confPath); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			// Skip comments and empty lines
			if line == "" || strings.HasPrefix(line, "#") {
				existingLines = append(existingLines, line)
				continue
			}
			// Check if this array already exists in conf
			if strings.HasPrefix(line, "ARRAY "+device) {
				// Skip existing entry for this device
				continue
			}
			// Check if the UUID matches (different device path)
			if strings.Contains(confLine, "UUID=") && strings.Contains(line, "UUID=") {
				confUUID := extractUUID(confLine)
				lineUUID := extractUUID(line)
				if confUUID != "" && confUUID == lineUUID {
					// Skip existing entry with same UUID
					continue
				}
			}
			existingLines = append(existingLines, line)
		}
	}

	// Build new conf content
	existingLines = append(existingLines, confLine)
	confContent := strings.Join(existingLines, "\n") + "\n"

	// Write to temp file first, then move (atomic write)
	tempPath := confPath + ".tmp"
	if err := os.WriteFile(tempPath, []byte(confContent), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Move temp file to actual conf (requires sudo via mv)
	mvCmd := exec.Command("sudo", "mv", tempPath, confPath)
	if err := mvCmd.Run(); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to move conf file: %w", err)
	}

	return nil
}

// extractUUID extracts the UUID from an ARRAY line
func extractUUID(line string) string {
	re := regexp.MustCompile(`UUID=([^\s]+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// extractUUIDFromDetail extracts the UUID from full mdadm --detail output
// Output format includes: "UUID : xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
func extractUUIDFromDetail(detailOutput string) string {
	lines := strings.Split(detailOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "UUID :") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// removeMDAdmConfEntry removes an array entry from /etc/mdadm/mdadm.conf
func removeMDAdmConfEntry(device, uuid string) error {
	confPath := "/etc/mdadm/mdadm.conf"

	// Read existing conf
	data, err := os.ReadFile(confPath)
	if err != nil {
		return fmt.Errorf("failed to read mdadm.conf: %w", err)
	}

	// Filter out the entry for this array
	var newLines []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Keep comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			newLines = append(newLines, scanner.Text())
			continue
		}
		// Skip if this line matches our device or UUID
		if strings.Contains(line, device) || (uuid != "" && strings.Contains(line, "UUID="+uuid)) {
			continue
		}
		newLines = append(newLines, scanner.Text())
	}

	// Build new conf content
	confContent := strings.Join(newLines, "\n") + "\n"

	// Write to temp file first, then move (atomic write)
	tempPath := confPath + ".tmp"
	if err := os.WriteFile(tempPath, []byte(confContent), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Move temp file to actual conf (requires sudo via mv)
	mvCmd := exec.Command("sudo", "mv", tempPath, confPath)
	if err := mvCmd.Run(); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to move conf file: %w", err)
	}

	return nil
}
