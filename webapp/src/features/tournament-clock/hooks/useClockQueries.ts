import { QueryClient, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { adjustClock, fetchClock, pauseClock, resumeClock, setClockLevel } from "../api/clockApi";
import { ClockState } from "@/types";

export const clockKeys = {
  all: ["clock"] as const,
  detail: (semesterId: string, eventId: number) => [...clockKeys.all, semesterId, eventId] as const,
};

export type ClockQueryData = ClockState & { offsetMs: number };

/**
 * Accepts an incoming clock value only when its version is at least the
 * cached version. This is the version guard: it stops a stale poll — issued
 * before a control action, landing after it — from clobbering fresher state
 * (#453 design: server response always carries the authoritative version).
 */
export function acceptClockUpdate<T extends { version: number }>(oldData: T | undefined, newData: T): T {
  if (!oldData || newData.version >= oldData.version) {
    return newData;
  }
  return oldData;
}

/**
 * `serverTime` minus local receive time. Consumers add this to their own
 * `Date.now()` before calling `deriveClock`, which is the whole safeguard
 * against a projector laptop with a badly-set system clock.
 */
export function computeOffsetMs(serverTime: string, receivedAt: number): number {
  return new Date(serverTime).getTime() - receivedAt;
}

/** Exported for testing the receive-time ordering; consumed internally by `clockQueryOptions`. */
export async function fetchClockWithOffset(semesterId: string, eventId: number): Promise<ClockQueryData> {
  const state = await fetchClock(semesterId, eventId);
  const receivedAt = Date.now();
  return { ...state, offsetMs: computeOffsetMs(state.serverTime, receivedAt) };
}

/**
 * The query config shared by `useEventClock` and its tests. `structuralSharing`
 * is what applies the version guard to every fetch that lands in this query's
 * cache slot, including the 2s poll — see `acceptClockUpdate`.
 */
export function clockQueryOptions(semesterId: string, eventId: number) {
  return {
    queryKey: clockKeys.detail(semesterId, eventId),
    queryFn: () => fetchClockWithOffset(semesterId, eventId),
    refetchInterval: 2000,
    structuralSharing: (oldData: unknown, newData: unknown) =>
      acceptClockUpdate(oldData as ClockQueryData | undefined, newData as ClockQueryData),
  };
}

/**
 * Replaces the cached clock with a control action's response. Never
 * fabricates `version` — it is carried through verbatim from the server.
 *
 * Applies `acceptClockUpdate` itself rather than relying on the query's
 * registered `structuralSharing` (set by `clockQueryOptions`): that option
 * only exists once something has fetched this key through `useEventClock`,
 * and a mutation can resolve before that first fetch has happened.
 */
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
  queryClient.setQueryData<ClockQueryData>(key, acceptClockUpdate(current, incoming));
}

export type ClockMutationContext = { previous: ClockQueryData | undefined };

/**
 * Optimistically patches `pausedAt` ahead of a pause/resume response landing.
 * Leaves `version` untouched — the whole point being that we never guess at a
 * value only the server can assign. The frozen instant is corrected by the
 * cached `offsetMs`, not the raw local clock, so a projector with a
 * badly-set system clock still freezes at the right server-time instant.
 */
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
    const pausedAt = paused ? new Date(Date.now() + previous.offsetMs).toISOString() : null;
    queryClient.setQueryData<ClockQueryData>(key, () => ({ ...previous, pausedAt }));
  }
  return { previous };
}

/**
 * Restores the pre-mutation cached value after a failed control action.
 * Goes through `acceptClockUpdate` so a fresher value that landed in the
 * meantime (e.g. a poll response) is not clobbered by the stale snapshot.
 */
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
  queryClient.setQueryData<ClockQueryData>(key, acceptClockUpdate(current, context.previous));
}

/** Polls an event's tournament clock every 2s, deriving level/remaining time on the client from the returned state. */
export function useEventClock(semesterId: string | undefined, eventId: number | undefined) {
  return useQuery({
    ...clockQueryOptions(semesterId ?? "", eventId ?? 0),
    enabled: !!semesterId && eventId !== undefined,
  });
}

export function usePauseClock() {
  const queryClient = useQueryClient();

  return useMutation({
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
