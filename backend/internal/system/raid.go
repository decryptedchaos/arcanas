/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package system

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"arcanas/internal/models"
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
			// Extract state from "State : clean"
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				array.State = strings.TrimSpace(parts[1])
				// Remove any trailing comma or whitespace
				array.State = strings.TrimSuffix(array.State, ",")
				array.State = strings.TrimSpace(array.State)
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
				devices = append(devices, match)
			}
		}
	}

	return devices
}

func CreateRAIDArray(req models.RAIDCreateRequest) error {
	// Validate RAID level
	validLevels := map[string]bool{
		"raid0": true, "raid1": true, "raid5": true, "raid6": true, "raid10": true,
	}
	if !validLevels[req.Level] {
		return fmt.Errorf("invalid RAID level: %s", req.Level)
	}

	// Create mdadm command with sudo
	var cmd *exec.Cmd
	switch req.Level {
	case "raid0":
		cmd = exec.Command("sudo", "mdadm", "--create", "--verbose", "/dev/md0", "--level=0", "--raid-devices="+strconv.Itoa(len(req.Devices)))
	case "raid1":
		cmd = exec.Command("sudo", "mdadm", "--create", "--verbose", "/dev/md0", "--level=1", "--raid-devices="+strconv.Itoa(len(req.Devices)))
	case "raid5":
		cmd = exec.Command("sudo", "mdadm", "--create", "--verbose", "/dev/md0", "--level=5", "--raid-devices="+strconv.Itoa(len(req.Devices)))
	case "raid6":
		cmd = exec.Command("sudo", "mdadm", "--create", "--verbose", "/dev/md0", "--level=6", "--raid-devices="+strconv.Itoa(len(req.Devices)))
	case "raid10":
		cmd = exec.Command("sudo", "mdadm", "--create", "--verbose", "/dev/md0", "--level=10", "--raid-devices="+strconv.Itoa(len(req.Devices)))
	}

	cmd.Args = append(cmd.Args, req.Devices...)
	return cmd.Run()
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

	// Stop the array using sudo
	cmd := exec.Command("sudo", "mdadm", "--stop", device)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop array %s: %w", device, err)
	}

	// Zero the superblock on all member devices to complete removal
	// Get the member devices before stopping
	detailCmd := exec.Command("sudo", "mdadm", "--detail", "--brief", device)
	output, err := detailCmd.Output()
	if err == nil {
		// Extract member devices from output like: ARRAY /dev/md0 devices=/dev/sda,/dev/sdb
		re := regexp.MustCompile(`devices=([^\s]+)`)
		matches := re.FindStringSubmatch(string(output))
		if len(matches) >= 2 {
			devices := strings.Split(matches[1], ",")
			for _, dev := range devices {
				// Zero the superblock on each device
				zeroCmd := exec.Command("sudo", "mdadm", "--zero-superblock", dev)
				zeroCmd.Run() // Ignore errors, device might not have a superblock
			}
		}
	}

	return nil
}

func AddDiskToRAID(arrayName, device string) error {
	cmd := exec.Command("sudo", "mdadm", "--add", "/dev/"+arrayName, device)
	return cmd.Run()
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
	// Find mount point and usage using df
	cmd := exec.Command("df", "--output=target,used", device)
	output, err := cmd.Output()
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
