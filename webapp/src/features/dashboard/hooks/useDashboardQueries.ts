import { useQuery } from "@tanstack/react-query";
import {
  fetchEngagementDashboard,
  fetchEventActivity,
  fetchMembershipsDashboard,
  fetchSignups,
  fetchSpotlight,
  fetchTopRankings,
  fetchTrialConversion,
} from "../api/dashboardApi";

export const dashboardKeys = {
  all: ["dashboard"] as const,
  memberships: (semesterId: string) => [...dashboardKeys.all, "memberships", semesterId] as const,
  engagement: (semesterId: string) => [...dashboardKeys.all, "engagement", semesterId] as const,
  spotlight: (semesterId: string) => [...dashboardKeys.all, "spotlight", semesterId] as const,
  eventActivity: (semesterId: string) => [...dashboardKeys.all, "eventActivity", semesterId] as const,
  signups: (semesterId: string) => [...dashboardKeys.all, "signups", semesterId] as const,
  conversion: (semesterId: string) => [...dashboardKeys.all, "conversion", semesterId] as const,
  rankings: (semesterId: string) => [...dashboardKeys.all, "rankings", semesterId] as const,
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

export function useSpotlight(semesterId: string) {
  return useQuery({
    queryKey: dashboardKeys.spotlight(semesterId),
    queryFn: () => fetchSpotlight(semesterId),
    enabled: !!semesterId,
  });
}

export function useEventActivity(semesterId: string) {
  return useQuery({
    queryKey: dashboardKeys.eventActivity(semesterId),
    queryFn: () => fetchEventActivity(semesterId),
    enabled: !!semesterId,
  });
}

export function useSignups(semesterId: string) {
  return useQuery({
    queryKey: dashboardKeys.signups(semesterId),
    queryFn: () => fetchSignups(semesterId),
    enabled: !!semesterId,
  });
}

export function useTrialConversion(semesterId: string) {
  return useQuery({
    queryKey: dashboardKeys.conversion(semesterId),
    queryFn: () => fetchTrialConversion(semesterId),
    enabled: !!semesterId,
  });
}

export function useTopRankings(semesterId: string) {
  return useQuery({
    queryKey: dashboardKeys.rankings(semesterId),
    queryFn: () => fetchTopRankings(semesterId),
    enabled: !!semesterId,
  });
}
