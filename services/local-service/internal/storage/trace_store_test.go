package storage

import (
	"context"
	"path/filepath"
	"testing"
)

func TestInMemoryTraceAndEvalStoresPersistAndList(t *testing.T) {
	traceStore := newInMemoryTraceStore()
	evalStore := newInMemoryEvalStore()
	if err := traceStore.WriteTraceRecord(context.Background(), TraceRecord{
		TraceID:          "trace_001",
		TaskID:           "task_001",
		RunID:            "run_001",
		LoopRound:        2,
		LLMInputSummary:  "input",
		LLMOutputSummary: "output",
		CreatedAt:        "2026-04-17T10:00:00Z",
	}); err != nil {
		t.Fatalf("write trace record failed: %v", err)
	}
	if err := evalStore.WriteEvalSnapshot(context.Background(), EvalSnapshotRecord{
		EvalSnapshotID: "eval_001",
		TraceID:        "trace_001",
		TaskID:         "task_001",
		Status:         "passed",
		MetricsJSON:    `{"latency_ms":321}`,
		CreatedAt:      "2026-04-17T10:00:00Z",
	}); err != nil {
		t.Fatalf("write eval snapshot failed: %v", err)
	}
	traces, total, err := traceStore.ListTraceRecords(context.Background(), "task_001", 10, 0)
	if err != nil || total != 1 || len(traces) != 1 {
		t.Fatalf("expected one trace record, total=%d len=%d err=%v", total, len(traces), err)
	}
	evals, total, err := evalStore.ListEvalSnapshots(context.Background(), "task_001", 10, 0)
	if err != nil || total != 1 || len(evals) != 1 {
		t.Fatalf("expected one eval snapshot, total=%d len=%d err=%v", total, len(evals), err)
	}
}

func TestSQLiteTraceAndEvalStoresPersistAndList(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "trace-eval.db")
	traceStore, err := NewSQLiteTraceStore(databasePath)
	if err != nil {
		t.Fatalf("new sqlite trace store failed: %v", err)
	}
	defer func() { _ = traceStore.Close() }()
	evalStore, err := NewSQLiteEvalStore(databasePath)
	if err != nil {
		t.Fatalf("new sqlite eval store failed: %v", err)
	}
	defer func() { _ = evalStore.Close() }()
	if err := traceStore.WriteTraceRecord(context.Background(), TraceRecord{
		TraceID:          "trace_sql_001",
		TaskID:           "task_sql_001",
		RunID:            "run_sql_001",
		LoopRound:        3,
		LLMInputSummary:  "input summary",
		LLMOutputSummary: "output summary",
		LatencyMS:        321,
		Cost:             0.012,
		RuleHitsJSON:     `{"doom_loop":"triggered"}`,
		ReviewResult:     "human_review_required",
		CreatedAt:        "2026-04-17T10:00:00Z",
	}); err != nil {
		t.Fatalf("write sqlite trace failed: %v", err)
	}
	if err := evalStore.WriteEvalSnapshot(context.Background(), EvalSnapshotRecord{
		EvalSnapshotID: "eval_sql_001",
		TraceID:        "trace_sql_001",
		TaskID:         "task_sql_001",
		Status:         "human_review_required",
		MetricsJSON:    `{"doom_loop_triggered":true}`,
		CreatedAt:      "2026-04-17T10:00:00Z",
	}); err != nil {
		t.Fatalf("write sqlite eval failed: %v", err)
	}
	traces, total, err := traceStore.ListTraceRecords(context.Background(), "task_sql_001", 10, 0)
	if err != nil || total != 1 || len(traces) != 1 {
		t.Fatalf("expected one sqlite trace record, total=%d len=%d err=%v", total, len(traces), err)
	}
	if traces[0].ReviewResult != "human_review_required" {
		t.Fatalf("expected review result to round-trip, got %+v", traces[0])
	}
	evals, total, err := evalStore.ListEvalSnapshots(context.Background(), "task_sql_001", 10, 0)
	if err != nil || total != 1 || len(evals) != 1 {
		t.Fatalf("expected one sqlite eval snapshot, total=%d len=%d err=%v", total, len(evals), err)
	}
	if evals[0].Status != "human_review_required" {
		t.Fatalf("expected eval status to round-trip, got %+v", evals[0])
	}
}
