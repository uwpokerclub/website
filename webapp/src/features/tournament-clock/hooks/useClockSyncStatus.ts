import { useRef } from "react";

/** Consecutive failed polls before a room is told it may be missing updates. */
export const OFFLINE_THRESHOLD = 3;

/**
 * Tracks consecutive clock-poll failures from a query's `dataUpdatedAt`/
 * `errorUpdatedAt` timestamps and reports whether the room has gone quiet.
 *
 * This is not a claim that the displayed time is wrong — deriveClock is
 * exact, so a room that stops hearing from the server keeps showing the
 * right level and time. It only means a control action taken elsewhere
 * (pause, ±1 min, skip level) may not have been heard yet.
 */
export function useClockSyncStatus(dataUpdatedAt: number, errorUpdatedAt: number): boolean {
  const lastErrorUpdatedAt = useRef(errorUpdatedAt);
  const consecutiveFailures = useRef(0);
  const lastDataUpdatedAt = useRef(dataUpdatedAt);

  if (dataUpdatedAt !== lastDataUpdatedAt.current) {
    lastDataUpdatedAt.current = dataUpdatedAt;
    consecutiveFailures.current = 0;
  } else if (errorUpdatedAt !== lastErrorUpdatedAt.current) {
    lastErrorUpdatedAt.current = errorUpdatedAt;
    consecutiveFailures.current += 1;
  }

  return consecutiveFailures.current >= OFFLINE_THRESHOLD;
}
