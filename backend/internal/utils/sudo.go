/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package utils

import (
	"os/exec"
	"strings"
)

// SudoCommand executes a command with sudo
func SudoCommand(command string, args ...string) ([]byte, error) {
	fullArgs := append([]string{"sudo", command}, args...)
	cmd := exec.Command(fullArgs[0], fullArgs[1:]...)
	return cmd.Output()
}

// SudoCommandWithInput executes a command with sudo and provides input
func SudoCommandWithInput(command string, args []string, input string) error {
	fullArgs := append([]string{"sudo", command}, args...)
	cmd := exec.Command(fullArgs[0], fullArgs[1:]...)
	cmd.Stdin = strings.NewReader(input)
	return cmd.Run()
}

// SudoRunCommand runs a command with sudo (for reload operations)
func SudoRunCommand(command string, args ...string) error {
	fullArgs := append([]string{"sudo", command}, args...)
	cmd := exec.Command(fullArgs[0], fullArgs[1:]...)
	return cmd.Run()
}

// SudoReadFile reads a file using sudo cat
func SudoReadFile(path string) ([]byte, error) {
	return SudoCommand("cat", path)
}

// SudoWriteFile writes to a file using sudo tee
func SudoWriteFile(path string, content string) error {
	return SudoCommandWithInput("tee", []string{path}, content)
}

// SudoAppendFile appends to a file using sudo tee -a
func SudoAppendFile(path string, content string) error {
	return SudoCommandWithInput("tee", []string{"-a", path}, content)
}

// SudoSystemctlReload reloads a service using sudo systemctl
func SudoSystemctlReload(service string) error {
	return SudoRunCommand("systemctl", "reload", service)
}

// SudoServiceReload reloads a service using sudo service
func SudoServiceReload(service string) error {
	return SudoRunCommand("service", service, "reload")
}
