import { useQuery } from "@tanstack/react-query";
import { fetchMembershipsDashboard } from "../api/dashboardApi";

export const dashboardKeys = {
  all: ["dashboard"] as const,
  memberships: (semesterId: string) => [...dashboardKeys.all, "memberships", semesterId] as const,
};

export function useMembershipsDashboard(semesterId: string) {
  return useQuery({
    queryKey: dashboardKeys.memberships(semesterId),
    queryFn: () => fetchMembershipsDashboard(semesterId),
    enabled: !!semesterId,
  });
}
