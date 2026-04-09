package bootstrap

import (
	"path/filepath"
	"testing"

	"github.com/cialloclaw/cialloclaw/services/local-service/internal/config"
)

func TestNewWiresStorageBackedMemoryService(t *testing.T) {
	cfg := config.Config{
		RPC: config.RPCConfig{
			Transport:        "named_pipe",
			NamedPipeName:    `\\.\pipe\cialloclaw-rpc-test`,
			DebugHTTPAddress: ":0",
		},
		WorkspaceRoot: filepath.Join(t.TempDir(), "workspace"),
		DatabasePath:  filepath.Join(t.TempDir(), "data", "local.db"),
		Model: config.ModelConfig{
			Provider: "openai_responses",
			ModelID:  "gpt-5.4",
			Endpoint: "https://api.openai.com/v1/responses",
		},
	}

	app, err := New(cfg)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	defer func() { _ = app.Close() }()

	if app.storage == nil {
		t.Fatal("expected storage service to be wired")
	}
	if app.storage.MemoryStore() == nil {
		t.Fatal("expected storage memory store to be available")
	}
	if !app.storage.Capabilities().SupportsMemoryStore {
		t.Fatalf("expected storage capabilities to expose memory store: %+v", app.storage.Capabilities())
	}
}
