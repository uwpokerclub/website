import { useEffect, useState } from "react";
import { deriveClock, DerivedClock, EventClockState } from "../utils/deriveClock";
import { ClockQueryData } from "./useClockQueries";

const TICK_MS = 250;

/**
 * Recomputes the derived clock state on a local tick, independent of the 2s
 * network poll behind `data` — see the epic design doc's "two independent
 * clocks" requirement (#448/#454).
 */
export function useDerivedClock(data: ClockQueryData | undefined, levelDurationsMs: number[]): DerivedClock | null {
  const [now, setNow] = useState(() => Date.now());
  const isPaused = data?.pausedAt != null;

  // While paused, deriveClock freezes on pausedAt regardless of `now`, so
  // ticking would just burn CPU/battery for a value that never changes —
  // this can run for hours with the screen wake lock held.
  useEffect(() => {
    if (isPaused) {
      return;
    }
    const id = setInterval(() => setNow(Date.now()), TICK_MS);
    return () => clearInterval(id);
  }, [isPaused]);

  if (!data) {
    return null;
  }

  const state: EventClockState = {
    levelIndex: data.levelIndex,
    levelEndsAt: Date.parse(data.levelEndsAt),
    pausedAt: data.pausedAt ? Date.parse(data.pausedAt) : null,
    version: data.version,
  };

  return deriveClock(state, levelDurationsMs, now + data.offsetMs);
}
