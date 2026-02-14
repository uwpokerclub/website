import { sendAPIRequest } from "../../../lib/sendAPIRequest";
import { User } from "../../../types/user";
import { Membership } from "../../../types/membership";
import { APIErrorResponse } from "../../../types/error";
import type { CreateMemberFormData } from "../validation/registrationSchema";

/**
 * Result type for API operations that may fail
 */
export type ApiResult<T> = { success: true; data: T } | { success: false; error: string };

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

/**
 * Search for existing members by name, email, or student ID
 * @param query - Search query string
 * @returns Array of matching members or error
 */
export async function searchMembers(query: string): Promise<ApiResult<User[]>> {
  const trimmedQuery = query.trim();

  if (!trimmedQuery) {
    return { success: true, data: [] };
  }

  const paramType = getSearchParamType(trimmedQuery);
  const searchParam = encodeURIComponent(trimmedQuery);

  const { status, data } = await sendAPIRequest<{ data: User[]; total: number } | APIErrorResponse>(
    `v2/members?${paramType}=${searchParam}`,
  );

  if (status >= 200 && status < 300) {
    const listResponse = data as { data: User[]; total: number };
    return { success: true, data: listResponse?.data ?? [] };
  }

  const errorResponse = data as APIErrorResponse | undefined;
  return {
    success: false,
    error: errorResponse?.message ?? "Failed to search members",
  };
}

/**
 * Create a new member in the system
 * @param memberData - Member data from the form
 * @returns Created member or error
 */
export async function createMember(memberData: CreateMemberFormData): Promise<ApiResult<User>> {
  // Convert string ID to number for API (backend expects numeric ID)
  const payload = {
    id: parseInt(memberData.id, 10),
    firstName: memberData.firstName,
    lastName: memberData.lastName,
    email: memberData.email,
    faculty: memberData.faculty || "",
    questId: memberData.questId || "",
  };

  const { status, data } = await sendAPIRequest<User | APIErrorResponse>("v2/members", "POST", payload);

  if (status === 201) {
    return { success: true, data: data as User };
  }

  if (status === 409) {
    return {
      success: false,
      error: "A member with this student ID or email already exists",
    };
  }

  const errorResponse = data as APIErrorResponse | undefined;
  return {
    success: false,
    error: errorResponse?.message ?? "Failed to create member",
  };
}

/**
 * Create a membership for a member in a specific semester
 * @param semesterId - The semester to create membership in
 * @param memberId - The member's ID (student ID)
 * @param paid - Whether the membership is paid
 * @param discounted - Whether the membership is discounted
 * @returns Created membership or error
 */
export async function createMembership(
  semesterId: string,
  memberId: string,
  paid: boolean,
  discounted: boolean,
): Promise<ApiResult<Membership>> {
  // Validate business rule: cannot be discounted if not paid
  if (discounted && !paid) {
    return {
      success: false,
      error: "A membership cannot be discounted if it is not paid",
    };
  }

  // API uses 'userId' field for the member ID
  const payload = {
    userId: parseInt(memberId, 10),
    paid,
    discounted,
  };

  const { status, data } = await sendAPIRequest<Membership | APIErrorResponse>(
    `v2/semesters/${semesterId}/memberships`,
    "POST",
    payload,
  );

  if (status === 201) {
    return { success: true, data: data as Membership };
  }

  if (status === 409) {
    return {
      success: false,
      error: "This member already has a membership for the current semester",
    };
  }

  if (status === 400) {
    const errorResponse = data as APIErrorResponse | undefined;
    return {
      success: false,
      error: errorResponse?.message ?? "Invalid membership data",
    };
  }

  const errorResponse = data as APIErrorResponse | undefined;
  return {
    success: false,
    error: errorResponse?.message ?? "Failed to create membership",
  };
}

/**
 * Register a new member with membership in a single operation
 * This is a convenience function that creates a member and then their membership
 * @param memberData - Member data from the form
 * @param semesterId - The semester to create membership in
 * @param paid - Whether the membership is paid
 * @param discounted - Whether the membership is discounted
 * @returns Created membership or error
 */
export async function registerNewMemberWithMembership(
  memberData: CreateMemberFormData,
  semesterId: string,
  paid: boolean,
  discounted: boolean,
): Promise<ApiResult<{ member: User; membership: Membership }>> {
  // Step 1: Create the member
  const memberResult = await createMember(memberData);

  if (!memberResult.success) {
    return { success: false, error: memberResult.error };
  }

  // Step 2: Create the membership
  const membershipResult = await createMembership(semesterId, memberResult.data.id, paid, discounted);

  if (!membershipResult.success) {
    return { success: false, error: membershipResult.error };
  }

  return {
    success: true,
    data: {
      member: memberResult.data,
      membership: membershipResult.data,
    },
  };
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

/**
 * Update an existing member's information
 * @param memberId - The member's ID (student ID)
 * @param data - Updated member data
 * @returns Updated member or error
 */
export async function updateMember(memberId: string, data: UpdateMemberRequest): Promise<ApiResult<User>> {
  const { status, data: responseData } = await sendAPIRequest<User | APIErrorResponse>(
    `v2/members/${memberId}`,
    "PATCH",
    data as unknown as Record<string, unknown>,
  );

  if (status === 200) {
    return { success: true, data: responseData as User };
  }

  if (status === 404) {
    return { success: false, error: "Member not found" };
  }

  const errorResponse = responseData as APIErrorResponse | undefined;
  return {
    success: false,
    error: errorResponse?.message ?? "Failed to update member",
  };
}

/**
 * Update an existing membership
 * @param semesterId - The semester ID
 * @param membershipId - The membership ID
 * @param data - Updated membership data (paid, discounted)
 * @returns Updated membership or error
 */
export async function updateMembership(
  semesterId: string,
  membershipId: string,
  data: UpdateMembershipRequest,
): Promise<ApiResult<Membership>> {
  const { status, data: responseData } = await sendAPIRequest<Membership | APIErrorResponse>(
    `v2/semesters/${semesterId}/memberships/${membershipId}`,
    "PATCH",
    data as unknown as Record<string, unknown>,
  );

  if (status === 200) {
    return { success: true, data: responseData as Membership };
  }

  if (status === 404) {
    return { success: false, error: "Membership not found" };
  }

  if (status === 400) {
    const errorResponse = responseData as APIErrorResponse | undefined;
    return {
      success: false,
      error: errorResponse?.message ?? "Invalid membership data",
    };
  }

  const errorResponse = responseData as APIErrorResponse | undefined;
  return {
    success: false,
    error: errorResponse?.message ?? "Failed to update membership",
  };
}

/**
 * Delete an existing membership
 * @param semesterId - The semester ID
 * @param membershipId - The membership ID
 * @returns Success or error
 */
export async function deleteMembership(semesterId: string, membershipId: string): Promise<ApiResult<void>> {
  const { status, data } = await sendAPIRequest<void | APIErrorResponse>(
    `v2/semesters/${semesterId}/memberships/${membershipId}`,
    "DELETE",
  );

  if (status === 204) {
    return { success: true, data: undefined };
  }

  if (status === 404) {
    return { success: false, error: "Membership not found" };
  }

  const errorResponse = data as APIErrorResponse | undefined;
  return {
    success: false,
    error: errorResponse?.message ?? "Failed to delete membership",
  };
}
