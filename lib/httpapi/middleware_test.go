package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripBasePath(t *testing.T) {
	tests := []struct {
		name           string
		basePath       string
		requestPath    string
		expectedPath   string
		redirectPath   string
		expectedRedirect string
	}{
		{
			name:         "no base path",
			basePath:     "",
			requestPath:  "/status",
			expectedPath: "/status",
		},
		{
			name:         "with base path - match",
			basePath:     "/api/v1",
			requestPath:  "/api/v1/status",
			expectedPath: "/status",
		},
		{
			name:         "with base path - no match",
			basePath:     "/api/v1",
			requestPath:  "/other/status",
			expectedPath: "/other/status",
		},
		{
			name:         "with base path - root",
			basePath:     "/api/v1",
			requestPath:  "/api/v1",
			expectedPath: "/",
		},
		{
			name:         "base path with trailing slash",
			basePath:     "/api/v1/",
			requestPath:  "/api/v1/status",
			expectedPath: "/status",
		},
		{
			name:         "base path without leading slash",
			basePath:     "api/v1",
			requestPath:  "/api/v1/status",
			expectedPath: "/status",
		},
		{
			name:             "redirect with base path",
			basePath:         "/api/v1",
			requestPath:      "/api/v1/old",
			expectedPath:     "/old",
			redirectPath:     "/new",
			expectedRedirect: "/api/v1/new",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check that the path was modified correctly
				assert.Equal(t, tt.expectedPath, r.URL.Path)
				
				// If test specifies a redirect, do it
				if tt.redirectPath != "" {
					http.Redirect(w, r, tt.redirectPath, http.StatusFound)
				} else {
					w.WriteHeader(http.StatusOK)
				}
			})

			middleware := StripBasePath(tt.basePath)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest("GET", tt.requestPath, nil)
			rec := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rec, req)

			// Check redirect if expected
			if tt.redirectPath != "" {
				assert.Equal(t, http.StatusFound, rec.Code)
				assert.Equal(t, tt.expectedRedirect, rec.Header().Get("Location"))
			} else {
				assert.Equal(t, http.StatusOK, rec.Code)
			}
		})
	}
}