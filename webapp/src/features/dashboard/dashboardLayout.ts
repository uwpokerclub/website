import { ROLES } from "@/types/roles";

export type DashboardCardId =
  | "spotlight"
  | "quickActions"
  | "termAtAGlance"
  | "signupTimeline"
  | "eventActivity"
  | "engagement"
  | "leaderboard";

export const CARD_TITLES: Record<DashboardCardId, string> = {
  spotlight: "Event Spotlight",
  quickActions: "Quick Actions",
  termAtAGlance: "Memberships",
  signupTimeline: "Signup Timeline",
  eventActivity: "Event Activity",
  engagement: "Engagement & Retention",
  leaderboard: "Leaderboard",
};

const OPS_LAYOUT: readonly DashboardCardId[] = [
  "spotlight",
  "quickActions",
  "eventActivity",
  "leaderboard",
  "termAtAGlance",
  "engagement",
  "signupTimeline",
] as const;

const RECORDS_LAYOUT: readonly DashboardCardId[] = [
  "termAtAGlance",
  "signupTimeline",
  "spotlight",
  "eventActivity",
  "engagement",
  "leaderboard",
  "quickActions",
] as const;

const LEADERSHIP_LAYOUT: readonly DashboardCardId[] = [
  "engagement",
  "eventActivity",
  "termAtAGlance",
  "spotlight",
  "signupTimeline",
  "leaderboard",
  "quickActions",
] as const;

const LAYOUTS_BY_ROLE: Record<string, readonly DashboardCardId[]> = {
  [ROLES.EXECUTIVE]: OPS_LAYOUT,
  [ROLES.TOURNAMENT_DIRECTOR]: OPS_LAYOUT,
  [ROLES.SECRETARY]: RECORDS_LAYOUT,
  [ROLES.TREASURER]: RECORDS_LAYOUT,
  [ROLES.VICE_PRESIDENT]: LEADERSHIP_LAYOUT,
  [ROLES.PRESIDENT]: LEADERSHIP_LAYOUT,
  [ROLES.WEBMASTER]: LEADERSHIP_LAYOUT,
};

export function resolveDashboardLayout(role: string | null | undefined): readonly DashboardCardId[] {
  if (!role) {
    return OPS_LAYOUT;
  }

  return LAYOUTS_BY_ROLE[role] ?? OPS_LAYOUT;
}
