import { useQuery } from "@tanstack/react-query";
import { fetchExecutiveTeam } from "../api/executiveTeamApi";

export const executiveTeamKeys = {
  all: ["executive-team"] as const,
  bySemester: (semesterId: string) => [...executiveTeamKeys.all, semesterId] as const,
};

export function useExecutiveTeam(semesterId: string | undefined, enabled: boolean) {
  return useQuery({
    queryKey: executiveTeamKeys.bySemester(semesterId ?? ""),
    queryFn: () => fetchExecutiveTeam(semesterId!),
    enabled: Boolean(semesterId && enabled),
    retry: false,
  });
}
