import type {
  AgentNotepadConvertToTaskParams,
  AgentNotepadListParams,
  RequestMeta,
  TodoBucket,
} from "@cialloclaw/protocol";
import { convertNotepadToTask, listNotepad } from "@/rpc/methods";
import { getMockNoteBuckets, getMockNoteExperience, runMockConvertNoteToTask } from "./notePage.mock";
import type { NoteBucketsData, NoteConvertOutcome, NoteDetailExperience, NoteListItem } from "./notePage.types";

function createRequestMeta(scope: string): RequestMeta {
  return {
    client_time: new Date().toISOString(),
    trace_id: `trace_${scope}_${Date.now()}`,
  };
}

function createFallbackExperience(item: NoteListItem["item"]): NoteDetailExperience {
  return {
    agentSuggestion: {
      detail: item.agent_suggestion ?? "当前仅拿到协议里的基础数据，建议先补充说明后再转交给 Agent。",
      label: "下一步建议",
    },
    canConvertToTask: item.bucket !== "closed",
    detailStatus: item.bucket === "closed" ? "已结束" : "待处理",
    detailStatusTone: item.status === "overdue" ? "overdue" : item.status === "completed" || item.status === "cancelled" ? "done" : "normal",
    effectiveScope: null,
    endedAt: item.status === "completed" || item.status === "cancelled" ? item.due_at : null,
    isRecurringEnabled: item.bucket === "recurring_rule",
    nextOccurrenceAt: item.bucket === "recurring_rule" ? item.due_at : null,
    noteText: item.title,
    noteType: item.bucket === "recurring_rule" ? "recurring" : item.bucket === "closed" ? "archive" : "reminder",
    plannedAt: item.due_at,
    previewStatus: item.bucket === "closed" ? (item.status === "completed" ? "已完成" : "已取消") : item.status === "overdue" ? "已逾期" : item.status === "due_today" ? "今天要做" : item.bucket === "later" ? "未到时间" : item.bucket === "recurring_rule" ? "规则生效中" : "近期要做",
    prerequisite: null,
    recentInstanceStatus: null,
    relatedResources: [],
    repeatRule: item.bucket === "recurring_rule" ? "重复规则待补充" : null,
    summaryLabel: item.bucket,
    timeHint: item.due_at ? new Date(item.due_at).toLocaleString() : "未设置时间",
    title: item.title,
    typeLabel: item.type,
  };
}

function mapItems(items: Awaited<ReturnType<typeof listNotepad>>["items"]): NoteListItem[] {
  return items.map((item) => ({
    experience: getMockNoteExperience(item.item_id) ?? createFallbackExperience(item),
    item,
  }));
}

async function listNotepadByBucket(group: TodoBucket) {
  const params: AgentNotepadListParams = {
    group,
    limit: group === "closed" ? 24 : 12,
    offset: 0,
    request_meta: createRequestMeta(`notepad_${group}`),
  };

  return listNotepad(params);
}

export async function loadNoteBuckets(): Promise<NoteBucketsData> {
  try {
    const [upcomingResult, laterResult, recurringResult, closedResult] = await Promise.all([
      listNotepadByBucket("upcoming"),
      listNotepadByBucket("later"),
      listNotepadByBucket("recurring_rule"),
      listNotepadByBucket("closed"),
    ]);

    return {
      closed: mapItems(closedResult.items),
      later: mapItems(laterResult.items),
      recurring_rule: mapItems(recurringResult.items),
      source: "rpc",
      upcoming: mapItems(upcomingResult.items),
    };
  } catch (error) {
    console.warn("Notepad RPC unavailable, using local mock fallback.", error);
    return getMockNoteBuckets();
  }
}

export async function convertNoteToTask(itemId: string): Promise<NoteConvertOutcome> {
  const params: AgentNotepadConvertToTaskParams = {
    confirmed: true,
    item_id: itemId,
    request_meta: createRequestMeta(`notepad_convert_${itemId}`),
  };

  try {
    return {
      result: await convertNotepadToTask(params),
      source: "rpc",
    };
  } catch (error) {
    console.warn("Convert note to task RPC unavailable, using local mock fallback.", error);
    return runMockConvertNoteToTask(itemId);
  }
}
