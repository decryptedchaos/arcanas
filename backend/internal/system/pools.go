/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package system

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"arcanas/internal/models"
)

func GetStoragePools() ([]models.StoragePool, error) {
	// Initialize as empty slice to avoid null JSON encoding
	pools := make([]models.StoragePool, 0)

	// Get all mounts (ext4/xfs/btrfs + mergerfs) at once to simplify parsing
	cmd := exec.Command("findmnt", "-J")
	output, err := cmd.Output()
	if err != nil {
		return pools, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return pools, nil
	}

	// Recursive function to process filesystems (including nested children)
	var processFilesystem func(fs map[string]interface{}) []models.StoragePool
	processFilesystem = func(fs map[string]interface{}) []models.StoragePool {
		var result []models.StoragePool

		target, _ := fs["target"].(string)
		fstype, _ := fs["fstype"].(string)
		source, _ := fs["source"].(string)

		// Skip system and special mounts that shouldn't be managed
		if strings.HasPrefix(target, "/boot") ||
			strings.HasPrefix(target, "/efi") ||
			strings.HasPrefix(target, "/[") ||
			target == "/" ||
			strings.HasPrefix(target, "/run") ||
			strings.HasPrefix(target, "/sys") ||
			strings.HasPrefix(target, "/proc") ||
			strings.HasPrefix(target, "/dev") {
			// Process children recursively
			if children, ok := fs["children"].([]interface{}); ok {
				for _, childRaw := range children {
					if child, ok := childRaw.(map[string]interface{}); ok {
						result = append(result, processFilesystem(child)...)
					}
				}
			}
			return result
		}

		// Check if this is an arcanas-managed mount or an MD device
		isArcanasMount := strings.HasPrefix(target, "/srv/") ||
			strings.HasPrefix(target, "/mnt/arcanas-disk-")
		isMDDevice := strings.HasPrefix(source, "/dev/md")

		// Only process arcanas mounts or MD devices with filesystems
		if !isArcanasMount && !isMDDevice {
			// Process children recursively
			if children, ok := fs["children"].([]interface{}); ok {
				for _, childRaw := range children {
					if child, ok := childRaw.(map[string]interface{}); ok {
						result = append(result, processFilesystem(child)...)
					}
				}
			}
			return result
		}

		// Skip the base directories
		if target == "/srv/" || target == "/mnt/arcanas-disk" {
			return result
		}

		// Determine pool name and type based on mount location
		var poolName string
		var poolType string

		if strings.HasPrefix(target, "/srv/") {
			// New architecture: /srv/{poolname}
			poolName = strings.TrimPrefix(target, "/srv/")
			poolType = "direct"
		} else if strings.HasPrefix(target, "/mnt/arcanas-disk-") {
			// Legacy architecture: /mnt/arcanas-disk-{device}
			deviceName := strings.TrimPrefix(target, "/mnt/arcanas-disk-")
			poolName = deviceName // Use device name as pool name (e.g., "md0")
			poolType = "legacy"
		} else if isMDDevice {
			// MD device mounted elsewhere (e.g., /mnt/md0, /media/raid)
			// Use mount point basename as pool name
			poolName = filepath.Base(target)
			poolType = "md"
		} else {
			poolName = filepath.Base(target)
			poolType = "directory"
		}

		pool := models.StoragePool{
			Name:       poolName,
			MountPoint: target,
			State:      "active",
			CreatedAt:  time.Now(),
		}

		// Determine pool type based on filesystem (override default)
		if fstype == "fuse.mergerfs" {
			pool.Type = "mergerfs"
		} else if isMDDevice {
			pool.Type = "md" // RAID array
		} else if fstype == "ext4" || fstype == "xfs" || fstype == "btrfs" {
			if poolType == "legacy" {
				pool.Type = "legacy" // MD devices at old location
			} else {
				pool.Type = poolType
			}
		} else {
			pool.Type = poolType
		}

		// Get source device(s) from fstab
		fstabData, err := os.ReadFile("/etc/fstab")
		if err == nil {
			fstabLines := strings.Split(string(fstabData), "\n")
			for _, line := range fstabLines {
				if strings.Contains(line, target) {
					fields := strings.Fields(line)
					if len(fields) >= 1 {
						src := fields[0]
						if src != "" && src != "none" {
							// Split by colon for mergerfs, single device otherwise
							if strings.Contains(line, "fuse.mergerfs") {
								pool.Devices = strings.Split(src, ":")
							} else {
								pool.Devices = []string{src}
							}
							break
						}
					}
				}
			}
		}

		// Fallback: use source from findmnt
		if len(pool.Devices) == 0 && source != "" {
			pool.Devices = []string{source}
		}

		// Get size and usage
		pool.Size, pool.Used = GetPathUsage(pool.MountPoint)
		pool.Available = pool.Size - pool.Used

		// Set export mode based on current state
		// Mounted pools default to "file" mode (for NFS/Samba)
		// Unmounted pools would be "iscsi" or "available"
		pool.ExportMode = "file" // Default: mounted filesystem for file sharing

		// Fix permissions on MD device mounts for Samba access
		// Only for MD devices mounted outside Arcanas-managed locations
		if isMDDevice && !strings.HasPrefix(target, "/srv/") && !strings.HasPrefix(target, "/mnt/arcanas-disk-") {
			// Set ownership to nobody:nogroup for Samba access
			cmd := exec.Command("sudo", "chown", "-R", "nobody:nogroup", target)
			if output, err := cmd.CombinedOutput(); err != nil {
				fmt.Printf("Warning: failed to set MD device ownership: %v, output: %s\n", err, string(output))
			}

			// Set permissions to 0777 for full access
			cmd = exec.Command("sudo", "chmod", "-R", "0777", target)
			if output, err := cmd.CombinedOutput(); err != nil {
				fmt.Printf("Warning: failed to set MD device permissions: %v, output: %s\n", err, string(output))
			}

			// Set setgid bit for proper group inheritance
			cmd = exec.Command("sudo", "chmod", "g+s", target)
			if output, err := cmd.CombinedOutput(); err != nil {
				fmt.Printf("Warning: failed to set setgid bit: %v, output: %s\n", err, string(output))
			}
		}

		result = append(result, pool)

		// Process children recursively
		if children, ok := fs["children"].([]interface{}); ok {
			for _, childRaw := range children {
				if child, ok := childRaw.(map[string]interface{}); ok {
					result = append(result, processFilesystem(child)...)
				}
			}
		}

		return result
	}

	// Process all top-level filesystems
	if filesystems, ok := result["filesystems"].([]interface{}); ok {
		for _, fsRaw := range filesystems {
			if fs, ok := fsRaw.(map[string]interface{}); ok {
				pools = append(pools, processFilesystem(fs)...)
			}
		}
	}

	// Check for existing pool directories (even if not mounted)
	arcanasDir := "/srv"
	if dirEntries, err := os.ReadDir(arcanasDir); err == nil {
		for _, entry := range dirEntries {
			if !entry.IsDir() {
				continue
			}

			poolName := entry.Name()

			// Skip system directories
			if poolName == "ftp" || poolName == "http" {
				continue
			}

			poolPath := filepath.Join(arcanasDir, poolName)

			// Skip if already detected as mounted pool
			alreadyMounted := false
			for _, pool := range pools {
				if pool.Name == poolName {
					alreadyMounted = true
					break
				}
			}
			if alreadyMounted {
				continue
			}

			// Add as unmounted pool
			pool := models.StoragePool{
				Name:       poolName,
				Type:       "directory",
				MountPoint: poolPath,
				State:      "inactive",
				CreatedAt:  time.Now(),
			}

			// Try to get size and usage if path exists
			if _, err := os.Stat(poolPath); err == nil {
				pool.Size, pool.Used = GetPathUsage(poolPath)
				pool.Available = pool.Size - pool.Used
			}

			pools = append(pools, pool)
		}
	}

	return pools, nil
}

