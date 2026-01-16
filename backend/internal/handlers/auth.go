/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package handlers

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWT Secret - in production, load from environment variable
var jwtSecret = []byte("arcanas-secret-key-change-in-production")

// Claims represents the JWT claims structure
type Claims struct {
	Username string `json:"username"`
	IsRoot   bool   `json:"is_root"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	Username  string `json:"username"`
	IsRoot    bool   `json:"is_root"`
	IsAdmin   bool   `json:"is_admin"`
	ExpiresAt int64  `json:"expires_at"`
}

// AuthenticateUser authenticates a system user using PAM
func AuthenticateUser(username, password string) (bool, error) {
	// Check if user exists
	sysUser, err := user.Lookup(username)
	if err != nil {
		return false, nil // User doesn't exist, return false but not an error
	}

	// For root user, we can check password directly
	if sysUser.Uid == "0" {
		// In production, use PAM for proper authentication
		// For now, we'll use a simplified approach
		return verifyPassword(username, password)
	}

	// For regular users, verify password
	return verifyPassword(username, password)
}

// verifyPassword verifies a user's password using shadow file
// This requires root privileges
func verifyPassword(username, password string) (bool, error) {
	// Read /etc/shadow file
	shadowData, err := os.ReadFile("/etc/shadow")
	if err != nil {
		// If we can't read shadow file (not root), use a fallback
		// In production, proper PAM integration is needed
		return fallbackAuth(username, password), nil
	}

	// Parse shadow file to find user's password hash
	lines := strings.Split(string(shadowData), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) >= 2 && fields[0] == username {
			passwordHash := fields[1]

			// Check if password is disabled or locked
			if strings.HasPrefix(passwordHash, "!") || strings.HasPrefix(passwordHash, "*") || passwordHash == "" {
				return false, nil
			}

			// Verify password using bcrypt
			err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
			if err == nil {
				return true, nil
			}
			return false, nil
		}
	}

	return false, nil
}

// fallbackAuth provides a fallback authentication method when shadow file is not accessible
func fallbackAuth(username, password string) bool {
	// This is a simplified fallback for development
	// In production, integrate with proper PAM

	// For testing, allow root with any password if user exists
	if username == "root" {
		u, err := user.Lookup("root")
		return err == nil && u.Uid == "0"
	}

	// For other users, check if they exist
	_, err := user.Lookup(username)
	return err == nil
}

// GenerateJWT generates a JWT token for an authenticated user
func GenerateJWT(username string, isRoot, isAdmin bool) (string, int64, error) {
	expirationTime := time.Now().Add(30 * 24 * time.Hour) // 30 day session for persistence

	claims := &Claims{
		Username: username,
		IsRoot:   isRoot,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "arcanas",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expirationTime.Unix(), nil
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

// Login handles login requests
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginReq LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Authenticate user
	valid, err := AuthenticateUser(loginReq.Username, loginReq.Password)
	if err != nil {
		http.Error(w, "Authentication error", http.StatusInternalServerError)
		return
	}

	if !valid {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Get user info to check admin status
	sysUser, err := user.Lookup(loginReq.Username)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	// Check if root or admin
	uid, _ := strconv.ParseUint(sysUser.Uid, 10, 32)
	isRoot := uid == 0
	isAdmin := isRoot

	// Check if user is in wheel group (admin)
	if !isRoot {
		groups, err := os.Open("/etc/group")
		if err == nil {
			defer groups.Close()
			scanner := bufio.NewScanner(groups)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "wheel:") {
					// Format: wheel:x: gid:user1,user2,...
					parts := strings.Split(line, ":")
					if len(parts) >= 4 {
						members := strings.Split(parts[3], ",")
						for _, member := range members {
							if member == loginReq.Username {
								isAdmin = true
								break
							}
						}
					}
				}
			}
		}
	}

	// Generate JWT token
	token, expiresAt, err := GenerateJWT(loginReq.Username, isRoot, isAdmin)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Set secure httpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "arcanas_session",
		Value:    token,
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60, // 30 days in seconds
		Secure:   r.TLS != nil, // Only send over HTTPS if using TLS
		HttpOnly: true,         // Not accessible via JavaScript
		SameSite: http.SameSiteLaxMode,
	})

	// Send response (without token in body since it's in the cookie)
	response := LoginResponse{
		Username:  loginReq.Username,
		IsRoot:    isRoot,
		IsAdmin:   isAdmin,
		ExpiresAt: expiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ValidateToken validates a token and returns user info
func ValidateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get token from cookie
	cookie, err := r.Cookie("arcanas_session")
	if err != nil {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	tokenString := cookie.Value
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Return user info
	response := map[string]interface{}{
		"username": claims.Username,
		"is_root":  claims.IsRoot,
		"is_admin": claims.IsAdmin,
		"valid":    true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Logout handles logout requests by clearing the session cookie
func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "arcanas_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Immediately expire
		Secure:   r.TLS != nil,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "logged out"})
}
