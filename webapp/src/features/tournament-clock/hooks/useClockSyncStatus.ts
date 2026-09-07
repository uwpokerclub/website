import { useRef } from "react";

/** Consecutive failed polls before a room is told it may be missing updates. */
export const OFFLINE_THRESHOLD = 3;

/**
 * Tracks consecutive clock-poll failures from a query's success/failure
 * counters and reports whether the room has gone quiet.
 *
 * These counters must come from the queryFn itself, not from the query's
 * `dataUpdatedAt`/`errorUpdatedAt` — those are also bumped by setQueryData,
 * which mutation-driven cache writes (optimistic pause, rollback, reconcile)
 * use, so they can't tell a real poll apart from a local write. See #455.
 *
 * This is not a claim that the displayed time is wrong — deriveClock is
 * exact, so a room that stops hearing from the server keeps showing the
 * right level and time. It only means a control action taken elsewhere
 * (pause, ±1 min, skip level) may not have been heard yet.
 */
export function useClockSyncStatus(pollSuccessCount: number, pollFailureCount: number): boolean {
  const lastFailureCount = useRef(pollFailureCount);
  const consecutiveFailures = useRef(0);
  const lastSuccessCount = useRef(pollSuccessCount);

  if (pollSuccessCount !== lastSuccessCount.current) {
    lastSuccessCount.current = pollSuccessCount;
    consecutiveFailures.current = 0;
  } else if (pollFailureCount !== lastFailureCount.current) {
    lastFailureCount.current = pollFailureCount;
    consecutiveFailures.current += 1;
  }

  return consecutiveFailures.current >= OFFLINE_THRESHOLD;
}
