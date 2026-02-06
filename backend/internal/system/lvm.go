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
	"arcanas/internal/utils"
)

// CreateVolumeGroup creates a new LVM volume group from physical devices
func CreateVolumeGroup(req models.VolumeGroupCreateRequest) (models.VolumeGroup, error) {
	vg := models.VolumeGroup{
		Name:      req.Name,
		Devices:   req.Devices,
		CreatedAt: time.Now(),
	}

	// Validate VG name
	if req.Name == "" {
		return vg, fmt.Errorf("volume group name is required")
	}
	if strings.Contains(req.Name, " ") || strings.Contains(req.Name, "/") {
		return vg, fmt.Errorf("volume group name cannot contain spaces or slashes")
	}

	// Check for LVM tools
	if err := checkLVMTools(); err != nil {
		return vg, err
	}

	// Wipe any existing LVM signatures from devices
	for _, device := range req.Devices {
		output, err := utils.SudoCombinedOutput("pvremove", "-ff", "-y", device)
		if err != nil {
			// pvremove returns error if device isn't a PV, that's ok
			if !strings.Contains(string(output), "No physical volume label") {
				return vg, fmt.Errorf("failed to wipe device %s: %v, output: %s", device, err, string(output))
			}
		}
	}

	// Create physical volumes
	for _, device := range req.Devices {
		output, err := utils.SudoCombinedOutput("pvcreate", "-ff", "-y", device)
		if err != nil {
			return vg, fmt.Errorf("failed to create physical volume on %s: %v, output: %s", device, err, string(output))
		}
	}

	// Create volume group
	pvList := strings.Join(req.Devices, " ")
	output, err := utils.SudoCombinedOutput("vgcreate", req.Name, pvList)
	if err != nil {
		// Cleanup: remove PVs we just created
		for _, device := range req.Devices {
			_ = utils.SudoRunCommand("pvremove", "-ff", "-y", device)
		}
		return vg, fmt.Errorf("failed to create volume group: %v, output: %s", err, string(output))
	}

	// Get VG info for size
	vgInfo, err := getVGInfo(req.Name)
	if err != nil {
		return vg, fmt.Errorf("created VG but failed to get info: %w", err)
	}

	vg.Size = vgInfo.Size
	vg.Free = vgInfo.Free

	return vg, nil
}

// DeleteVolumeGroup removes a VG and all its LVs
func DeleteVolumeGroup(name string) error {
	// First, remove all LVs in the VG
	lvs, err := getLogicalVolumes(name)
	if err != nil {
		return fmt.Errorf("failed to list logical volumes: %w", err)
	}

	for _, lv := range lvs {
		lvPath := fmt.Sprintf("/dev/%s/%s", name, lv)
		if err := deleteLogicalVolume(lvPath); err != nil {
			return fmt.Errorf("failed to remove LV %s: %w", lv, err)
		}
	}

	// Remove the VG
	output, err := utils.SudoCombinedOutput("vgremove", "-ff", "-y", name)
	if err != nil {
		return fmt.Errorf("failed to remove volume group: %v, output: %s", err, string(output))
	}

	return nil
}

// GetVolumeGroups returns all volume groups (excluding system VGs)
func GetVolumeGroups() ([]models.VolumeGroup, error) {
	if err := checkLVMTools(); err != nil {
		return []models.VolumeGroup{}, nil // Return empty if LVM not available
	}

	// Use --units b to get raw bytes for accurate parsing
	output, err := utils.SudoCombinedOutput("vgs", "--noheadings", "--separator", "|", "--units", "b", "-o", "vg_name,vg_size,vg_free")
	if err != nil {
		return nil, fmt.Errorf("failed to list volume groups: %v", err)
	}

	var vgs []models.VolumeGroup
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}

		vgName := strings.TrimSpace(parts[0])
		// Parse bytes directly (numeric output from --units b with "B" suffix)
		vgSizeStr := strings.TrimSuffix(strings.TrimSpace(parts[1]), "B")
		vgFreeStr := strings.TrimSuffix(strings.TrimSpace(parts[2]), "B")
		vgSize, _ := strconv.ParseInt(vgSizeStr, 10, 64)
		vgFree, _ := strconv.ParseInt(vgFreeStr, 10, 64)

		// Check if this is a system VG (contains LVs mounted at critical paths)
		// System VGs should not be exposed to the user
		if isSystemVG(vgName) {
			continue
		}

		// Get devices in this VG
		devices, _ := getVGDevices(vgName)

		// Count LUNs (LVs) in this VG
		lvs, _ := getLogicalVolumes(vgName)

		vgs = append(vgs, models.VolumeGroup{
			Name:      vgName,
			Size:      vgSize,
			Free:      vgFree,
			Devices:   devices,
			LUNCount:  len(lvs),
			CreatedAt: time.Now(), // We don't have actual creation time
		})
	}

	return vgs, nil
}