func CreateStoragePool(req models.StoragePoolCreateRequest) error {
	switch req.Type {
	case "jbod", "mergerfs":
		// MD devices are now handled via LVM, not auto-mounted
		// JBOD with mergerfs only applies to raw physical disks (sda, sdb, etc.)
		if len(req.Devices) == 1 {
			return fmt.Errorf("single device pools should use 'lvm' type. Create a VG from the device first, then create an LV pool")
		}
		return createMergerFSPool(req)
	case "bind":
		return createBindMountPool(req)
	case "lvm", "lvm_lv":
		// Mount an existing LVM Logical Volume as a storage pool
		// req.Devices[0] should be the LV path (e.g., /dev/vg-name/lv-name)
		if len(req.Devices) != 1 {
			return fmt.Errorf("LVM pools require exactly one logical volume device")
		}
		return createLVMountPool(req)
	default:
		return fmt.Errorf("unsupported pool type: %s", req.Type)
	}
}

// isRAIDDevice checks if a device is an MD RAID array (md0, md1, etc.)
func isRAIDDevice(device string) bool {
	// Check if device path starts with /dev/md
	return strings.HasPrefix(device, "/dev/md")
}

// createLVMountPool mounts an existing LVM Logical Volume
// This is used when you've already created VG+LV and want to mount it as a storage pool
func createLVMountPool(req models.StoragePoolCreateRequest) error {
	lvPath := req.Devices[0] // e.g., /dev/vg-raid/lv-data
	mountPoint := "/srv/" + req.Name

	// Verify the LV exists
	cmd := exec.Command("sudo", "lvs", lvPath, "-o", "LV_SIZE")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("logical volume not found: %s. Create it first using: sudo lvcreate -L 100G -n lv-name vg-name", lvPath)
	}

	// Ensure data directory exists
	cmd = exec.Command("sudo", "mkdir", "-p", "/srv")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create data directory /srv: %v", err)
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

	// Add to fstab for persistence (use LV path directly since it persists)
	cmd = exec.Command("grep", "-q", mountPoint, "/etc/fstab")
	if err := cmd.Run(); err != nil {
		fstabEntry := fmt.Sprintf("%s %s ext4 defaults 0 0\n", lvPath, mountPoint)
		cmd = exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' >> /etc/fstab", fstabEntry))
		cmd.Run()
	}

	fmt.Printf("Successfully mounted LV pool %s from %s at %s (size: %s)\n", req.Name, lvPath, mountPoint, strings.TrimSpace(string(output)))
	return nil
}

