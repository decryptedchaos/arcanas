/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package system

import (
	"fmt"
	"os/exec"
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
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Name :") {
			array.Name = strings.Fields(line)[2]
		} else if strings.HasPrefix(line, "Raid Level :") {
			array.Level = strings.Fields(line)[3]
		} else if strings.HasPrefix(line, "Array Size :") {
			sizeStr := strings.Fields(line)[2]
			if size, err := parseSize(sizeStr); err == nil {
				array.Size = size
			}
		} else if strings.HasPrefix(line, "State :") {
			array.State = strings.Fields(line)[2]
		} else if strings.HasPrefix(line, "UUID :") {
			array.UUID = strings.Fields(line)[2]
		}
	}

	// Get device list using sudo
	cmd = exec.Command("sudo", "mdadm", "--detail", "--brief", device)
	briefOutput, _ := cmd.Output()
	array.Devices = parseRAIDDevices(string(briefOutput))

	// Get mount point and usage
	array.MountPoint, array.Used = getMountPointAndUsage(device)

	array.CreatedAt = time.Now() // TODO: Get actual creation time
	array.Health = calculateRAIDHealth(array.State)

	return array, nil
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
	// Stop the array using sudo
	cmd := exec.Command("sudo", "mdadm", "--stop", "/dev/md"+name)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Remove from mdadm.conf
	// TODO: Remove from config file properly

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
