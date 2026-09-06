/** @jest-environment jsdom */
import { act, renderHook } from "@testing-library/react";
import { useDerivedClock } from "./useDerivedClock";
import { ClockQueryData } from "./useClockQueries";

function clockData(overrides: Partial<ClockQueryData> = {}): ClockQueryData {
  return {
    levelIndex: 0,
    levelEndsAt: "2026-01-01T00:10:00.000Z",
    pausedAt: null,
    version: 1,
    serverTime: "2026-01-01T00:00:00.000Z",
    offsetMs: 0,
    ...overrides,
  };
}

describe("useDerivedClock", () => {
  afterEach(() => {
    jest.useRealTimers();
  });

  it("returns null when there is no data yet", () => {
    const { result } = renderHook(() => useDerivedClock(undefined, [10 * 60_000]));

    expect(result.current).toBeNull();
  });

  it("derives the remaining time for the current level from the server state", () => {
    jest.useFakeTimers().setSystemTime(new Date("2026-01-01T00:05:00.000Z"));

    const { result } = renderHook(() => useDerivedClock(clockData(), [10 * 60_000]));

    expect(result.current?.levelIndex).toBe(0);
    expect(result.current?.remainingMs).toBe(5 * 60_000);
  });

  it("recomputes on its own tick as time passes, without new server data", () => {
    jest.useFakeTimers().setSystemTime(new Date("2026-01-01T00:05:00.000Z"));

    const { result } = renderHook(() => useDerivedClock(clockData(), [10 * 60_000]));
    expect(result.current?.remainingMs).toBe(5 * 60_000);

    act(() => {
      jest.advanceTimersByTime(3000);
    });

    expect(result.current?.remainingMs).toBe(5 * 60_000 - 3000);
  });

  it("rolls forward to the next level once the current level's end time has passed", () => {
    jest.useFakeTimers().setSystemTime(new Date("2026-01-01T00:09:59.000Z"));

    const { result } = renderHook(() => useDerivedClock(clockData(), [10 * 60_000, 15 * 60_000]));
    expect(result.current?.levelIndex).toBe(0);

    act(() => {
      jest.advanceTimersByTime(2000);
    });

    expect(result.current?.levelIndex).toBe(1);
    expect(result.current?.remainingMs).toBe(15 * 60_000 - 1000);
  });

  it("applies the server-time offset when computing remaining time", () => {
    // Local clock reads 1 minute behind the server.
    jest.useFakeTimers().setSystemTime(new Date("2026-01-01T00:05:00.000Z"));

    const { result } = renderHook(() => useDerivedClock(clockData({ offsetMs: 60_000 }), [10 * 60_000]));

    expect(result.current?.remainingMs).toBe(4 * 60_000);
  });

  it("freezes remaining time while paused, ignoring the tick", () => {
    jest.useFakeTimers().setSystemTime(new Date("2026-01-01T00:05:00.000Z"));

    const { result } = renderHook(() =>
      useDerivedClock(clockData({ pausedAt: "2026-01-01T00:05:00.000Z" }), [10 * 60_000]),
    );

    act(() => {
      jest.advanceTimersByTime(10_000);
    });

    expect(result.current?.remainingMs).toBe(5 * 60_000);
  });

  it("returns null when there are no levels", () => {
    const { result } = renderHook(() => useDerivedClock(clockData(), []));

    expect(result.current).toBeNull();
  });
});
