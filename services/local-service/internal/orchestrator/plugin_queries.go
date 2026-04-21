package orchestrator

import (
	"errors"
	"strings"

	"github.com/cialloclaw/cialloclaw/services/local-service/internal/plugin"
)

type pluginRuntimeRef struct {
	Name string
	Kind plugin.RuntimeKind
}

type pluginCapabilitySummary struct {
	ToolName    string
	DisplayName string
	Description string
	Source      string
	RiskHint    string
}

type pluginContractField struct {
	Name        string
	Type        string
	Required    bool
	Description string
	Example     string
}

type pluginDataContract struct {
	SchemaRef  string
	SchemaJSON map[string]any
	Fields     []pluginContractField
}

type pluginDeliveryMapping struct {
	EmitsToolCall       bool
	ArtifactTypes       []string
	DeliveryTypes       []string
	CitationSourceTypes []string
}

type pluginToolContract struct {
	ToolName       string
	DisplayName    string
	Description    string
	Source         string
	RiskHint       string
	TimeoutSec     int
	SupportsDryRun bool
	InputContract  pluginDataContract
	OutputContract pluginDataContract
	DeliveryMap    pluginDeliveryMapping
}

type pluginCatalogEntry struct {
	PluginID     string
	Name         string
	DisplayName  string
	Summary      string
	Version      string
	Source       string
	Entry        string
	Enabled      bool
	Permissions  []string
	RuntimeRefs  []pluginRuntimeRef
	Capabilities []pluginCapabilitySummary
	Tools        []pluginToolContract
}