func createMergerFSPool(req models.StoragePoolCreateRequest) error {
	// Check if mergerfs is available
	if _, err := exec.LookPath("mergerfs"); err != nil {
		return fmt.Errorf("mergerfs is not installed. Install with:\nUbuntu/Debian: sudo apt install mergerfs\nFedora/CentOS: sudo dnf install mergerfs\nArch: sudo pacman -S mergerfs\nOr download from: https://github.com/trapexit/mergerfs/releases")
	}

	// Prepare each disk (Format -> persistent mount)
	var sourcePaths []string
	for _, device := range req.Devices {
		mountPath, err := prepareDiskForPool(device)
		if err != nil {
			return fmt.Errorf("failed to prepare device %s: %v", device, err)
		}
		sourcePaths = append(sourcePaths, mountPath)
	}

	// Ensure data directory exists and has proper permissions
	cmd := exec.Command("sudo", "mkdir", "-p", "/srv")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create data directory /srv: %v, output: %s", err, string(output))
	}

	// Set ownership of data directory to arcanas user if it exists
	cmd = exec.Command("sudo", "chown", "-R", "arcanas:arcanas", "/srv")
	if output, err := cmd.CombinedOutput(); err != nil {
		// Log warning but don't fail if arcanas user doesn't exist
		fmt.Printf("Warning: failed to set data directory ownership: %v, output: %s\n", err, string(output))
	}

	// Create mount point for the POOL
	mountPoint := "/srv/" + req.Name
	cmd = exec.Command("sudo", "mkdir", "-p", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create pool mount point %s: %v, output: %s", mountPoint, err, string(output))
	}

	// Build mergerfs command using MOUNT POINTS, not raw devices
	// MergerFS requires paths to be separated by colons
	// Use a special syntax to ensure proper parsing: /path1:/path2
	devicesStr := strings.Join(sourcePaths, ":")

	config := req.Config
	if config == "" {
		config = "defaults,allow_other,use_ino"
	}

	// Mount with mergerfs using systemd-run for persistence
	// This ensures mergerfs runs as a independent service, not a child of arcanas
	// Using systemd-run --user would be ideal but requires user session, using system scope
	unitName := "mergerfs-" + strings.ReplaceAll(req.Name, "-", "--") + ".mount"
	cmd = exec.Command("sudo", "systemd-run", "--unit="+unitName, "--scope", "mergerfs", devicesStr, mountPoint, "-o", config)
	if _, err := cmd.CombinedOutput(); err != nil {
		// Fallback: try direct mount (may not persist across arcanas restarts)
		fmt.Printf("systemd-run failed, trying direct mount: %v\n", err)
		cmd = exec.Command("sudo", "mergerfs", devicesStr, mountPoint, "-o", config)
		if output, err := cmd.CombinedOutput(); err != nil {
			// Cleanup mount point on failure
			exec.Command("sudo", "rmdir", mountPoint).Run()
			return fmt.Errorf("failed to mount mergerfs: %v, output: %s", err, string(output))
		}
	}

	// Verify the mount was successful
	if !isMounted(mountPoint) {
		exec.Command("sudo", "rmdir", mountPoint).Run()
		return fmt.Errorf("mergerfs mount verification failed for %s", mountPoint)
	}

	// Set permissions for the mounted pool so Samba and other services can access it
	// Set ownership to allow Samba to write (nogroup is the default guest account)
	cmd = exec.Command("sudo", "chown", "-R", "nobody:nogroup", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Log warning but don't fail - this might not be critical
		fmt.Printf("Warning: failed to set pool ownership: %v, output: %s\n", err, string(output))
	}

	// Set permissions to allow read/write access
	cmd = exec.Command("sudo", "chmod", "-R", "0777", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Log warning but don't fail
		fmt.Printf("Warning: failed to set pool permissions: %v, output: %s\n", err, string(output))
	}

	// Add pool to fstab for persistence
	fstabEntry := fmt.Sprintf("%s %s fuse.mergerfs %s 0 0\n", devicesStr, mountPoint, config)
	cmd = exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' >> /etc/fstab", fstabEntry))
	if err := cmd.Run(); err != nil {
		// Log warning but don't fail - fstab update is secondary
		fmt.Printf("Warning: failed to add to fstab: %v\n", err)
	}

	return nil
}

