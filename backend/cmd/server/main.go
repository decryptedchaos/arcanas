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
	// Auto-build frontend if needed (only for development)
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

	// Create SPA handler from embedded filesystem
	spaHandler := createEmbeddedSPAHandler()
	mux.Handle("/", spaHandler)
	log.Println("Serving frontend from embedded binary")

	// Determine ports
	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "4000"
	}

	log.Printf("Arcanas API running on :%s", apiPort)
	log.Printf("Arcanas available at: http://localhost:%s", apiPort)
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

// createEmbeddedSPAHandler creates a handler that serves static files with SPA fallback
func createEmbeddedSPAHandler() http.Handler {
	// Always try embedded files first (production)
	fileServer := http.FileServer(http.Dir("static"))
	log.Println("Using bundled frontend files")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't handle API routes
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Try to serve the file from static directory
		path := filepath.Join("static", r.URL.Path)
		if _, err := os.Stat(path); err == nil {
			// File exists, serve it
			fileServer.ServeHTTP(w, r)
			return
		}

		// If static files don't exist, fall back to filesystem (development)
		fallbackPath := filepath.Join("../frontend/build", r.URL.Path)
		if _, err := os.Stat(fallbackPath); err == nil {
			log.Println("Falling back to filesystem frontend files")
			fallbackServer := http.FileServer(http.Dir("../frontend/build"))
			fallbackServer.ServeHTTP(w, r)
			return
		}

		// File doesn't exist anywhere, serve index.html for SPA routing
		if _, err := os.Stat("static/index.html"); err == nil {
			http.ServeFile(w, r, filepath.Join("static", "index.html"))
		} else if _, err := os.Stat("../frontend/build/index.html"); err == nil {
			http.ServeFile(w, r, filepath.Join("../frontend/build", "index.html"))
		} else {
			http.NotFound(w, r)
		}
	})
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
