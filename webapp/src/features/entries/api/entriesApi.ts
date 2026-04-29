import { apiClient } from "@/lib/apiClient";
import { Entry } from "@/types";

/**
 * Raw participant shape returned by the entries endpoint. Components typically
 * transform this into the flatter `Entry` shape via `participantToEntry`.
 */
export interface ParticipantResponse {
  membershipId: string;
  membership?: {
    id: string;
    user?: {
      id?: string;
      firstName?: string;
      lastName?: string;
    };
  };
  signedOutAt: Date;
  placement?: number;
  eventId: string;
}

export interface CreateEntryResult {
  membershipId: string;
  status: "created" | "error";
  error?: string;
}

export async function fetchEntries(
  semesterId: string,
  eventId: number,
  params: { limit: number; offset: number; search?: string },
): Promise<{ data: ParticipantResponse[]; total: number }> {
  let query = `?limit=${params.limit}&offset=${params.offset}`;
  if (params.search) {
    query += `&search=${encodeURIComponent(params.search)}`;
  }
  const response = await apiClient<{ data: ParticipantResponse[]; total: number }>(
    `v2/semesters/${semesterId}/events/${eventId}/entries${query}`,
  );
  return { data: response.data ?? [], total: response.total ?? 0 };
}

export async function registerEntries(
  semesterId: string,
  eventId: number,
  membershipIds: string[],
): Promise<CreateEntryResult[]> {
  return apiClient<CreateEntryResult[]>(`v2/semesters/${semesterId}/events/${eventId}/entries`, {
    method: "POST",
    body: membershipIds,
  });
}

export async function unregisterEntry(semesterId: string, eventId: number, membershipId: string): Promise<void> {
  return apiClient<void>(`v2/semesters/${semesterId}/events/${eventId}/entries/${membershipId}`, {
    method: "DELETE",
  });
}

export async function signInEntry(semesterId: string, eventId: number, membershipId: string): Promise<void> {
  return apiClient<void>(`v2/semesters/${semesterId}/events/${eventId}/entries/${membershipId}/sign-in`, {
    method: "POST",
  });
}

export async function signOutEntry(semesterId: string, eventId: number, membershipId: string): Promise<void> {
  return apiClient<void>(`v2/semesters/${semesterId}/events/${eventId}/entries/${membershipId}/sign-out`, {
    method: "POST",
  });
}

/**
 * Convert the raw participant shape into the flatter Entry shape used by
 * EntriesTable.
 */
export function participantToEntry(participant: ParticipantResponse): Entry {
  return {
    id: participant.membership?.user?.id ?? "",
    membershipId: participant.membershipId,
    eventId: participant.eventId,
    firstName: participant.membership?.user?.firstName ?? "",
    lastName: participant.membership?.user?.lastName ?? "",
    signedOutAt: participant.signedOutAt,
    placement: participant.placement,
  };
}