// prepareDiskForPool ensures a device is formatted and mounted, returning its mount point
func prepareDiskForPool(device string) (string, error) {
	//1. Check if already mounted
	cmd := exec.Command("findmnt", "-n", "-o", "TARGET", "--source", device)
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return strings.TrimSpace(string(output)), nil
	}

	//2. Check if device has a filesystem already
	cmd = exec.Command("blkid", "-o", "value", "-s", "TYPE", device)
	fsOutput, err := cmd.Output()
	hasFilesystem := err == nil && len(strings.TrimSpace(string(fsOutput))) > 0

	if hasFilesystem {
		fmt.Printf("Device %s already has filesystem, skipping format\n", device)
	} else {
		// Format disk (ext4) if needed
		// Use -F to force if it looks like it has a partition table
		fmt.Printf("Formatting device %s...\n", device)
		formatCmd := exec.Command("sudo", "mkfs.ext4", "-F", device)
		if out, err := formatCmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("mkfs.ext4 failed: %v %s", err, string(out))
		}
	}

	//3. Get UUID for device (more reliable than device path)
	// This handles md0 -> md127 rename issue
	uuidCmd := exec.Command("blkid", "-s", "UUID", "-o", "value", device)
	uuidOutput, err := uuidCmd.Output()
	uuid := strings.TrimSpace(string(uuidOutput))
	if uuid == "" || err != nil {
		// Fallback to device path if UUID not available
		fmt.Printf("Warning: could not get UUID for %s, using device path\n", device)
		uuid = device
	} else {
		// Prepend "UUID=" for fstab
		uuid = "UUID=" + uuid
	}

	//4. Create persistent mount point
	// Naming: /mnt/arcanas-disk-{devname} (e.g. sdb)
	devName := strings.TrimPrefix(device, "/dev/")
	mountPath := "/mnt/arcanas-disk-" + devName

	if err := exec.Command("sudo", "mkdir", "-p", mountPath).Run(); err != nil {
		return "", fmt.Errorf("failed to make dir %s: %v", mountPath, err)
	}

	//5. Mount it using UUID (or device path as fallback)
	mountSource := uuid
	if strings.HasPrefix(uuid, "UUID=") {
		// For UUID mounting, we need to resolve the actual device
		// But mount accepts UUID= syntax directly
		mountSource = uuid
	} else {
		mountSource = device
	}

	if out, err := exec.Command("sudo", "mount", mountSource, mountPath).CombinedOutput(); err != nil {
		return "", fmt.Errorf("mount failed: %v %s", err, string(out))
	}

	// Set ownership and permissions on the mount point for Samba access
	// This ensures shares created from these paths have proper access
	cmd = exec.Command("sudo", "chown", "-R", "nobody:nogroup", mountPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Warning: failed to set disk mount ownership: %v, output: %s\n", err, string(output))
	}

	cmd = exec.Command("sudo", "chmod", "-R", "0777", mountPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Warning: failed to set disk mount permissions: %v, output: %s\n", err, string(output))
	}

	//6. Add to fstab (for disk persistence)
	// Check if entry already exists to avoid duplicates
	cmd = exec.Command("grep", "-q", mountPath, "/etc/fstab")
	err = cmd.Run()
	if err != nil {
		// Not found in fstab, add it
		// Use UUID for resilience against device renaming (md0 -> md127)
		fstabEntry := fmt.Sprintf("%s %s ext4 defaults 0 0\n", uuid, mountPath)
		exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' >> /etc/fstab", fstabEntry)).Run()
	}

	return mountPath, nil
}

func createBindMountPool(req models.StoragePoolCreateRequest) error {
	if len(req.Devices) != 1 {
		return fmt.Errorf("bind mount pools require exactly one device")
	}

	// Verify source path exists
	sourcePath := req.Devices[0]
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source path %s does not exist", sourcePath)
	}

	// Ensure data directory exists and has proper permissions
	cmd := exec.Command("sudo", "mkdir", "-p", "/srv")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create data directory /srv: %v, output: %s", err, string(output))
	}

	// Set ownership of data directory to arcanas user if it exists
	cmd = exec.Command("sudo", "chown", "-R", "arcanas:arcanas", "/srv")
	if output, err := cmd.CombinedOutput(); err != nil {
		// Log warning but don't fail if arcanas user doesn't exist
		fmt.Printf("Warning: failed to set data directory ownership: %v, output: %s\n", err, string(output))
	}

	// Create mount point using sudo
	mountPoint := "/srv/" + req.Name
	cmd = exec.Command("sudo", "mkdir", "-p", mountPoint)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create mount point: %v", err)
	}

	// Create bind mount
	cmd = exec.Command("sudo", "mount", "--bind", sourcePath, mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Cleanup mount point on failure
		exec.Command("sudo", "rmdir", mountPoint).Run()
		return fmt.Errorf("failed to create bind mount: %v, output: %s", err, string(output))
	}

	// Set permissions for the mounted pool
	cmd = exec.Command("sudo", "chown", "-R", "nobody:nogroup", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Warning: failed to set pool ownership: %v, output: %s\n", err, string(output))
	}

	cmd = exec.Command("sudo", "chmod", "-R", "0777", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Warning: failed to set pool permissions: %v, output: %s\n", err, string(output))
	}

	// Add to fstab for persistence
	fstabEntry := fmt.Sprintf("%s %s none bind 0 0\n", sourcePath, mountPoint)
	cmd = exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' >> /etc/fstab", fstabEntry))
	if err := cmd.Run(); err != nil {
		// Log warning but don't fail - fstab update is secondary
		fmt.Printf("Warning: failed to add to fstab: %v\n", err)
	}

	return nil
}

