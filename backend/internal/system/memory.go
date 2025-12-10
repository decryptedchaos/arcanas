/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package system

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"arcanas/internal/models"
)

func GetMemoryStats() (models.MemoryStats, error) {
	// Get memory info from /proc/meminfo
	memInfo, err := parseMeminfo()
	if err != nil {
		return models.MemoryStats{}, err
	}

	// Calculate usage percentage
	usage := 0.0
	if memInfo["MemTotal"] > 0 {
		usage = float64(memInfo["MemUsed"]) / float64(memInfo["MemTotal"]) * 100
	}

	return models.MemoryStats{
		Total:     memInfo["MemTotal"],
		Used:      memInfo["MemUsed"],
		Available: memInfo["MemAvailable"],
		Usage:     usage,
		Swap: models.SwapInfo{
			Total: memInfo["SwapTotal"],
			Used:  memInfo["SwapUsed"],
		},
	}, nil
}

func parseMeminfo() (map[string]int64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	memInfo := make(map[string]int64)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			key := fields[0]
			value, err := strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				continue
			}

			// Convert from KB to bytes
			value *= 1024

			switch key {
			case "MemTotal:":
				memInfo["MemTotal"] = value
			case "MemFree:":
				memInfo["MemFree"] = value
			case "MemAvailable:":
				memInfo["MemAvailable"] = value
			case "Buffers:":
				memInfo["Buffers"] = value
			case "Cached:":
				memInfo["Cached"] = value
			case "SwapTotal:":
				memInfo["SwapTotal"] = value
			case "SwapFree:":
				memInfo["SwapFree"] = value
			}
		}
	}

	// Calculate used memory
	memInfo["MemUsed"] = memInfo["MemTotal"] - memInfo["MemFree"] - memInfo["Buffers"] - memInfo["Cached"]

	// Calculate used swap
	memInfo["SwapUsed"] = memInfo["SwapTotal"] - memInfo["SwapFree"]

	return memInfo, nil
}
