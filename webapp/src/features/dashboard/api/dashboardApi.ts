import { apiClient } from "@/lib/apiClient";

export interface MembershipStats {
  total: number;
  paid: number;
  unpaid: number;
  discounted: number;
  executive: number;
  new: number;
  returning: number;
}

export interface MembershipsDashboardResponse {
  current: MembershipStats;
  comparison: {
    semester: { id: string; name: string };
    stats: MembershipStats;
    /**
     * The comparison term's membership count at the same elapsed point this term has
     * reached, or null when that term has no dated memberships. Null means
     * unknowable, never zero — see the backend's MembershipTotalAsOf.
     */
    totalAsOf: number | null;
  } | null;
}

export interface EngagementStats {
  players: number;
  medianEventsAttended: number;
  playedOnceCount: number;
  playedOnceShare: number;
  tenPlusCount: number;
}

export interface EngagementDashboardResponse {
  current: EngagementStats;
  comparison: {
    semester: { id: string; name: string };
    stats: EngagementStats;
  } | null;
}

export async function fetchMembershipsDashboard(semesterId: string): Promise<MembershipsDashboardResponse> {
  return apiClient<MembershipsDashboardResponse>(`v2/semesters/${semesterId}/dashboard/memberships`);
}

export async function fetchEngagementDashboard(semesterId: string): Promise<EngagementDashboardResponse> {
  return apiClient<EngagementDashboardResponse>(`v2/semesters/${semesterId}/dashboard/engagement`);
}

export interface SpotlightEvent {
  id: number;
  name: string;
  format: string;
  startDate: string;
  state: number;
  entries: number;
  rebuys: number;
}

export interface EventActivityStats {
  eventsRun: number;
  eventsScheduled: number;
  totalEntries: number;
  averageFieldSize: number;
}

export interface EventSeriesPoint {
  id: number;
  name: string;
  startDate: string;
  entries: number;
}

export interface EventActivityResponse {
  current: EventActivityStats & { series: EventSeriesPoint[] };
  comparison: { semester: { id: string; name: string }; averageFieldSize: number } | null;
}

export interface SignupPoint {
  date: string;
  admin: number;
  discord: number;
  /** Null, legacy, and unrecognized membership sources are grouped here by the API. */
  unknown: number;
}

export interface SignupsResponse {
  series: SignupPoint[];
  eventDates: string[];
  dataStartsAt: string | null;
  total: number;
}

export interface ConversionStats {
  players: number;
  paid: number;
  trialSpent: number;
  trialOpen: number;
  executive: number;
}

export interface ConversionResponse {
  current: ConversionStats;
  freeTrialLimit: number;
  comparison: { semester: { id: string; name: string }; stats: ConversionStats; freeTrialLimit: number } | null;
}

export interface RankingEntry {
  id: number;
  firstName: string;
  lastName: string;
  points: number;
  position: number;
}

export async function fetchSpotlight(semesterId: string): Promise<SpotlightEvent | null> {
  return apiClient<SpotlightEvent | null>(`v2/semesters/${semesterId}/dashboard/spotlight`);
}

export async function fetchEventActivity(semesterId: string): Promise<EventActivityResponse> {
  return apiClient<EventActivityResponse>(`v2/semesters/${semesterId}/dashboard/events`);
}

export async function fetchSignups(semesterId: string): Promise<SignupsResponse> {
  return apiClient<SignupsResponse>(`v2/semesters/${semesterId}/dashboard/signups`);
}

export async function fetchTrialConversion(semesterId: string): Promise<ConversionResponse> {
  return apiClient<ConversionResponse>(`v2/semesters/${semesterId}/dashboard/conversion`);
}

export async function fetchTopRankings(semesterId: string): Promise<RankingEntry[]> {
  // The rankings endpoint returns models.ListResponse — { data, total } — not a bare
  // array. Unwrap here so the card never has to know, matching RankingsPage.
  const response = await apiClient<{ data: RankingEntry[]; total: number }>(
    `v2/semesters/${semesterId}/rankings?limit=5`,
  );

  return response.data ?? [];
}
