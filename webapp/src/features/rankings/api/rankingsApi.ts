import { apiClient, ApiError } from "@/lib/apiClient";
import { Ranking } from "@/types";

export interface FetchRankingsParams {
  limit: number;
  offset: number;
  search?: string;
}

export async function fetchRankings(
  semesterId: string,
  params: FetchRankingsParams,
): Promise<{ data: Ranking[]; total: number }> {
  const query = new URLSearchParams({
    limit: String(params.limit),
    offset: String(params.offset),
  });
  if (params.search) query.set("search", params.search);

  const response = await apiClient<{ data: Ranking[]; total: number }>(
    `v2/semesters/${semesterId}/rankings?${query.toString()}`,
  );
  return { data: response.data ?? [], total: response.total ?? 0 };
}

// Direct fetch (not apiClient) because the response is a CSV blob, not JSON.
export async function exportRankings(semesterId: string): Promise<{ blob: Blob; filename: string }> {
  const res = await fetch(`${import.meta.env.VITE_API_URL}/v2/semesters/${semesterId}/rankings/export`, {
    credentials: "include",
  });

  if (!res.ok) {
    throw new ApiError(res.status, "export-failed", `Failed to export rankings: ${res.statusText}`);
  }

  const blob = await res.blob();
  const filename = res.headers.get("Content-Disposition")?.split("filename=")[1]?.replace(/"/g, "") || "rankings.csv";
  return { blob, filename };
}