// PluginList returns the task-adjacent plugin catalog view documented by the
// protocol without exposing raw runtime caches as the primary list object.
func (s *Service) PluginList(params map[string]any) (map[string]any, error) {
	query := strings.TrimSpace(stringValue(params, "query", ""))
	pageParams := mapValue(params, "page")
	limit := clampListLimit(intValue(pageParams, "limit", 20))
	offset := clampListOffset(intValue(pageParams, "offset", 0))
	kinds, err := normalizePluginFilterValues(params["kinds"], validPluginRuntimeKind)
	if err != nil {
		return nil, err
	}
	health, err := normalizePluginFilterValues(params["health"], validPluginHealth)
	if err != nil {
		return nil, err
	}

	runtimeIndex := pluginRuntimeIndex(s.plugin)
	items := make([]map[string]any, 0)
	for _, entry := range builtinPluginCatalog() {
		runtimes := pluginRuntimesForEntry(entry, runtimeIndex)
		if !matchesPluginListQuery(entry, runtimes, query, kinds, health) {
			continue
		}
		items = append(items, pluginListItem(entry, runtimes))
	}

	total := len(items)
	if offset >= total {
		return map[string]any{
			"items": []map[string]any{},
			"page":  pageMap(limit, offset, total),
		}, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return map[string]any{
		"items": items[offset:end],
		"page":  pageMap(limit, offset, total),
	}, nil
}

// PluginDetailGet exposes one plugin-centric detail payload with optional
// runtime, metric, and event slices so later UI work does not need to infer the
// plugin catalog from worker declarations.
func (s *Service) PluginDetailGet(params map[string]any) (map[string]any, error) {
	pluginID := strings.TrimSpace(stringValue(params, "plugin_id", ""))
	if pluginID == "" {
		return nil, errors.New("plugin_id is required")
	}
	entry, ok := pluginCatalogEntryByID(pluginID)
	if !ok {
		return nil, errors.New("plugin_id is invalid")
	}

	includeRuntime := boolValue(params, "include_runtime", true)
	includeMetrics := boolValue(params, "include_metrics", true)
	includeEvents := boolValue(params, "include_events", true)
	runtimeIndex := pluginRuntimeIndex(s.plugin)
	metricIndex := pluginMetricIndex(s.plugin)
	events := pluginEventSlice(s.plugin)

	result := map[string]any{
		"plugin":        pluginManifestItem(entry),
		"runtimes":      []map[string]any{},
		"metrics":       []map[string]any{},
		"recent_events": []map[string]any{},
		"tools":         pluginToolItems(entry.Tools),
	}
	if includeRuntime {
		result["runtimes"] = pluginRuntimeItems(pluginRuntimesForEntry(entry, runtimeIndex))
	}
	if includeMetrics {
		result["metrics"] = pluginMetricItems(pluginMetricsForEntry(entry, metricIndex))
	}
	if includeEvents {
		result["recent_events"] = pluginEventItems(pluginEventsForEntry(entry, events))
	}
	return result, nil
}

func builtinPluginCatalog() []pluginCatalogEntry {
	return []pluginCatalogEntry{
		{
			PluginID:    "playwright",
			Name:        "playwright",
			DisplayName: "Playwright Browser Automation",
			Summary:     "Read, search, and interact with web pages through the controlled Playwright runtime.",
			Version:     "builtin-1",
			Source:      "builtin",
			Entry:       "worker://playwright_worker",
			Enabled:     true,
			Permissions: []string{"workspace:read", "web:read"},
			RuntimeRefs: []pluginRuntimeRef{
				{Name: "playwright_worker", Kind: plugin.RuntimeKindWorker},
				{Name: "playwright_sidecar", Kind: plugin.RuntimeKindSidecar},
			},
			Capabilities: []pluginCapabilitySummary{
				{ToolName: "page_read", DisplayName: "页面读取", Description: "Read visible page text and page metadata.", Source: "worker", RiskHint: "green"},
				{ToolName: "page_search", DisplayName: "页面搜索", Description: "Search the current page for relevant strings.", Source: "worker", RiskHint: "green"},
				{ToolName: "page_interact", DisplayName: "页面交互", Description: "Execute controlled browser interactions.", Source: "worker", RiskHint: "yellow"},
				{ToolName: "structured_dom", DisplayName: "结构化 DOM", Description: "Capture structured DOM snapshots for downstream reasoning.", Source: "worker", RiskHint: "green"},
			},
			Tools: []pluginToolContract{
				pluginToolContract{
					ToolName:       "page_read",
					DisplayName:    "页面读取",
					Description:    "Read visible page text and metadata from the active browser page.",
					Source:         "worker",
					RiskHint:       "green",
					TimeoutSec:     30,
					SupportsDryRun: false,
					InputContract: pluginDataContract{
						SchemaRef: "tools/page_read/input",
						Fields: []pluginContractField{
							{Name: "url", Type: "string", Required: true, Description: "Target page URL.", Example: "https://example.com"},
						},
					},
					OutputContract: pluginDataContract{
						SchemaRef: "tools/page_read/output",
						Fields: []pluginContractField{
							{Name: "title", Type: "string", Required: true, Description: "Resolved page title."},
							{Name: "text_content", Type: "string", Required: true, Description: "Visible page text."},
						},
					},
					DeliveryMap: pluginDeliveryMapping{EmitsToolCall: true, ArtifactTypes: []string{}, DeliveryTypes: []string{"task_detail"}, CitationSourceTypes: []string{"web"}},
				},
			},
		},
		{
			PluginID:    "ocr",
			Name:        "ocr",
			DisplayName: "OCR Worker",
			Summary:     "Extract text from files, images and PDFs.",
			Version:     "builtin-1",
			Source:      "builtin",
			Entry:       "worker://ocr_worker",
			Enabled:     true,
			Permissions: []string{"workspace:read"},
			RuntimeRefs: []pluginRuntimeRef{
				{Name: "ocr_worker", Kind: plugin.RuntimeKindWorker},
			},
			Capabilities: []pluginCapabilitySummary{
				{ToolName: "extract_text", DisplayName: "文本提取", Description: "Extract body text from files, images, and PDFs.", Source: "worker", RiskHint: "green"},
				{ToolName: "ocr_image", DisplayName: "图片 OCR", Description: "Run OCR against one image.", Source: "worker", RiskHint: "green"},
				{ToolName: "ocr_pdf", DisplayName: "PDF OCR", Description: "Extract text or OCR one PDF.", Source: "worker", RiskHint: "green"},
			},
			Tools: []pluginToolContract{
				{
					ToolName:       "ocr_image",
					DisplayName:    "图片 OCR",
					Description:    "通过 OCR worker 对图片执行文字识别",
					Source:         "worker",
					RiskHint:       "green",
					TimeoutSec:     30,
					SupportsDryRun: false,
					InputContract: pluginDataContract{
						SchemaRef: "tools/ocr_image/input",
						Fields: []pluginContractField{
							{Name: "path", Type: "string", Required: true, Description: "待识别图片路径", Example: "D:/workspace/invoice.png"},
							{Name: "language", Type: "string", Required: false, Description: "OCR 语言提示", Example: "zh-CN"},
						},
					},
					OutputContract: pluginDataContract{
						SchemaRef: "tools/ocr_image/output",
						Fields: []pluginContractField{
							{Name: "path", Type: "string", Required: true, Description: "原始输入路径"},
							{Name: "text", Type: "string", Required: true, Description: "识别后的正文文本"},
							{Name: "language", Type: "string", Required: false, Description: "识别语言"},
							{Name: "page_count", Type: "integer", Required: true, Description: "页数"},
							{Name: "source", Type: "string", Required: true, Description: "来源运行时，默认 ocr_worker"},
						},
					},
					DeliveryMap: pluginDeliveryMapping{EmitsToolCall: true, ArtifactTypes: []string{}, DeliveryTypes: []string{"task_detail"}, CitationSourceTypes: []string{}},
				},
			},
		},
		{
			PluginID:    "media",
			Name:        "media",
			DisplayName: "Media Worker",
			Summary:     "Normalize recordings, transcode media, and extract representative frames.",
			Version:     "builtin-1",
			Source:      "builtin",
			Entry:       "worker://media_worker",
			Enabled:     true,
			Permissions: []string{"workspace:read", "workspace:write"},
			RuntimeRefs: []pluginRuntimeRef{
				{Name: "media_worker", Kind: plugin.RuntimeKindWorker},
			},
			Capabilities: []pluginCapabilitySummary{
				{ToolName: "transcode_media", DisplayName: "媒体转码", Description: "Transcode media into a normalized output format.", Source: "worker", RiskHint: "yellow"},
				{ToolName: "normalize_recording", DisplayName: "录音归一化", Description: "Normalize recorded audio for downstream use.", Source: "worker", RiskHint: "green"},
				{ToolName: "extract_frames", DisplayName: "抽帧", Description: "Extract representative frames from one media file.", Source: "worker", RiskHint: "green"},
			},
			Tools: []pluginToolContract{
				{
					ToolName:       "extract_frames",
					DisplayName:    "抽帧",
					Description:    "Extract representative frames from one media file.",
					Source:         "worker",
					RiskHint:       "green",
					TimeoutSec:     90,
					SupportsDryRun: false,
					InputContract: pluginDataContract{
						SchemaRef: "tools/extract_frames/input",
						Fields: []pluginContractField{
							{Name: "path", Type: "string", Required: true, Description: "Source media path.", Example: "D:/workspace/demo.mp4"},
						},
					},
					OutputContract: pluginDataContract{
						SchemaRef: "tools/extract_frames/output",
						Fields: []pluginContractField{
							{Name: "output_paths", Type: "string[]", Required: true, Description: "Extracted frame file paths."},
							{Name: "source", Type: "string", Required: true, Description: "Source runtime identifier."},
						},
					},
					DeliveryMap: pluginDeliveryMapping{EmitsToolCall: true, ArtifactTypes: []string{"image"}, DeliveryTypes: []string{"task_detail"}, CitationSourceTypes: []string{"file"}},
				},
			},
		},
	}
}

func pluginCatalogEntryByID(pluginID string) (pluginCatalogEntry, bool) {
	needle := strings.TrimSpace(pluginID)
	for _, entry := range builtinPluginCatalog() {
		if entry.PluginID == needle {
			return entry, true
		}
	}
	return pluginCatalogEntry{}, false
}

func pluginRuntimeIndex(service *plugin.Service) map[string]plugin.RuntimeState {
	if service == nil {
		return map[string]plugin.RuntimeState{}
	}
	index := make(map[string]plugin.RuntimeState)
	for _, item := range service.RuntimeStates() {
		index[pluginRefKey(item.Kind, item.Name)] = item
	}
	return index
}

func pluginMetricIndex(service *plugin.Service) map[string]plugin.MetricSnapshot {
	if service == nil {
		return map[string]plugin.MetricSnapshot{}
	}
	index := make(map[string]plugin.MetricSnapshot)
	for _, item := range service.MetricSnapshots() {
		index[pluginRefKey(item.Kind, item.Name)] = item
	}
	return index
}

func pluginEventSlice(service *plugin.Service) []plugin.RuntimeEvent {
	if service == nil {
		return nil
	}
	return service.RuntimeEvents()
}

func pluginRuntimesForEntry(entry pluginCatalogEntry, runtimeIndex map[string]plugin.RuntimeState) []plugin.RuntimeState {
	result := make([]plugin.RuntimeState, 0, len(entry.RuntimeRefs))
	for _, ref := range entry.RuntimeRefs {
		if runtime, ok := runtimeIndex[pluginRefKey(ref.Kind, ref.Name)]; ok {
			result = append(result, runtime)
		}
	}
	return result
}

func pluginMetricsForEntry(entry pluginCatalogEntry, metricIndex map[string]plugin.MetricSnapshot) []plugin.MetricSnapshot {
	result := make([]plugin.MetricSnapshot, 0, len(entry.RuntimeRefs))
	for _, ref := range entry.RuntimeRefs {
		if metric, ok := metricIndex[pluginRefKey(ref.Kind, ref.Name)]; ok {
			result = append(result, metric)
		}
	}
	return result
}

func pluginEventsForEntry(entry pluginCatalogEntry, events []plugin.RuntimeEvent) []plugin.RuntimeEvent {
	if len(events) == 0 {
		return nil
	}
	allowed := make(map[string]struct{}, len(entry.RuntimeRefs))
	for _, ref := range entry.RuntimeRefs {
		allowed[pluginRefKey(ref.Kind, ref.Name)] = struct{}{}
	}
	result := make([]plugin.RuntimeEvent, 0, len(events))
	for _, event := range events {
		if _, ok := allowed[pluginRefKey(event.Kind, event.Name)]; ok {
			result = append(result, event)
		}
	}
	return result
}

func matchesPluginListQuery(entry pluginCatalogEntry, runtimes []plugin.RuntimeState, query string, kinds []string, health []string) bool {
	if query != "" {
		haystack := strings.ToLower(strings.Join([]string{entry.PluginID, entry.Name, entry.DisplayName, entry.Summary}, " "))
		if !strings.Contains(haystack, strings.ToLower(query)) {
			return false
		}
	}
	if len(kinds) > 0 {
		foundKind := false
		for _, runtime := range runtimes {
			if containsPluginFilterValue(kinds, string(runtime.Kind)) {
				foundKind = true
				break
			}
		}
		if !foundKind {
			return false
		}
	}
	if len(health) > 0 {
		foundHealth := false
		for _, runtime := range runtimes {
			if containsPluginFilterValue(health, string(runtime.Health)) {
				foundHealth = true
				break
			}
		}
		if !foundHealth {
			return false
		}
	}
	return true
}

func normalizePluginFilterValues(value any, validate func(string) bool) ([]string, error) {
	if value == nil {
		return nil, nil
	}
	items, ok := normalizeStringSlice(value)
	if !ok {
		return nil, errors.New("plugin filter values must be string arrays")
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		normalized := strings.TrimSpace(item)
		if normalized == "" || !validate(normalized) {
			return nil, errors.New("plugin filter values contain unsupported entries")
		}
		result = append(result, normalized)
	}
	return result, nil
}

func validPluginRuntimeKind(value string) bool {
	switch value {
	case string(plugin.RuntimeKindWorker), string(plugin.RuntimeKindSidecar):
		return true
	default:
		return false
	}
}

func validPluginHealth(value string) bool {
	switch value {
	case string(plugin.RuntimeHealthUnknown), string(plugin.RuntimeHealthHealthy), string(plugin.RuntimeHealthDegraded), string(plugin.RuntimeHealthFailed), string(plugin.RuntimeHealthStopped), string(plugin.RuntimeHealthUnavailable):
		return true
	default:
		return false
	}
}

func containsPluginFilterValue(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func pluginRefKey(kind plugin.RuntimeKind, name string) string {
	return string(kind) + "::" + strings.TrimSpace(name)
}

func pluginListItem(entry pluginCatalogEntry, runtimes []plugin.RuntimeState) map[string]any {
	result := pluginManifestItem(entry)
	result["runtimes"] = pluginRuntimeItems(runtimes)
	return result
}

func pluginManifestItem(entry pluginCatalogEntry) map[string]any {
	return map[string]any{
		"plugin_id":    entry.PluginID,
		"name":         entry.Name,
		"display_name": entry.DisplayName,
		"summary":      entry.Summary,
		"version":      entry.Version,
		"source":       entry.Source,
		"entry":        entry.Entry,
		"enabled":      entry.Enabled,
		"permissions":  append([]string(nil), entry.Permissions...),
		"capabilities": pluginCapabilityItems(entry.Capabilities),
	}
}

func pluginCapabilityItems(items []pluginCapabilitySummary) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{
			"tool_name":    item.ToolName,
			"display_name": item.DisplayName,
			"description":  item.Description,
			"source":       item.Source,
			"risk_hint":    item.RiskHint,
		})
	}
	return result
}

