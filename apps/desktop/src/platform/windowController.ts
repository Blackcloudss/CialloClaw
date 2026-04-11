import { Window } from "@tauri-apps/api/window";

export type DesktopWindowLabel = "dashboard" | "control-panel";

// 该文件封装桌面窗口控制能力。
export function focusWindow(label: DesktopWindowLabel) {
  return openOrFocusDesktopWindow(label);
}

export async function openOrFocusDesktopWindow(label: DesktopWindowLabel) {
  const windowHandle = await Window.getByLabel(label);

  if (windowHandle === null) {
    throw new Error(`Desktop window not found: ${label}`);
  }

  await windowHandle.show();
  await windowHandle.setFocus();

  return label;
}

// openWindow 处理当前模块的相关逻辑。
export function openWindow(label: string) {
  if (typeof window !== "undefined") {
    window.location.assign(`./${label}.html`);
  }

  return Promise.resolve(label);
}
