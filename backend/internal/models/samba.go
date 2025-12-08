package models

import "time"

type SambaShare struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Comment      string    `json:"comment"`
	Users        []string  `json:"users"`
	Groups       []string  `json:"groups"`
	GuestOK      bool      `json:"guest_ok"`
	ReadOnly     bool      `json:"read_only"`
	Browseable   bool      `json:"browseable"`
	Available    bool      `json:"available"`
	Size         string    `json:"size"`
	Used         string    `json:"used"`
	Connections  int       `json:"connections"`
	Created      time.Time `json:"created"`
	LastModified time.Time `json:"last_modified"`
}

type SambaConnection struct {
	ID        int       `json:"id"`
	User      string    `json:"user"`
	Share     string    `json:"share"`
	IP        string    `json:"ip"`
	PID       string    `json:"pid"`
	Connected time.Time `json:"connected"`
}
