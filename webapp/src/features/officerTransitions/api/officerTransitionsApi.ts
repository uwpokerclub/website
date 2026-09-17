import { ApiError, apiClient } from "@/lib/apiClient";
import type { User } from "@/types/user";
import type {
  OfficerRole,
  OfficerTransition,
  OfficerTransitionFormData,
  ResolvedQuestId,
  StageOfficerTransitionResponse,
} from "../types";

function normalizeQuestId(value: string): string {
  return value.trim().toLowerCase();
}

export async function resolveQuestId(value: string, signal?: AbortSignal): Promise<ResolvedQuestId> {
  const questId = normalizeQuestId(value);
  const params = new URLSearchParams({ questId, limit: "2" });
  const response = await fetch(`${import.meta.env.VITE_API_URL}/v2/members?${params}`, {
    credentials: "include",
    signal,
  });

  if (!response.ok) {
    let message = "Unable to resolve Quest ID";
    let type = "unknown";
    try {
      const error = await response.json();
      message = error.message || message;
      type = error.type || type;
    } catch {
      // The endpoint may return a non-JSON error response.
    }
    throw new ApiError(response.status, type, message);
  }

  const result = (await response.json()) as { data: User[]; total: number };
  const member = result.data[0];
  if (result.total !== 1 || !member || normalizeQuestId(member.questId) !== questId) {
    throw new Error("Quest ID must resolve to exactly one member");
  }

  return {
    questId,
    nominee: { firstName: member.firstName, lastName: member.lastName },
  };
}

export function stageOfficerTransition(data: OfficerTransitionFormData): Promise<StageOfficerTransitionResponse> {
  return apiClient<StageOfficerTransitionResponse>("v2/officer-transitions", {
    method: "POST",
    body: {
      presidentQuestId: normalizeQuestId(data.presidentQuestId),
      vicePresidentQuestId: normalizeQuestId(data.vicePresidentQuestId),
      secretaryQuestId: normalizeQuestId(data.secretaryQuestId),
      treasurerQuestId: normalizeQuestId(data.treasurerQuestId),
    },
  });
}

export function fetchCurrentOfficerTransition(): Promise<OfficerTransition> {
  return apiClient<OfficerTransition>("v2/officer-transitions/current");
}

export function cancelOfficerTransition(id: string): Promise<void> {
  return apiClient<void>(`v2/officer-transitions/${id}/cancel`, { method: "POST" });
}

export function reissueOfficerTransitionLink(id: string, role: OfficerRole): Promise<{ activationToken: string }> {
  return apiClient<{ activationToken: string }>(`v2/officer-transitions/${id}/reissue`, {
    method: "POST",
    body: { role },
  });
}

export { normalizeQuestId };
