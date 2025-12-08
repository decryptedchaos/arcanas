package main

import (
	"arcanas/internal/routes"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	// Auto-build frontend if needed
	frontendDist := "../frontend/build"
	isDevMode := os.Getenv("DEV_MODE") == "true"

	// Always build in dev mode, or build once if directory doesn't exist
	if isDevMode {
		log.Println("Development mode: rebuilding frontend...")
		if err := buildFrontend(); err != nil {
			log.Printf("Failed to build frontend: %v", err)
			log.Println("Please run 'cd ../frontend && npm run build' manually")
		} else {
			log.Println("Frontend built successfully")
		}
	} else if _, err := os.Stat(frontendDist); os.IsNotExist(err) {
		log.Println("Frontend not built, building automatically...")
		if err := buildFrontend(); err != nil {
			log.Printf("Failed to build frontend: %v", err)
			log.Println("Please run 'cd ../frontend && npm run build' manually")
		} else {
			log.Println("Frontend built successfully")
		}
	} else {
		log.Println("Frontend already built, skipping build")
	}

	// Setup all routes
	mux := routes.SetupRoutes()

	// Check again after potential build
	if _, err := os.Stat(frontendDist); err == nil {
		// Create SPA handler that serves index.html for non-API routes
		spaHandler := createSPAHandler(frontendDist)
		mux.Handle("/", spaHandler)
		log.Println("Serving frontend from filesystem")
	} else {
		log.Printf("Frontend directory not found at: %s", frontendDist)
	}

	// Determine ports
	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "4000"
	}

	log.Printf("Arcanas API running on :%s", apiPort)
	if _, err := os.Stat(frontendDist); err == nil {
		log.Printf("Arcanas available at: http://localhost:%s", apiPort)
	} else {
		log.Println("Arcanas not available - build failed")
	}
	log.Fatal(http.ListenAndServe(":"+apiPort, mux))
}

func buildFrontend() error {
	// Change to frontend directory
	frontendDir := "../frontend"
	if err := os.Chdir(frontendDir); err != nil {
		return err
	}
	defer os.Chdir("../backend") // Change back

	// Check if npm is available
	if _, err := exec.LookPath("npm"); err != nil {
		return err
	}

	// Run npm install (if needed)
	if _, err := os.Stat("node_modules"); os.IsNotExist(err) {
		log.Println("Running npm install...")
		if err := exec.Command("npm", "install").Run(); err != nil {
			return err
		}
	}

	// Run npm run build
	log.Println("Running npm run build...")
	return exec.Command("npm", "run", "build").Run()
}

// createSPAHandler creates a handler that serves static files with SPA fallback
func createSPAHandler(dir string) http.Handler {
	fileServer := http.FileServer(http.Dir(dir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't handle API routes
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Try to serve the file
		path := filepath.Join(dir, r.URL.Path)

		// Check if file exists
		if _, err := os.Stat(path); err == nil {
			// File exists, serve it
			fileServer.ServeHTTP(w, r)
			return
		}

		// File doesn't exist, serve index.html for SPA routing
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}
