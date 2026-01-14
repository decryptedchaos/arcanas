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
	"regexp"
	"strings"

	"arcanas/internal/models"
	"arcanas/internal/system"
)

func GetStoragePools(w http.ResponseWriter, r *http.Request) {
	pools, err := system.GetStoragePools()
	if err != nil {
		log.Printf("Error getting storage pools: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, pools)
}

func CreateStoragePool(w http.ResponseWriter, r *http.Request) {
	// Validate content type
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var req models.StoragePoolCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Validate pool name format (alphanumeric, underscore, hyphen only)
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(req.Name) {
		http.Error(w, "Pool name must contain only letters, numbers, underscores, and hyphens", http.StatusBadRequest)
		return
	}

	if len(req.Name) > 64 {
		http.Error(w, "Pool name must be 64 characters or less", http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		http.Error(w, "Pool type is required", http.StatusBadRequest)
		return
	}

	// Validate pool type
	validTypes := map[string]bool{"jbod": true, "mergerfs": true, "lvm": true, "bind": true}
	if !validTypes[req.Type] {
		http.Error(w, "Invalid pool type. Must be one of: jbod, mergerfs, lvm, bind", http.StatusBadRequest)
		return
	}

	if len(req.Devices) == 0 {
		http.Error(w, "At least one device is required", http.StatusBadRequest)
		return
	}

	// Validate device paths
	for _, device := range req.Devices {
		if device == "" {
			http.Error(w, "Device paths cannot be empty", http.StatusBadRequest)
			return
		}
		if !strings.HasPrefix(device, "/dev/") {
			http.Error(w, fmt.Sprintf("Invalid device path: %s. Must start with /dev/", device), http.StatusBadRequest)
			return
		}
	}

	if err := system.CreateStoragePool(req); err != nil {
		log.Printf("Error creating storage pool: %v", err)
		http.Error(w, "Failed to create storage pool", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSONResponse(w, map[string]string{
		"message": "Storage pool created successfully",
		"name":    req.Name,
	})
}

func UpdateStoragePool(w http.ResponseWriter, r *http.Request) {
	// Extract pool name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/storage-pools/")
	poolName := strings.TrimSuffix(path, "/")

	// Validate pool name to prevent path traversal
	if strings.Contains(poolName, "..") || strings.Contains(poolName, "/") {
		http.Error(w, "Invalid pool name", http.StatusBadRequest)
		return
	}

	// Validate content type
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var req models.StoragePoolCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request if provided
	if req.Name != "" {
		// Validate pool name format (alphanumeric, underscore, hyphen only)
		if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(req.Name) {
			http.Error(w, "Pool name must contain only letters, numbers, underscores, and hyphens", http.StatusBadRequest)
			return
		}

		if len(req.Name) > 64 {
			http.Error(w, "Pool name must be 64 characters or less", http.StatusBadRequest)
			return
		}
	}

	if req.Type != "" {
		// Validate pool type
		validTypes := map[string]bool{"jbod": true, "mergerfs": true, "lvm": true, "bind": true}
		if !validTypes[req.Type] {
			http.Error(w, "Invalid pool type. Must be one of: jbod, mergerfs, lvm, bind", http.StatusBadRequest)
			return
		}
	}

	if len(req.Devices) > 0 {
		// Validate device paths
		for _, device := range req.Devices {
			if device == "" {
				http.Error(w, "Device paths cannot be empty", http.StatusBadRequest)
				return
			}
			if !strings.HasPrefix(device, "/dev/") {
				http.Error(w, fmt.Sprintf("Invalid device path: %s. Must start with /dev/", device), http.StatusBadRequest)
				return
			}
		}
	}

	if err := system.UpdateStoragePool(poolName, req); err != nil {
		log.Printf("Error updating storage pool: %v", err)
		http.Error(w, "Failed to update storage pool", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, map[string]string{
		"message": "Storage pool updated successfully",
		"name":    poolName,
	})
}

func DeleteStoragePool(w http.ResponseWriter, r *http.Request) {
	// Extract pool name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/storage-pools/")
	poolName := strings.TrimSuffix(path, "/")

	// Validate pool name to prevent path traversal
	if strings.Contains(poolName, "..") || strings.Contains(poolName, "/") {
		http.Error(w, "Invalid pool name", http.StatusBadRequest)
		return
	}

	if poolName == "" {
		http.Error(w, "Pool name is required", http.StatusBadRequest)
		return
	}

	if err := system.DeleteStoragePool(poolName); err != nil {
		log.Printf("Error deleting storage pool: %v", err)
		http.Error(w, "Failed to delete storage pool", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, map[string]string{
		"message": "Storage pool deleted successfully",
		"name":    poolName,
	})
}

func FormatDisk(w http.ResponseWriter, r *http.Request) {
	// Validate content type
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var req models.DiskFormatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Device == "" {
		http.Error(w, "Device is required", http.StatusBadRequest)
		return
	}

	// Validate device path
	if !strings.HasPrefix(req.Device, "/dev/") {
		http.Error(w, "Invalid device path. Must start with /dev/", http.StatusBadRequest)
		return
	}

	if req.FSType == "" {
		http.Error(w, "Filesystem type is required", http.StatusBadRequest)
		return
	}

	// Validate filesystem type
	validFSTypes := map[string]bool{"ext4": true, "xfs": true, "btrfs": true, "ntfs": true, "fat32": true, "exfat": true}
	if !validFSTypes[req.FSType] {
		http.Error(w, "Invalid filesystem type. Must be one of: ext4, xfs, btrfs, ntfs, fat32, exfat", http.StatusBadRequest)
		return
	}

	if err := system.FormatDisk(req); err != nil {
		log.Printf("Error formatting disk: %v", err)
		http.Error(w, "Failed to format disk", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, map[string]string{
		"message": "Disk formatted successfully",
		"device":  req.Device,
		"fs_type": req.FSType,
	})
}

// CleanupLegacyPool handles cleanup of legacy pools from /var/lib/arcanas/
// TODO: Remove this handler after migration period (v1.0.0 or later)
// DEPRECATED: This is temporary migration helper code
func CleanupLegacyPool(w http.ResponseWriter, r *http.Request) {
	// Extract pool name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/storage-pools/cleanup/")
	poolName := strings.TrimSuffix(path, "/")

	// Validate pool name to prevent path traversal
	if strings.Contains(poolName, "..") || strings.Contains(poolName, "/") {
		http.Error(w, "Invalid pool name", http.StatusBadRequest)
		return
	}

	if poolName == "" {
		http.Error(w, "Pool name is required", http.StatusBadRequest)
		return
	}

	if err := system.CleanupLegacyPool(poolName); err != nil {
		log.Printf("Error cleaning up legacy pool: %v", err)
		http.Error(w, "Failed to cleanup legacy pool", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, map[string]string{
		"message": "Legacy pool cleaned up successfully",
		"name":    poolName,
	})
}