// GetAvailableDevicesForVG returns devices that can be used to create a VG
// We want:
// 1. MD RAID arrays that are NOT mounted and NOT already in a VG
// 2. Physical disks that are completely unused (not RAID members, not system disks, not mounted)
func GetAvailableDevicesForVG() ([]models.BackingStore, error) {
	// Get active MD RAID arrays from /proc/mdstat
	mdstatOutput, _ := utils.SudoCombinedOutput("cat", "/proc/mdstat")

	// Parse /proc/mdstat to get list of active md devices
	var mdDevices []string
	lines := strings.Split(string(mdstatOutput), "\n")
	for _, line := range lines {
		// Lines look like: "md0 : active raid1 sdc[0] sdb[1]"
		if strings.Contains(line, ": active ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				mdName := strings.TrimSuffix(parts[0], ":")
				if strings.HasPrefix(mdName, "md") {
					mdDevices = append(mdDevices, mdName)
				}
			}
		}
	}

	// Build a set of RAID member disks to exclude
	raidMemberDisks := make(map[string]bool)
	for _, line := range lines {
		if strings.Contains(line, ": active ") {
			// Extract member devices from lines like "md0 : active raid1 sdc[0] sdb[1]"
			parts := strings.Fields(line)
			for i, part := range parts {
				// Skip the first 3 parts (md0, :, active, level)
				if i <= 2 {
					continue
				}
				// Remove the [N] suffix from device names like "sdc[0]" -> "sdc"
				devName := strings.TrimSuffix(part, "]")
				devName = strings.TrimSuffix(devName, "[")
				raidMemberDisks[devName] = true
			}
		}
	}

	// Get lsblk output to find physical disks (without -d flag so we see all devices)
	// Use -i flag to force ASCII characters instead of tree drawing characters
	lsblkOutput, _ := utils.SudoCombinedOutput("sh", "-c", "lsblk -i -n -o NAME,TYPE,PKNAME,MOUNTPOINT,FSTYPE 2>/dev/null || true")

	// Get existing VG devices to exclude them
	existingVGDevices := make(map[string]bool)
	vgs, _ := GetVolumeGroups()
	for _, vg := range vgs {
		for _, dev := range vg.Devices {
			existingVGDevices[dev] = true
		}
	}

	var devices []models.BackingStore
	lsblkLines := strings.Split(strings.TrimSpace(string(lsblkOutput)), "\n")

	// Build map of md device names for quick lookup
	mdDeviceMap := make(map[string]bool)
	for _, md := range mdDevices {
		mdDeviceMap[md] = true
	}

	// Track seen devices to avoid duplicates
	seenDevices := make(map[string]bool)

	for _, line := range lsblkLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		name := fields[0]
		devType := fields[1]
		parent := fields[2]
		mountPoint := ""
		fstype := ""
		if len(fields) >= 4 {
			mountPoint = fields[3]
		}
		if len(fields) >= 5 {
			fstype = fields[4]
		}

		// Strip lsblk tree prefixes (e.g., "|-", "`-")
		name = strings.TrimPrefix(name, "|-")
		name = strings.TrimPrefix(name, "`-")
		name = strings.TrimSpace(name)

		// Normalize device path
		device := "/dev/" + name

		// Skip if device is already in a VG
		if existingVGDevices[device] {
			continue
		}

		// Check if this is an md device (RAID array)
		isMDDevice := mdDeviceMap[name]

		// Check if this is a physical disk that's completely unused
		isPhysicalDisk := devType == "disk" && parent == "" && !raidMemberDisks[name]

		// Only accept: md devices OR completely unused physical disks
		if !isMDDevice && !isPhysicalDisk {
			continue
		}

		// Skip physical disks that have a filesystem (might be in use)
		if !isMDDevice && fstype != "" && fstype != "swap" {
			continue
		}

		// Check mount status - verify with findmnt for accuracy
		isMounted := false
		actualMountPoint := ""
		mountOutput, mountErr := utils.SudoCombinedOutput("findmnt", "-n", "-o", "TARGET", "--source", device)
		if mountErr == nil {
			actualMountPoint = strings.TrimSpace(string(mountOutput))
			isMounted = actualMountPoint != ""
		}

		// Skip if mounted (unless it's swap, which is OK)
		if isMounted && mountPoint != "[SWAP]" {
			continue
		}

		// Skip if we've already seen this device (avoid duplicates from lsblk tree output)
		if seenDevices[device] {
			continue
		}
		seenDevices[device] = true

		// Determine device type
		displayType := "disk"
		if isMDDevice {
			displayType = "raid"
		}

		devices = append(devices, models.BackingStore{
			Path:       device,
			Type:       displayType,
			MountPoint: actualMountPoint,
			Available:  true,
			Reason:     "Available for volume group",
		})
	}

	return devices, nil
}

