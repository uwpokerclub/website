/** @jest-environment jsdom */
import { renderHook } from "@testing-library/react";
import { useClockSyncStatus } from "./useClockSyncStatus";

describe("useClockSyncStatus", () => {
  it("is synced when there have been no poll failures", () => {
    const { result } = renderHook(() => useClockSyncStatus(1000, 0));

    expect(result.current).toBe(false);
  });

  it("stays synced after fewer than three consecutive failed polls", () => {
    const { result, rerender } = renderHook(
      ({ dataUpdatedAt, errorUpdatedAt }) => useClockSyncStatus(dataUpdatedAt, errorUpdatedAt),
      {
        initialProps: { dataUpdatedAt: 1000, errorUpdatedAt: 0 },
      },
    );

    rerender({ dataUpdatedAt: 1000, errorUpdatedAt: 2000 });
    rerender({ dataUpdatedAt: 1000, errorUpdatedAt: 4000 });

    expect(result.current).toBe(false);
  });

  it("goes offline after three consecutive failed polls", () => {
    const { result, rerender } = renderHook(
      ({ dataUpdatedAt, errorUpdatedAt }) => useClockSyncStatus(dataUpdatedAt, errorUpdatedAt),
      {
        initialProps: { dataUpdatedAt: 1000, errorUpdatedAt: 0 },
      },
    );

    rerender({ dataUpdatedAt: 1000, errorUpdatedAt: 2000 });
    rerender({ dataUpdatedAt: 1000, errorUpdatedAt: 4000 });
    rerender({ dataUpdatedAt: 1000, errorUpdatedAt: 6000 });

    expect(result.current).toBe(true);
  });

  it("clears as soon as a poll succeeds again", () => {
    const { result, rerender } = renderHook(
      ({ dataUpdatedAt, errorUpdatedAt }) => useClockSyncStatus(dataUpdatedAt, errorUpdatedAt),
      {
        initialProps: { dataUpdatedAt: 1000, errorUpdatedAt: 0 },
      },
    );

    rerender({ dataUpdatedAt: 1000, errorUpdatedAt: 2000 });
    rerender({ dataUpdatedAt: 1000, errorUpdatedAt: 4000 });
    rerender({ dataUpdatedAt: 1000, errorUpdatedAt: 6000 });
    expect(result.current).toBe(true);

    rerender({ dataUpdatedAt: 8000, errorUpdatedAt: 6000 });

    expect(result.current).toBe(false);
  });

  it("does not double-count the same settled error across re-renders with unrelated prop changes", () => {
    const { result, rerender } = renderHook(
      ({ dataUpdatedAt, errorUpdatedAt }) => useClockSyncStatus(dataUpdatedAt, errorUpdatedAt),
      {
        initialProps: { dataUpdatedAt: 1000, errorUpdatedAt: 2000 },
      },
    );

    // Re-rendering with the same errorUpdatedAt (e.g. a parent re-render for
    // an unrelated reason) must not advance the failure count.
    rerender({ dataUpdatedAt: 1000, errorUpdatedAt: 2000 });
    rerender({ dataUpdatedAt: 1000, errorUpdatedAt: 2000 });
    rerender({ dataUpdatedAt: 1000, errorUpdatedAt: 2000 });

    expect(result.current).toBe(false);
  });
});
