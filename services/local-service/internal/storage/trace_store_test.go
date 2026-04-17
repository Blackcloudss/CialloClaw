package storage

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
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

func TestSQLiteEvalStoreEnforcesTraceForeignKey(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "trace-eval-fk.db")
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

	err = evalStore.WriteEvalSnapshot(context.Background(), EvalSnapshotRecord{
		EvalSnapshotID: "eval_orphan_001",
		TraceID:        "trace_missing",
		TaskID:         "task_sql_002",
		Status:         "passed",
		MetricsJSON:    `{"latency_ms":100}`,
		CreatedAt:      "2026-04-17T11:00:00Z",
	})
	if err == nil {
		t.Fatal("expected foreign key error when writing eval snapshot without trace record")
	}

	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatalf("open sqlite db failed: %v", err)
	}
	defer db.Close()
	rows, err := db.Query(`PRAGMA foreign_key_list(eval_snapshots);`)
	if err != nil {
		t.Fatalf("query foreign key list failed: %v", err)
	}
	defer rows.Close()

	hasTraceForeignKey := false
	for rows.Next() {
		var id, seq int
		var table, from, to, onUpdate, onDelete, match string
		if err := rows.Scan(&id, &seq, &table, &from, &to, &onUpdate, &onDelete, &match); err != nil {
			t.Fatalf("scan foreign key row failed: %v", err)
		}
		if table == "trace_records" && from == "trace_id" && to == "trace_id" {
			hasTraceForeignKey = true
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate foreign key rows failed: %v", err)
	}
	if !hasTraceForeignKey {
		t.Fatal("expected eval_snapshots to keep foreign key back to trace_records")
	}
}
