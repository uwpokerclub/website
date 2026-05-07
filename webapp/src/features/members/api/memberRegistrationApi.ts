import { apiClient } from "../../../lib/apiClient";
import { User } from "../../../types/user";
import { Membership } from "../../../types/membership";
import type { CreateMemberFormData } from "../validation/registrationSchema";

/**
 * Determines the search parameter type based on query format
 * - Contains '@' → email search
 * - All digits → id search
 * - Otherwise → name search
 */
function getSearchParamType(query: string): "email" | "id" | "name" {
  if (query.includes("@")) {
    return "email";
  }
  if (/^\d+$/.test(query)) {
    return "id";
  }
  return "name";
}

export async function searchMembers(query: string): Promise<User[]> {
  const trimmedQuery = query.trim();

  if (!trimmedQuery) {
    return [];
  }

  const paramType = getSearchParamType(trimmedQuery);
  const searchParam = encodeURIComponent(trimmedQuery);

  const response = await apiClient<{ data: User[]; total: number }>(`v2/members?${paramType}=${searchParam}`);
  return response.data ?? [];
}

export async function createMember(memberData: CreateMemberFormData): Promise<User> {
  const payload = {
    id: parseInt(memberData.id, 10),
    firstName: memberData.firstName,
    lastName: memberData.lastName,
    email: memberData.email,
    faculty: memberData.faculty || "",
    questId: memberData.questId || "",
  };

  return apiClient<User>("v2/members", { method: "POST", body: payload });
}

export async function createMembership(
  semesterId: string,
  memberId: string,
  paid: boolean,
  discounted: boolean,
): Promise<Membership> {
  // Validate business rule: cannot be discounted if not paid
  if (discounted && !paid) {
    throw new Error("A membership cannot be discounted if it is not paid");
  }

  const payload = {
    userId: parseInt(memberId, 10),
    paid,
    discounted,
  };

  return apiClient<Membership>(`v2/semesters/${semesterId}/memberships`, {
    method: "POST",
    body: payload,
  });
}

export async function registerNewMemberWithMembership(
  memberData: CreateMemberFormData,
  semesterId: string,
  paid: boolean,
  discounted: boolean,
): Promise<{ member: User; membership: Membership }> {
  const member = await createMember(memberData);
  const membership = await createMembership(semesterId, member.id, paid, discounted);

  return { member, membership };
}

export async function deleteMember(memberId: string): Promise<void> {
  return apiClient<void>(`v2/members/${memberId}`, { method: "DELETE" });
}

/**
 * Request type for updating a member
 */
export interface UpdateMemberRequest {
  firstName: string;
  lastName: string;
  email: string;
  faculty: string;
  questId: string;
}

/**
 * Request type for updating a membership
 */
export interface UpdateMembershipRequest {
  paid?: boolean;
  discounted?: boolean;
}

export async function updateMember(memberId: string, data: UpdateMemberRequest): Promise<User> {
  return apiClient<User>(`v2/members/${memberId}`, {
    method: "PATCH",
    body: data,
  });
}

export async function updateMembership(
  semesterId: string,
  membershipId: string,
  data: UpdateMembershipRequest,
): Promise<Membership> {
  return apiClient<Membership>(`v2/semesters/${semesterId}/memberships/${membershipId}`, {
    method: "PATCH",
    body: data,
  });
}

export async function deleteMembership(semesterId: string, membershipId: string): Promise<void> {
  return apiClient<void>(`v2/semesters/${semesterId}/memberships/${membershipId}`, { method: "DELETE" });
}

export async function fetchMemberships(
  semesterId: string,
  params: { limit: number; offset: number; search?: string },
): Promise<{ data: Membership[]; total: number }> {
  let query = `?limit=${params.limit}&offset=${params.offset}`;
  if (params.search) {
    query += `&search=${encodeURIComponent(params.search)}`;
  }
  const response = await apiClient<{ data: Membership[]; total: number }>(
    `v2/semesters/${semesterId}/memberships${query}`,
  );
  return { data: response.data ?? [], total: response.total ?? 0 };
}
