package storage

import (
	"context"
	"testing"
)

func TestServiceCurrentExecutionAssetsAndPluginResolution(t *testing.T) {
	service := NewService(nil)
	if err := service.EnsureBuiltinExecutionAssets(context.Background()); err != nil {
		t.Fatalf("ensure builtin execution assets: %v", err)
	}
	if err := service.PluginManifestStore().WritePluginManifest(context.Background(), PluginManifestRecord{
		PluginID:         "ocr",
		Name:             "OCR Worker",
		Version:          "builtin-v1",
		Entry:            "builtin://plugin/ocr",
		Source:           "builtin",
		Summary:          "OCR runtime manifest",
		CapabilitiesJSON: `["ocr_image","ocr_pdf"]`,
		PermissionsJSON:  `["artifact_read"]`,
		RuntimeNamesJSON: `["ocr_worker"]`,
	}); err != nil {
		t.Fatalf("write plugin manifest: %v", err)
	}

	refs, err := service.CurrentExecutionAssets(context.Background())
	if err != nil {
		t.Fatalf("current execution assets: %v", err)
	}
	if len(refs) != 3 {
		t.Fatalf("expected built-in skill/blueprint/prompt refs, got %+v", refs)
	}

	pluginRefs, err := service.PluginAssetsForCapabilities(context.Background(), []string{"ocr_image"})
	if err != nil {
		t.Fatalf("plugin assets for capabilities: %v", err)
	}
	if len(pluginRefs) != 1 || pluginRefs[0].AssetKind != ExtensionAssetKindPluginManifest || pluginRefs[0].AssetID != "ocr" {
		t.Fatalf("expected OCR plugin manifest ref, got %+v", pluginRefs)
	}
}

func TestExtensionAssetCatalogHandlesEmptyAndMalformedPluginData(t *testing.T) {
	service := NewService(nil)
	refs, err := service.CurrentExecutionAssets(context.Background())
	if err != nil {
		t.Fatalf("CurrentExecutionAssets returned error: %v", err)
	}
	if len(refs) != 0 {
		t.Fatalf("expected nil refs before built-in asset seeding, got %+v", refs)
	}
	if pluginRefs, err := service.PluginAssetsForCapabilities(context.Background(), []string{"   "}); err != nil || pluginRefs != nil {
		t.Fatalf("expected blank capability request to be ignored, refs=%+v err=%v", pluginRefs, err)
	}
	if err := service.PluginManifestStore().WritePluginManifest(context.Background(), PluginManifestRecord{
		PluginID:         "broken",
		Name:             "Broken Plugin",
		Version:          "v1",
		Entry:            "builtin://plugin/broken",
		Source:           "builtin",
		Summary:          "broken manifest",
		CapabilitiesJSON: `not-json`,
		PermissionsJSON:  `[]`,
		RuntimeNamesJSON: `[]`,
	}); err != nil {
		t.Fatalf("write malformed plugin manifest: %v", err)
	}
	if pluginRefs, err := service.PluginAssetsForCapabilities(context.Background(), []string{"ocr_image"}); err != nil || len(pluginRefs) != 0 {
		t.Fatalf("expected malformed plugin manifest capabilities to be ignored, refs=%+v err=%v", pluginRefs, err)
	}
}

func TestServiceCurrentExecutionAssetsSkipsNewerUnsupportedStoreRows(t *testing.T) {
	service := NewService(nil)
	ctx := context.Background()
	builtinSkill := builtinSkillManifestRecord("2026-04-22T10:00:00Z")
	builtinBlueprint := builtinBlueprintDefinitionRecord("2026-04-22T10:00:00Z")
	builtinPrompt := builtinPromptTemplateVersionRecord("2026-04-22T10:00:00Z")
	if err := service.SkillManifestStore().WriteSkillManifest(ctx, builtinSkill); err != nil {
		t.Fatalf("write builtin skill manifest: %v", err)
	}
	if err := service.BlueprintDefinitionStore().WriteBlueprintDefinition(ctx, builtinBlueprint); err != nil {
		t.Fatalf("write builtin blueprint definition: %v", err)
	}
	if err := service.PromptTemplateVersionStore().WritePromptTemplateVersion(ctx, builtinPrompt); err != nil {
		t.Fatalf("write builtin prompt template version: %v", err)
	}
	if err := service.SkillManifestStore().WriteSkillManifest(ctx, SkillManifestRecord{
		SkillManifestID: "skill_community_latest",
		Name:            "community_skill",
		Version:         "v2",
		Source:          extensionAssetSourceGitHub,
		Summary:         "community skill should stay outside the current execution boundary",
		ManifestJSON:    `{}`,
		CreatedAt:       "2026-04-22T11:00:00Z",
		UpdatedAt:       "2026-04-22T11:00:00Z",
	}); err != nil {
		t.Fatalf("write community skill manifest: %v", err)
	}

	refs, err := service.CurrentExecutionAssets(ctx)
	if err != nil {
		t.Fatalf("current execution assets: %v", err)
	}
	if len(refs) != 3 {
		t.Fatalf("expected fallback to supported builtin execution assets, got %+v", refs)
	}
	if refs[0].AssetKind != ExtensionAssetKindSkillManifest || refs[0].AssetID != builtinSkill.SkillManifestID {
		t.Fatalf("expected builtin skill manifest to remain the selected execution asset, got %+v", refs[0])
	}
}
