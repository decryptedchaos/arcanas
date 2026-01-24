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
	if err := json.NewEncoder(w).Encode(arrays); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func CreateRAIDArray(w http.ResponseWriter, r *http.Request) {
	var req models.RAIDCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request (name is optional, will be auto-generated)
	if req.Level == "" {
		http.Error(w, "RAID level is required", http.StatusBadRequest)
		return
	}
	if len(req.Devices) == 0 {
		http.Error(w, "At least one device is required", http.StatusBadRequest)
		return
	}

	createdName, err := system.CreateRAIDArray(req)
	if err != nil {
		log.Printf("Error creating RAID array: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": "RAID array created successfully",
		"name":    createdName,
	}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
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

	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": "RAID array deleted successfully",
		"name":    arrayName,
	}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
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

	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": "Disk added to RAID array successfully",
		"array":   arrayName,
		"device":  req.Device,
	}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// WipeRAIDSuperblock removes orphaned RAID metadata from a device
func WipeRAIDSuperblock(w http.ResponseWriter, r *http.Request) {
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

	if err := system.WipeRAIDSuperblock(req.Device); err != nil {
		log.Printf("Error wiping RAID superblock: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": "RAID superblock wiped successfully",
		"device":  req.Device,
	}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// ExamineRAIDDevice checks a device for RAID metadata
func ExamineRAIDDevice(w http.ResponseWriter, r *http.Request) {
	// Extract device from URL query parameter
	device := r.URL.Query().Get("device")
	if device == "" {
		http.Error(w, "Device parameter is required", http.StatusBadRequest)
		return
	}

	info, err := system.ExamineRAIDDevice(device)
	if err != nil {
		// Return 404 if no RAID metadata found
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}
