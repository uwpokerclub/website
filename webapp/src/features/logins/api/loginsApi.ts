import { sendAPIRequest } from "@/lib/sendAPIRequest";
import { APIErrorResponse } from "@/types/error";
import { LoginResponse, CreateLoginRequest, ChangePasswordRequest } from "../types";

/**
 * Result type for API operations that may fail
 */
export type ApiResult<T> = { success: true; data: T } | { success: false; error: string };

/**
 * Fetch all logins
 * @returns Array of logins or error
 */
export async function fetchLogins(): Promise<ApiResult<LoginResponse[]>> {
  const { status, data } = await sendAPIRequest<LoginResponse[] | APIErrorResponse>("v2/logins");

  if (status >= 200 && status < 300) {
    return { success: true, data: (data as LoginResponse[]) ?? [] };
  }

  const errorResponse = data as APIErrorResponse | undefined;
  return {
    success: false,
    error: errorResponse?.message ?? "Failed to fetch logins",
  };
}

/**
 * Create a new login
 * @param loginData - Login credentials and role
 * @returns Created login or error
 */
export async function createLogin(loginData: CreateLoginRequest): Promise<ApiResult<LoginResponse>> {
  const { status, data } = await sendAPIRequest<LoginResponse | APIErrorResponse>(
    "v2/logins",
    "POST",
    loginData as unknown as Record<string, unknown>,
  );

  if (status === 201) {
    return { success: true, data: data as LoginResponse };
  }

  if (status === 409) {
    return { success: false, error: "Username already exists" };
  }

  const errorResponse = data as APIErrorResponse | undefined;
  return {
    success: false,
    error: errorResponse?.message ?? "Failed to create login",
  };
}

/**
 * Change password for a login
 * @param username - Login username to update
 * @param passwordData - New password
 * @returns Success or error
 */
export async function changePassword(username: string, passwordData: ChangePasswordRequest): Promise<ApiResult<void>> {
  const { status, data } = await sendAPIRequest<void | APIErrorResponse>(
    `v2/logins/${username}/password`,
    "PATCH",
    passwordData as unknown as Record<string, unknown>,
  );

  if (status >= 200 && status < 300) {
    return { success: true, data: undefined };
  }

  if (status === 404) {
    return { success: false, error: "Login not found" };
  }

  const errorResponse = data as APIErrorResponse | undefined;
  return {
    success: false,
    error: errorResponse?.message ?? "Failed to change password",
  };
}

/**
 * Delete a login
 * @param username - Login username to delete
 * @returns Success or error
 */
export async function deleteLogin(username: string): Promise<ApiResult<void>> {
  const { status, data } = await sendAPIRequest<void | APIErrorResponse>(`v2/logins/${username}`, "DELETE");

  if (status === 204) {
    return { success: true, data: undefined };
  }

  if (status === 404) {
    return { success: false, error: "Login not found" };
  }

  const errorResponse = data as APIErrorResponse | undefined;
  return {
    success: false,
    error: errorResponse?.message ?? "Failed to delete login",
  };
}
