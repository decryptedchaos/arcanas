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

func GetRAIDArrays(w http.ResponseWriter, r *http.Request) {
	arrays, err := system.GetRAIDArrays()
	if err != nil {
		log.Printf("Error getting RAID arrays: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(arrays)
}

func CreateRAIDArray(w http.ResponseWriter, r *http.Request) {
	var req models.RAIDCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if req.Level == "" {
		http.Error(w, "RAID level is required", http.StatusBadRequest)
		return
	}
	if len(req.Devices) == 0 {
		http.Error(w, "At least one device is required", http.StatusBadRequest)
		return
	}

	if err := system.CreateRAIDArray(req); err != nil {
		log.Printf("Error creating RAID array: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "RAID array created successfully",
		"name":    req.Name,
	})
}

func DeleteRAIDArray(w http.ResponseWriter, r *http.Request) {
	// Extract array name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/raid-arrays/")
	arrayName := strings.TrimSuffix(path, "/")

	if arrayName == "" {
		http.Error(w, "Array name is required", http.StatusBadRequest)
		return
	}

	if err := system.DeleteRAIDArray(arrayName); err != nil {
		log.Printf("Error deleting RAID array: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "RAID array deleted successfully",
		"name":    arrayName,
	})
}

func AddDiskToRAID(w http.ResponseWriter, r *http.Request) {
	// Extract array name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/raid-arrays/")
	arrayName := strings.TrimSuffix(path, "/")

	var req struct {
		Device string `json:"device"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Device == "" {
		http.Error(w, "Device is required", http.StatusBadRequest)
		return
	}

	if err := system.AddDiskToRAID(arrayName, req.Device); err != nil {
		log.Printf("Error adding disk to RAID: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Disk added to RAID array successfully",
		"array":   arrayName,
		"device":  req.Device,
	})
}
