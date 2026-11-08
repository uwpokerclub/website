import { apiClient } from "@/lib/apiClient";
import type { ExecutiveTeamActivationLinkResponse, ExecutiveTeamResponse } from "../types";

export type ExecutiveTeamRole = "executive" | "tournament_director";

export interface AddExecutiveTeamMemberRequest {
  questId: string;
  role: ExecutiveTeamRole;
}

export interface AddExecutiveTeamMemberResponse {
  memberId: number;
  username: string;
  role: ExecutiveTeamRole;
  accountStatus: "active" | "pending_activation";
  activationToken?: string;
}

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

export function addExecutiveTeamMember(
  semesterId: string,
  data: AddExecutiveTeamMemberRequest,
): Promise<AddExecutiveTeamMemberResponse> {
  return apiClient<AddExecutiveTeamMemberResponse>(`v2/semesters/${semesterId}/executive-team`, {
    method: "POST",
    body: data,
  });
}
