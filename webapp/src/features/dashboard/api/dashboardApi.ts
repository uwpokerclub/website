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
