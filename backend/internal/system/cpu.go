/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package system

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"arcanas/internal/models"
)

var (
	lastCPUUsage struct {
		total float64
		idle  float64
	}
	cpuMutex sync.Mutex
)

func GetCPUStats() (models.CPUStats, error) {
	// Get CPU usage from /proc/stat
	usage, err := getCPUUsage()
	if err != nil {
		return models.CPUStats{}, err
	}

	// Get CPU info from /proc/cpuinfo
	model, cores, frequency, err := getCPUInfo()
	if err != nil {
		return models.CPUStats{}, err
	}

	// Get load average
	loadAvg, err := getLoadAverage()
	if err != nil {
		return models.CPUStats{}, err
	}

	// Get temperature (try common paths)
	temp, err := getCPUTemperature()
	if err != nil {
		temp = 0 // Default if can't read temp
	}

	// Get process count
	processes, err := getProcessCount()
	if err != nil {
		processes = models.ProcessInfo{Total: 0, Running: 0, Sleeping: 0}
	}

	return models.CPUStats{
		Usage:       usage,
		Cores:       cores,
		Model:       model,
		Frequency:   frequency,
		Temperature: temp,
		LoadAverage: loadAvg,
		Processes:   processes,
	}, nil
}

func getCPUUsage() (float64, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) >= 8 {
				user, _ := strconv.ParseFloat(fields[1], 64)
				nice, _ := strconv.ParseFloat(fields[2], 64)
				system, _ := strconv.ParseFloat(fields[3], 64)
				idle, _ := strconv.ParseFloat(fields[4], 64)
				iowait, _ := strconv.ParseFloat(fields[5], 64)
				irq, _ := strconv.ParseFloat(fields[6], 64)
				softirq, _ := strconv.ParseFloat(fields[7], 64)

				total := user + nice + system + idle + iowait + irq + softirq
				idleTime := idle + iowait

				cpuMutex.Lock()
				if lastCPUUsage.total == 0 {
					// First reading, just store values
					lastCPUUsage.total = total
					lastCPUUsage.idle = idleTime
					cpuMutex.Unlock()
					return 0, nil // Return 0 on first reading
				}

				// Calculate difference from last reading
				totalDiff := total - lastCPUUsage.total
				idleDiff := idleTime - lastCPUUsage.idle

				// Update last values
				lastCPUUsage.total = total
				lastCPUUsage.idle = idleTime
				cpuMutex.Unlock()

				if totalDiff > 0 {
					usage := ((totalDiff - idleDiff) / totalDiff) * 100
					// Cap at 100% and round to 1 decimal place
					if usage > 100 {
						usage = 100
					}
					return usage, nil
				}
			}
		}
	}

	return 0, fmt.Errorf("could not parse CPU usage")
}

func getCPUInfo() (string, int, string, error) {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return "", 0, "", err
	}
	defer file.Close()

	var model string
	var cores int
	var freq float64

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.SplitN(line, ":", 2)
		if len(fields) == 2 {
			key := strings.TrimSpace(fields[0])
			value := strings.TrimSpace(fields[1])

			switch key {
			case "model name":
				model = value
			case "processor":
				if value == "0" {
					cores = 1
				} else {
					cores++
				}
			case "cpu MHz":
				if f, err := strconv.ParseFloat(value, 64); err == nil {
					freq = f
				}
			}
		}
	}

	frequency := fmt.Sprintf("%.1fGHz", freq/1000)
	return model, cores, frequency, nil
}

func getLoadAverage() ([]float64, error) {
	file, err := os.Open("/proc/loadavg")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 {
			load1, _ := strconv.ParseFloat(fields[0], 64)
			load5, _ := strconv.ParseFloat(fields[1], 64)
			load15, _ := strconv.ParseFloat(fields[2], 64)
			return []float64{load1, load5, load15}, nil
		}
	}

	return nil, fmt.Errorf("could not parse load average")
}

func getCPUTemperature() (float64, error) {
	// Try multiple approaches to find CPU temperature on different systems

	// 1. Try to find CPU-specific hwmon sensors by checking device names
	if temp := tryHwmonCPUTemp(); temp > 0 {
		return temp, nil
	}

	// 2. Try thermal zones with type detection
	if temp := tryThermalZoneCPUTemp(); temp > 0 {
		return temp, nil
	}

	// 3. Try common fallback paths
	if temp := tryCommonPaths(); temp > 0 {
		return temp, nil
	}

	return 0, fmt.Errorf("could not read CPU temperature")
}

