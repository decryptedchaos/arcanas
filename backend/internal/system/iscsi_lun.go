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
	"arcanas/internal/utils"
)

// ISCSIConfig holds the iSCSI server configuration
const (
	DefaultISCSITargetIQN = "iqn.2024-01.com.nas:storage"
	ISCSIBackstoreDir     = "/var/lib/arcanas/iscsi"
)

// EnsureISCSITargetConfigured ensures the default iSCSI target exists with proper configuration
// This must be called before any LUN operations to make sure authentication is disabled
// and auto-ACL generation is disabled for manual ACL management
func EnsureISCSITargetConfigured() error {
	// Check if target exists
	checkOutput, err := utils.SudoCombinedOutput("targetcli", "/iscsi ls")
	if err != nil {
		return fmt.Errorf("failed to check iSCSI targets: %v", err)
	}

	targetExists := strings.Contains(string(checkOutput), DefaultISCSITargetIQN)

	// Create target if it doesn't exist
	if !targetExists {
		_, err := utils.SudoCombinedOutput("targetcli", fmt.Sprintf("/iscsi create %s", DefaultISCSITargetIQN))
		if err != nil {
			return fmt.Errorf("failed to create iSCSI target: %v", err)
		}
	}

	// Disable authentication and demo mode write protect, DISABLE auto-ACL generation
	// With generate_node_acls=0, we must manually create ACLs for each initiator
	// This allows per-client LUN access control
	_, err = utils.SudoCombinedOutput("targetcli", fmt.Sprintf("/iscsi/%s/tpg1 set attribute authentication=0 demo_mode_write_protect=0 generate_node_acls=0", DefaultISCSITargetIQN))
	if err != nil {
		return fmt.Errorf("failed to configure target attributes: %v", err)
	}

	// Configure conservative burst lengths to prevent false completion reporting
	// on slow links. Default MaxBurstLength of 256KB causes transfers to appear
	// complete instantly while data continues transferring at line speed.
	// These settings (32KB first, 64KB max) provide better pacing for slower networks.
	configfsTPGPath := fmt.Sprintf("/sys/kernel/config/target/iscsi/%s/tpgt_1/param", DefaultISCSITargetIQN)
	_, err = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo 32768 > %s/FirstBurstLength", configfsTPGPath))
	if err != nil {
		// Log but don't fail - may not exist on older kernels
		fmt.Printf("Warning: failed to set FirstBurstLength: %v\n", err)
	}
	_, err = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo 65536 > %s/MaxBurstLength", configfsTPGPath))
	if err != nil {
		// Log but don't fail - may not exist on older kernels
		fmt.Printf("Warning: failed to set MaxBurstLength: %v\n", err)
	}

	return nil
}

// LUNManager handles LUN creation and management with different backends
type LUNManager struct{}

// GetAvailableBackends returns information about available LUN backends
func GetAvailableBackends() ([]models.LUNBackendInfo, error) {
	var backends []models.LUNBackendInfo

	// LVM Backend - Flexible, resizable LUNs from volume groups
	vgs, err := getVolumeGroups()
	if err == nil && len(vgs) > 0 {
		backends = append(backends, models.LUNBackendInfo{
			Type:        "lvm",
			Name:        "LVM Volume (Flexible)",
			Description: "Create LUNs of any size from a volume group. Best for sharing storage among clients.",
			Available:   true,
			Resources:   vgs,
		})
	} else {
		backends = append(backends, models.LUNBackendInfo{
			Type:        "lvm",
			Name:        "LVM Volume (Flexible)",
			Description: "Create LUNs of any size from a volume group. Best for sharing storage among clients.",
			Available:   false,
			Resources:   []string{},
		})
	}

	// Block Backend - Raw device mapping (dedicated disk per LUN)
	devices, err := getAvailableBlockDevices()
	if err == nil && len(devices) > 0 {
		backends = append(backends, models.LUNBackendInfo{
			Type:        "block",
			Name:        "Block Device (Dedicated)",
			Description: "Use an entire disk, RAID, or partition for one LUN. Simple but uses the whole device.",
			Available:   true,
			Resources:   devices,
		})
	} else {
		backends = append(backends, models.LUNBackendInfo{
			Type:        "block",
			Name:        "Block Device (Dedicated)",
			Description: "Use an entire disk, RAID, or partition for one LUN. Simple but uses the whole device.",
			Available:   false,
			Resources:   []string{},
		})
	}

	// FileIO Backend - File-based (testing only)
	backends = append(backends, models.LUNBackendInfo{
		Type:        "fileio",
		Name:        "File-Based (Testing)",
		Description: "Create a file on disk as the LUN. For testing only. Slower performance.",
		Available:   true,
		Resources:   []string{ISCSIBackstoreDir},
	})

	return backends, nil
}

