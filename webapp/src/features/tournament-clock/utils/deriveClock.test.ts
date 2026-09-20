import { deriveClock, EventClockState } from "./deriveClock";

// This table mirrors server/internal/models/event_clock_test.go's
// TestEventClock_Derive case-for-case (#449). A change to one table without
// the other is a defect: it means the Go and TypeScript derivations disagree.
type Case = {
  name: string;
  levelIndex: number;
  levelEndsAt: number;
  pausedAt: number | null;
  levels: number[];
  now: number;
  wantLevelIndex: number;
  wantLevelEndsAt: number;
  wantRemainingMs: number;
  wantPausedAt: number | null;
};

const cases: Case[] = [
  {
    name: "pause freezes remaining regardless of elapsed wall time",
    levelIndex: 0,
    levelEndsAt: 1000,
    pausedAt: 800,
    levels: [600],
    now: 5000,
    wantLevelIndex: 0,
    wantLevelEndsAt: 1000,
    wantRemainingMs: 200,
    wantPausedAt: 800,
  },
  {
    name: "resume at the moment of unpause restores the exact frozen remaining",
    levelIndex: 0,
    levelEndsAt: 1000,
    pausedAt: null,
    levels: [600],
    now: 800,
    wantLevelIndex: 0,
    wantLevelEndsAt: 1000,
    wantRemainingMs: 200,
    wantPausedAt: null,
  },
  {
    name: "multi-level roll-forward across a long gap lands on the right level and remaining time",
    levelIndex: 0,
    levelEndsAt: 600,
    pausedAt: null,
    levels: [600, 900, 1200, 1500],
    now: 3200,
    wantLevelIndex: 3,
    wantLevelEndsAt: 4200,
    wantRemainingMs: 1000,
    wantPausedAt: null,
  },
  {
    name: "negative adjust crossing a level boundary carries the overflow",
    levelIndex: 0,
    levelEndsAt: 500,
    pausedAt: null,
    levels: [600, 900],
    now: 520,
    wantLevelIndex: 1,
    wantLevelEndsAt: 1400,
    wantRemainingMs: 880,
    wantPausedAt: null,
  },
  {
    name: "last level expired clamps at 0:00",
    levelIndex: 1,
    levelEndsAt: 1500,
    pausedAt: null,
    levels: [600, 900],
    now: 5000,
    wantLevelIndex: 1,
    wantLevelEndsAt: 1500,
    wantRemainingMs: 0,
    wantPausedAt: null,
  },
  {
    name: "a single level far shorter than the gap rolls to the end and clamps",
    levelIndex: 0,
    levelEndsAt: 600,
    pausedAt: null,
    levels: [600],
    now: 100000,
    wantLevelIndex: 0,
    wantLevelEndsAt: 600,
    wantRemainingMs: 0,
    wantPausedAt: null,
  },
];

describe("deriveClock", () => {
  test.each(cases)("$name", (tc) => {
    const state: EventClockState = {
      levelIndex: tc.levelIndex,
      levelEndsAt: epoch(tc.levelEndsAt),
      pausedAt: epochOrNull(tc.pausedAt),
      version: 0,
    };
    const levels = tc.levels.map((s) => s * 1000);

    const got = deriveClock(state, levels, epoch(tc.now));

    expect(got).toEqual({
      levelIndex: tc.wantLevelIndex,
      levelEndsAt: epoch(tc.wantLevelEndsAt),
      remainingMs: tc.wantRemainingMs * 1000,
      pausedAt: epochOrNull(tc.wantPausedAt),
      version: 0,
    });
  });

  it("returns null for an empty structure", () => {
    const state: EventClockState = { levelIndex: 0, levelEndsAt: epoch(0), pausedAt: null, version: 0 };

    expect(deriveClock(state, [], epoch(0))).toBeNull();
  });

  it("passes version through unchanged", () => {
    const state: EventClockState = { levelIndex: 0, levelEndsAt: epoch(600), pausedAt: null, version: 17 };

    const got = deriveClock(state, [600 * 1000, 900 * 1000], epoch(3000));

    expect(got?.version).toBe(17);
  });

  it("does not mutate the input state", () => {
    const state: EventClockState = { levelIndex: 0, levelEndsAt: epoch(600), pausedAt: null, version: 0 };

    deriveClock(state, [600 * 1000, 900 * 1000], epoch(3000));

    expect(state.levelIndex).toBe(0);
    expect(state.levelEndsAt).toBe(epoch(600));
  });
});

function epoch(seconds: number): number {
  return seconds * 1000;
}

function epochOrNull(seconds: number | null): number | null {
  return seconds === null ? null : epoch(seconds);
}
