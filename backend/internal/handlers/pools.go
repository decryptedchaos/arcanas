/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"arcanas/internal/models"
	"arcanas/internal/system"
)

func GetStoragePools(w http.ResponseWriter, r *http.Request) {
	pools, err := system.GetStoragePools()
	if err != nil {
		log.Printf("Error getting storage pools: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pools)
}

func CreateStoragePool(w http.ResponseWriter, r *http.Request) {
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
	if req.Type == "" {
		http.Error(w, "Pool type is required", http.StatusBadRequest)
		return
	}
	if len(req.Devices) == 0 {
		http.Error(w, "At least one device is required", http.StatusBadRequest)
		return
	}

	if err := system.CreateStoragePool(req); err != nil {
		log.Printf("Error creating storage pool: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Storage pool created successfully",
		"name":    req.Name,
	})
}

func UpdateStoragePool(w http.ResponseWriter, r *http.Request) {
	// Extract pool name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/storage-pools/")
	poolName := strings.TrimSuffix(path, "/")

	var req models.StoragePoolCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := system.UpdateStoragePool(poolName, req); err != nil {
		log.Printf("Error updating storage pool: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Storage pool updated successfully",
		"name":    poolName,
	})
}

func DeleteStoragePool(w http.ResponseWriter, r *http.Request) {
	// Extract pool name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/storage-pools/")
	poolName := strings.TrimSuffix(path, "/")

	if poolName == "" {
		http.Error(w, "Pool name is required", http.StatusBadRequest)
		return
	}

	if err := system.DeleteStoragePool(poolName); err != nil {
		log.Printf("Error deleting storage pool: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Storage pool deleted successfully",
		"name":    poolName,
	})
}

func FormatDisk(w http.ResponseWriter, r *http.Request) {
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
	if req.FSType == "" {
		http.Error(w, "Filesystem type is required", http.StatusBadRequest)
		return
	}

	if err := system.FormatDisk(req); err != nil {
		log.Printf("Error formatting disk: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Disk formatted successfully",
		"device":  req.Device,
		"fs_type": req.FSType,
	})
}
