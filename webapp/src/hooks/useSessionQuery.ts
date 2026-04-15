import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, ApiError } from "@/lib/apiClient";
import { UserSession } from "@/interfaces/responses";

export const sessionKeys = {
  all: ["session"] as const,
  current: () => [...sessionKeys.all, "current"] as const,
};

async function fetchSession(): Promise<UserSession> {
  return apiClient<UserSession>("v2/session");
}

/**
 * Login uses a direct fetch instead of apiClient because the backend returns
 * 401 for invalid credentials, and apiClient's 401 interceptor would redirect
 * to the login page instead of showing the error message.
 */
async function loginRequest(credentials: { username: string; password: string }): Promise<void> {
  const res = await fetch(`${import.meta.env.VITE_API_URL}/v2/session`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(credentials),
  });

  if (!res.ok) {
    let message = "An unexpected error occurred";
    let type = "unknown";
    try {
      const errorData = await res.json();
      message = errorData.message || message;
      type = errorData.type || type;
    } catch {
      /* response may not be JSON */
    }
    throw new ApiError(res.status, type, message);
  }
}

async function logoutRequest(): Promise<void> {
  await apiClient<void>("v2/session/logout", { method: "POST" });
}

export function useSession() {
  return useQuery({
    queryKey: sessionKeys.current(),
    queryFn: fetchSession,
    retry: false,
  });
}

export function useLogin() {
  return useMutation({
    mutationFn: loginRequest,
  });
}

export function useLogout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: logoutRequest,
    onSuccess: () => {
      queryClient.setQueryData(sessionKeys.current(), null);
    },
  });
}
