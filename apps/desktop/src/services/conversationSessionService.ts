import type { Task, TaskUpdatedNotification } from "@cialloclaw/protocol";

type BackendOwnedSessionCarrier = {
  task_id?: unknown;
  session_id?: unknown;
};

let currentConversationSessionId: string | null = null;
const taskSessionIds = new Map<string, string>();

function normalizeSessionId(value: unknown) {
  if (typeof value !== "string") {
    return null;
  }

  const trimmed = value.trim();
  return trimmed === "" ? null : trimmed;
}

function normalizeTaskId(value: unknown) {
  if (typeof value !== "string") {
    return null;
  }

  const trimmed = value.trim();
  return trimmed === "" ? null : trimmed;
}

function storeSession(taskId: unknown, sessionId: unknown) {
  const normalizedSessionId = normalizeSessionId(sessionId);
  if (normalizedSessionId === null) {
    return null;
  }

  currentConversationSessionId = normalizedSessionId;
  const normalizedTaskId = normalizeTaskId(taskId);
  if (normalizedTaskId !== null) {
    taskSessionIds.set(normalizedTaskId, normalizedSessionId);
  }
  return normalizedSessionId;
}

function rememberConversationSession(value: BackendOwnedSessionCarrier | null | undefined) {
  if (value == null) {
    return null;
  }

  return storeSession(value.task_id, value.session_id);
}

/**
 * Returns the latest hidden conversation session acknowledged by the backend.
 * The frontend does not generate session ids locally anymore.
 */
export function getCurrentConversationSessionId() {
  return currentConversationSessionId ?? undefined;
}

/**
 * Records the backend-owned session carried by a formal task payload.
 * The normalization path remains permissive so stale local services that still
 * omit `task.session_id` fail soft instead of breaking the desktop cache.
 */
export function rememberConversationSessionFromTask(task: Task | null | undefined) {
  return rememberConversationSession(task as BackendOwnedSessionCarrier | null | undefined);
}

/**
 * Records the backend-owned session carried by task.updated.
 * The notification contract now includes `session_id`, but permissive
 * normalization keeps older backend builds from breaking session recall.
 */
export function rememberConversationSessionFromTaskUpdated(payload: TaskUpdatedNotification | null | undefined) {
  return rememberConversationSession(payload as BackendOwnedSessionCarrier | null | undefined);
}

export function getConversationSessionIdForTask(taskId: string | null | undefined) {
  const normalizedTaskId = normalizeTaskId(taskId);
  if (normalizedTaskId === null) {
    return undefined;
  }

  return taskSessionIds.get(normalizedTaskId);
}
