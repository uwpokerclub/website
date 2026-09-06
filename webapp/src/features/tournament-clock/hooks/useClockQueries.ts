import { QueryClient, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { adjustClock, fetchClock, pauseClock, resumeClock, setClockLevel } from "../api/clockApi";
import { ClockState } from "@/types";

export const clockKeys = {
  all: ["clock"] as const,
  detail: (semesterId: string, eventId: number) => [...clockKeys.all, semesterId, eventId] as const,
};

export type ClockQueryData = ClockState & { offsetMs: number };

/** Returns whichever of two versioned clock values is newer, preferring the incoming one on a tie. */
export function acceptClockUpdate<T extends { version: number }>(oldData: T | undefined, newData: T): T {
  if (!oldData || newData.version >= oldData.version) {
    return newData;
  }
  return oldData;
}

/** Shared mutationKey for pause/resume, so the poll can tell when one is in flight. */
export const CLOCK_PAUSE_RESUME_MUTATION_KEY = ["clock-pause-resume"];

/**
 * Poll interval that pauses while a pause/resume mutation is in flight.
 *
 * Without this, the 2s poll can race an in-flight optimistic pause/resume:
 * a poll response that reflects the pre-mutation state carries the same
 * version as the optimistic value already in the cache, so the version
 * guard's `>=` lets it clobber the fresher optimistic state — flashing the
 * display back to the old paused/running state until the mutation's own
 * response corrects it a moment later. Suppressing the poll for the
 * mutation's duration removes the race instead of trying to out-guess it
 * after the fact (setQueryData re-applies this same query's structural
 * sharing, so a data-level guard can't tell a legitimate rollback/reconcile
 * write apart from a stale poll).
 */
export function clockRefetchInterval(queryClient: QueryClient): number | false {
  return queryClient.isMutating({ mutationKey: CLOCK_PAUSE_RESUME_MUTATION_KEY }) > 0 ? false : 2000;
}

/** Returns the difference between the server's clock and the local receive time, in milliseconds. */
export function computeOffsetMs(serverTime: string, receivedAt: number): number {
  return new Date(serverTime).getTime() - receivedAt;
}

/** Fetches an event's clock state and attaches the computed server-time offset. */
export async function fetchClockWithOffset(semesterId: string, eventId: number): Promise<ClockQueryData> {
  const state = await fetchClock(semesterId, eventId);
  const receivedAt = Date.now();
  return { ...state, offsetMs: computeOffsetMs(state.serverTime, receivedAt) };
}

/** Base React Query options for fetching an event's clock. See useEventClock for the poll interval. */
export function clockQueryOptions(semesterId: string, eventId: number) {
  return {
    queryKey: clockKeys.detail(semesterId, eventId),
    queryFn: () => fetchClockWithOffset(semesterId, eventId),
    structuralSharing: (oldData: unknown, newData: unknown) =>
      acceptClockUpdate(oldData as ClockQueryData | undefined, newData as ClockQueryData),
  };
}

/** Merges a control action's response into the clock query cache. */
export function reconcileClockResponse(
  queryClient: QueryClient,
  semesterId: string,
  eventId: number,
  response: ClockState,
): void {
  const key = clockKeys.detail(semesterId, eventId);
  const receivedAt = Date.now();
  const incoming: ClockQueryData = { ...response, offsetMs: computeOffsetMs(response.serverTime, receivedAt) };
  const current = queryClient.getQueryData<ClockQueryData>(key);
  // A mutation can resolve before the first query fetch runs, so structuralSharing isn't registered yet.
  queryClient.setQueryData<ClockQueryData>(key, acceptClockUpdate(current, incoming));
}

export type ClockMutationContext = { previous: ClockQueryData | undefined };

/** Optimistically sets pausedAt in the clock query cache ahead of a pause/resume request resolving. */
export async function withOptimisticPause(
  queryClient: QueryClient,
  semesterId: string,
  eventId: number,
  paused: boolean,
): Promise<ClockMutationContext> {
  const key = clockKeys.detail(semesterId, eventId);
  await queryClient.cancelQueries({ queryKey: key });
  const previous = queryClient.getQueryData<ClockQueryData>(key);
  if (previous) {
    // Corrected by offsetMs, not the raw local clock, to stay right under clock skew.
    const pausedAt = paused ? new Date(Date.now() + previous.offsetMs).toISOString() : null;
    queryClient.setQueryData<ClockQueryData>(key, () => ({ ...previous, pausedAt }));
  }
  return { previous };
}

/** Restores the clock query cache to its pre-mutation value after a failed control action. */
export function rollback(
  queryClient: QueryClient,
  semesterId: string,
  eventId: number,
  context: ClockMutationContext | undefined,
): void {
  if (!context?.previous) {
    return;
  }
  const key = clockKeys.detail(semesterId, eventId);
  const current = queryClient.getQueryData<ClockQueryData>(key);
  // Guarded so a fresher poll response that landed during the mutation isn't clobbered.
  queryClient.setQueryData<ClockQueryData>(key, acceptClockUpdate(current, context.previous));
}

/** Polls an event's tournament clock every 2s, pausing while a pause/resume mutation is in flight. */
export function useEventClock(semesterId: string | undefined, eventId: number | undefined) {
  const queryClient = useQueryClient();

  return useQuery({
    ...clockQueryOptions(semesterId ?? "", eventId ?? 0),
    refetchInterval: () => clockRefetchInterval(queryClient),
    enabled: !!semesterId && eventId !== undefined,
  });
}

export function usePauseClock() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationKey: CLOCK_PAUSE_RESUME_MUTATION_KEY,
    mutationFn: ({ semesterId, eventId }: { semesterId: string; eventId: number }) => pauseClock(semesterId, eventId),
    onMutate: ({ semesterId, eventId }) => withOptimisticPause(queryClient, semesterId, eventId, true),
    onError: (_err, { semesterId, eventId }, context) => rollback(queryClient, semesterId, eventId, context),
    onSuccess: (response, { semesterId, eventId }) =>
      reconcileClockResponse(queryClient, semesterId, eventId, response),
  });
}

export function useResumeClock() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationKey: CLOCK_PAUSE_RESUME_MUTATION_KEY,
    mutationFn: ({ semesterId, eventId }: { semesterId: string; eventId: number }) => resumeClock(semesterId, eventId),
    onMutate: ({ semesterId, eventId }) => withOptimisticPause(queryClient, semesterId, eventId, false),
    onError: (_err, { semesterId, eventId }, context) => rollback(queryClient, semesterId, eventId, context),
    onSuccess: (response, { semesterId, eventId }) =>
      reconcileClockResponse(queryClient, semesterId, eventId, response),
  });
}

export function useAdjustClock() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      semesterId,
      eventId,
      deltaSeconds,
    }: {
      semesterId: string;
      eventId: number;
      deltaSeconds: number;
    }) => adjustClock(semesterId, eventId, deltaSeconds),
    onSuccess: (response, { semesterId, eventId }) =>
      reconcileClockResponse(queryClient, semesterId, eventId, response),
  });
}

export function useSetClockLevel() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ semesterId, eventId, index }: { semesterId: string; eventId: number; index: number }) =>
      setClockLevel(semesterId, eventId, index),
    onSuccess: (response, { semesterId, eventId }) =>
      reconcileClockResponse(queryClient, semesterId, eventId, response),
  });
}