// CreateLUN creates a new LUN with the specified backend
func CreateLUN(req models.LUNCreateRequest) (models.ISCSILUN, error) {
	// Ensure the iSCSI target exists and is configured with no authentication
	if err := EnsureISCSITargetConfigured(); err != nil {
		return models.ISCSILUN{}, fmt.Errorf("failed to configure iSCSI target: %w", err)
	}

	lun := models.ISCSILUN{
		Name:        req.Name,
		SizeGB:      req.SizeGB,
		BackendType: req.BackendType,
		Status:      "active",
		Created:     time.Now(),
		AllowedIQNs: req.AllowedIQNs,
	}

	var backendPath string
	var err error

	switch req.BackendType {
	case "lvm":
		if req.VolumeGroup == "" {
			return lun, fmt.Errorf("volume_group is required for LVM backend")
		}
		backendPath, err = createLVMLogicalVolume(req.VolumeGroup, req.Name, req.SizeGB)
		lun.BackendPath = backendPath // Store the LV path (/dev/R0/c1)
		lun.LVPath = backendPath      // Also store in LVPath field for tracking

	case "block":
		if req.DevicePath == "" {
			return lun, fmt.Errorf("device_path is required for block backend")
		}
		backendPath = req.DevicePath
		lun.BackendPath = backendPath

	case "fileio":
		filePath := req.FilePath
		if filePath == "" {
			// Auto-generate file path
			timestamp := time.Now().Unix()
			filePath = fmt.Sprintf("%s/lun_%s_%d.img", ISCSIBackstoreDir, sanitizeName(req.Name), timestamp)
		}
		backendPath, err = createFileIOBackingStore(filePath, req.SizeGB)
		lun.BackendPath = backendPath

	default:
		return lun, fmt.Errorf("unsupported backend type: %s (must be lvm, block, or fileio)", req.BackendType)
	}

	if err != nil {
		return lun, fmt.Errorf("failed to create %s backend: %w", req.BackendType, err)
	}

	// Get next available LUN number
	lunNum, err := getNextAvailableLUN()
	if err != nil {
		// Cleanup created backend on failure
		lvPathToDelete := ""
		if req.BackendType == "lvm" {
			lvPathToDelete = backendPath
		}
		_ = deleteLUNBackend(req.BackendType, backendPath, lvPathToDelete)
		return lun, fmt.Errorf("failed to get LUN number: %w", err)
	}
	lun.LUN = lunNum

	// Create targetcli backstore and LUN
	if err := createTargetcliLUN(lun.LUN, backendPath, req.BackendType); err != nil {
		// Cleanup created backend on failure
		lvPathToDelete := ""
		if req.BackendType == "lvm" {
			lvPathToDelete = backendPath
		}
		_ = deleteLUNBackend(req.BackendType, backendPath, lvPathToDelete)
		return lun, fmt.Errorf("failed to create targetcli LUN: %w", err)
	}

	// Set up ACLs if specific IQNs are specified
	if len(req.AllowedIQNs) > 0 {
		for _, iqn := range req.AllowedIQNs {
			if err := setLUNACL(lun.LUN, iqn); err != nil {
				// Log warning but don't fail
				fmt.Printf("Warning: failed to set ACL for %s on LUN %d: %v\n", iqn, lun.LUN, err)
			}
		}
	}

	return lun, nil
}

