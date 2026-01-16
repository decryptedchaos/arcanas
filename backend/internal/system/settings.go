/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package system

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"arcanas/internal/models"
	"arcanas/internal/utils"
)

// ReadHostname reads the system hostname
func ReadHostname() (string, error) {
	// Try hostnamectl first
	cmd := exec.Command("hostnamectl", "status")
	output, err := cmd.Output()
	if err == nil {
		for _, line := range strings.Split(string(output), "\n") {
			if strings.Contains(line, "Static hostname") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					return parts[2], nil
				}
			}
		}
	}

	// Fall back to reading /etc/hostname
	data, err := os.ReadFile("/etc/hostname")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// WriteHostname sets the system hostname
func WriteHostname(hostname string) error {
	// Update /etc/hostname using sudo
	if err := utils.SudoWriteFile("/etc/hostname", hostname+"\n"); err != nil {
		return fmt.Errorf("failed to write hostname: %w", err)
	}

	// Apply using hostnamectl with sudo
	if err := utils.SudoRunCommand("hostnamectl", "set-hostname", hostname); err != nil {
		return fmt.Errorf("failed to apply hostname: %w", err)
	}

	return nil
}

// ReadTimezone reads the system timezone
func ReadTimezone() (string, error) {
	// Try timedatectl first
	cmd := exec.Command("timedatectl", "status")
	output, err := cmd.Output()
	if err == nil {
		for _, line := range strings.Split(string(output), "\n") {
			if strings.Contains(line, "Time zone") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					return parts[2], nil
				}
			}
		}
	}

	// Fall back to /etc/timezone
	data, err := os.ReadFile("/etc/timezone")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// SetTimezone sets the system timezone
func SetTimezone(timezone string) error {
	if err := utils.SudoRunCommand("timedatectl", "set-timezone", timezone); err != nil {
		return fmt.Errorf("failed to set timezone: %w", err)
	}
	return nil
}

// GetCommonTimezones returns a list of common timezones
func GetCommonTimezones() []string {
	return []string{
		"UTC",
		"America/New_York",
		"America/Chicago",
		"America/Denver",
		"America/Los_Angeles",
		"America/Anchorage",
		"America/Honolulu",
		"Europe/London",
		"Europe/Paris",
		"Europe/Berlin",
		"Europe/Moscow",
		"Asia/Tokyo",
		"Asia/Shanghai",
		"Asia/Dubai",
		"Asia/Kolkata",
		"Australia/Sydney",
		"Pacific/Auckland",
	}
}

