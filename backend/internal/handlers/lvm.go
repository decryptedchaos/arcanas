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
	"strings"

	"arcanas/internal/models"
	"arcanas/internal/system"
)

// GetVolumeGroups returns all LVM volume groups
func GetVolumeGroups(w http.ResponseWriter, r *http.Request) {
	vgs, err := system.GetVolumeGroups()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get volume groups: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vgs)
}

// CreateVolumeGroup creates a new LVM volume group
func CreateVolumeGroup(w http.ResponseWriter, r *http.Request) {
	var req models.VolumeGroupCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate
	if req.Name == "" {
		http.Error(w, "Volume group name is required", http.StatusBadRequest)
		return
	}
	if len(req.Devices) == 0 {
		http.Error(w, "At least one device is required", http.StatusBadRequest)
		return
	}

	// Validate name format
	if strings.Contains(req.Name, " ") || strings.Contains(req.Name, "/") {
		http.Error(w, "Volume group name cannot contain spaces or slashes", http.StatusBadRequest)
		return
	}

	vg, err := system.CreateVolumeGroup(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create volume group: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(vg)
}

// DeleteVolumeGroup removes a volume group
func DeleteVolumeGroup(w http.ResponseWriter, r *http.Request) {
	// Extract VG name from URL path
	// Format: /api/volume-groups/{name}
	path := strings.TrimPrefix(r.URL.Path, "/api/volume-groups/")
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		http.Error(w, "Volume group name is required", http.StatusBadRequest)
		return
	}

	vgName := path

	// Validate name to prevent path traversal
	if strings.Contains(vgName, "..") || strings.Contains(vgName, "/") {
		http.Error(w, "Invalid volume group name", http.StatusBadRequest)
		return
	}

	if err := system.DeleteVolumeGroup(vgName); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete volume group: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
		"name":   vgName,
	})
}

// GetVGDevices returns devices available for creating volume groups
func GetVGDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := system.GetAvailableDevicesForVG()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get available devices: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

// GetLogicalVolumes returns all LVM logical volumes
func GetLogicalVolumes(w http.ResponseWriter, r *http.Request) {
	lvs, err := system.GetLogicalVolumes()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get logical volumes: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lvs)
}

// CreateLogicalVolume creates a new LVM logical volume
func CreateLogicalVolume(w http.ResponseWriter, r *http.Request) {
	var req models.LVCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate
	if req.Name == "" {
		http.Error(w, "Logical volume name is required", http.StatusBadRequest)
		return
	}
	if req.VGName == "" {
		http.Error(w, "Volume group name is required", http.StatusBadRequest)
		return
	}
	if req.SizeGB <= 0 {
		http.Error(w, "Size must be greater than 0", http.StatusBadRequest)
		return
	}

	// Validate name format
	if strings.Contains(req.Name, " ") || strings.Contains(req.Name, "/") {
		http.Error(w, "Logical volume name cannot contain spaces or slashes", http.StatusBadRequest)
		return
	}

	lv, err := system.CreateLogicalVolume(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create logical volume: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(lv)
}

// DeleteLogicalVolume removes a logical volume
func DeleteLogicalVolume(w http.ResponseWriter, r *http.Request) {
	// Extract LV path from URL path
	// Format: /api/logical-volumes/{path}
	path := strings.TrimPrefix(r.URL.Path, "/api/logical-volumes/")
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		http.Error(w, "Logical volume path is required", http.StatusBadRequest)
		return
	}

	lvPath := path

	// Validate path to prevent path traversal
	if strings.Contains(lvPath, "..") {
		http.Error(w, "Invalid logical volume path", http.StatusBadRequest)
		return
	}

	// Accept both "/dev/vg/lv" and "vg/lv" formats
	if !strings.HasPrefix(lvPath, "/dev/") {
		lvPath = "/dev/" + lvPath
	}

	if err := system.DeleteLogicalVolumeByName(lvPath); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete logical volume: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
		"path":   lvPath,
	})
}