func createLVMPool(req models.StoragePoolCreateRequest) error {
	// Check if LVM tools are available
	if _, err := exec.LookPath("lvcreate"); err != nil {
		return fmt.Errorf("LVM tools not installed. Install with: sudo apt install lvm2")
	}

	if len(req.Devices) == 0 {
		return fmt.Errorf("at least one device is required for LVM")
	}

	// Ensure data directory exists and has proper permissions
	cmd := exec.Command("sudo", "mkdir", "-p", "/srv")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create data directory /srv: %v, output: %s", err, string(output))
	}

	// Create volume group
	vgName := "vg_" + req.Name
	cmd = exec.Command("vgcreate", append([]string{vgName}, req.Devices...)...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create volume group: %v", err)
	}

	// Create logical volume (use 90% of available space)
	lvName := "lv_" + req.Name
	cmd = exec.Command("lvcreate", "-l", "90%VG", "-n", lvName, vgName)
	if err := cmd.Run(); err != nil {
		// Cleanup volume group on failure
		exec.Command("vgremove", "-f", vgName).Run()
		return fmt.Errorf("failed to create logical volume: %v", err)
	}

	// Create filesystem
	lvPath := fmt.Sprintf("/dev/%s/%s", vgName, lvName)
	cmd = exec.Command("mkfs.ext4", lvPath)
	if err := cmd.Run(); err != nil {
		// Cleanup on failure
		exec.Command("lvremove", "-f", lvPath).Run()
		exec.Command("vgremove", "-f", vgName).Run()
		return fmt.Errorf("failed to create filesystem: %v", err)
	}

	// Create mount point using sudo
	mountPoint := "/srv/" + req.Name
	cmd = exec.Command("sudo", "mkdir", "-p", mountPoint)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create mount point: %v", err)
	}

	// Mount the logical volume
	cmd = exec.Command("sudo", "mount", lvPath, mountPoint)
	if err := cmd.Run(); err != nil {
		exec.Command("sudo", "rmdir", mountPoint).Run()
		return fmt.Errorf("failed to mount logical volume: %v", err)
	}

	// Set permissions for the mounted pool so Samba and other services can access it
	// Set ownership to allow Samba to write (nogroup is the default guest account)
	cmd = exec.Command("sudo", "chown", "-R", "nobody:nogroup", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Log warning but don't fail - this might not be critical
		fmt.Printf("Warning: failed to set pool ownership: %v, output: %s\n", err, string(output))
	}

	// Set permissions to allow read/write access
	cmd = exec.Command("sudo", "chmod", "-R", "0777", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Log warning but don't fail
		fmt.Printf("Warning: failed to set pool permissions: %v, output: %s\n", err, string(output))
	}

	// Add to fstab for persistence
	fstabEntry := fmt.Sprintf("%s %s ext4 defaults 0 0\n", lvPath, mountPoint)
	cmd = exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' >> /etc/fstab", fstabEntry))
	if err := cmd.Run(); err != nil {
		// Log warning but don't fail - fstab update is secondary
		fmt.Printf("Warning: failed to add to fstab: %v\n", err)
	}

	return nil
}

func UpdateStoragePool(name string, req models.StoragePoolCreateRequest) error {
	// For mergerfs, we need to remount with new config
	mountPoint := "/srv/" + name

	// Unmount first
	cmd := exec.Command("sudo", "umount", mountPoint)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to unmount pool: %v", err)
	}

	// Remount with new config
	devicesStr := strings.Join(req.Devices, ":")
	config := req.Config
	if config == "" {
		config = "defaults,allow_other,use_ino"
	}

	cmd = exec.Command("sudo", "mergerfs", devicesStr, mountPoint, "-o", config)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remount mergerfs pool: %v", err)
	}

	// Update fstab entry
	// First remove old entry
	cmd = exec.Command("sudo", "sed", "-i", fmt.Sprintf("\\|%s|d", mountPoint), "/etc/fstab")
	if err := cmd.Run(); err != nil {
		// Log warning but don't fail - entry might not exist
		fmt.Printf("Warning: failed to remove from fstab: %v\n", err)
	}

	// Add new entry
	fstabEntry := fmt.Sprintf("%s %s fuse.mergerfs %s 0 0\n", devicesStr, mountPoint, config)
	cmd = exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' >> /etc/fstab", fstabEntry))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update fstab: %v", err)
	}

	return nil
}

