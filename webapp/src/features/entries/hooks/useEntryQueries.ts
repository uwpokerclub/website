import { useQuery, useMutation, useQueryClient, keepPreviousData } from "@tanstack/react-query";
import { fetchEntries, unregisterEntry, signInEntry, signOutEntry } from "../api/entriesApi";

export const entryKeys = {
  all: ["entries"] as const,
  byEvent: (semesterId: string, eventId: number) => [...entryKeys.all, semesterId, eventId] as const,
  list: (semesterId: string, eventId: number, params: { limit: number; offset: number; search?: string }) =>
    [...entryKeys.byEvent(semesterId, eventId), params] as const,
};

export function useEntries(
  semesterId: string | undefined,
  eventId: number | undefined,
  params: { limit: number; offset: number; search?: string },
) {
  return useQuery({
    queryKey: entryKeys.list(semesterId ?? "", eventId ?? 0, params),
    queryFn: () => fetchEntries(semesterId!, eventId!, params),
    enabled: !!semesterId && eventId !== undefined,
    placeholderData: keepPreviousData,
  });
}

export function useUnregisterEntry() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      semesterId,
      eventId,
      membershipId,
    }: {
      semesterId: string;
      eventId: number;
      membershipId: string;
    }) => unregisterEntry(semesterId, eventId, membershipId),
    onSuccess: (_data, { semesterId, eventId }) => {
      queryClient.invalidateQueries({ queryKey: entryKeys.byEvent(semesterId, eventId) });
    },
  });
}

export function useSignInEntry() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      semesterId,
      eventId,
      membershipId,
    }: {
      semesterId: string;
      eventId: number;
      membershipId: string;
    }) => signInEntry(semesterId, eventId, membershipId),
    onSuccess: (_data, { semesterId, eventId }) => {
      queryClient.invalidateQueries({ queryKey: entryKeys.byEvent(semesterId, eventId) });
    },
  });
}

export function useSignOutEntry() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      semesterId,
      eventId,
      membershipId,
    }: {
      semesterId: string;
      eventId: number;
      membershipId: string;
    }) => signOutEntry(semesterId, eventId, membershipId),
    onSuccess: (_data, { semesterId, eventId }) => {
      queryClient.invalidateQueries({ queryKey: entryKeys.byEvent(semesterId, eventId) });
    },
  });
}