// DeleteLUN removes a LUN and cleans up its backend
func DeleteLUN(lunNum int, lun models.ISCSILUN) error {
	// Delete LUN from targetcli
	if err := deleteTargetcliLUN(lunNum); err != nil {
		return fmt.Errorf("failed to delete targetcli LUN: %w", err)
	}

	// Delete the backend storage
	if err := deleteLUNBackend(lun.BackendType, lun.BackendPath, lun.LVPath); err != nil {
		return fmt.Errorf("failed to delete backend: %w", err)
	}

	return nil
}

// ResizeLUN resizes an LVM-backed LUN
func ResizeLUN(lun models.ISCSILUN, newSizeGB float64) error {
	if lun.BackendType != "lvm" {
		return fmt.Errorf("only LVM-backed LUNs can be resized (current: %s)", lun.BackendType)
	}

	// Parse VG and LV names from path (e.g., /dev/vg-raid0/client1)
	parts := strings.Split(lun.BackendPath, "/")
	if len(parts) < 3 {
		return fmt.Errorf("invalid LVM path: %s", lun.BackendPath)
	}
	vgName := parts[2]
	lvName := parts[3]

	// Resize the logical volume
	sizeMB := int(newSizeGB * 1024)
	output, err := utils.SudoCombinedOutput("lvresize", "-L", fmt.Sprintf("%dM", sizeMB), fmt.Sprintf("%s/%s", vgName, lvName))
	if err != nil {
		return fmt.Errorf("failed to resize LV: %v, output: %s", err, string(output))
	}

	// Resize the filesystem (assume ext4)
	output, err = utils.SudoCombinedOutput("resize2fs", lun.BackendPath)
	if err != nil {
		return fmt.Errorf("failed to resize filesystem: %v, output: %s", err, string(output))
	}

	return nil
}

// getVolumeGroups returns available LVM volume groups
func getVolumeGroups() ([]string, error) {
	output, err := utils.SudoCombinedOutput("vgs", "--noheadings", "-o", "vg_name")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var vgs []string
	for _, line := range lines {
		vg := strings.TrimSpace(line)
		if vg != "" {
			vgs = append(vgs, vg)
		}
	}
	return vgs, nil
}

// getAvailableBlockDevices returns devices that can be used as block backends
func getAvailableBlockDevices() ([]string, error) {
	// Get block devices that are not mounted and not in use
	output, err := utils.SudoCombinedOutput("sh", "-c", "lsblk -d -n -o NAME,TYPE,MOUNTPOINT | awk '$2 == \"disk\" || $2 == \"raid1\" || $2 == \"raid5\" {print \"/dev/\"$1}'")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var devices []string
	for _, line := range lines {
		dev := strings.TrimSpace(line)
		if dev != "" && strings.HasPrefix(dev, "/dev/") {
			devices = append(devices, dev)
		}
	}
	return devices, nil
}

// createLVMLogicalVolume creates an LVM logical volume for a LUN
func createLVMLogicalVolume(vgName, lvName string, sizeGB float64) (string, error) {
	// Check if lvm tools are available
	if _, err := exec.LookPath("lvcreate"); err != nil {
		return "", fmt.Errorf("LVM tools not installed. Install with: sudo apt install lvm2")
	}

	lvName = sanitizeName(lvName)
	lvPath := fmt.Sprintf("/dev/%s/%s", vgName, lvName)

	// Check if LV already exists
	output, err := utils.SudoCombinedOutput("lvs", "--noheadings", "-o", "lv_name", vgName)
	if err == nil {
		existingLVs := strings.Fields(strings.TrimSpace(string(output)))
		for _, existingLV := range existingLVs {
			if existingLV == lvName {
				// LV already exists - check if we should reuse or error
				return "", fmt.Errorf("logical volume '%s' already exists in volume group '%s'. Please delete the existing LUN first or use a different name", lvName, vgName)
			}
		}
	}

	sizeMB := int(sizeGB * 1024)

	// Create logical volume with -W flag to auto-wipe old signatures and -y to auto-confirm
	output, err = utils.SudoCombinedOutput("lvcreate", "-y", "-W", "y", "-L", fmt.Sprintf("%dM", sizeMB), "-n", lvName, vgName)
	if err != nil {
		return "", fmt.Errorf("failed to create LV: %v, output: %s", err, string(output))
	}

	// Format with ext4
	output, err = utils.SudoCombinedOutput("mkfs.ext4", "-F", lvPath)
	if err != nil {
		// Cleanup LV on failure
		_ = utils.SudoRunCommand("lvremove", "-f", lvPath)
		return "", fmt.Errorf("failed to format LV: %v, output: %s", err, string(output))
	}

	return lvPath, nil
}

