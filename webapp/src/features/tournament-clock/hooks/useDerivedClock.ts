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

  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), TICK_MS);
    return () => clearInterval(id);
  }, []);

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
