import { apiClient } from "@/lib/apiClient";
import { LoginResponse, CreateLoginRequest, UpdateLoginRequest } from "../types";

export interface FetchLoginsParams {
  limit: number;
  offset: number;
  search?: string;
}

export async function fetchLogins(params: FetchLoginsParams): Promise<{ data: LoginResponse[]; total: number }> {
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

export async function updateLogin(username: string, data: UpdateLoginRequest): Promise<void> {
  return apiClient<void>(`v2/logins/${username}`, {
    method: "PATCH",
    body: data,
  });
}

export async function deleteLogin(username: string): Promise<void> {
  return apiClient<void>(`v2/logins/${username}`, { method: "DELETE" });
}
