import { create } from "zustand";
import type { DashboardModuleRoute } from "@/features/dashboard/shared/dashboardRoutes";

type DashboardState = {
  hoveredModule: DashboardModuleRoute | null;
  setHoveredModule: (module: DashboardModuleRoute | null) => void;
};

export const useDashboardStore = create<DashboardState>((set) => ({
  hoveredModule: null,
  setHoveredModule: (hoveredModule) => set({ hoveredModule }),
}));
