/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package handlers

import (
	"net/http"
	"time"

	"arcanas/internal/models"
	"arcanas/internal/system"
)

func GetSystemStats(w http.ResponseWriter, r *http.Request) {
	// Get real CPU stats
	cpuStats, err := system.GetCPUStats()
	if err != nil {
		handleError(w, err, "Failed to get CPU stats", http.StatusInternalServerError)
		return
	}

	// Get real memory stats
	memoryStats, err := system.GetMemoryStats()
	if err != nil {
		handleError(w, err, "Failed to get memory stats", http.StatusInternalServerError)
		return
	}

	// Get real network stats
	networkStats, err := system.GetNetworkStats()
	if err != nil {
		handleError(w, err, "Failed to get network stats", http.StatusInternalServerError)
		return
	}

	// Get real storage stats
	storageStats, err := system.GetStorageStats()
	if err != nil {
		handleError(w, err, "Failed to get storage stats", http.StatusInternalServerError)
		return
	}

	// Get real system info
	systemInfo, err := system.GetSystemInfo()
	if err != nil {
		handleError(w, err, "Failed to get system info", http.StatusInternalServerError)
		return
	}

	stats := map[string]interface{}{
		"cpu":     cpuStats,
		"memory":  memoryStats,
		"network": networkStats,
		"storage": storageStats,
		"system":  systemInfo,
	}

	writeJSONResponse(w, stats)
}

func GetDiskIORates(w http.ResponseWriter, r *http.Request) {
	// Get physical disk I/O rates from system (excludes md devices)
	ioRates, err := system.GetDiskIORates()
	if err != nil {
		handleError(w, err, "Failed to get disk I/O rates", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, ioRates)
}

func GetArrayIORates(w http.ResponseWriter, r *http.Request) {
	// Get RAID array I/O rates from system (actual data throughput)
	ioRates, err := system.GetArrayIORates()
	if err != nil {
		handleError(w, err, "Failed to get array I/O rates", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, ioRates)
}

func GetNetworkIORates(w http.ResponseWriter, r *http.Request) {
	// Get real network I/O rates from system
	ioRates, err := system.GetNetworkIORates()
	if err != nil {
		handleError(w, err, "Failed to get network I/O rates", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, ioRates)
}

func GetCPUStats(w http.ResponseWriter, r *http.Request) {
	cpu, err := system.GetCPUStats()
	if err != nil {
		handleError(w, err, "Failed to get CPU stats", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, cpu)
}

func GetMemoryStats(w http.ResponseWriter, r *http.Request) {
	memory, err := system.GetMemoryStats()
	if err != nil {
		handleError(w, err, "Failed to get memory stats", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, memory)
}

func GetNetworkStats(w http.ResponseWriter, r *http.Request) {
	network, err := system.GetNetworkStats()
	if err != nil {
		handleError(w, err, "Failed to get network stats", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, network)
}

func GetStorageHealth(w http.ResponseWriter, r *http.Request) {
	storage := models.StorageStats{
		Disks: []models.DiskHealth{
			{
				Device:      "/dev/sda",
				Model:       "Samsung SSD 870 EVO 2TB",
				Size:        2000,
				Used:        1200,
				Temperature: 42.0,
				Health:      95,
				SmartStatus: "healthy",
			},
			{
				Device:      "/dev/sdb",
				Model:       "WD Red Pro 4TB",
				Size:        4000,
				Used:        2800,
				Temperature: 38.0,
				Health:      92,
				SmartStatus: "healthy",
			},
		},
	}

	writeJSONResponse(w, storage)
}

func GetSystemProcesses(w http.ResponseWriter, r *http.Request) {
	processes := models.ProcessInfo{
		Total:    245,
		Running:  3,
		Sleeping: 242,
	}

	writeJSONResponse(w, processes)
}

func GetSystemLogs(w http.ResponseWriter, r *http.Request) {
	// Mock system logs
	logs := []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-1 * time.Hour),
			"level":     "info",
			"message":   "System startup completed successfully",
			"service":   "kernel",
		},
		{
			"timestamp": time.Now().Add(-30 * time.Minute),
			"level":     "warning",
			"message":   "High memory usage detected",
			"service":   "systemd",
		},
		{
			"timestamp": time.Now().Add(-15 * time.Minute),
			"level":     "info",
			"message":   "NFS service started",
			"service":   "nfsd",
		},
	}

	writeJSONResponse(w, logs)
}

func RebootSystem(w http.ResponseWriter, r *http.Request) {
	// In production, actually reboot the system
	w.WriteHeader(http.StatusOK)
	writeJSONResponse(w, map[string]string{
		"status":  "rebooting",
		"message": "System reboot initiated",
	})
}

func ShutdownSystem(w http.ResponseWriter, r *http.Request) {
	// In production, actually shutdown the system
	w.WriteHeader(http.StatusOK)
	writeJSONResponse(w, map[string]string{
		"status":  "shutting_down",
		"message": "System shutdown initiated",
	})
}
