package httpapi

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"
)

//go:embed chat/*
var chatStaticFiles embed.FS

// FileServerWithIndexFallback creates a file server that serves the given filesystem
// and falls back to index.html for any path that doesn't match a file
func FileServerWithIndexFallback() http.Handler {
	// First, try to get the embedded files
	subFS, err := fs.Sub(chatStaticFiles, "chat")
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, fmt.Sprintf("failed to get subfs: %s", err), http.StatusInternalServerError)
		})
	}
	chatFS := http.FS(subFS)
	fileServer := http.FileServer(chatFS)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		trimmedPath := strings.TrimPrefix(path, "/")
		fmt.Println("path", trimmedPath)
		if trimmedPath == "" {
			trimmedPath = "index.html"
		}

		// Try to serve the file directly
		_, err := chatFS.Open(trimmedPath)
		if err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		// If file doesn't exist, serve 404.html for any path
		if os.IsNotExist(err) {
			r2 := new(http.Request)
			*r2 = *r
			r2.URL.Path = "/404.html"
			fileServer.ServeHTTP(w, r2)
			return
		}

		// For other errors, return the error as is
		http.Error(w, fmt.Sprintf("failed to serve file: %s", err), http.StatusInternalServerError)
	})
}
