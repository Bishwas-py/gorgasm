// Package main implements the server for the Go fullstack framework
package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// Configure static file server
	fs := http.FileServer(http.Dir("./static"))

	// Custom file server for handling MIME types
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Set the correct MIME type based on file extension
		path := r.URL.Path
		ext := filepath.Ext(path)

		// Set specific MIME types for certain extensions
		if ext == ".wasm" {
			w.Header().Set("Content-Type", "application/wasm")
		}

		// Serve the file using the file server
		fs.ServeHTTP(w, r)
	})

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	log.Printf("Server starting on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
