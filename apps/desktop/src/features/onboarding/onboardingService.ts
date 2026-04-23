import { Window, getCurrentWindow } from "@tauri-apps/api/window";
import { shellBallWindowLabels } from "@/platform/shellBallWindowController";
import { loadStoredValue, removeStoredValue, saveStoredValue } from "@/platform/storage";
import { desktopOnboardingEvents, desktopOnboardingLocalEvents } from "./onboarding.events";

export type DesktopOnboardingStep =
  | "welcome"
  | "shell_ball_intro"
  | "shell_ball_hold_voice"
  | "shell_ball_double_click"
  | "dashboard_overview"
  | "tray_hint"
  | "control_panel_api_key"
  | "done";

export type DesktopOnboardingSource = "first_launch" | "manual";

export type DesktopOnboardingStatus = {
  first_seen_at: string | null;
  completed: boolean;
  completed_at: string | null;
  skipped: boolean;
  skipped_at: string | null;
};

export type DesktopOnboardingSession = {
  isOpen: boolean;
  source: DesktopOnboardingSource;
  step: DesktopOnboardingStep;
  started_at: string;
};

const DESKTOP_ONBOARDING_STATUS_KEY = "cialloclaw.desktop.onboarding.status";
const DESKTOP_ONBOARDING_SESSION_KEY = "cialloclaw.desktop.onboarding.session";

const DESKTOP_ONBOARDING_WINDOW_LABELS = [
  shellBallWindowLabels.ball,
  "dashboard",
  "control-panel",
] as const;

function createDefaultDesktopOnboardingStatus(): DesktopOnboardingStatus {
  return {
    first_seen_at: null,
    completed: false,
    completed_at: null,
    skipped: false,
    skipped_at: null,
  };
}

function dispatchLocalSessionChanged(session: DesktopOnboardingSession | null) {
  window.dispatchEvent(
    new CustomEvent<DesktopOnboardingSession | null>(desktopOnboardingLocalEvents.sessionChanged, {
      detail: session,
    }),
  );
}

async function broadcastSession(session: DesktopOnboardingSession | null) {
  dispatchLocalSessionChanged(session);

  const currentWindowLabel = getCurrentWindow().label;
  await Promise.all(
    DESKTOP_ONBOARDING_WINDOW_LABELS.map(async (label) => {
      if (label === currentWindowLabel) {
        return;
      }

      try {
        const targetWindow = await Window.getByLabel(label);
        if (targetWindow === null) {
          return;
        }

        await targetWindow.emit(desktopOnboardingEvents.sessionChanged, session);
      } catch (error) {
        console.warn("desktop onboarding session sync failed", error);
      }
    }),
  );
}

export function loadDesktopOnboardingStatus(): DesktopOnboardingStatus {
  return {
    ...createDefaultDesktopOnboardingStatus(),
    ...(loadStoredValue<DesktopOnboardingStatus>(DESKTOP_ONBOARDING_STATUS_KEY) ?? {}),
  };
}

export function saveDesktopOnboardingStatus(status: DesktopOnboardingStatus) {
  saveStoredValue(DESKTOP_ONBOARDING_STATUS_KEY, status);
}

export function loadDesktopOnboardingSession() {
  return loadStoredValue<DesktopOnboardingSession>(DESKTOP_ONBOARDING_SESSION_KEY);
}

export function shouldAutoStartDesktopOnboarding() {
  const status = loadDesktopOnboardingStatus();
  return !status.completed && !status.skipped;
}

export async function setDesktopOnboardingSession(session: DesktopOnboardingSession | null) {
  if (session === null) {
    removeStoredValue(DESKTOP_ONBOARDING_SESSION_KEY);
  } else {
    saveStoredValue(DESKTOP_ONBOARDING_SESSION_KEY, session);
  }

  await broadcastSession(session);
}

export async function startDesktopOnboarding(source: DesktopOnboardingSource) {
  const now = new Date().toISOString();
  const status = loadDesktopOnboardingStatus();

  saveDesktopOnboardingStatus({
    ...status,
    first_seen_at: status.first_seen_at ?? now,
  });

  const session: DesktopOnboardingSession = {
    isOpen: true,
    source,
    step: "welcome",
    started_at: now,
  };

  await setDesktopOnboardingSession(session);
  return session;
}

export async function advanceDesktopOnboarding(step: DesktopOnboardingStep) {
  const currentSession = loadDesktopOnboardingSession();
  if (currentSession === null) {
    return null;
  }

  const nextSession: DesktopOnboardingSession = {
    ...currentSession,
    step,
  };

  await setDesktopOnboardingSession(nextSession);
  return nextSession;
}

export async function completeDesktopOnboarding() {
  const now = new Date().toISOString();
  const currentStatus = loadDesktopOnboardingStatus();

  saveDesktopOnboardingStatus({
    ...currentStatus,
    first_seen_at: currentStatus.first_seen_at ?? now,
    completed: true,
    completed_at: now,
  });

  await setDesktopOnboardingSession(null);
}

export async function skipDesktopOnboarding() {
  const now = new Date().toISOString();
  const currentStatus = loadDesktopOnboardingStatus();

  saveDesktopOnboardingStatus({
    ...currentStatus,
    first_seen_at: currentStatus.first_seen_at ?? now,
    skipped: true,
    skipped_at: now,
  });

  await setDesktopOnboardingSession(null);
}
