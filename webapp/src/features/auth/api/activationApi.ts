import { ApiError } from "@/lib/apiClient";

export interface ActivationMember {
  firstName: string;
  lastName: string;
}

async function activationRequest<T>(path: "verify" | "complete", body: Record<string, string>): Promise<T> {
  const response = await fetch(`${import.meta.env.VITE_API_URL}/v2/activations/${path}`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });

  if (!response.ok) {
    let message = "Unable to activate your account";
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

  return response.status === 204 || response.status === 201 ? (undefined as T) : response.json();
}

export function verifyActivation(token: string): Promise<ActivationMember> {
  return activationRequest<ActivationMember>("verify", { token });
}

export function completeActivation(token: string, password: string): Promise<void> {
  return activationRequest<void>("complete", { token, password });
}
