import { ROLES } from "@/types/roles";

export type DashboardCardId =
  | "spotlight"
  | "quickActions"
  | "termAtAGlance"
  | "signupTimeline"
  | "eventActivity"
  | "engagement"
  | "leaderboard"
  | "trialConversion";

export const CARD_TITLES: Record<DashboardCardId, string> = {
  spotlight: "Event spotlight",
  quickActions: "Quick actions",
  termAtAGlance: "Memberships",
  signupTimeline: "Signup timeline",
  eventActivity: "Event activity",
  engagement: "Engagement & retention",
  leaderboard: "Leaderboard",
  trialConversion: "Trial conversion",
};

export type CardLane = "wide" | "rail";

// Cards carrying a chart or a distribution need the wide lane; everything else reads
// fine in the rail, and routing it there is what keeps the page gap-free at any width.
export const CARD_LANES: Record<DashboardCardId, CardLane> = {
  spotlight: "rail",
  quickActions: "rail",
  termAtAGlance: "rail",
  signupTimeline: "wide",
  eventActivity: "wide",
  engagement: "wide",
  leaderboard: "rail",
  trialConversion: "rail",
};

const OPS_LAYOUT: readonly DashboardCardId[] = [
  "spotlight",
  "quickActions",
  "eventActivity",
  "leaderboard",
  "termAtAGlance",
  "engagement",
  "signupTimeline",
  "trialConversion",
] as const;

const RECORDS_LAYOUT: readonly DashboardCardId[] = [
  "trialConversion",
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
  "trialConversion",
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
