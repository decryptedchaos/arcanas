/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package models

import "time"

type DiskStats struct {
	Device     string    `json:"device"`
	Model      string    `json:"model"`
	Size       int64     `json:"size"`
	Used       int64     `json:"used"`
	Available  int64     `json:"available"`
	Usage      float64   `json:"usage"`
	Mountpoint string    `json:"mountpoint"`
	Filesystem string    `json:"filesystem"`
	ReadOnly   bool      `json:"read_only"`
	Smart      SmartInfo `json:"smart"`
}

type SmartInfo struct {
	Status      string    `json:"status"`
	Health      int       `json:"health"`
	Temperature int       `json:"temperature"`
	PassedTests int       `json:"passed_tests"`
	FailedTests int       `json:"failed_tests"`
	LastTest    time.Time `json:"last_test"`
}

type Partition struct {
	Device     string  `json:"device"`
	Mountpoint string  `json:"mountpoint"`
	Size       int64   `json:"size"`
	Used       int64   `json:"used"`
	Available  int64   `json:"available"`
	Usage      float64 `json:"usage"`
	Filesystem string  `json:"filesystem"`
}
