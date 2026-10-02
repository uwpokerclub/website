import { ComponentType } from "react";
import { DashboardCardId } from "./dashboardLayout";
import { TermAtAGlanceCard } from "./components/TermAtAGlanceCard";
import { EngagementRetentionCard } from "./components/EngagementRetentionCard";
import { TrialConversionCard } from "./components/TrialConversionCard";
import { LeaderboardCard } from "./components/LeaderboardCard";
import { QuickActionsCard } from "./components/QuickActionsCard";
import { EventSpotlightCard } from "./components/EventSpotlightCard";
import { EventActivityCard } from "./components/EventActivityCard";
import { SignupTimelineCard } from "./components/SignupTimelineCard";

type DashboardCardComponentProps = {
  semesterId: string;
};

/**
 * Every card id maps to a real component. The type is a total Record rather than a
 * Partial, so adding an id to DashboardCardId without building its card is a compile
 * error instead of a silent "Coming soon." tile.
 */
export const DASHBOARD_CARDS: Record<DashboardCardId, ComponentType<DashboardCardComponentProps>> = {
  spotlight: EventSpotlightCard,
  quickActions: QuickActionsCard,
  termAtAGlance: TermAtAGlanceCard,
  signupTimeline: SignupTimelineCard,
  eventActivity: EventActivityCard,
  engagement: EngagementRetentionCard,
  leaderboard: LeaderboardCard,
  trialConversion: TrialConversionCard,
};
