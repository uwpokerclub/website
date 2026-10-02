import { ComponentType } from "react";
import { DashboardCardId } from "./dashboardLayout";
import { TermAtAGlanceCard } from "./components/TermAtAGlanceCard";
import { EngagementRetentionCard } from "./components/EngagementRetentionCard";
import { TrialConversionCard } from "./components/TrialConversionCard";
import { LeaderboardCard } from "./components/LeaderboardCard";
import { QuickActionsCard } from "./components/QuickActionsCard";
import { EventSpotlightCard } from "./components/EventSpotlightCard";

type DashboardCardComponentProps = {
  semesterId: string;
};

/**
 * Maps a card id to its real implementation once one exists. Unregistered ids
 * fall back to a placeholder in Dashboard.tsx. Each card issue adds one entry
 * here and threads semesterId into its own React Query keys.
 */
export const DASHBOARD_CARDS: Partial<Record<DashboardCardId, ComponentType<DashboardCardComponentProps>>> = {
  termAtAGlance: TermAtAGlanceCard,
  engagement: EngagementRetentionCard,
  trialConversion: TrialConversionCard,
  leaderboard: LeaderboardCard,
  quickActions: QuickActionsCard,
  spotlight: EventSpotlightCard,
};