// checkLVMTools verifies LVM tools are installed
func checkLVMTools() error {
	if _, err := utils.SudoCombinedOutput("which", "vgcreate"); err != nil {
		return fmt.Errorf("LVM tools not installed. Install with: sudo apt install lvm2")
	}
	return nil
}

// isSystemVG checks if a VG is a system VG (contains LVs mounted at critical paths)
func isSystemVG(vgName string) bool {
	// Get all LV paths in this VG
	output, err := utils.SudoCombinedOutput("lvs", "--noheadings", "-o", "lv_path", vgName)
	if err != nil {
		return false
	}

	// System mount points that indicate a system VG
	systemMounts := []string{"/", "/boot", "/home", "/usr", "/var"}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		lvPath := line

		// Check if this LV is mounted
		mountOutput, mountErr := utils.SudoCombinedOutput("findmnt", "-n", "-o", "TARGET", "--source", lvPath)
		if mountErr != nil {
			continue
		}

		mountPoint := strings.TrimSpace(string(mountOutput))
		if mountPoint == "" {
			continue
		}

		// Check if this LV is mounted at a system path
		for _, sysMount := range systemMounts {
			if mountPoint == sysMount || strings.HasPrefix(mountPoint, sysMount+"/") {
				return true
			}
		}
	}

	return false
}

// getVGInfo gets detailed info about a VG
func getVGInfo(name string) (struct{ Size, Free int64 }, error) {
	output, err := utils.SudoCombinedOutput("vgs", "--noheadings", "--units", "b", "-o", "vg_size,vg_free", name)
	if err != nil {
		return struct{ Size, Free int64 }{}, err
	}

	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) < 2 {
		return struct{ Size, Free int64 }{}, fmt.Errorf("unexpected vgs output format")
	}

	// Trim the "B" suffix that --units b adds (e.g., "123456789B")
	sizeStr := strings.TrimSuffix(strings.TrimSpace(parts[0]), "B")
	freeStr := strings.TrimSuffix(strings.TrimSpace(parts[1]), "B")
	size, _ := strconv.ParseInt(sizeStr, 10, 64)
	free, _ := strconv.ParseInt(freeStr, 10, 64)

	return struct{ Size, Free int64 }{Size: size, Free: free}, nil
}

// getVGDevices returns the physical devices in a VG
func getVGDevices(name string) ([]string, error) {
	output, err := utils.SudoCombinedOutput("pvdisplay", "-C", "--noheadings", "-o", "pv_name", fmt.Sprintf("--select=%s", name))
	if err != nil {
		return []string{}, nil // Return empty on error
	}

	var devices []string
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			devices = append(devices, line)
		}
	}
	return devices, nil
}

// getLogicalVolumes returns all LVs in a VG
func getLogicalVolumes(vgName string) ([]string, error) {
	output, err := utils.SudoCombinedOutput("lvs", "--noheadings", "-o", "lv_name", vgName)
	if err != nil {
		return []string{}, nil
	}

	var lvs []string
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			lvs = append(lvs, line)
		}
	}
	return lvs, nil
}

// deleteLogicalVolume removes an LV
func deleteLogicalVolume(lvPath string) error {
	output, err := utils.SudoCombinedOutput("lvremove", "-ff", "-y", lvPath)
	if err != nil {
		return fmt.Errorf("failed to remove LV: %v, output: %s", err, string(output))
	}
	return nil
}

