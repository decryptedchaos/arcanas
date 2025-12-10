/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package models

import "time"

type NFSExport struct {
	ID                int         `json:"id"`
	Path              string      `json:"path"`
	Clients           []NFSClient `json:"clients"`
	Filesystem        string      `json:"filesystem"`
	Size              string      `json:"size"`
	Used              string      `json:"used"`
	ActiveConnections int         `json:"active_connections"`
	Created           time.Time   `json:"created"`
	LastModified      time.Time   `json:"last_modified"`
}

type NFSClient struct {
	Network string `json:"network"`
	Options string `json:"options"`
	Access  string `json:"access"`
}

type NFSConnection struct {
	ID         int       `json:"id"`
	ExportID   int       `json:"export_id"`
	ClientIP   string    `json:"client_ip"`
	Connected  time.Time `json:"connected"`
	BytesRead  int64     `json:"bytes_read"`
	BytesWrite int64     `json:"bytes_write"`
}
