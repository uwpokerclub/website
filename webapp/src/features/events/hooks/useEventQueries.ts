import { useQuery, useMutation, useQueryClient, keepPreviousData } from "@tanstack/react-query";
import {
  fetchEvent,
  fetchEvents,
  createEvent,
  updateEvent,
  endEvent,
  restartEvent,
  rebuyEvent,
  CreateEventRequest,
  UpdateEventRequest,
} from "../api/eventApi";
import { entryKeys } from "@/features/entries/hooks/useEntryQueries";

export const eventKeys = {
  all: ["events"] as const,
  lists: () => [...eventKeys.all, "list"] as const,
  list: (semesterId: string, params: { limit: number; offset: number; search?: string }) =>
    [...eventKeys.lists(), semesterId, params] as const,
  details: () => [...eventKeys.all, "detail"] as const,
  detail: (semesterId: string, eventId: number) => [...eventKeys.details(), semesterId, eventId] as const,
};

export function useEvents(semesterId: string | undefined, params: { limit: number; offset: number; search?: string }) {
  return useQuery({
    queryKey: semesterId ? eventKeys.list(semesterId, params) : eventKeys.lists(),
    queryFn: () => fetchEvents(semesterId!, params),
    enabled: !!semesterId,
    placeholderData: keepPreviousData,
  });
}

export function useEvent(semesterId: string | undefined, eventId: number | undefined) {
  return useQuery({
    queryKey: eventKeys.detail(semesterId ?? "", eventId ?? 0),
    queryFn: () => fetchEvent(semesterId!, eventId!),
    enabled: !!semesterId && eventId !== undefined,
  });
}

export function useCreateEvent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ semesterId, data }: { semesterId: string; data: CreateEventRequest }) =>
      createEvent(semesterId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: eventKeys.lists() });
    },
  });
}

export function useUpdateEvent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ semesterId, eventId, data }: { semesterId: string; eventId: number; data: UpdateEventRequest }) =>
      updateEvent(semesterId, eventId, data),
    onSuccess: (_data, { semesterId, eventId }) => {
      queryClient.invalidateQueries({ queryKey: eventKeys.detail(semesterId, eventId) });
      queryClient.invalidateQueries({ queryKey: eventKeys.lists() });
    },
  });
}

export function useEndEvent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ semesterId, eventId }: { semesterId: string; eventId: number }) => endEvent(semesterId, eventId),
    onSuccess: (_data, { semesterId, eventId }) => {
      queryClient.invalidateQueries({ queryKey: eventKeys.detail(semesterId, eventId) });
      queryClient.invalidateQueries({ queryKey: eventKeys.lists() });
      // Ending an event sets entry placements; refresh the entries list
      queryClient.invalidateQueries({ queryKey: entryKeys.byEvent(semesterId, eventId) });
    },
  });
}

export function useRestartEvent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ semesterId, eventId }: { semesterId: string; eventId: number }) => restartEvent(semesterId, eventId),
    onSuccess: (_data, { semesterId, eventId }) => {
      queryClient.invalidateQueries({ queryKey: eventKeys.detail(semesterId, eventId) });
      queryClient.invalidateQueries({ queryKey: eventKeys.lists() });
      // Restarting an event clears placements; refresh the entries list
      queryClient.invalidateQueries({ queryKey: entryKeys.byEvent(semesterId, eventId) });
    },
  });
}

export function useRebuyEvent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ semesterId, eventId }: { semesterId: string; eventId: number }) => rebuyEvent(semesterId, eventId),
    onSuccess: (_data, { semesterId, eventId }) => {
      queryClient.invalidateQueries({ queryKey: eventKeys.detail(semesterId, eventId) });
      queryClient.invalidateQueries({ queryKey: eventKeys.lists() });
    },
  });
}
