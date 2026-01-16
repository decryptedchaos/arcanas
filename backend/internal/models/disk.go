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

// SmartFullInfo contains complete SMART data for a disk
type SmartFullInfo struct {
	Device       string           `json:"device"`
	Model        string           `json:"model"`
	Serial       string           `json:"serial"`
	Firmware     string           `json:"firmware"`
	Status       string           `json:"status"`
	Health       int              `json:"health"`
	Temperature  int              `json:"temperature"`
	PowerOnHours int              `json:"power_on_hours"`
	PowerCycles  int              `json:"power_cycles"`
	Attributes   []SmartAttribute `json:"attributes"`
	SelfTests    []SmartTestEntry  `json:"self_tests"`
	PassedTests  int              `json:"passed_tests"`
	FailedTests  int              `json:"failed_tests"`
	Errors       []SmartError      `json:"errors"`
	Enabled      bool             `json:"enabled"`
	Offline      bool             `json:"offline"`
}

// SmartAttribute represents a single SMART attribute
type SmartAttribute struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Flag        string `json:"flag"`
	Value       int    `json:"value"`
	Worst       int    `json:"worst"`
	Threshold   int    `json:"threshold"`
	RawValue    int64  `json:"raw_value"`
	Failed      bool   `json:"failed"`
}

// SmartTestEntry represents a self-test log entry
type SmartTestEntry struct {
	TestNum  int    `json:"test_num"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	Remaining int   `json:"remaining"`
	LBA      int64  `json:"lba"`
	Hours    int    `json:"hours"`
}

// SmartError represents an error log entry
type SmartError struct {
	ErrorNum int    `json:"error_num"`
	Type     string `json:"type"`
	LBA      int64  `json:"lba"`
	State    string `json:"state"`
	Hours    int    `json:"hours"`
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