func DeleteStoragePool(name string) error {
	// First, find the actual mount point of this pool
	// Pools can be at different locations:
	// - /srv/{name} (new direct mount)
	// - /mnt/arcanas-disk-{name} (legacy)
	// - Other locations for MD devices
	pools, err := GetStoragePools()
	if err != nil {
		return fmt.Errorf("failed to get storage pools: %v", err)
	}

	var mountPoint string
	var foundPool bool
	for _, pool := range pools {
		if pool.Name == name {
			mountPoint = pool.MountPoint
			foundPool = true
			break
		}
	}

	if !foundPool {
		return fmt.Errorf("pool %s not found", name)
	}

	// Check if samba is running and stop it temporarily
	sambaRunning := false
	if cmd := exec.Command("sudo", "systemctl", "is-active", "smbd").Run(); cmd == nil {
		sambaRunning = true
		fmt.Printf("Stopping Samba for pool deletion...\n")
		if output, err := exec.Command("sudo", "systemctl", "stop", "smbd").CombinedOutput(); err != nil {
			return fmt.Errorf("failed to stop Samba: %v, output: %s", err, string(output))
		}
	}

	// Check if it's currently mounted and unmount if needed
	if isMounted(mountPoint) {
		// Try normal unmount first
		cmd := exec.Command("sudo", "umount", mountPoint)
		output, err := cmd.CombinedOutput()
		if err != nil {
			// If normal unmount fails, try lazy unmount
			fmt.Printf("Normal unmount failed, trying lazy unmount...\n")
			cmd = exec.Command("sudo", "umount", "-l", mountPoint)
			output, err = cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to unmount %s: %v, output: %s", mountPoint, err, string(output))
			}
		}
		fmt.Printf("Successfully unmounted %s\n", mountPoint)
	}

	// Remove from fstab if exists
	cmd := exec.Command("sudo", "sed", "-i", fmt.Sprintf("\\|%s|d", mountPoint), "/etc/fstab")
	if err := cmd.Run(); err != nil {
		// Log warning but don't fail - fstab entry might not exist
		fmt.Printf("Warning: failed to remove from fstab: %v\n", err)
	}

	// Also try to remove legacy mount points from fstab
	// Legacy pools might have entries like /dev/md0, UUID=xxx, etc.
	// We need to find and remove any fstab entry that points to our mount point
	cmd = exec.Command("sudo", "grep", "-v", "^#", "/etc/fstab")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		var newLines []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			fields := strings.Fields(line)
			// Check if this fstab entry mounts to our pool's mount point
			if len(fields) >= 2 && fields[1] == mountPoint {
				// Skip this line (remove it from fstab)
				continue
			}
			newLines = append(newLines, line)
		}
		// Write back the cleaned fstab
		if len(newLines) < len(lines) {
			newFstab := strings.Join(newLines, "\n") + "\n"
			cmd = exec.Command("sudo", "tee", "/etc/fstab")
			cmd.Stdin = strings.NewReader(newFstab)
			if output, err := cmd.CombinedOutput(); err != nil {
				fmt.Printf("Warning: failed to update fstab: %v, output: %s\n", err, string(output))
			}
		}
	}

	// Remove pool directory and all contents
	cmd = exec.Command("sudo", "rm", "-rf", mountPoint)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove pool directory %s: %v, output: %s", mountPoint, err, string(output))
	}

	// Restart Samba if it was running
	if sambaRunning {
		fmt.Printf("Restarting Samba...\n")
		if output, err := exec.Command("sudo", "systemctl", "start", "smbd").CombinedOutput(); err != nil {
			return fmt.Errorf("failed to restart Samba: %v, output: %s", err, string(output))
		}
	}

	fmt.Printf("Successfully deleted pool %s (was at %s)\n", name, mountPoint)
	return nil
}

func isMounted(mountPoint string) bool {
	cmd := exec.Command("findmnt", "-n", mountPoint)
	err := cmd.Run()
	return err == nil
}

