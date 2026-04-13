export class ApiError extends Error {
  constructor(
    public status: number,
    public type: string,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

interface RequestOptions {
  method?: string;
  body?: Record<string, unknown>;
}

export async function apiClient<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = "GET", body } = options;

  const res = await fetch(`${import.meta.env.VITE_API_URL}/${path}`, {
    credentials: "include",
    method,
    headers: body ? { "Content-Type": "application/json" } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (res.status === 401) {
    window.location.href = "/admin/login";
    throw new ApiError(401, "unauthorized", "Session expired");
  }

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

  if (res.status === 204) {
    return undefined as T;
  }

  return res.json();
}
