/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Hello(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Hello from Go backend"}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}
