import { useQuery, keepPreviousData } from "@tanstack/react-query";
import { fetchRankings, type FetchRankingsParams } from "../api/rankingsApi";

export const rankingKeys = {
  all: ["rankings"] as const,
  bySemester: (semesterId: string) => [...rankingKeys.all, semesterId] as const,
  list: (semesterId: string, params: FetchRankingsParams) => [...rankingKeys.bySemester(semesterId), params] as const,
};

export function useRankings(semesterId: string | undefined, params: FetchRankingsParams) {
  return useQuery({
    queryKey: rankingKeys.list(semesterId ?? "", params),
    queryFn: () => fetchRankings(semesterId!, params),
    enabled: !!semesterId,
    placeholderData: keepPreviousData,
  });
}
