import { apiClient } from "@/lib/apiClient";
import type { ExecutiveTeamActivationLinkResponse, ExecutiveTeamResponse } from "../types";

export function fetchExecutiveTeam(semesterId: string): Promise<ExecutiveTeamResponse> {
  return apiClient<ExecutiveTeamResponse>(`v2/semesters/${semesterId}/executive-team`);
}

export function createExecutiveTeamActivationLink(
  semesterId: string,
  username: string,
): Promise<ExecutiveTeamActivationLinkResponse> {
  return apiClient<ExecutiveTeamActivationLinkResponse>(
    `v2/semesters/${semesterId}/executive-team/${encodeURIComponent(username)}/activation-link`,
    { method: "POST" },
  );
}