// unmountWithRetry attempts multiple unmount strategies
func unmountWithRetry(mountPoint string) error {
	// Strategy 1: Normal unmount
	cmd := exec.Command("sudo", "umount", mountPoint)
	output, err := cmd.CombinedOutput()
	if err == nil {
		fmt.Printf("Successfully unmounted %s\n", mountPoint)
		return nil
	}

	fmt.Printf("Normal unmount failed: %s\n", string(output))

	// Strategy 2: Lazy unmount (detach filesystem)
	cmd = exec.Command("sudo", "umount", "-l", mountPoint)
	output, err = cmd.CombinedOutput()
	if err == nil {
		fmt.Printf("Successfully lazy-unmounted %s\n", mountPoint)
		return nil
	}

	fmt.Printf("Lazy unmount failed: %s\n", string(output))

	// Strategy 3: Force unmount with fuser (kill processes using the mount)
	_ = exec.Command("sudo", "fuser", "-km", mountPoint).Run()

	// Try lazy unmount again after killing processes
	cmd = exec.Command("sudo", "umount", "-l", mountPoint)
	output, err = cmd.CombinedOutput()
	if err == nil {
		fmt.Printf("Successfully unmounted %s after killing processes\n", mountPoint)
		return nil
	}

	return fmt.Errorf("failed to unmount %s after all attempts: %v, last output: %s", mountPoint, err, string(output))
}

// removeNFSExportsForPath removes any NFS exports that contain the given path
func removeNFSExportsForPath(path string) error {
	// Read exports file
	exports, err := exec.Command("cat", "/etc/exports").CombinedOutput()
	if err != nil {
		return nil // No exports file, nothing to remove
	}

	lines := strings.Split(string(exports), "\n")
	var newLines []string
	removed := false

	for _, line := range lines {
		// Skip lines that export our path
		if strings.Contains(line, path) && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			removed = true
			continue
		}
		newLines = append(newLines, line)
	}

	if !removed {
		return nil // No changes needed
	}

	// Write back the modified exports
	newExports := strings.Join(newLines, "\n")
	cmd := exec.Command("sudo", "tee", "/etc/exports")
	cmd.Stdin = strings.NewReader(newExports)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to update exports: %v, output: %s", err, string(output))
	}

	// Reload NFS exports
	_ = exec.Command("sudo", "exportfs", "-ra").Run()

	return nil
}

// CleanupLegacyPool removes a pool from the old /var/lib/arcanas/ location
// This is useful for migrating pools to the new /srv/ location
// TODO: Remove this function after migration period (v1.0.0 or later)
// DEPRECATED: This is temporary migration helper code
func CleanupLegacyPool(poolName string) error {
	legacyMountPoint := "/var/lib/arcanas/" + poolName

	// Check if it's mounted
	if isMounted(legacyMountPoint) {
		cmd := exec.Command("sudo", "umount", legacyMountPoint)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to unmount legacy pool %s: %w, output: %s", legacyMountPoint, err, string(output))
		}
	}

	// Remove from fstab if exists
	cmd := exec.Command("sudo", "sed", "-i", fmt.Sprintf("\\|%s|d", legacyMountPoint), "/etc/fstab")
	if err := cmd.Run(); err != nil {
		// Log warning but don't fail
		fmt.Printf("Warning: failed to remove legacy pool from fstab: %v\n", err)
	}

	// Remove the pool directory if it exists
	if _, err := os.Stat(legacyMountPoint); err == nil {
		cmd = exec.Command("sudo", "rm", "-rf", legacyMountPoint)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to remove legacy pool directory %s: %w, output: %s", legacyMountPoint, err, string(output))
		}
	}

	return nil
}

func FormatDisk(req models.DiskFormatRequest) error {
	// Validate filesystem type
	validFSTypes := map[string]bool{
		"ext4": true, "xfs": true, "btrfs": true,
	}
	if !validFSTypes[req.FSType] {
		return fmt.Errorf("unsupported filesystem type: %s", req.FSType)
	}

	// Format the disk
	var cmd *exec.Cmd
	if req.Label != "" {
		cmd = exec.Command("mkfs."+req.FSType, "-L", req.Label, req.Device)
	} else {
		cmd = exec.Command("mkfs."+req.FSType, req.Device)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to format disk: %v", err)
	}

	return nil
}

func GetPathUsage(mountPoint string) (int64, int64) {
	var used int64

	// Try to get used space using du -s -b
	// This works reliably for both regular and FUSE filesystems
	duCmd := exec.Command("du", "-s", "-b", mountPoint)
	duOutput, err := duCmd.Output()
	if err == nil {
		lines := strings.Split(string(duOutput), "\n")
		if len(lines) > 0 {
			fields := strings.Fields(lines[0])
			if len(fields) > 0 {
				used, _ = strconv.ParseInt(fields[0], 10, 64)
			}
		}
	}

	// Try df first for total size (works for regular filesystems)
	cmd := exec.Command("df", "-B", "1", "--output=size,used", mountPoint)
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		if len(lines) >= 2 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 1 {
				size, _ := strconv.ParseInt(fields[0], 10, 64)
				duUsed, _ := strconv.ParseInt(fields[1], 10, 64)
				if size > 0 {
					if duUsed > used {
						used = duUsed
					}
					return size, used
				}
			}
		}
	}

	// For FUSE filesystems where df returns 0 or unreliable values,
	// we can't reliably get total capacity. Return used for both
	// or 0, used if we couldn't get any values
	return used, used
}

