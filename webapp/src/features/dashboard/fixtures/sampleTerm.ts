import type {
  EngagementDashboardResponse,
  EventActivityResponse,
  MembershipsDashboardResponse,
  RankingEntry,
  SpotlightEvent,
} from "../api/dashboardApi";

/** Shared, internally consistent sample responses for dashboard component tests. */
const FALL_2025 = { id: "0f2d6a4e-0000-4000-8000-000000000001", name: "Fall 2025" };

export const SAMPLE_TERM = {
  spotlight: {
    id: 412,
    name: "Thursday Night NLH",
    format: "No-limit hold'em · 10k stacks",
    startDate: "2026-11-12T23:04:00Z",
    state: 0,
    entries: 47,
    rebuys: 12,
  } satisfies SpotlightEvent,

  memberships: {
    current: { total: 967, paid: 712, unpaid: 189, discounted: 34, executive: 32, new: 561, returning: 406 },
    comparison: {
      semester: FALL_2025,
      stats: { total: 891, paid: 654, unpaid: 182, discounted: 29, executive: 26, new: 498, returning: 393 },
      // Fall 2025 had 847 by this point last year, so 967 is genuinely ahead.
      totalAsOf: 847,
    },
  } satisfies MembershipsDashboardResponse,

  engagement: {
    current: { players: 824, medianEventsAttended: 2, playedOnceCount: 313, playedOnceShare: 0.38, tenPlusCount: 41 },
    comparison: {
      semester: FALL_2025,
      stats: { players: 891, medianEventsAttended: 3, playedOnceCount: 303, playedOnceShare: 0.34, tenPlusCount: 54 },
    },
  } satisfies EngagementDashboardResponse,

  eventActivity: {
    current: {
      eventsRun: 11,
      eventsScheduled: 3,
      totalEntries: 486,
      averageFieldSize: 44.2,
      series: [
        { id: 1, name: "Week 1", startDate: "2026-09-12T23:00:00Z", entries: 31 },
        { id: 2, name: "Week 2", startDate: "2026-09-19T23:00:00Z", entries: 52 },
        { id: 3, name: "Week 3", startDate: "2026-09-26T23:00:00Z", entries: 41 },
        { id: 4, name: "Week 4", startDate: "2026-10-03T23:00:00Z", entries: 63 },
        { id: 5, name: "Week 5", startDate: "2026-10-10T23:00:00Z", entries: 47 },
        { id: 6, name: "Week 6", startDate: "2026-10-17T23:00:00Z", entries: 68 },
        { id: 7, name: "Week 7", startDate: "2026-10-24T23:00:00Z", entries: 38 },
        { id: 8, name: "Week 8", startDate: "2026-10-31T23:00:00Z", entries: 56 },
        { id: 9, name: "Week 9", startDate: "2026-11-07T23:00:00Z", entries: 27 },
        { id: 10, name: "Week 10", startDate: "2026-11-14T23:00:00Z", entries: 59 },
        { id: 11, name: "Week 11", startDate: "2026-11-21T23:00:00Z", entries: 44 },
      ],
    },
    comparison: { semester: FALL_2025, averageFieldSize: 49.1 },
  } satisfies EventActivityResponse,

  rankings: [
    { id: 1, firstName: "Robin", lastName: "Chen", points: 412, position: 1 },
    { id: 2, firstName: "Asha", lastName: "Patel", points: 388, position: 2 },
    { id: 3, firstName: "Michael", lastName: "Osei", points: 355, position: 3 },
    { id: 4, firstName: "Jenny", lastName: "Lam", points: 340, position: 4 },
    { id: 5, firstName: "Sam", lastName: "Novak", points: 327, position: 5 },
  ] satisfies RankingEntry[],
};
