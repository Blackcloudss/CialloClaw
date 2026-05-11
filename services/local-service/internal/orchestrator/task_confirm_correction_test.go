package orchestrator

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cialloclaw/cialloclaw/services/local-service/internal/model"
	"github.com/cialloclaw/cialloclaw/services/local-service/internal/runengine"
	"github.com/cialloclaw/cialloclaw/services/local-service/internal/taskcontext"
)

func TestTaskConfirmCorrectionContextSummaryNormalizesPromptFiles(t *testing.T) {
	files := []string{
		" first\nreport.md ",
		strings.Repeat("very-long-file-name-", 20) + ".txt",
		"",
		"third.txt",
		"fourth.txt",
		"fifth.txt",
		"sixth.txt",
		"seventh.txt",
		"eighth.txt",
		"ninth.txt",
	}

	summary := taskConfirmCorrectionContextSummary(taskcontext.TaskContextSnapshot{
		Source:    "floating_ball",
		Trigger:   "file_drop",
		InputType: "file",
		Files:     files,
	})
	if strings.Contains(summary, "first\nreport.md") {
		t.Fatalf("expected file names to be whitespace-normalized, got %q", summary)
	}
	if !strings.Contains(summary, "files=first report.md,") {
		t.Fatalf("expected normalized first file in summary, got %q", summary)
	}
	if strings.Contains(summary, "ninth.txt") {
		t.Fatalf("expected prompt file summary to cap oversized file lists, got %q", summary)
	}
	filesLine := ""
	for _, line := range strings.Split(summary, "\n") {
		if strings.HasPrefix(line, "files=") {
			filesLine = line
			break
		}
	}
	if filesLine == "" {
		t.Fatalf("expected files line in summary, got %q", summary)
	}
	if len(filesLine) > 246 {
		t.Fatalf("expected prompt file summary to stay bounded, got %d chars in %q", len(filesLine), filesLine)
	}
}

func TestServiceConfirmTaskFallsBackWhenCorrectionModelReturnsMalformedPayload(t *testing.T) {
	service, _ := newTestServiceWithModelClient(t, stubModelClient{
		generateText: func(request model.GenerateTextRequest) (model.GenerateTextResponse, error) {
			return model.GenerateTextResponse{
				TaskID:     request.TaskID,
				RunID:      request.RunID,
				RequestID:  "req_bad_correction_payload",
				Provider:   "openai_responses",
				ModelID:    "gpt-5.4",
				OutputText: "not-json",
			}, nil
		},
	})

	startResult, err := service.SubmitInput(map[string]any{
		"session_id": "sess_correction_bad_payload",
		"source":     "floating_ball",
		"trigger":    "hover_text_input",
		"input": map[string]any{
			"type": "text",
			"text": "帮我总结这个错误",
		},
		"options": map[string]any{
			"confirm_required": true,
		},
	})
	if err != nil {
		t.Fatalf("submit input failed: %v", err)
	}

	confirmResult, err := service.ConfirmTask(map[string]any{
		"task_id":         startResult["task"].(map[string]any)["task_id"].(string),
		"confirmed":       false,
		"correction_text": "改成解释这个错误",
	})
	if err != nil {
		t.Fatalf("confirm task correction failed: %v", err)
	}

	intentValue := confirmResult["task"].(map[string]any)["intent"].(map[string]any)
	if intentValue["name"] != "agent_loop" {
		t.Fatalf("expected malformed model payload to fall back to agent_loop, got %+v", intentValue)
	}
	arguments := intentValue["arguments"].(map[string]any)
	if arguments["goal"] != "改成解释这个错误" {
		t.Fatalf("expected fallback correction to preserve the user goal, got %+v", arguments)
	}
}

func TestServiceConfirmTaskAcceptsDirectIntentObjectFromCorrectionModel(t *testing.T) {
	service, _ := newTestServiceWithModelClient(t, stubModelClient{
		generateText: func(request model.GenerateTextRequest) (model.GenerateTextResponse, error) {
			return model.GenerateTextResponse{
				TaskID:     request.TaskID,
				RunID:      request.RunID,
				RequestID:  "req_direct_intent_payload",
				Provider:   "openai_responses",
				ModelID:    "gpt-5.4",
				OutputText: `{"name":"translate","arguments":{"target_language":"en"}}`,
			}, nil
		},
	})

	startResult, err := service.StartTask(map[string]any{
		"session_id": "sess_direct_intent_payload",
		"source":     "floating_ball",
		"trigger":    "text_selected_click",
		"input": map[string]any{
			"type": "text_selection",
			"text": "你好，世界",
		},
	})
	if err != nil {
		t.Fatalf("start task failed: %v", err)
	}

	confirmResult, err := service.ConfirmTask(map[string]any{
		"task_id":         startResult["task"].(map[string]any)["task_id"].(string),
		"confirmed":       false,
		"correction_text": "改成翻译成英文",
	})
	if err != nil {
		t.Fatalf("confirm task correction failed: %v", err)
	}

	task := confirmResult["task"].(map[string]any)
	intentValue := task["intent"].(map[string]any)
	if intentValue["name"] != "translate" {
		t.Fatalf("expected direct model intent payload to be accepted, got %+v", intentValue)
	}
	if task["status"] != "confirming_intent" || task["current_step"] != "intent_confirmation" {
		t.Fatalf("expected direct model intent payload to stay behind confirmation, got %+v", task)
	}
}

func TestBuildTaskConfirmCorrectionPromptUsesNormalizedFileSummary(t *testing.T) {
	files := make([]string, 0, 10)
	for index := 0; index < 10; index++ {
		files = append(files, fmt.Sprintf("artifact-%02d\nreport.md", index))
	}

	prompt := buildTaskConfirmCorrectionPrompt(
		runengine.TaskRecord{
			TaskID:     "task_prompt_files",
			Status:     "confirming_intent",
			SourceType: "dragged_file",
			Intent: map[string]any{
				"name": "summarize",
			},
		},
		taskcontext.TaskContextSnapshot{
			Source:    "floating_ball",
			Trigger:   "file_drop",
			InputType: "file",
			Files:     files,
		},
		"改成提炼行动项",
	)
	if strings.Contains(prompt, "artifact-00\nreport.md") {
		t.Fatalf("expected prompt file list to remove embedded newlines, got %q", prompt)
	}
	if !strings.Contains(prompt, "files=artifact-00 report.md,artifact-01 report.md") {
		t.Fatalf("expected normalized file summary in prompt, got %q", prompt)
	}
	if strings.Contains(prompt, "artifact-09 report.md") {
		t.Fatalf("expected prompt file summary to cap oversized lists, got %q", prompt)
	}
}
