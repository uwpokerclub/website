import { useLayoutEffect, useState } from "react";
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
  //
  // A layout effect (not a plain effect) so the reset below lands before the
  // browser paints. `now` stops updating while paused; without an immediate
  // reset here, resuming would render once with that stale value first — a
  // visible flash inflated by however long the clock sat paused, since
  // remaining is computed against `now` once pausedAt goes back to null.
  useLayoutEffect(() => {
    if (isPaused) {
      return;
    }
    setNow(Date.now());
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
