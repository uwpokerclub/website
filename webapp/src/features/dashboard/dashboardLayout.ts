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

/** Width in columns of the dashboard's 12-column grid. */
export type ColumnSpan = 4 | 6 | 8 | 12;

export type CardSize = {
  /** The narrowest span the card's content still reads well at. */
  min: ColumnSpan;
  /** The span the card takes when its row has room. */
  preferred: ColumnSpan;
};

// Charts need the width to show a series; the engagement distribution and its range
// tracks read best wide but survive half a row. Trial conversion compares two terms
// side by side when it has two thirds of a row, and stacks tall without it. Everything
// else is a compact stat or a short list that a third of a row holds comfortably, and
// that widens when its row has room to spare.
export const CARD_SIZES: Record<DashboardCardId, CardSize> = {
  spotlight: { min: 4, preferred: 4 },
  quickActions: { min: 4, preferred: 4 },
  termAtAGlance: { min: 4, preferred: 4 },
  signupTimeline: { min: 8, preferred: 12 },
  eventActivity: { min: 8, preferred: 8 },
  engagement: { min: 6, preferred: 8 },
  leaderboard: { min: 4, preferred: 4 },
  trialConversion: { min: 4, preferred: 8 },
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
