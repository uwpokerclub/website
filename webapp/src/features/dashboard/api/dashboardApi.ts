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

export async function fetchMembershipsDashboard(semesterId: string): Promise<MembershipsDashboardResponse> {
  return apiClient<MembershipsDashboardResponse>(`v2/semesters/${semesterId}/dashboard/memberships`);
}
