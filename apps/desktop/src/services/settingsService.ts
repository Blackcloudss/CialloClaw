// settingsService centralizes desktop settings persistence.
import type { SettingsSnapshot } from "@cialloclaw/protocol";
import { loadStoredValue, saveStoredValue } from "../platform/storage";

// SETTINGS_KEY is the single storage key for the desktop snapshot.
const SETTINGS_KEY = "cialloclaw.settings";

// DesktopSettings keeps the local alias aligned with the protocol snapshot.
export type DesktopSettings = SettingsSnapshot;

// loadSettings returns the stored snapshot or a complete default snapshot.
export function loadSettings(): DesktopSettings {
  return (
    loadStoredValue<DesktopSettings>(SETTINGS_KEY) ?? {
      settings: {
        general: {
          language: "zh-CN",
          auto_launch: true,
          theme_mode: "follow_system",
          voice_notification_enabled: true,
          voice_type: "default_female",
          download: {
            workspace_path: "D:/CialloClawWorkspace",
            ask_before_save_each_file: true,
          },
        },
        floating_ball: {
          auto_snap: true,
          idle_translucent: true,
          position_mode: "draggable",
          size: "medium",
        },
        memory: {
          enabled: true,
          lifecycle: "30d",
          work_summary_interval: {
            unit: "day",
            value: 7,
          },
          profile_refresh_interval: {
            unit: "week",
            value: 2,
          },
        },
        task_automation: {
          inspect_on_startup: true,
          inspect_on_file_change: true,
          inspection_interval: {
            unit: "minute",
            value: 15,
          },
          task_sources: ["D:/workspace/todos"],
          remind_before_deadline: true,
          remind_when_stale: false,
        },
        data_log: {
          provider: "openai",
          budget_auto_downgrade: true,
          provider_api_key_configured: false,
        },
      },
    }
  );
}

// saveSettings persists the latest desktop settings snapshot.
export function saveSettings(settings: DesktopSettings) {
  saveStoredValue(SETTINGS_KEY, settings);
}