// UnmountStoragePool unmounts a storage pool, freeing the underlying device for other uses (e.g., iSCSI)
func UnmountStoragePool(poolName string) error {
	mountPoint := "/srv/" + poolName

	// Check if pool exists
	if _, err := os.Stat(mountPoint); os.IsNotExist(err) {
		return fmt.Errorf("pool %s does not exist", poolName)
	}

	// Check if it's currently mounted
	if !isMounted(mountPoint) {
		return fmt.Errorf("pool %s is not mounted", poolName)
	}

	// Remove any NFS exports that might be using this mount
	// This is important for freeing the device for iSCSI use
	if err := removeNFSExportsForPath(mountPoint); err != nil {
		fmt.Printf("Warning: failed to remove NFS exports: %v\n", err)
	}

	// Stop Samba temporarily to release file handles
	sambaRunning := false
	if cmd := exec.Command("sudo", "systemctl", "is-active", "smbd").Run(); cmd == nil {
		sambaRunning = true
		fmt.Printf("Stopping Samba temporarily for pool unmount...\n")
		_ = exec.Command("sudo", "systemctl", "stop", "smbd").Run()
	}

	// Use robust unmount with retries
	err := unmountWithRetry(mountPoint)

	// Restart Samba if it was running
	if sambaRunning {
		fmt.Printf("Restarting Samba...\n")
		if output, startErr := exec.Command("sudo", "systemctl", "start", "smbd").CombinedOutput(); startErr != nil {
			fmt.Printf("Warning: failed to restart Samba: %v, output: %s\n", startErr, string(output))
		}
	}

	if err != nil {
		return err
	}

	fmt.Printf("Successfully unmounted %s - device is now available for iSCSI/LVM\n", mountPoint)
	return nil
}

// MountStoragePool remounts a previously unmounted storage pool
func MountStoragePool(poolName string) error {
	mountPoint := "/srv/" + poolName

	// Check if pool directory exists
	if _, err := os.Stat(mountPoint); os.IsNotExist(err) {
		return fmt.Errorf("pool %s does not exist", poolName)
	}

	// Check if already mounted
	if isMounted(mountPoint) {
		return fmt.Errorf("pool %s is already mounted", poolName)
	}

	// Check if there's an fstab entry for this pool
	cmd := exec.Command("grep", "-q", mountPoint, "/etc/fstab")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("no fstab entry found for pool %s - cannot remount automatically", poolName)
	}

	// Mount using fstab entry
	cmd = exec.Command("sudo", "mount", mountPoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to mount %s: %v, output: %s", mountPoint, err, string(output))
	}

	// Verify the mount was successful
	if !isMounted(mountPoint) {
		return fmt.Errorf("mount verification failed for %s", mountPoint)
	}

	fmt.Printf("Successfully mounted %s\n", mountPoint)
	return nil
}

// SetPoolExportMode changes how a storage pool is exported
// Modes:
// - "file" - Mounted filesystem for NFS/Samba sharing
// - "iscsi" - Unmounted, exported as iSCSI LUN
// - "available" - Unmounted, available for either use
func SetPoolExportMode(poolName, mode string) error {
	mountPoint := "/srv/" + poolName

	// Validate mode
	validModes := map[string]bool{"file": true, "iscsi": true, "available": true}
	if !validModes[mode] {
		return fmt.Errorf("invalid export mode: %s (must be file, iscsi, or available)", mode)
	}

	// Get current pool state
	pools, err := GetStoragePools()
	if err != nil {
		return fmt.Errorf("failed to get pools: %w", err)
	}

	var currentPool *models.StoragePool
	for i := range pools {
		if pools[i].Name == poolName {
			currentPool = &pools[i]
			break
		}
	}

	if currentPool == nil {
		return fmt.Errorf("pool %s not found", poolName)
	}

	currentlyMounted := isMounted(mountPoint)

	// Handle mode transitions
	switch mode {
	case "file":
		// Need to mount for file sharing
		if !currentlyMounted {
			if err := MountStoragePool(poolName); err != nil {
				return fmt.Errorf("failed to mount pool: %w", err)
			}
		}

	case "iscsi":
		// Need to unmount for iSCSI use
		if currentlyMounted {
			if err := UnmountStoragePool(poolName); err != nil {
				return fmt.Errorf("failed to unmount pool: %w", err)
			}
		}
		// Note: The pool will show up in iSCSI backend options when unmounted

	case "available":
		// Just unmount, don't use for anything
		if currentlyMounted {
			if err := UnmountStoragePool(poolName); err != nil {
				return fmt.Errorf("failed to unmount pool: %w", err)
			}
		}
	}

	fmt.Printf("Pool %s export mode changed to %s\n", poolName, mode)
	return nil
}