func tryHwmonCPUTemp() float64 {
	// Find hwmon devices that are likely CPU-related
	hwmonDirs, err := os.ReadDir("/sys/class/hwmon")
	if err != nil {
		return 0
	}

	// First, specifically look for known CPU temperature drivers
	cpuDrivers := map[string]int{
		"k10temp":  1, // AMD CPU
		"coretemp": 2, // Intel CPU
		"zenpower": 3, // AMD Zen CPU
	}

	var bestTemp float64 = 0
	var bestPriority int = 999

	for _, hwmon := range hwmonDirs {
		hwmonPath := "/sys/class/hwmon/" + hwmon.Name()

		// Check if this hwmon device has a name file
		nameFile := hwmonPath + "/name"
		if nameData, err := os.ReadFile(nameFile); err == nil {
			deviceName := strings.TrimSpace(string(nameData))

			// Check if this is a known CPU driver
			if priority, exists := cpuDrivers[deviceName]; exists {
				if temp := readHwmonTemp(hwmonPath); temp > 0 {
					if priority < bestPriority {
						bestTemp = temp
						bestPriority = priority
					}
				}
			}
		}
	}

	// If we found a specific CPU driver, use it
	if bestTemp > 0 {
		return bestTemp
	}

	// Fallback: try any hwmon device but be more selective
	for _, hwmon := range hwmonDirs {
		hwmonPath := "/sys/class/hwmon/" + hwmon.Name()

		nameFile := hwmonPath + "/name"
		if nameData, err := os.ReadFile(nameFile); err == nil {
			deviceName := strings.TrimSpace(string(nameData))

			// Skip known non-CPU devices
			skipDevices := []string{"nvme", "drivetemp"}
			shouldSkip := false
			for _, skip := range skipDevices {
				if strings.Contains(strings.ToLower(deviceName), skip) {
					shouldSkip = true
					break
				}
			}

			if !shouldSkip {
				if temp := readHwmonTemp(hwmonPath); temp > 0 {
					return temp
				}
			}
		}
	}

	return 0
}

func readHwmonTemp(hwmonPath string) float64 {
	// Try temp1_input (primary sensor)
	if temp, err := readTempFile(hwmonPath + "/temp1_input"); err == nil && isReasonableCPUTemp(temp) {
		return temp
	}

	// Try other temp inputs
	for i := 2; i <= 10; i++ {
		tempFile := fmt.Sprintf(hwmonPath+"/temp%d_input", i)
		if temp, err := readTempFile(tempFile); err == nil && isReasonableCPUTemp(temp) {
			return temp
		}
	}

	return 0
}

func tryThermalZoneCPUTemp() float64 {
	// Try thermal zones with type checking
	thermalDirs, err := os.ReadDir("/sys/class/thermal")
	if err != nil {
		return 0
	}

	for _, thermal := range thermalDirs {
		if !strings.HasPrefix(thermal.Name(), "thermal_zone") {
			continue
		}

		thermalPath := "/sys/class/thermal/" + thermal.Name()

		// Check thermal zone type
		typeFile := thermalPath + "/type"
		if typeData, err := os.ReadFile(typeFile); err == nil {
			zoneType := strings.TrimSpace(string(typeData))
			// Look for CPU-related thermal zones
			if strings.Contains(strings.ToLower(zoneType), "cpu") ||
				strings.Contains(strings.ToLower(zoneType), "acpitz") {
				if temp, err := readTempFile(thermalPath + "/temp"); err == nil && isReasonableCPUTemp(temp) {
					return temp
				}
			}
		}
	}

	return 0
}

func tryCommonPaths() float64 {
	// Try some common paths as last resort
	commonPaths := []string{
		"/sys/class/thermal/thermal_zone0/temp",
		"/sys/class/thermal/thermal_zone1/temp",
		"/sys/class/hwmon/hwmon0/temp1_input",
		"/sys/class/hwmon/hwmon1/temp1_input",
	}

	for _, path := range commonPaths {
		if temp, err := readTempFile(path); err == nil && isReasonableCPUTemp(temp) {
			return temp
		}
	}

	return 0
}

func readTempFile(path string) (float64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	temp, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
	if err != nil {
		return 0, err
	}

	// Convert from millidegrees to degrees
	return temp / 1000, nil
}

func isReasonableCPUTemp(temp float64) bool {
	// Check if temperature is reasonable for a CPU
	return temp >= 10 && temp <= 110
}

func getProcessCount() (models.ProcessInfo, error) {
	// Get total process count
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return models.ProcessInfo{}, err
	}

	lines := strings.Split(string(output), "\n")
	total := len(lines) - 1 // Subtract header

	// Get running processes
	cmd = exec.Command("ps", "-eo", "stat")
	output, err = cmd.Output()
	if err != nil {
		return models.ProcessInfo{}, err
	}

	running := 0
	sleeping := 0
	lines = strings.Split(string(output), "\n")
	for _, line := range lines[1:] { // Skip header
		if strings.Contains(line, "R") {
			running++
		} else if strings.Contains(line, "S") {
			sleeping++
		}
	}

	return models.ProcessInfo{
		Total:    total,
		Running:  running,
		Sleeping: sleeping,
	}, nil
}
