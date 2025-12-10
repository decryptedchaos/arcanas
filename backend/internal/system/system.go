/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package system

import (
	"bufio"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"arcanas/internal/models"
)

func GetSystemInfo() (models.SystemInfo, error) {
	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Get uptime
	uptime, err := getUptime()
	if err != nil {
		uptime = 0
	}

	// Get OS info
	osInfo, kernel, err := getOSInfo()
	if err != nil {
		osInfo = "Unknown"
		kernel = "Unknown"
	}

	return models.SystemInfo{
		Hostname:     hostname,
		Uptime:       uptime,
		OS:           osInfo,
		Kernel:       kernel,
		Architecture: runtime.GOARCH,
	}, nil
}

func getUptime() (int64, error) {
	file, err := os.Open("/proc/uptime")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 1 {
			uptime, _ := strconv.ParseFloat(fields[0], 64)
			return int64(uptime), nil
		}
	}

	return 0, err
}

func getOSInfo() (string, string, error) {
	// Get OS release info
	osRelease := "Unknown"
	kernel := "Unknown"

	// Try to read /etc/os-release
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				osRelease = strings.Trim(strings.Split(line, "=")[1], `"`)
				break
			}
		}
	}

	// Get kernel version from uname
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err == nil {
		kernel = strings.TrimSpace(string(output))
	}

	return osRelease, kernel, nil
}
