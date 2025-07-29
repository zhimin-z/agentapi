package httpapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"

	"github.com/coder/agentapi/lib/httpapi"
	"github.com/coder/agentapi/lib/logctx"
	"github.com/coder/agentapi/lib/msgfmt"
	"github.com/coder/agentapi/lib/termexec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func normalizeSchema(t *testing.T, schema any) any {
	t.Helper()
	switch val := (schema).(type) {
	case *any:
		normalizeSchema(t, *val)
	case []any:
		for i := range val {
			normalizeSchema(t, &val[i])
		}
		sort.SliceStable(val, func(i, j int) bool {
			return fmt.Sprintf("%v", val[i]) < fmt.Sprintf("%v", val[j])
		})
	case map[string]any:
		for k := range val {
			valUnderKey := val[k]
			normalizeSchema(t, &valUnderKey)
			val[k] = valUnderKey
		}
	}
	return schema
}

// Ensure the OpenAPI schema on disk is up to date.
// To update the schema, run `go run main.go server --print-openapi dummy > openapi.json`.
func TestOpenAPISchema(t *testing.T) {
	t.Parallel()

	ctx := logctx.WithLogger(context.Background(), slog.New(slog.NewTextHandler(os.Stdout, nil)))
	srv := httpapi.NewServer(ctx, msgfmt.AgentTypeClaude, nil, 0, "/chat", "")
	currentSchemaStr := srv.GetOpenAPI()
	var currentSchema any
	if err := json.Unmarshal([]byte(currentSchemaStr), &currentSchema); err != nil {
		t.Fatalf("failed to unmarshal current schema: %s", err)
	}

	diskSchemaFile, err := os.OpenFile("../../openapi.json", os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("failed to open disk schema: %s", err)
	}
	defer func() {
		_ = diskSchemaFile.Close()
	}()

	diskSchemaBytes, err := io.ReadAll(diskSchemaFile)
	if err != nil {
		t.Fatalf("failed to read disk schema: %s", err)
	}
	var diskSchema any
	if err := json.Unmarshal(diskSchemaBytes, &diskSchema); err != nil {
		t.Fatalf("failed to unmarshal disk schema: %s", err)
	}

	normalizeSchema(t, &currentSchema)
	normalizeSchema(t, &diskSchema)

	require.Equal(t, currentSchema, diskSchema)
}

func TestBasePath(t *testing.T) {
	t.Parallel()

	ctx := logctx.WithLogger(context.Background(), slog.New(slog.NewTextHandler(os.Stdout, nil)))
	proc, err := termexec.StartProcess(ctx, termexec.StartProcessConfig{
		Program: "sleep",
		Args:    []string{"inf"},
	})
	require.NoError(t, err)

	// Given: a server with a non-empty base path
	require.NoError(t, err)
	hndlr := httpapi.NewServer(ctx, msgfmt.AgentTypeCustom, proc, 0, "/chat", "/subpath").Handler()
	srv := httptest.NewServer(hndlr)
	t.Cleanup(srv.Close)

	// When: we make a request to "/"
	resp, err := srv.Client().Get(srv.URL)
	require.NoError(t, err)

	// Then: we get redirected to /belongs/to/me/embed/chat
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	location := resp.Header.Get("Location")
	assert.Equal(t, "/subpath/chat/embed", location)
}