// createFileIOBackingStore creates a file-based backing store for testing
func createFileIOBackingStore(filePath string, sizeGB float64) (string, error) {
	// Ensure directory exists
	if err := utils.SudoRunCommand("mkdir", "-p", ISCSIBackstoreDir); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	sizeMB := int(sizeGB * 1024)
	output, err := utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("dd if=/dev/zero of=%s bs=1M count=%d status='progress'", filePath, sizeMB))
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v, output: %s", err, string(output))
	}

	return filePath, nil
}

// deleteLUNBackend removes the backend storage for a LUN
func deleteLUNBackend(backendType, backendPath, lvPath string) error {
	// First, delete the targetcli backstore object (for block/fileio types)
	if backendType == "block" || backendType == "fileio" {
		// backendPath is the targetcli backstore path (e.g., /backstores/block/bs_lun0)
		// Parse the path to get parent directory and backstore name
		parts := strings.Split(backendPath, "/")
		if len(parts) >= 4 {
			parentPath := strings.Join(parts[:3], "/")
			backstoreName := parts[3]
			output, err := utils.SudoCombinedOutput("targetcli", fmt.Sprintf("%s delete %s", parentPath, backstoreName))
			if err != nil {
				return fmt.Errorf("failed to delete backstore: %v, output: %s", err, string(output))
			}
		}
	}

	// For fileio backends, also delete the file
	if backendType == "fileio" {
		if err := utils.SudoRunCommand("rm", "-f", backendPath); err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}

	// If lvPath is set, delete the underlying LV (for LVM-backed LUNs)
	// This handles the case where backendType is "block" but the device is an LVM LV
	if lvPath != "" {
		if err := utils.SudoRunCommand("lvremove", "-f", lvPath); err != nil {
			return fmt.Errorf("failed to delete LV: %w", err)
		}
	}

	return nil
}

