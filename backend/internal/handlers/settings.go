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

	"arcanas/internal/models"
	"arcanas/internal/system"
)

// GetSystemSettings returns the current system settings
func GetSystemSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hostname, err := system.ReadHostname()
	if err != nil {
		log.Printf("Error reading hostname: %v", err)
	}

	timezone, err := system.ReadTimezone()
	if err != nil {
		log.Printf("Error reading timezone: %v", err)
	}

	settings := models.SystemSettings{
		Hostname: hostname,
		Timezone: timezone,
		Locale:   "en_US.UTF-8", // Default for now
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(settings); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateSystemSettings updates system settings
func UpdateSystemSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var settings models.SystemSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update hostname if changed
	if settings.Hostname != "" {
		if err := system.WriteHostname(settings.Hostname); err != nil {
			log.Printf("Error setting hostname: %v", err)
			http.Error(w, "Failed to set hostname: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Update timezone if changed
	if settings.Timezone != "" {
		if err := system.SetTimezone(settings.Timezone); err != nil {
			log.Printf("Error setting timezone: %v", err)
			http.Error(w, "Failed to set timezone: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Return updated settings
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// GetNetworkConfig returns network configuration for all interfaces
func GetNetworkConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	interfaces, err := system.ReadNetworkConfig()
	if err != nil {
		log.Printf("Error reading network config: %v", err)
		http.Error(w, "Failed to read network config", http.StatusInternalServerError)
		return
	}

	// Convert to NetworkInterface format
	var result []models.NetworkInterface
	for _, info := range interfaces {
		iface := models.NetworkInterface{
			Name:   info.Name,
			Type:   info.Type,
			MAC:    info.MAC,
			IPv4: models.NetConfig{
				Address: info.IPv4.Address,
				Netmask: info.IPv4.Netmask,
				Gateway: info.IPv4.Gateway,
			},
			IPv6: models.NetConfig{
				Address: info.IPv6.Address,
			},
			DHCP:  false, // Would need to check dhclient config
			Up:    info.Up,
		}

		// Get DNS
		dns, err := system.GetDNS()
		if err == nil {
			iface.DNS = dns
		}

		result = append(result, iface)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateNetworkConfig updates network configuration for an interface
func UpdateNetworkConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Interface string `json:"interface"`
		Config     models.NetworkInterface `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Write network configuration using netplan
	// Note: This is a simplified implementation
	if req.Config.IPv4.Method == "dhcp" {
		// Configure DHCP
		if err := system.WriteNetworkConfig(req.Interface, system.NetworkConfig{DHCP: true}, req.Config.IPv4, req.Config.IPv6, req.Config.DNS); err != nil {
			log.Printf("Error configuring DHCP: %v", err)
			http.Error(w, "Failed to configure DHCP: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else if req.Config.IPv4.Method == "static" {
		// Configure static IP
		if err := system.WriteNetworkConfig(req.Interface, system.NetworkConfig{DHCP: false}, req.Config.IPv4, req.Config.IPv6, req.Config.DNS); err != nil {
			log.Printf("Error configuring static IP: %v", err)
			http.Error(w, "Failed to configure static IP: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "configured"})
}

// GetTimezone returns timezone information
func GetTimezone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	current, err := system.ReadTimezone()
	if err != nil {
		log.Printf("Error reading timezone: %v", err)
		http.Error(w, "Failed to read timezone", http.StatusInternalServerError)
		return
	}

	info := models.TimezoneInfo{
		Current:   current,
		Available: system.GetCommonTimezones(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateTimezone updates the system timezone
func UpdateTimezone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Timezone string `json:"timezone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Timezone == "" {
		http.Error(w, "Timezone is required", http.StatusBadRequest)
		return
	}

	if err := system.SetTimezone(req.Timezone); err != nil {
		log.Printf("Error setting timezone: %v", err)
		http.Error(w, "Failed to set timezone: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"timezone": req.Timezone})
}
