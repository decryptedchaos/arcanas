/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// writeJSONResponse writes JSON response with proper error handling
func writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleError handles common error scenarios with consistent responses
func handleError(w http.ResponseWriter, err error, message string, statusCode int) {
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	if message == "" {
		message = "Internal server error"
	}

	log.Printf("Error: %s - %v", message, err)
	http.Error(w, fmt.Sprintf("%s: %v", message, err), statusCode)
}

// writeJSONStatusResponse writes JSON status response with proper error handling
func writeJSONStatusResponse(w http.ResponseWriter, status, message string) {
	writeJSONResponse(w, map[string]string{
		"status":  status,
		"message": message,
	})
}
