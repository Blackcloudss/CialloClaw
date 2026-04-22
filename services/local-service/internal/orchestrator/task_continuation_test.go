package orchestrator

import (
	"strings"
	"testing"
	"time"

	contextsvc "github.com/cialloclaw/cialloclaw/services/local-service/internal/context"
	"github.com/cialloclaw/cialloclaw/services/local-service/internal/runengine"
)

func TestResolveTaskContinuationContextUsesSingleActiveSession(t *testing.T) {
	service := newTestService()
	activeTask := service.runEngine.CreateTask(runengine.CreateTaskInput{
		SessionID:   "sess_active",
		Title:       "Analyze the current failure",
		SourceType:  "hover_input",
		Status:      "waiting_auth",
		CurrentStep: "waiting_authorization",
		RiskLevel:   "yellow",
	})

	continuationContext := service.resolveTaskContinuationContext("")
	if continuationContext.SessionID != activeTask.SessionID {
		t.Fatalf("expected active session %s, got %+v", activeTask.SessionID, continuationContext)
	}
	if continuationContext.SessionMode != "implicit_active" {
		t.Fatalf("expected implicit_active session mode, got %+v", continuationContext)
	}
	if len(continuationContext.Candidates) != 1 || continuationContext.Candidates[0].TaskID != activeTask.TaskID {
		t.Fatalf("expected active task to remain the only continuation candidate, got %+v", continuationContext.Candidates)
	}
}

func TestBuildTaskContinuationPromptRedactsSensitivePayloads(t *testing.T) {
	snapshot := contextsvc.TaskContextSnapshot{
		Trigger:       "hover_text_input",
		InputType:     "text",
		Text:          "Focus on the database timeout and keep the scope narrow.",
		SelectionText: "panic: dial tcp timeout",
		ErrorText:     "database timeout",
		Files:         []string{"logs/network.log"},
		PageTitle:     "Internal dashboard",
		PageURL:       "https://internal.example/tasks/1",
		ScreenSummary: "database panel shows critical errors",
	}
	candidate := runengine.TaskRecord{
		TaskID:      "task_001",
		SessionID:   "sess_secret",
		Title:       "Investigate the production timeout",
		Status:      "processing",
		CurrentStep: "agent_loop",
		SourceType:  "hover_input",
		UpdatedAt:   time.Now().Add(-30 * time.Second),
		Intent:      map[string]any{"name": "agent_loop"},
		Snapshot: contextsvc.TaskContextSnapshot{
			Files:     []string{"logs/private.log"},
			PageTitle: "Production dashboard",
		},
	}

	prompt := buildTaskContinuationPrompt(snapshot, map[string]any{"name": "write_file"}, taskContinuationContext{
		SessionMode: "implicit_active",
		Candidates:  []runengine.TaskRecord{candidate},
	})

	for _, sensitive := range []string{
		snapshot.Text,
		snapshot.SelectionText,
		snapshot.ErrorText,
		snapshot.PageURL,
		"logs/network.log",
		"logs/private.log",
		candidate.SessionID,
		candidate.Title,
	} {
		if strings.Contains(prompt, sensitive) {
			t.Fatalf("expected prompt to redact %q, got %s", sensitive, prompt)
		}
	}
	if !strings.Contains(prompt, "task_id=task_001") {
		t.Fatalf("expected prompt to retain stable task identifiers, got %s", prompt)
	}
}

func TestHeuristicTaskContinuationDecisionDoesNotAutoMergeBareFileDrop(t *testing.T) {
	decision := heuristicTaskContinuationDecision(
		contextsvc.TaskContextSnapshot{
			Trigger:   "file_drop",
			InputType: "file",
			Files:     []string{"logs/network.log"},
		},
		nil,
		taskContinuationContext{
			Candidates: []runengine.TaskRecord{{
				TaskID:      "task_001",
				Status:      "processing",
				CurrentStep: "agent_loop",
				SourceType:  "hover_input",
				UpdatedAt:   time.Now().Add(-10 * time.Second),
				Intent:      map[string]any{"name": "write_file"},
			}},
		},
	)

	if decision.Decision != "new_task" {
		t.Fatalf("expected bare file drop to stay a new task when fallback runs, got %+v", decision)
	}
}
