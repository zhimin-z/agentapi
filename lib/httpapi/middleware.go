package httpapi

import (
	"net/http"
	"strings"
)

// responseWriter wraps http.ResponseWriter to intercept redirects
type basePathResponseWriter struct {
	http.ResponseWriter
	basePath string
}

func (w *basePathResponseWriter) WriteHeader(statusCode int) {
	// Intercept redirects and prepend base path to Location header
	if statusCode >= 300 && statusCode < 400 {
		if location := w.Header().Get("Location"); location != "" {
			// Only modify relative redirects
			if !strings.HasPrefix(location, "http://") && !strings.HasPrefix(location, "https://") {
				if !strings.HasPrefix(location, w.basePath) {
					w.Header().Set("Location", w.basePath+location)
				}
			}
		}
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

// StripBasePath creates a middleware that strips the base path from incoming requests
func StripBasePath(basePath string) func(http.Handler) http.Handler {
	// Normalize base path: ensure it starts with / and doesn't end with /
	if basePath != "" {
		if !strings.HasPrefix(basePath, "/") {
			basePath = "/" + basePath
		}
		basePath = strings.TrimSuffix(basePath, "/")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if basePath != "" && strings.HasPrefix(r.URL.Path, basePath) {
				// Strip the base path
				r.URL.Path = strings.TrimPrefix(r.URL.Path, basePath)
				if r.URL.Path == "" {
					r.URL.Path = "/"
				}
				
				// Wrap response writer to handle redirects
				w = &basePathResponseWriter{
					ResponseWriter: w,
					basePath:       basePath,
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}