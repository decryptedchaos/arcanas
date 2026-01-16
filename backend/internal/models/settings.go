/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package models

// SystemSettings contains general system settings
type SystemSettings struct {
	Hostname string `json:"hostname"`
	Timezone string `json:"timezone"`
	Locale   string `json:"locale"`
}

// NetConfig represents network configuration for an interface
type NetConfig struct {
	Method  string `json:"method"`  // dhcp, static, link-local, auto
	Address string `json:"address"`
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway"`
}

// NetworkInterface represents a network interface configuration
type NetworkInterface struct {
	Name      string    `json:"name"`
	Type      string    `json:"type"`       // eth, wlan, bond, bridge
	MAC       string    `json:"mac"`
	IPv4      NetConfig `json:"ipv4"`
	IPv6      NetConfig `json:"ipv6"`
	DNS       []string  `json:"dns"`
	DHCP      bool      `json:"dhcp"`
	Up        bool      `json:"up"`
	Speed     string    `json:"speed"`
	Duplex    string    `json:"duplex"`
}

// TimezoneInfo contains timezone information
type TimezoneInfo struct {
	Current   string   `json:"current"`
	Available []string `json:"available"` // Common timezones
}