// getNextAvailableLUN finds the next available LUN number
func getNextAvailableLUN() (int, error) {
	// Get existing LUNs from targetcli
	output, err := utils.SudoCombinedOutput("sh", "-c", "echo '/iscsi/"+DefaultISCSITargetIQN+"/tpg1/luns ls' | sudo targetcli")
	if err != nil {
		// Target might not exist yet, return LUN 0
		return 0, nil
	}

	// Find the highest LUN number in use
	maxLUN := -1
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		// Match lines starting with "o- lun" followed by a digit (e.g., "o- lun0")
		// This avoids matching the header line "o- luns"
		if strings.HasPrefix(line, "  o- lun") || strings.HasPrefix(line, "o- lun") {
			// Extract LUN number - format: "o- lun0" or "  o- lun0"
			// Use regex to extract the number after "lun"
			re := regexp.MustCompile(`lun(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 2 {
				if lunNum, err := strconv.Atoi(matches[1]); err == nil && lunNum > maxLUN {
					maxLUN = lunNum
				}
			}
		}
	}

	return maxLUN + 1, nil
}

// setBlockBackstoreCacheMode configures write-through caching for a block backstore
// This prevents false transfer completion reporting where the OS reports
// completion before data is actually written to disk (write-back cache behavior)
func setBlockBackstoreCacheMode(backstoreName string) error {
	// The iblock attributes are exposed via configfs
	// Path format: /sys/kernel/config/target/core/iblock_<device_num>/<backstore>/attrib/
	// We need to find the iblock device directory for this backstore

	// List all iblock devices and find the one matching our backstore name
	output, err := utils.SudoCombinedOutput("sh", "-c", "ls /sys/kernel/config/target/core/ | grep '^iblock'")
	if err != nil {
		return fmt.Errorf("failed to list iblock devices: %v", err)
	}

	iblockDevices := strings.Fields(strings.TrimSpace(string(output)))
	var configfsPath string

	// Find the configfs path for our backstore
	for _, iblockDev := range iblockDevices {
		checkPath := fmt.Sprintf("/sys/kernel/config/target/core/%s/%s/attrib", iblockDev, backstoreName)
		// Check if this directory exists (the backstore is registered under this iblock device)
		checkOutput, checkErr := utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("[ -d %s ] && echo 'exists'", checkPath))
		if checkErr == nil && strings.TrimSpace(string(checkOutput)) == "exists" {
			configfsPath = checkPath
			break
		}
	}

	if configfsPath == "" {
		return fmt.Errorf("could not find configfs path for backstore %s", backstoreName)
	}

	// Set emulate_tpu to 0 (disable thin provisioning unmap emulation)
	// This can improve performance and prevent issues with TRIM commands
	emulateTpuPath := fmt.Sprintf("%s/emulate_tpu", configfsPath)
	_, err = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo 0 > %s", emulateTpuPath))
	if err != nil {
		return fmt.Errorf("failed to set emulate_tpu: %v", err)
	}

	// Note: LIO/iblock doesn't have a direct write_through attribute like fileio does.
	// The caching behavior is primarily controlled by the underlying block device's
	// queue settings and filesystem mount options. The emulate_tpu setting above
	// helps ensure more predictable I/O behavior.
	//
	// For true write-through behavior, you would need to set the block device's
	// queue cache mode to write-through using:
	//   echo write-through > /sys/block/<device>/queue/write_cache
	// This can be done on the backing device (e.g., /dev/dm-0 for LVM)

	return nil
}

// createTargetcliLUN creates the backstore and LUN in targetcli
func createTargetcliLUN(lunNum int, backendPath, backendType string) error {
	var backstorePath string
	var backstoreName string
	var output []byte
	var err error

	// Create backstore based on type
	if backendType == "fileio" {
		// FileIO backstore - use write-through mode to prevent false completion reporting
		backstoreName = fmt.Sprintf("bs_lun%d", lunNum)
		output, err = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/backstores/fileio create %s %s' | sudo targetcli", backstoreName, backendPath))
		if err != nil {
			return fmt.Errorf("failed to create fileio backstore: %v, output: %s", err, string(output))
		}
		backstorePath = "/backstores/fileio/" + backstoreName
	} else {
		// Block backstore (for both LVM and raw devices)
		backstoreName = fmt.Sprintf("bs_lun%d", lunNum)
		output, err = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/backstores/block create %s %s' | sudo targetcli", backstoreName, backendPath))
		if err != nil {
			return fmt.Errorf("failed to create block backstore: %v, output: %s", err, string(output))
		}
		backstorePath = "/backstores/block/" + backstoreName

		// Set write-through caching to prevent false transfer completion reporting
		// Without this, write-back cache causes transfers to appear complete instantly
		// while data continues transferring in background at line speed
		if err := setBlockBackstoreCacheMode(backstoreName); err != nil {
			// Log warning but don't fail - cache setting is best-effort
			fmt.Printf("Warning: failed to set cache mode for %s: %v\n", backstoreName, err)
		}
	}

	// Create LUN
	output, err = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1/luns create %s' | sudo targetcli", DefaultISCSITargetIQN, backstorePath))
	if err != nil {
		// Cleanup backstore on failure
		_ = utils.SudoRunCommand("sh", "-c", fmt.Sprintf("echo '%s delete' | sudo targetcli", backstorePath))
		return fmt.Errorf("failed to create LUN: %v, output: %s", err, string(output))
	}

	return nil
}

// deleteTargetcliLUN removes a LUN from targetcli
func deleteTargetcliLUN(lunNum int) error {
	// Use the correct targetcli command format: targetcli '/path/to/luns delete lunN'
	output, err := utils.SudoCombinedOutput("targetcli", fmt.Sprintf("/iscsi/%s/tpg1/luns delete lun%d", DefaultISCSITargetIQN, lunNum))
	if err != nil {
		return fmt.Errorf("failed to delete LUN: %v, output: %s", err, string(output))
	}
	return nil
}

// setLUNACL configures ACL for a specific LUN and initiator
func setLUNACL(lunNum int, initiatorIQN string) error {
	// Create ACL entry if it doesn't exist
	output, err := utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1/acls create %s' | sudo targetcli", DefaultISCSITargetIQN, initiatorIQN))
	if err != nil && !strings.Contains(string(output), "already exists") {
		return fmt.Errorf("failed to create ACL: %v, output: %s", err, string(output))
	}

	// Map the LUN for this initiator
	output, err = utils.SudoCombinedOutput("sh", "-c", fmt.Sprintf("echo '/iscsi/%s/tpg1/acls/%s/luns create /backstores/block/bs_lun%d' | sudo targetcli", DefaultISCSITargetIQN, initiatorIQN, lunNum))
	if err != nil {
		return fmt.Errorf("failed to map LUN: %v, output: %s", err, string(output))
	}

	return nil
}

// sanitizeName converts a name to a valid LVM/LUN name
func sanitizeName(name string) string {
	// Replace invalid characters with underscores
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, ":", "_")

	// Limit length
	if len(name) > 32 {
		name = name[:32]
	}

	// Remove any remaining non-alphanumeric characters (except underscore)
	result := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			result += string(r)
		}
	}

	return result
}

// ============================================================================
// ACL Management Functions
// ============================================================================

// GetISCSIACLs returns all ACL entries for the iSCSI target
func GetISCSIACLs() ([]models.ISCSIACL, error) {
	output, err := utils.SudoCombinedOutput("targetcli", fmt.Sprintf("/iscsi/%s/tpg1/acls ls", DefaultISCSITargetIQN))
	if err != nil {
		return nil, fmt.Errorf("failed to list ACLs: %v", err)
	}

	var acls []models.ISCSIACL
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Match ACL directories (e.g., "o- iqn.1993-08.org.debian:01:abc123...")
		// Skip the header line "o- acls ..."
		if strings.HasPrefix(line, "o- ") && strings.Contains(line, "iqn.") && !strings.Contains(line, "acls") {
			// Extract IQN from the line
			// Format: "o- iqn.2016-04.com.open-iscsi:c8e34e60ec9 ..."
			// Split by whitespace: parts[0] = "o-", parts[1] = "iqn.2016-04.com.open-iscsi:c8e34e60ec9"
			parts := strings.Fields(line)
			if len(parts) > 1 {
				iqn := parts[1] // The IQN is the second field after "o-"

				// Get mapped LUNs for this ACL
				mappedLUNs, err := getMappedLUNsForACL(iqn)
				if err != nil {
					mappedLUNs = []int{}
				}

				acls = append(acls, models.ISCSIACL{
					InitiatorIQN: iqn,
					Name:         iqn, // Default to IQN if no name set
					MappedLUNs:   mappedLUNs,
					Created:      time.Now(),
				})
			}
		}
	}

	return acls, nil
}

// getMappedLUNsForACL returns the LUN numbers mapped to a specific ACL
func getMappedLUNsForACL(iqn string) ([]int, error) {
	// Use targetcli ls command to get mapped LUNs
	// This shows the actual mapped_lun entries, not configfs directories
	output, err := utils.SudoCombinedOutput("targetcli", fmt.Sprintf("/iscsi/%s/tpg1/acls/%s ls", DefaultISCSITargetIQN, iqn))
	if err != nil {
		return nil, fmt.Errorf("failed to list ACL: %v", err)
	}

	var luns []int
	lines := strings.Split(string(output), "\n")

	// Parse output for mapped_lun entries
	// Format: "  o- mapped_lun0 ................................................................................. [lun0 block/bs_lun0 (rw)]"
	re := regexp.MustCompile(`mapped_lun(\d+)`)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "mapped_lun") {
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 2 {
				lunNum, _ := strconv.Atoi(matches[1])
				luns = append(luns, lunNum)
			}
		}
	}

	return luns, nil
}

// CreateISCSIACL creates a new ACL entry for an initiator IQN
func CreateISCSIACL(iqn, name string) error {
	// Validate IQN format
	if !strings.HasPrefix(iqn, "iqn.") && !strings.HasPrefix(iqn, "eui.") {
		return fmt.Errorf("invalid IQN format: must start with 'iqn.' or 'eui.'")
	}

	// Check if ACL already exists
	acls, err := GetISCSIACLs()
	if err == nil {
		for _, acl := range acls {
			if acl.InitiatorIQN == iqn {
				return fmt.Errorf("ACL for IQN %s already exists", iqn)
			}
		}
	}

	// Create the ACL
	output, err := utils.SudoCombinedOutput("targetcli", fmt.Sprintf("/iscsi/%s/tpg1/acls create %s", DefaultISCSITargetIQN, iqn))
	if err != nil {
		return fmt.Errorf("failed to create ACL: %v, output: %s", err, string(output))
	}

	// targetcli automatically maps all LUNs to new ACLs
	// Remove all automatically mapped LUNs so the user has explicit control
	// Use targetcli's delete command which properly removes the entire mapping
	for attempt := 0; attempt < 3; attempt++ {
		// Get current mapped LUNs
		mapped, err := getMappedLUNsForACL(iqn)
		if err != nil || len(mapped) == 0 {
			break
		}

		// Delete each mapped LUN using targetcli
		// This properly removes the entire mapped_lun entry, not just the symlink
		removedAny := false
		for _, lunNum := range mapped {
			output, err := utils.SudoCombinedOutput("targetcli", fmt.Sprintf("/iscsi/%s/tpg1/acls/%s delete %d", DefaultISCSITargetIQN, iqn, lunNum))
			if err == nil {
				removedAny = true
			}
			// Output might contain "Deleted Mapped LUN X" message, ignore
			_ = output
		}

		if !removedAny {
			break
		}

		// Small delay before checking if all were removed
		_ = utils.SudoRunCommand("sleep", "0.1")

		// Verify all mappings are gone
		mapped, _ = getMappedLUNsForACL(iqn)
		if len(mapped) == 0 {
			break
		}
	}

	// Set a human-readable name as an attribute if provided
	if name != "" {
		_, err = utils.SudoCombinedOutput("targetcli", fmt.Sprintf("/iscsi/%s/tpg1/acls/%s set attribute name=%s", DefaultISCSITargetIQN, iqn, name))
		if err != nil {
			// Don't fail on attribute setting, just log it
			fmt.Printf("Warning: failed to set ACL name attribute: %v\n", err)
		}
	}

	return nil
}

// DeleteISCSIACL removes an ACL entry
func DeleteISCSIACL(iqn string) error {
	output, err := utils.SudoCombinedOutput("targetcli", fmt.Sprintf("/iscsi/%s/tpg1/acls delete %s", DefaultISCSITargetIQN, iqn))
	if err != nil {
		return fmt.Errorf("failed to delete ACL: %v, output: %s", err, string(output))
	}
	return nil
}

// MapLUNToACL maps a LUN to an ACL (gives the initiator access to that LUN)
func MapLUNToACL(iqn string, sourceLUN int, targetLUN int) error {
	// Validate ACL exists
	acls, err := GetISCSIACLs()
	if err != nil {
		return fmt.Errorf("failed to list ACLs: %v", err)
	}

	found := false
	for _, acl := range acls {
		if acl.InitiatorIQN == iqn {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("ACL for IQN %s does not exist", iqn)
	}

	// If targetLUN is 0 or negative, default to sourceLUN
	if targetLUN < 0 {
		targetLUN = sourceLUN
	}

	// Check if this LUN is already mapped
	mapped, _ := getMappedLUNsForACL(iqn)
	for _, lunNum := range mapped {
		if lunNum == targetLUN {
			// Already mapped, verify it's mapped to the correct source LUN
			// Use targetcli ls to check the mapping
			output, err := utils.SudoCombinedOutput("targetcli", fmt.Sprintf("/iscsi/%s/tpg1/acls/%s ls", DefaultISCSITargetIQN, iqn))
			if err == nil {
				// Parse output to verify the mapping
				expectedLine := fmt.Sprintf("mapped_lun%d", targetLUN)
				for _, line := range strings.Split(string(output), "\n") {
					line = strings.TrimSpace(line)
					if strings.Contains(line, expectedLine) && strings.Contains(line, fmt.Sprintf("lun%d", sourceLUN)) {
						// Already correctly mapped
						return nil
					}
				}
			}
			// Mapping exists but points to wrong LUN, delete it first
			_ = UnmapLUNFromACL(iqn, targetLUN)
		}
	}

	// Map the LUN using configfs directly
	// Create the lun_<targetLUN> directory and symlink to the actual LUN
	// This is more reliable than targetcli's create command across different versions
	configfsLUNDir := fmt.Sprintf("/sys/kernel/config/target/iscsi/%s/tpgt_1/acls/%s/lun_%d", DefaultISCSITargetIQN, iqn, targetLUN)
	sourceLUNPath := fmt.Sprintf("/sys/kernel/config/target/iscsi/%s/tpgt_1/lun/lun_%d", DefaultISCSITargetIQN, sourceLUN)

	// Create the LUN directory
	if err := utils.SudoRunCommand("mkdir", "-p", configfsLUNDir); err != nil {
		return fmt.Errorf("failed to create LUN directory: %v", err)
	}

	// Create symlink pointing to the source LUN
	symlinkPath := fmt.Sprintf("%s/link", configfsLUNDir)
	output, err := utils.SudoCombinedOutput("ln", "-s", sourceLUNPath, symlinkPath)
	if err != nil {
		// Clean up directory if symlink creation fails
		_ = utils.SudoRunCommand("rmdir", configfsLUNDir)
		return fmt.Errorf("failed to create LUN symlink: %v, output: %s", err, string(output))
	}

	return nil
}

// UnmapLUNFromACL removes a LUN mapping from an ACL
func UnmapLUNFromACL(iqn string, targetLUN int) error {
	// Use targetcli's delete command to properly remove the mapped LUN
	// This completely removes the mapped_lun entry, not just the symlink
	output, err := utils.SudoCombinedOutput("targetcli", fmt.Sprintf("/iscsi/%s/tpg1/acls/%s delete %d", DefaultISCSITargetIQN, iqn, targetLUN))
	if err != nil {
		// Check if it's just "not found" (already unmapped) - that's OK
		if strings.Contains(string(output), "not found") || strings.Contains(string(output), "cannot find") {
			return nil
		}
		return fmt.Errorf("failed to unmap LUN %d: %v, output: %s", targetLUN, err, string(output))
	}
	return nil
}

// GetACLForLUN returns all ACLs that have a specific LUN mapped
func GetACLsForLUN(lunNum int) ([]models.ISCSIACL, error) {
	acls, err := GetISCSIACLs()
	if err != nil {
		return nil, err
	}

	var result []models.ISCSIACL
	for _, acl := range acls {
		for _, mappedLUN := range acl.MappedLUNs {
			if mappedLUN == lunNum {
				result = append(result, acl)
				break
			}
		}
	}

	return result, nil
}
