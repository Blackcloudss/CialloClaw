import type { Task, TaskUpdatedNotification } from "@cialloclaw/protocol";

type TaskWithOptionalSessionId = Task & {
  session_id?: unknown;
};

type TaskUpdatedWithOptionalSessionId = TaskUpdatedNotification & {
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

/**
 * Returns the latest hidden conversation session acknowledged by the backend.
 * The frontend does not generate session ids locally anymore.
 */
export function getCurrentConversationSessionId() {
  return currentConversationSessionId ?? undefined;
}

/**
 * Records the backend-owned session carried by a formal task payload when
 * available, while remaining compatible with older payloads that do not expose
 * `task.session_id` yet.
 */
export function rememberConversationSessionFromTask(task: Task | null | undefined) {
  if (task == null) {
    return null;
  }

  return storeSession(task.task_id, (task as TaskWithOptionalSessionId).session_id);
}

/**
 * Records the backend-owned session when task.updated starts exposing it.
 * Current stable payloads may still omit `session_id`, so this remains a no-op
 * until the backend contract lands.
 */
export function rememberConversationSessionFromTaskUpdated(payload: TaskUpdatedNotification | null | undefined) {
  if (payload == null) {
    return null;
  }

  return storeSession(payload.task_id, (payload as TaskUpdatedWithOptionalSessionId).session_id);
}

export function getConversationSessionIdForTask(taskId: string | null | undefined) {
  const normalizedTaskId = normalizeTaskId(taskId);
  if (normalizedTaskId === null) {
    return undefined;
  }

  return taskSessionIds.get(normalizedTaskId);
}
