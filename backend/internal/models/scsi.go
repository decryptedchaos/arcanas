/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package models

import "time"

type SCSITarget struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	Sessions     int       `json:"sessions"`
	LUNCount     int       `json:"lun_count"`
	Size         string    `json:"size"`
	BackingStore string    `json:"backing_store"`
	InitiatorIPs []string  `json:"initiator_ips"`
	Created      time.Time `json:"created"`
	LastAccess   time.Time `json:"last_access"`
}

type LUN struct {
	ID          int    `json:"id"`
	TargetID    int    `json:"target_id"`
	LUN         int    `json:"lun"`
	Device      string `json:"device"`
	Size        int64  `json:"size"`
	BackingFile string `json:"backing_file"`
}

type Session struct {
	ID         int       `json:"id"`
	TargetID   int       `json:"target_id"`
	Initiator  string    `json:"initiator"`
	IP         string    `json:"ip"`
	Connected  time.Time `json:"connected"`
	BytesRead  int64     `json:"bytes_read"`
	BytesWrite int64     `json:"bytes_write"`
}
