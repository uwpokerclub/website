import { apiClient } from "@/lib/apiClient";
import { LoginResponse, CreateLoginRequest, ChangePasswordRequest } from "../types";

export async function fetchLogins(params: {
  limit: number;
  offset: number;
  search?: string;
}): Promise<{ data: LoginResponse[]; total: number }> {
  let query = `?limit=${params.limit}&offset=${params.offset}`;
  if (params.search) {
    query += `&search=${encodeURIComponent(params.search)}`;
  }
  const response = await apiClient<{ data: LoginResponse[]; total: number }>(`v2/logins${query}`);
  return { data: response.data ?? [], total: response.total ?? 0 };
}

export async function createLogin(loginData: CreateLoginRequest): Promise<LoginResponse> {
  return apiClient<LoginResponse>("v2/logins", {
    method: "POST",
    body: loginData,
  });
}

export async function changePassword(username: string, passwordData: ChangePasswordRequest): Promise<void> {
  return apiClient<void>(`v2/logins/${username}/password`, {
    method: "PATCH",
    body: passwordData,
  });
}

export async function deleteLogin(username: string): Promise<void> {
  return apiClient<void>(`v2/logins/${username}`, { method: "DELETE" });
}