// parseVGSize parses size output from vgs (can be like "100.00G" or bytes with --units b)
func parseVGSize(sizeStr string) int64 {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" || sizeStr == "0" {
		return 0
	}

	// If already in bytes (numeric), parse directly
	if strings.IndexFunc(sizeStr, func(r rune) bool { return r < '0' || r > '9' }) == -1 {
		size, _ := strconv.ParseInt(sizeStr, 10, 64)
		return size
	}

	// Parse human-readable format (e.g., "100.00G", "512M")
	numStr := ""
	unit := ""
	for i, r := range sizeStr {
		if r >= '0' && r <= '9' || r == '.' {
			numStr += string(r)
		} else {
			unit = sizeStr[i:]
			break
		}
	}

	num, _ := strconv.ParseFloat(numStr, 64)
	multiplier := int64(1)
	switch strings.ToUpper(unit) {
	case "T", "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "G", "GB":
		multiplier = 1024 * 1024 * 1024
	case "M", "MB":
		multiplier = 1024 * 1024
	case "K", "KB":
		multiplier = 1024
	}

	return int64(num * float64(multiplier))
}

// getDeviceReason returns a human-readable reason for device availability
func getDeviceReason(device string, available bool, mountPoint string) string {
	if available {
		return "Available for volume group"
	}
	if mountPoint != "" {
		return fmt.Sprintf("Mounted at %s", mountPoint)
	}
	return "In use"
}

// CreateLogicalVolume creates a new LV from a VG
func CreateLogicalVolume(req models.LVCreateRequest) (models.LogicalVolume, error) {
	var lv models.LogicalVolume

	// Validate inputs
	if req.Name == "" {
		return lv, fmt.Errorf("LV name is required")
	}
	if req.VGName == "" {
		return lv, fmt.Errorf("VG name is required")
	}
	if req.SizeGB <= 0 {
		return lv, fmt.Errorf("size must be greater than 0")
	}

	// Check if VG exists
	vgInfo, err := getVGInfo(req.VGName)
	if err != nil {
		return lv, fmt.Errorf("VG not found: %w", err)
	}

	// Check if there's enough free space
	requiredBytes := int64(req.SizeGB * 1024 * 1024 * 1024)
	if vgInfo.Free < requiredBytes {
		return lv, fmt.Errorf("not enough free space in VG (have %d GB, need %.2f GB)",
			vgInfo.Free/(1024*1024*1024), req.SizeGB)
	}

	// Create the LV (force wipe signatures and auto-confirm to avoid interactive prompts)
	lvPath := fmt.Sprintf("/dev/%s/%s", req.VGName, req.Name)
	output, err := utils.SudoCombinedOutput("lvcreate", "-L", fmt.Sprintf("%.0fG", req.SizeGB), "-n", req.Name, "-W", "y", "-y", req.VGName)
	if err != nil {
		return lv, fmt.Errorf("failed to create LV: %v, output: %s", err, string(output))
	}

	// Get LV size for the response
	sizeCmd := exec.Command("sudo", "lvs", "--noheadings", "--units", "b", "-o", "lv_size", lvPath)
	sizeOutput, err := sizeCmd.Output()
	if err != nil {
		return lv, fmt.Errorf("created LV but failed to get size: %w", err)
	}
	actualSize, _ := strconv.ParseInt(strings.TrimSpace(string(sizeOutput)), 10, 64)

	lv = models.LogicalVolume{
		Name:       req.Name,
		Path:       lvPath,
		VGName:     req.VGName,
		Size:       actualSize,
		MountPoint: "",
		Available:  true,
		UsedFor:    "available",
		CreatedAt:  time.Now(),
	}

	return lv, nil
}

// MountLVAsPool mounts an existing LV as a storage pool at /srv/{poolName}
func MountLVAsPool(lvPath, poolName string) error {
	mountPoint := "/srv/" + poolName

	// Verify the LV exists
	cmd := exec.Command("sudo", "lvs", lvPath, "-o", "LV_SIZE")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("logical volume not found: %s", lvPath)
	}

	// Ensure /srv directory exists
	cmd = exec.Command("sudo", "mkdir", "-p", "/srv")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create /srv directory: %v", err)
	}

	// Check if LV has a filesystem, format if needed
	cmd = exec.Command("blkid", "-o", "value", "-s", "TYPE", lvPath)
	fsOutput, err := cmd.Output()
	hasFilesystem := err == nil && len(strings.TrimSpace(string(fsOutput))) > 0

	if !hasFilesystem {
		fmt.Printf("Formatting LV %s with ext4...\n", lvPath)
		formatCmd := exec.Command("sudo", "mkfs.ext4", "-F", lvPath)
		if out, err := formatCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("mkfs.ext4 failed: %v %s", err, string(out))
		}
	}

	// Create mount point
	cmd = exec.Command("sudo", "mkdir", "-p", mountPoint)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create mount point %s: %v", mountPoint, err)
	}

	// Mount the LV
	if out, err := exec.Command("sudo", "mount", lvPath, mountPoint).CombinedOutput(); err != nil {
		return fmt.Errorf("mount failed: %v %s", err, string(out))
	}

	// Set permissions
	cmd = exec.Command("sudo", "chown", "-R", "nobody:nogroup", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Warning: failed to set pool ownership: %v, output: %s\n", err, string(output))
	}

	cmd = exec.Command("sudo", "chmod", "-R", "0777", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Warning: failed to set pool permissions: %v, output: %s\n", err, string(output))
	}

	// Add to fstab for persistence
	cmd = exec.Command("grep", "-q", mountPoint, "/etc/fstab")
	if err := cmd.Run(); err != nil {
		fstabEntry := fmt.Sprintf("%s %s ext4 defaults 0 0\n", lvPath, mountPoint)
		cmd = exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' >> /etc/fstab", fstabEntry))
		cmd.Run()
	}

	fmt.Printf("Successfully mounted LV pool %s from %s at %s (size: %s)\n", poolName, lvPath, mountPoint, strings.TrimSpace(string(output)))
	return nil
}

