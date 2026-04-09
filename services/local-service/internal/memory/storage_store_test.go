package memory

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/cialloclaw/cialloclaw/services/local-service/internal/platform"
	storagesvc "github.com/cialloclaw/cialloclaw/services/local-service/internal/storage"
)

func TestNewServiceFromStorageWritesAndSearchesThroughStorageBoundary(t *testing.T) {
	storageService := storagesvc.NewService(platform.NewLocalStorageAdapter(filepath.Join(t.TempDir(), "memory.db")))
	defer func() { _ = storageService.Close() }()

	service := NewServiceFromStorage(storageService.MemoryStore(), storageService.Capabilities().MemoryStoreBackend)
	if service.RetrievalBackend() != "sqlite_wal" {
		t.Fatalf("unexpected retrieval backend: %q", service.RetrievalBackend())
	}

	err := service.WriteSummary(context.Background(), MemorySummary{
		MemorySummaryID: "mem_001",
		TaskID:          "task_old_001",
		RunID:           "run_old_001",
		Summary:         "user prefers markdown summary",
		CreatedAt:       "2026-04-08T10:00:00Z",
	})
	if err != nil {
		t.Fatalf("WriteSummary returned error: %v", err)
	}

	hits, err := service.Search(context.Background(), RetrievalQuery{
		TaskID: "task_001",
		RunID:  "run_001",
		Query:  "markdown summary",
		Limit:  5,
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(hits) != 1 || hits[0].MemoryID != "mem_001" {
		t.Fatalf("unexpected hits: %+v", hits)
	}

	references, err := service.RecentReferenceItems(context.Background(), 5)
	if err != nil {
		t.Fatalf("RecentReferenceItems returned error: %v", err)
	}
	if len(references) != 1 || references[0].MemoryID != "mem_001" {
		t.Fatalf("unexpected references: %+v", references)
	}
}
