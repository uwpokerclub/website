import { useQuery } from "@tanstack/react-query";
import { fetchEngagementDashboard, fetchMembershipsDashboard } from "../api/dashboardApi";

export const dashboardKeys = {
  all: ["dashboard"] as const,
  memberships: (semesterId: string) => [...dashboardKeys.all, "memberships", semesterId] as const,
  engagement: (semesterId: string) => [...dashboardKeys.all, "engagement", semesterId] as const,
};

export function useMembershipsDashboard(semesterId: string) {
  return useQuery({
    queryKey: dashboardKeys.memberships(semesterId),
    queryFn: () => fetchMembershipsDashboard(semesterId),
    enabled: !!semesterId,
  });
}

export function useEngagementDashboard(semesterId: string) {
  return useQuery({
    queryKey: dashboardKeys.engagement(semesterId),
    queryFn: () => fetchEngagementDashboard(semesterId),
    enabled: !!semesterId,
  });
}
