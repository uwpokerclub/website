import { AuthContext } from "@/contexts";
import { Actions, APIError, Resources, SubResources, UserSession } from "@/interfaces/responses";
import { Role } from "@/types/roles";
import { ReactNode, useEffect, useState } from "react";

interface AuthProviderProps {
  children: ReactNode;
}

export default function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<UserSession | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  const fetchSession = async (abortSignal: AbortSignal) => {
    setLoading(true);

    try {
      const response = await fetch("/api/v2/session", {
        credentials: "include",
        signal: abortSignal,
      });

      if (response.ok) {
        const data: UserSession = await response.json();
        setUser(data);
      } else {
        const data: APIError = await response.json();
        setError(data.message);
      }
    } catch (err) {
      if (!(err instanceof DOMException && err.name === "AbortError")) {
        setError("Failed to fetch session. Please contact the webmaster.");
      }
    } finally {
      if (!abortSignal.aborted) {
        setLoading(false);
      }
    }
  };

  const login = async (username: string, password: string, cb: () => void) => {
    const abortController = new AbortController();
    setLoading(true);
    try {
      const response = await fetch("/api/v2/session", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ username, password }),
        credentials: "include",
        signal: abortController.signal,
      });

      if (response.ok) {
        await fetchSession(abortController.signal);

        cb();
      } else {
        const data: APIError = await response.json();
        setError(data.message);
      }
    } catch (err) {
      if (!(err instanceof DOMException && err.name === "AbortError")) {
        setError("Network error during login. Please contact the webmaster.");
      }
    } finally {
      if (!abortController.signal.aborted) {
        setLoading(false);
      }
    }
  };

  const logout = async (cb: () => void) => {
    setLoading(true);
    const abortController = new AbortController();
    try {
      const response = await fetch("/api/v2/session/logout", {
        method: "POST",
        credentials: "include",
        signal: abortController.signal,
      });

      if (response.ok) {
        setUser(null);
        setError("");
        cb();
      }
    } catch (err) {
      if (!(err instanceof DOMException && err.name === "AbortError")) {
        setError("Network error during logout. Please contact the webmaster.");
      }
    } finally {
      setLoading(false);
    }
  };

  const hasPermission = (action: Actions, resource: Resources, subResource?: SubResources) => {
    if (!user || !user.permissions) {
      return false;
    }

    if (subResource) {
      const resourcePerm = user.permissions[resource] as { [key: string]: { [key: string]: boolean } };
      return resourcePerm?.[subResource]?.[action] ?? false;
    }
    const resourcePerm = user.permissions[resource] as { [key: string]: boolean };
    return resourcePerm?.[action] ?? false;
  };

  const hasRoles = (roles: Role[]) => {
    if (!user || !user.role) {
      return false;
    }

    return roles.some((role) => user.role === role);
  };

  const value = {
    user,
    loading,
    error,
    login,
    logout,
    hasPermission,
    hasRoles,
  };

  useEffect(() => {
    const abortContoller = new AbortController();

    fetchSession(abortContoller.signal);

    return () => abortContoller.abort();
  }, []);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
