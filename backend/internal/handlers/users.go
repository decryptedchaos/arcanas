package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type UserInfo struct {
	Name     string          `json:"name"`
	UID      int             `json:"uid"`
	HomeDir  string          `json:"home_dir"`
	Username string          `json:"username"`
	Services map[string]bool `json:"services"`
	Groups   []string        `json:"groups"`
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	// Get service memberships
	serviceMemberships, err := getServiceMemberships()
	if err != nil {
		fmt.Printf("Warning: Could not get service memberships: %v\n", err)
		serviceMemberships = make(map[string]map[string]bool)
	}

	// Read all users from /etc/passwd
	content, err := os.ReadFile("/etc/passwd")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read /etc/passwd: %v", err), http.StatusInternalServerError)
		return
	}

	var users []UserInfo
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, ":")
		if len(fields) < 7 {
			continue
		}

		username := fields[0]
		uidStr := fields[2]
		homeDir := fields[5]

		// Parse UID
		uid, err := strconv.Atoi(uidStr)
		if err != nil {
			continue
		}

		// Filter criteria:
		// 1. UID >= 1000 (regular users)
		// 2. Has valid home directory in /home/ or /var/lib/arcanas/
		// 3. Home directory exists
		if uid < 1000 {
			continue
		}

		// Skip system users that might have high UIDs
		systemUsers := []string{"nobody", "daemon", "sys", "sync", "shutdown", "halt", "operator", "nfsnobody"}
		skip := false
		for _, sysUser := range systemUsers {
			if username == sysUser {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		// Check if home directory exists and is in expected locations
		if !strings.HasPrefix(homeDir, "/home/") && !strings.HasPrefix(homeDir, "/var/lib/arcanas/") {
			continue
		}

		// Check if home directory actually exists
		if _, err := os.Stat(homeDir); os.IsNotExist(err) {
			continue
		}

		// Try to get user's full name from GECOS field
		gecos := fields[4]
		name := strings.Split(gecos, ",")[0]
		if name == "" || name == username {
			name = username // Fallback to username if no full name
		}

		// Get user's groups
		groups, err := getUserGroups(username)
		if err != nil {
			groups = []string{}
		}

		// Build service membership map
		services := make(map[string]bool)
		for service, members := range serviceMemberships {
			services[service] = members[username]
		}

		users = append(users, UserInfo{
			Name:     name,
			UID:      uid,
			HomeDir:  homeDir,
			Username: username,
			Services: services,
			Groups:   groups,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// getServiceMemberships returns a map of service -> map[username]bool
func getServiceMemberships() (map[string]map[string]bool, error) {
	services := make(map[string]map[string]bool)

	// Samba users
	sambaUsers, err := getSambaUsers()
	if err == nil {
		services["samba"] = sambaUsers
	}

	// NFS users (check /etc/exports for user-specific exports)
	nfsUsers, err := getNFSUsers()
	if err == nil {
		services["nfs"] = nfsUsers
	}

	// SSH users (users with SSH keys or allowed in sshd_config)
	sshUsers, err := getSSHUsers()
	if err == nil {
		services["ssh"] = sshUsers
	}

	// Add more services as needed
	// services["ftp"] = getFTPUsers()
	// services["webdav"] = getWebDAVUsers()

	return services, nil
}

// getSambaUsers returns a map of usernames that exist in Samba
func getSambaUsers() (map[string]bool, error) {
	// Method 1: Check /etc/samba/smbpasswd if it exists
	if _, err := os.Stat("/etc/samba/smbpasswd"); err == nil {
		return getSambaUsersFromSmbPasswd()
	}

	// Method 2: Use pdbedit command if available
	if _, err := exec.LookPath("pdbedit"); err == nil {
		return getSambaUsersFromPdbEdit()
	}

	// Method 3: Check if user is in common Samba groups
	return getSambaUsersFromGroups()
}

func getSambaUsersFromSmbPasswd() (map[string]bool, error) {
	content, err := os.ReadFile("/etc/samba/smbpasswd")
	if err != nil {
		return nil, err
	}

	users := make(map[string]bool)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) > 0 {
			users[fields[0]] = true
		}
	}
	return users, nil
}

func getSambaUsersFromPdbEdit() (map[string]bool, error) {
	cmd := exec.Command("sudo", "pdbedit", "-L")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	users := make(map[string]bool)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format is typically: username:UID:...
		fields := strings.Split(line, ":")
		if len(fields) > 0 {
			users[fields[0]] = true
		}
	}
	return users, nil
}

func getSambaUsersFromGroups() (map[string]bool, error) {
	users := make(map[string]bool)

	// Common Samba-related groups
	sambaGroups := []string{"sambashare", "samba", "users"}

	for _, group := range sambaGroups {
		cmd := exec.Command("getent", "group", group)
		output, err := cmd.Output()
		if err != nil {
			continue // Skip if group doesn't exist
		}

		// Parse group members: group:x:gid:user1,user2,user3
		line := strings.TrimSpace(string(output))
		if line == "" {
			continue
		}

		fields := strings.Split(line, ":")
		if len(fields) < 4 {
			continue
		}

		members := strings.Split(fields[3], ",")
		for _, member := range members {
			member = strings.TrimSpace(member)
			if member != "" {
				users[member] = true
			}
		}
	}

	return users, nil
}

func getNFSUsers() (map[string]bool, error) {
	users := make(map[string]bool)

	// Parse /etc/exports for user-specific exports
	content, err := os.ReadFile("/etc/exports")
	if err != nil {
		return users, nil // Return empty map if file doesn't exist
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Look for user-specific exports like /home/username
		if strings.Contains(line, "/home/") {
			// Extract username from path like /home/username
			parts := strings.Fields(line)
			if len(parts) > 0 {
				path := parts[0]
				if strings.HasPrefix(path, "/home/") {
					username := strings.TrimPrefix(path, "/home/")
					username = strings.Split(username, "/")[0] // Get just the username
					if username != "" {
						users[username] = true
					}
				}
			}
		}
	}

	return users, nil
}

func getSSHUsers() (map[string]bool, error) {
	users := make(map[string]bool)

	// Check for authorized keys in user home directories
	homeDir := "/home"
	entries, err := os.ReadDir(homeDir)
	if err != nil {
		return users, nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		username := entry.Name()
		sshDir := fmt.Sprintf("%s/%s/.ssh", homeDir, username)

		// Check if .ssh directory exists
		if _, err := os.Stat(sshDir); os.IsNotExist(err) {
			continue
		}

		// Check for authorized_keys file
		authKeysFile := fmt.Sprintf("%s/authorized_keys", sshDir)
		if _, err := os.Stat(authKeysFile); err == nil {
			users[username] = true
		}
	}

	return users, nil
}

func getUserGroups(username string) ([]string, error) {
	cmd := exec.Command("groups", username)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Output format: username : group1 group2 group3
	line := strings.TrimSpace(string(output))
	parts := strings.Split(line, ":")
	if len(parts) < 2 {
		return []string{}, nil
	}

	groups := strings.Fields(strings.TrimSpace(parts[1]))
	return groups, nil
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user struct {
		Username string   `json:"username"`
		Name     string   `json:"name"`
		Password string   `json:"password"`
		Groups   []string `json:"groups"`
		Shell    string   `json:"shell"`
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		handleError(w, err, "Failed to decode user data", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if user.Username == "" {
		handleError(w, fmt.Errorf("username is required"), "Username is required", http.StatusBadRequest)
		return
	}
	if user.Password == "" {
		handleError(w, fmt.Errorf("password is required"), "Password is required", http.StatusBadRequest)
		return
	}

	// Default shell if not provided
	if user.Shell == "" {
		user.Shell = "/bin/bash"
	}

	// Create user with useradd command
	args := []string{"useradd", "-m", "-s", user.Shell}
	if user.Name != "" {
		args = append(args, "-c", user.Name)
	}
	if len(user.Groups) > 0 {
		args = append(args, "-G", strings.Join(user.Groups, ","))
	}
	args = append(args, user.Username)

	cmd := exec.Command("sudo", args...)
	if err := cmd.Run(); err != nil {
		handleError(w, err, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Set password using echo and chpasswd
	passwordScript := fmt.Sprintf("echo '%s:%s' | sudo chpasswd", user.Username, user.Password)
	cmd = exec.Command("sh", "-c", passwordScript)
	if err := cmd.Run(); err != nil {
		handleError(w, err, "Failed to set password", http.StatusInternalServerError)
		return
	}

	writeJSONStatusResponse(w, "success", "User created successfully")
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Extract username from URL path - handle both /api/users/username and /api/users/username/services
	path := strings.TrimPrefix(r.URL.Path, "/api/users/")
	username := strings.Split(path, "/")[0]
	if username == "" {
		handleError(w, fmt.Errorf("missing username"), "Username is required", http.StatusBadRequest)
		return
	}

	var user struct {
		Name     string   `json:"name"`
		Groups   []string `json:"groups"`
		Password string   `json:"password,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		handleError(w, err, "Failed to decode user data", http.StatusBadRequest)
		return
	}

	// Log what we're trying to do
	fmt.Printf("DEBUG: Updating user '%s' with name '%s' and groups %v\n", username, user.Name, user.Groups)

	// Check if user exists first
	checkCmd := exec.Command("id", username)
	if err := checkCmd.Run(); err != nil {
		fmt.Printf("DEBUG: User '%s' does not exist: %v\n", username, err)
		handleError(w, err, "User does not exist", http.StatusNotFound)
		return
	}

	// Update user info with usermod command
	if user.Name != "" {
		fmt.Printf("DEBUG: Updating name for user '%s' to '%s'\n", username, user.Name)
		cmd := exec.Command("sudo", "usermod", "-c", user.Name, username)
		if err := cmd.Run(); err != nil {
			fmt.Printf("DEBUG: Failed to update name: %v\n", err)
			handleError(w, err, "Failed to update user name", http.StatusInternalServerError)
			return
		}
	}

	// Update groups if provided and not empty
	if len(user.Groups) > 0 {
		// Filter out empty groups
		validGroups := []string{}
		for _, group := range user.Groups {
			if strings.TrimSpace(group) != "" {
				validGroups = append(validGroups, strings.TrimSpace(group))
			}
		}

		if len(validGroups) > 0 {
			fmt.Printf("DEBUG: Updating groups for user '%s' to %v\n", username, validGroups)
			cmd := exec.Command("sudo", "usermod", "-G", strings.Join(validGroups, ","), username)
			if err := cmd.Run(); err != nil {
				fmt.Printf("DEBUG: Failed to update groups: %v\n", err)
				handleError(w, err, "Failed to update user groups", http.StatusInternalServerError)
				return
			}
		}
	}

	// Update password if provided
	if user.Password != "" {
		fmt.Printf("DEBUG: Updating password for user '%s'\n", username)
		// Use echo and chpasswd to update password
		passwordScript := fmt.Sprintf("echo '%s:%s' | sudo chpasswd", username, user.Password)
		cmd := exec.Command("sh", "-c", passwordScript)
		if err := cmd.Run(); err != nil {
			fmt.Printf("DEBUG: Failed to update password: %v\n", err)
			handleError(w, err, "Failed to update user password", http.StatusInternalServerError)
			return
		}

		// Check if user has Samba enabled and sync password
		serviceMemberships, err := getServiceMemberships()
		if err == nil {
			if sambaUsers, exists := serviceMemberships["samba"]; exists && sambaUsers[username] {
				// Update Samba password to match system password using proper syntax
				fmt.Printf("DEBUG: Syncing Samba password for user '%s'\n", username)
				sambaCmd := exec.Command("sudo", "sh", "-c", fmt.Sprintf("(echo '%s'; echo '%s') | smbpasswd -a %s", user.Password, user.Password, username))
				_ = sambaCmd.Run() // Ignore errors, Samba sync is optional

				// Enable the Samba user
				enableCmd := exec.Command("sudo", "smbpasswd", "-e", username)
				_ = enableCmd.Run()
			}
		}
	}

	fmt.Printf("DEBUG: Successfully updated user '%s'\n", username)
	writeJSONStatusResponse(w, "success", "User updated successfully")
}

func UpdateUserServices(w http.ResponseWriter, r *http.Request) {
	// Extract username from URL path like /api/users/username/services
	path := strings.TrimPrefix(r.URL.Path, "/api/users/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "services" {
		handleError(w, fmt.Errorf("invalid path"), "Invalid services endpoint", http.StatusBadRequest)
		return
	}

	username := parts[0]
	if username == "" {
		handleError(w, fmt.Errorf("missing username"), "Username is required", http.StatusBadRequest)
		return
	}

	var services struct {
		Services map[string]bool `json:"services"`
	}

	if err := json.NewDecoder(r.Body).Decode(&services); err != nil {
		handleError(w, err, "Failed to decode services data", http.StatusBadRequest)
		return
	}

	// Check if user exists first
	checkCmd := exec.Command("id", username)
	if err := checkCmd.Run(); err != nil {
		handleError(w, err, "User does not exist", http.StatusNotFound)
		return
	}

	// Update Samba service
	if sambaEnabled, exists := services.Services["samba"]; exists {
		if sambaEnabled {
			// Add user to Samba
			cmd := exec.Command("sudo", "pdbedit", "-a", username)
			cmd.Stdin = strings.NewReader("temp123\ntemp123\n") // Temporary password
			if err := cmd.Run(); err != nil {
				// User might already exist, try to enable instead
				cmd = exec.Command("sudo", "smbpasswd", "-e", username)
				_ = cmd.Run() // Ignore errors, user might already be enabled
			}

			// Configure Samba to use system authentication (password sync)
			// Update smb.conf to use system password database
			updateCmd := exec.Command("sudo", "sed", "-i", "s/# passdb backend = tdbsam/passdb backend = tdbsam/", "/etc/samba/smb.conf")
			_ = updateCmd.Run()

			// Add password sync configuration if not present
			syncCmd := exec.Command("sudo", "sh", "-c", "grep -q 'unix password sync' /etc/samba/smb.conf || echo 'unix password sync = yes' >> /etc/samba/smb.conf")
			_ = syncCmd.Run()

			pwdProgCmd := exec.Command("sudo", "sh", "-c", "grep -q 'passwd program' /etc/samba/smb.conf || echo 'passwd program = /usr/bin/passwd %u' >> /etc/samba/smb.conf")
			_ = pwdProgCmd.Run()

			pwdChatCmd := exec.Command("sudo", "sh", "-c", "grep -q 'passwd chat' /etc/samba/smb.conf || echo 'passwd chat = *Enter\\snew\\sUNIX\\spassword:* %n\\n *Retype\\snew\\sUNIX\\spassword:* %n\\n *password\\supdated\\ssuccessfully.* .' >> /etc/samba/smb.conf")
			_ = pwdChatCmd.Run()

			// Use PAM authentication for Samba (this is the key fix)
			pamCmd := exec.Command("sudo", "sed", "-i", "s/security = user/security = user\\n   encrypt passwords = yes\\n   obey pam restrictions = yes/", "/etc/samba/smb.conf")
			_ = pamCmd.Run()

			// Restart Samba to apply changes
			restartCmd := exec.Command("sudo", "systemctl", "restart", "smbd")
			_ = restartCmd.Run()
			restartNmbCmd := exec.Command("sudo", "systemctl", "restart", "nmbd")
			_ = restartNmbCmd.Run()
		} else {
			// Remove/disable user from Samba
			cmd := exec.Command("sudo", "smbpasswd", "-d", username)
			_ = cmd.Run() // Ignore errors if user doesn't exist
		}
	}

	// Update NFS service (create/remove user home directory export)
	if nfsEnabled, exists := services.Services["nfs"]; exists {
		if nfsEnabled {
			// Add NFS export for user's home directory
			homeDir := fmt.Sprintf("/home/%s", username)
			// Ensure home directory exists
			mkdirCmd := exec.Command("sudo", "mkdir", "-p", homeDir)
			_ = mkdirCmd.Run() // Ignore errors, directory might already exist
			chownCmd := exec.Command("sudo", "chown", username+":", homeDir)
			_ = chownCmd.Run() // Ignore errors, ownership might already be correct

			// Add to /etc/exports if not already present
			exportsEntry := fmt.Sprintf("%s *(rw,sync,no_subtree_check)", homeDir)
			appendCmd := exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo '%s' >> /etc/exports", exportsEntry))
			_ = appendCmd.Run() // Ignore errors, entry might already exist

			// Reload NFS exports
			reloadCmd := exec.Command("sudo", "exportfs", "-ra")
			_ = reloadCmd.Run() // Ignore errors, export might already be reloaded
		} else {
			// Remove NFS export for user's home directory
			homeDir := fmt.Sprintf("/home/%s", username)
			// Remove from /etc/exports
			removeCmd := exec.Command("sudo", "sed", "-i", fmt.Sprintf("/^%s /d", homeDir), "/etc/exports")
			_ = removeCmd.Run() // Ignore errors, entry might not exist

			// Reload NFS exports
			reloadCmd := exec.Command("sudo", "exportfs", "-ra")
			_ = reloadCmd.Run() // Ignore errors, export might already be reloaded
		}
	}

	// Update SSH service (add/remove SSH keys)
	if sshEnabled, exists := services.Services["ssh"]; exists {
		if sshEnabled {
			// Ensure user can SSH in (create .ssh directory if needed)
			sshDir := fmt.Sprintf("/home/%s/.ssh", username)
			mkdirCmd := exec.Command("sudo", "mkdir", "-p", sshDir)
			_ = mkdirCmd.Run() // Ignore errors, directory might already exist
			chownCmd := exec.Command("sudo", "chown", username+":", sshDir)
			_ = chownCmd.Run() // Ignore errors, ownership might already be correct
			chmodCmd := exec.Command("sudo", "chmod", "700", sshDir)
			_ = chmodCmd.Run() // Ignore errors, permissions might already be correct
		} else {
			// Disable SSH access (could lock account or remove SSH keys)
			// For now, we'll just remove SSH keys
			authKeysFile := fmt.Sprintf("/home/%s/.ssh/authorized_keys", username)
			removeCmd := exec.Command("sudo", "rm", "-f", authKeysFile)
			_ = removeCmd.Run() // Ignore errors, file might not exist
		}
	}

	writeJSONStatusResponse(w, "success", "User services updated successfully")
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Extract username from URL
	username := strings.TrimPrefix(r.URL.Path, "/api/users/")
	if username == "" {
		handleError(w, fmt.Errorf("missing username"), "Username is required", http.StatusBadRequest)
		return
	}

	// Delete user with userdel command
	cmd := exec.Command("sudo", "userdel", "-r", username)
	if err := cmd.Run(); err != nil {
		handleError(w, err, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	writeJSONStatusResponse(w, "success", "User deleted successfully")
}
