import { AuthContext } from "@/contexts";
import { Actions, Resources, SubResources } from "@/interfaces/responses";
import { useSession, useLogin, useLogout, sessionKeys } from "@/hooks/useSessionQuery";
import { useQueryClient } from "@tanstack/react-query";
import { Role } from "@/types/roles";
import { ReactNode, useCallback } from "react";

interface AuthProviderProps {
  children: ReactNode;
}

export default function AuthProvider({ children }: AuthProviderProps) {
  const queryClient = useQueryClient();
  const { data: user = null, isLoading, error: queryError } = useSession();
  const { mutateAsync: doLogin, error: loginError, reset: resetLogin } = useLogin();
  const { mutateAsync: doLogout, error: logoutError, reset: resetLogout } = useLogout();

  const login = useCallback(
    async (username: string, password: string, cb: () => void) => {
      resetLogout();
      try {
        await doLogin({ username, password });
        // Wait for session to be fetched before navigating so RequireAuth sees the user
        await queryClient.invalidateQueries({ queryKey: sessionKeys.current() });
        cb();
      } catch {
        // Error is captured by the mutation state and surfaced via the context error field
      }
    },
    [doLogin, resetLogout, queryClient],
  );

  const logout = useCallback(
    async (cb: () => void) => {
      resetLogin();
      try {
        await doLogout();
        cb();
      } catch {
        // Error is captured by the mutation state and surfaced via the context error field
      }
    },
    [doLogout, resetLogin],
  );

  const hasPermission = useCallback(
    (action: Actions, resource: Resources, subResource?: SubResources) => {
      if (!user || !user.permissions) {
        return false;
      }

      if (subResource) {
        const resourcePerm = user.permissions[resource] as { [key: string]: { [key: string]: boolean } };
        return resourcePerm?.[subResource]?.[action] ?? false;
      }
      const resourcePerm = user.permissions[resource] as { [key: string]: boolean };
      return resourcePerm?.[action] ?? false;
    },
    [user],
  );

  const hasRoles = useCallback(
    (roles: Role[]) => {
      if (!user || !user.role) {
        return false;
      }

      return roles.some((role) => user.role === role);
    },
    [user],
  );

  // Combine query and mutation errors — login errors take priority over stale logout/query errors
  const error = loginError?.message ?? queryError?.message ?? logoutError?.message ?? "";

  const value = {
    user,
    loading: isLoading,
    error,
    login,
    logout,
    hasPermission,
    hasRoles,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
