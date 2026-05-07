import { useQuery, useMutation, useQueryClient, keepPreviousData } from "@tanstack/react-query";
import { fetchLogins, createLogin, updateLogin, deleteLogin, type FetchLoginsParams } from "../api/loginsApi";
import type { CreateLoginRequest, UpdateLoginRequest } from "../types";

export const loginKeys = {
  all: ["logins"] as const,
  list: (params: FetchLoginsParams) => [...loginKeys.all, params] as const,
};

export function useLogins(params: FetchLoginsParams) {
  return useQuery({
    queryKey: loginKeys.list(params),
    queryFn: () => fetchLogins(params),
    placeholderData: keepPreviousData,
  });
}

export function useCreateLogin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateLoginRequest) => createLogin(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: loginKeys.all });
    },
  });
}

export function useUpdateLogin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ username, data }: { username: string; data: UpdateLoginRequest }) => updateLogin(username, data),
    onSuccess: () => {
      // Role changes affect the list view, so invalidate even when only the password changed.
      queryClient.invalidateQueries({ queryKey: loginKeys.all });
    },
  });
}

export function useDeleteLogin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (username: string) => deleteLogin(username),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: loginKeys.all });
    },
  });
}
