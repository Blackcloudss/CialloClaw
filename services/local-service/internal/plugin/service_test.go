package plugin

import "testing"

func TestServiceRuntimeLifecycleAndSnapshots(t *testing.T) {
	service := NewService()
	if len(service.Workers()) != 3 || len(service.Sidecars()) != 1 {
		t.Fatalf("expected declared workers and sidecars, got workers=%+v sidecars=%+v", service.Workers(), service.Sidecars())
	}
	if !service.HasSidecar("playwright_sidecar") {
		t.Fatal("expected declared sidecar to be discoverable")
	}
	if _, ok := service.SidecarSpec("playwright_sidecar"); !ok {
		t.Fatal("expected sidecar spec to resolve")
	}
	service.MarkRuntimeStarting(RuntimeKindWorker, "ocr_worker")
	service.MarkRuntimeHealthy(RuntimeKindWorker, "ocr_worker")
	service.MarkRuntimeFailed(RuntimeKindSidecar, "playwright_sidecar", assertError("transport lost"))
	service.MarkRuntimeUnavailable(RuntimeKindWorker, "media_worker", "binary missing")
	service.MarkRuntimeStopped(RuntimeKindWorker, "media_worker")

	runtime, ok := service.RuntimeState(RuntimeKindWorker, "ocr_worker")
	if !ok || runtime.Status != RuntimeStatusRunning || runtime.Health != RuntimeHealthHealthy {
		t.Fatalf("expected runtime state to reflect healthy worker, got %+v ok=%v", runtime, ok)
	}
	failedRuntime, ok := service.RuntimeState(RuntimeKindSidecar, "playwright_sidecar")
	if !ok || failedRuntime.Health != RuntimeHealthFailed || failedRuntime.LastError == "" {
		t.Fatalf("expected sidecar failure state, got %+v ok=%v", failedRuntime, ok)
	}
	metrics := service.MetricSnapshots()
	if len(metrics) == 0 {
		t.Fatal("expected metric snapshots to be available")
	}
	events := service.RuntimeEvents()
	if len(events) < 4 {
		t.Fatalf("expected runtime events to be buffered, got %+v", events)
	}
}

func TestServiceEventPayloadsAreCloned(t *testing.T) {
	service := NewService()
	service.MarkRuntimeUnavailable(RuntimeKindWorker, "ocr_worker", "binary missing")
	events := service.RuntimeEvents()
	events[0].Payload["error"] = "mutated"
	freshEvents := service.RuntimeEvents()
	if freshEvents[0].Payload["error"] != "binary missing" {
		t.Fatalf("expected runtime events to return cloned payloads, got %+v", freshEvents)
	}
}

func assertError(message string) error { return testError(message) }

type testError string

func (e testError) Error() string { return string(e) }