// GetLogicalVolumes returns all LVs with their VG and mount information
func GetLogicalVolumes() ([]models.LogicalVolume, error) {
	var lvs []models.LogicalVolume

	// Get all VGs
	vgs, err := GetVolumeGroups()
	if err != nil {
		return lvs, err
	}

	// Get LVs from each VG
	for _, vg := range vgs {
		// Get LV names in this VG
		lvNames, err := getLogicalVolumes(vg.Name)
		if err != nil {
			continue
		}

		for _, lvName := range lvNames {
			lvPath := fmt.Sprintf("/dev/%s/%s", vg.Name, lvName)

			// Get LV size
			sizeOutput, err := utils.SudoCombinedOutput("lvs", "--noheadings", "--units", "b", "-o", "lv_size", lvPath)
			lvSize := int64(0)
			if err == nil {
				sizeStr := strings.TrimSpace(string(sizeOutput))
				lvSize, _ = strconv.ParseInt(sizeStr, 10, 64)
			}

			// Check if mounted
			mountPoint := ""
			isMounted := false
			mountOutput, mountErr := utils.SudoCombinedOutput("findmnt", "-n", "-o", "TARGET", "--source", lvPath)
			if mountErr == nil {
				mountPoint = strings.TrimSpace(string(mountOutput))
				isMounted = mountPoint != ""
			}

			// Determine used_for based on mount point
			usedFor := "available"
			if isMounted {
				// Check if mounted as storage pool (at /srv/)
				if strings.HasPrefix(mountPoint, "/srv/") {
					usedFor = "pool"
				} else {
					usedFor = "mounted"
				}
			}

			// Check if used by iSCSI
			// TODO: Add iSCSI LUN checking

			lvs = append(lvs, models.LogicalVolume{
				Name:       lvName,
				Path:       lvPath,
				VGName:     vg.Name,
				Size:       lvSize,
				MountPoint: mountPoint,
				Available: !isMounted,
				UsedFor:    usedFor,
				CreatedAt:  time.Now(),
			})
		}
	}

	return lvs, nil
}

// DeleteLogicalVolume removes an LV and unmounts it first if mounted
func DeleteLogicalVolumeByName(lvPath string) error {
	// Check if it's mounted
	mountPoint, err := getMountPointForDevice(lvPath)
	if err == nil && mountPoint != "" {
		// Try to unmount first
		fmt.Printf("LV is mounted at %s, unmounting...\n", mountPoint)
		umountCmd := exec.Command("sudo", "umount", mountPoint)
		if err := umountCmd.Run(); err != nil {
			// Try lazy unmount
			umountCmd = exec.Command("sudo", "umount", "-l", mountPoint)
			if err := umountCmd.Run(); err != nil {
				return fmt.Errorf("failed to unmount LV at %s: %w (cannot delete mounted LV)", mountPoint, err)
			}
		}

		// Remove from fstab
		cmd := exec.Command("sudo", "sed", "-i", fmt.Sprintf("\\|%s|d", mountPoint), "/etc/fstab")
		cmd.Run() // Ignore errors
	}

	// Delete the LV
	if err := deleteLogicalVolume(lvPath); err != nil {
		return err
	}

	fmt.Printf("Successfully deleted LV %s\n", lvPath)
	return nil
}

// getMountPointForDevice returns the mount point of a device
func getMountPointForDevice(device string) (string, error) {
	cmd := exec.Command("findmnt", "-n", "-o", "TARGET", "--source", device)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
