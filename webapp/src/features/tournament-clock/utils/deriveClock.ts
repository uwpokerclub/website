// Direct port of server/internal/models/EventClock.Derive (#449). All time
// values are epoch milliseconds and durations are milliseconds, matching
// Date.now(). Keep this in step with the Go version and its test table
// (server/internal/models/event_clock_test.go) — a change to one without the
// other is a defect.

// The server's ClockState wire format (server/internal/models/event_clock.go)
// serializes levelEndsAt/pausedAt as RFC3339 strings, not numbers — callers
// must convert with `new Date(...).getTime()` before calling deriveClock.
export type EventClockState = {
  levelIndex: number;
  levelEndsAt: number;
  pausedAt: number | null;
  version: number;
};

export type DerivedClock = {
  levelIndex: number;
  levelEndsAt: number;
  pausedAt: number | null;
  remainingMs: number;
  version: number;
};

/**
 * Rolls the clock forward to `now` given the ordered level durations (ms).
 * Returns null if `levels` is empty. Pure: does not mutate `state`.
 */
export function deriveClock(state: EventClockState, levels: number[], now: number): DerivedClock | null {
  if (levels.length === 0) {
    return null;
  }

  const effectiveNow = state.pausedAt ?? now;

  let levelIndex = state.levelIndex;
  let levelEndsAt = state.levelEndsAt;
  while (state.pausedAt === null && now >= levelEndsAt && levelIndex < levels.length - 1) {
    levelIndex++;
    levelEndsAt += levels[levelIndex];
  }

  const remainingMs = Math.max(0, levelEndsAt - effectiveNow);

  return {
    levelIndex,
    levelEndsAt,
    pausedAt: state.pausedAt,
    remainingMs,
    version: state.version,
  };
}