func pluginToolItems(items []pluginToolContract) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{
			"tool_name":        item.ToolName,
			"display_name":     item.DisplayName,
			"description":      item.Description,
			"source":           item.Source,
			"risk_hint":        item.RiskHint,
			"timeout_sec":      item.TimeoutSec,
			"supports_dry_run": item.SupportsDryRun,
			"input_contract":   pluginDataContractItem(item.InputContract),
			"output_contract":  pluginDataContractItem(item.OutputContract),
			"delivery_mapping": pluginDeliveryMappingItem(item.DeliveryMap),
		})
	}
	return result
}

func pluginDataContractItem(contract pluginDataContract) map[string]any {
	return map[string]any{
		"schema_ref":  contract.SchemaRef,
		"schema_json": cloneMap(contract.SchemaJSON),
		"fields":      pluginContractFieldItems(contract.Fields),
	}
}

func pluginContractFieldItems(items []pluginContractField) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		field := map[string]any{
			"name":        item.Name,
			"type":        item.Type,
			"required":    item.Required,
			"description": item.Description,
		}
		if strings.TrimSpace(item.Example) != "" {
			field["example"] = item.Example
		}
		result = append(result, field)
	}
	return result
}

func pluginDeliveryMappingItem(mapping pluginDeliveryMapping) map[string]any {
	return map[string]any{
		"emits_tool_call":       mapping.EmitsToolCall,
		"artifact_types":        append([]string(nil), mapping.ArtifactTypes...),
		"delivery_types":        append([]string(nil), mapping.DeliveryTypes...),
		"citation_source_types": append([]string(nil), mapping.CitationSourceTypes...),
	}
}
