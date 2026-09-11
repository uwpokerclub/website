import { ComponentType } from "react";
import { DashboardCardId } from "./dashboardLayout";

type DashboardCardComponentProps = {
  semesterId: string;
};

/**
 * Maps a card id to its real implementation once one exists. Unregistered ids
 * fall back to a placeholder in Dashboard.tsx. Each card issue adds one entry
 * here and threads semesterId into its own React Query keys.
 */
export const DASHBOARD_CARDS: Partial<Record<DashboardCardId, ComponentType<DashboardCardComponentProps>>> = {};