// ReadNetworkConfig reads network interface configuration
func ReadNetworkConfig() (map[string]NetworkInterfaceInfo, error) {
	result := make(map[string]NetworkInterfaceInfo)

	// Get interface list from ip command
	cmd := exec.Command("ip", "-o", "link", "show")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list interfaces: %w", err)
	}

	interfaces := strings.Split(string(output), "\n")
	for _, ifaceLine := range interfaces {
		ifaceLine = strings.TrimSpace(ifaceLine)
		if ifaceLine == "" {
			continue
		}

		// Parse: 2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc...
		fields := strings.Fields(ifaceLine)
		if len(fields) < 2 {
			continue
		}

		// Extract interface name (e.g., "eth0:")
		name := strings.TrimSuffix(fields[1], ":")
		if name == "lo" {
			continue
		}

		info := NetworkInterfaceInfo{
			Name: name,
			Type: getInterfaceType(name),
			Up:   strings.Contains(ifaceLine, "UP"),
		}

		// Find MAC address in the line (looks for "link/ether XX:XX:XX:XX:XX:XX")
		for i, field := range fields {
			if strings.HasPrefix(field, "link/ether") && len(fields) > i+1 {
				info.MAC = field[10:] // Remove "link/ether" prefix
				break
			}
		}

		result[name] = info
	}

	// Get IP addresses
	cmd = exec.Command("ip", "-o", "addr", "show")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		currentIface := ""
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}

			// Line starts with interface: "2: eth0    inet ..."
			if strings.Contains(fields[1], ":") {
				// This is the interface declaration line
				parts := strings.Split(fields[1], ":")
				if len(parts) == 2 {
					currentIface = parts[1]
				}
				continue
			}

			// Check for inet/inet6 addresses
			if fields[1] == "inet" && len(fields) >= 3 && currentIface != "" {
				if info, ok := result[currentIface]; ok {
					// Format: 192.168.1.100/24 brd 192.168.1.255 scope global eth0
					info.IPv4.Address = strings.Split(fields[2], "/")[0]
					// Extract CIDR and convert to netmask
					if len(fields[2]) > 0 {
						parts := strings.Split(fields[2], "/")
						if len(parts) == 2 {
							info.IPv4.Netmask = parts[1] // Store as CIDR for now
						}
					}
					result[currentIface] = info
				}
			} else if fields[1] == "inet6" && len(fields) >= 3 && currentIface != "" {
				if info, ok := result[currentIface]; ok {
					info.IPv6.Address = strings.Split(fields[2], "/")[0]
					result[currentIface] = info
				}
			}
		}
	}

	// Get gateway
	cmd = exec.Command("ip", "route")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "default") {
				// Format: default via 192.168.1.1 dev eth0 src 192.168.1.100
				fields := strings.Fields(line)
				for i, field := range fields {
					if field == "via" && i+1 < len(fields) {
						gateway := fields[i+1]
						// Find the interface for this gateway
						for j, f := range fields {
							if f == "dev" && j+1 < len(fields) {
								ifaceName := fields[j+1]
								if info, ok := result[ifaceName]; ok {
									info.IPv4.Gateway = gateway
									result[ifaceName] = info
								}
								break
							}
						}
						break
					}
				}
			}
		}
	}

	return result, nil
}

// NetworkInterfaceInfo contains network interface information
type NetworkInterfaceInfo struct {
	Name    string
	Type    string
	MAC     string
	IPv4    IPConfig
	IPv6    IPConfig
	DNS     []string
	DHCP    bool
	Up      bool
	Speed   string
	Duplex  string
}

// IPConfig contains IP configuration
type IPConfig struct {
	Address string
	Netmask string
	Gateway string
}

// getInterfaceType determines the interface type
func getInterfaceType(name string) string {
	if strings.HasPrefix(name, "eth") || strings.HasPrefix(name, "en") {
		return "eth"
	}
	if strings.HasPrefix(name, "wl") || strings.HasPrefix(name, "wlp") || strings.HasPrefix(name, "wlan") {
		return "wlan"
	}
	if strings.HasPrefix(name, "bond") {
		return "bond"
	}
	if strings.HasPrefix(name, "br") {
		return "bridge"
	}
	return "unknown"
}

// detectNetworkMethod detects which network management method is available
func detectNetworkMethod() string {
	// Check for nmcli (NetworkManager) - most universal
	if _, err := exec.LookPath("nmcli"); err == nil {
		return "nmcli"
	}
	// Check for netplan (Ubuntu/Debian)
	if _, err := os.Stat("/etc/netplan"); err == nil {
		return "netplan"
	}
	// Check for /etc/network/interfaces (traditional Debian)
	if _, err := os.Stat("/etc/network/interfaces"); err == nil {
		return "ifupdown"
	}
	// Default to nmcli as it's most common
	return "nmcli"
}

// WriteNetworkConfig writes network configuration using the detected method
func WriteNetworkConfig(interfaceName string, config NetworkConfig, ipv4 models.NetConfig, ipv6 models.NetConfig, dns []string) error {
	method := detectNetworkMethod()

	switch method {
	case "nmcli":
		return writeNetworkConfigNMCLI(interfaceName, config, ipv4, dns)
	case "netplan":
		return writeNetworkConfigNetplan(interfaceName, config, ipv4, dns)
	case "ifupdown":
		return writeNetworkConfigIfupdown(interfaceName, config, ipv4, dns)
	default:
		return fmt.Errorf("no supported network configuration method found")
	}
}

// writeNetworkConfigNMCLI uses NetworkManager's nmcli (most distro-neutral)
func writeNetworkConfigNMCLI(interfaceName string, config NetworkConfig, ipv4 models.NetConfig, dns []string) error {
	// Set IPv4 configuration
	if config.DHCP {
		if err := utils.SudoRunCommand("nmcli", "connection", "modify", interfaceName, "ipv4.method", "auto"); err != nil {
			return fmt.Errorf("failed to set DHCP: %w", err)
		}
	} else {
		if err := utils.SudoRunCommand("nmcli", "connection", "modify", interfaceName, "ipv4.method", "manual"); err != nil {
			return fmt.Errorf("failed to set manual: %w", err)
		}
		if err := utils.SudoRunCommand("nmcli", "connection", "modify", interfaceName, "ipv4.addresses", fmt.Sprintf("%s/%s", ipv4.Address, ipv4.Netmask)); err != nil {
			return fmt.Errorf("failed to set address: %w", err)
		}
		if ipv4.Gateway != "" {
			if err := utils.SudoRunCommand("nmcli", "connection", "modify", interfaceName, "ipv4.gateway", ipv4.Gateway); err != nil {
				return fmt.Errorf("failed to set gateway: %w", err)
			}
		}
	}

	// Set DNS
	if len(dns) > 0 {
		dnsStr := strings.Join(dns, ",")
		if err := utils.SudoRunCommand("nmcli", "connection", "modify", interfaceName, "ipv4.dns", dnsStr); err != nil {
			return fmt.Errorf("failed to set DNS: %w", err)
		}
	}

	// Apply changes
	if err := utils.SudoRunCommand("nmcli", "connection", "up", interfaceName); err != nil {
		return fmt.Errorf("failed to apply network config: %w", err)
	}

	return nil
}

// writeNetworkConfigNetplan uses netplan (Ubuntu/Debian)
func writeNetworkConfigNetplan(interfaceName string, config NetworkConfig, ipv4 models.NetConfig, dns []string) error {
	netplanPath := "/etc/netplan/arcanas-" + interfaceName + ".yaml"

	var newConfig bytes.Buffer
	newConfig.WriteString("network:\n")
	newConfig.WriteString("  version: 2\n")
	newConfig.WriteString("  renderer: networkd\n")
	newConfig.WriteString("  ethernets:\n")
	newConfig.WriteString(fmt.Sprintf("    %s:\n", interfaceName))

	if config.DHCP {
		newConfig.WriteString("      dhcp4: true\n")
	} else {
		newConfig.WriteString("      dhcp4: false\n")
		newConfig.WriteString(fmt.Sprintf("      addresses: [%s/%s]\n", ipv4.Address, ipv4.Netmask))
		if ipv4.Gateway != "" {
			newConfig.WriteString(fmt.Sprintf("      gateway4: %s\n", ipv4.Gateway))
		}
	}

	if len(dns) > 0 {
		newConfig.WriteString("      nameservers:\n")
		newConfig.WriteString("        addresses: [")
		for i, addr := range dns {
			if i > 0 {
				newConfig.WriteString(", ")
			}
			newConfig.WriteString(addr)
		}
		newConfig.WriteString("]\n")
	}

	if err := utils.SudoWriteFile(netplanPath, newConfig.String()); err != nil {
		return fmt.Errorf("failed to write netplan config: %w", err)
	}

	if err := utils.SudoRunCommand("netplan", "apply"); err != nil {
		return fmt.Errorf("failed to apply netplan config: %w", err)
	}

	return nil
}

// writeNetworkConfigIfupdown uses /etc/network/interfaces (traditional Debian)
func writeNetworkConfigIfupdown(interfaceName string, config NetworkConfig, ipv4 models.NetConfig, dns []string) error {
	// This is a simplified version - in production, you'd need to parse/edit the existing file
	// For now, return not implemented error
	return fmt.Errorf("ifupdown network configuration not yet implemented")
}

// NetworkConfig is a helper struct for network configuration
type NetworkConfig struct {
	DHCP bool
}

// GetDNS reads DNS configuration
func GetDNS() ([]string, error) {
	// Try reading from /etc/resolv.conf
	data, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to read resolv.conf: %w", err)
	}

	var dnsServers []string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "nameserver") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				dnsServers = append(dnsServers, parts[1])
			}
		}
	}

	return dnsServers, nil
}
