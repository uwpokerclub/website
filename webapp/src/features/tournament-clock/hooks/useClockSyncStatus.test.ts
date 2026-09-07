/** @jest-environment jsdom */
import { renderHook } from "@testing-library/react";
import { useClockSyncStatus } from "./useClockSyncStatus";

describe("useClockSyncStatus", () => {
  it("is synced when there have been no poll failures", () => {
    const { result } = renderHook(() => useClockSyncStatus(1, 0));

    expect(result.current).toBe(false);
  });

  it("stays synced after fewer than three consecutive failed polls", () => {
    const { result, rerender } = renderHook(
      ({ pollSuccessCount, pollFailureCount }) => useClockSyncStatus(pollSuccessCount, pollFailureCount),
      {
        initialProps: { pollSuccessCount: 1, pollFailureCount: 0 },
      },
    );

    rerender({ pollSuccessCount: 1, pollFailureCount: 1 });
    rerender({ pollSuccessCount: 1, pollFailureCount: 2 });

    expect(result.current).toBe(false);
  });

  it("goes offline after three consecutive failed polls", () => {
    const { result, rerender } = renderHook(
      ({ pollSuccessCount, pollFailureCount }) => useClockSyncStatus(pollSuccessCount, pollFailureCount),
      {
        initialProps: { pollSuccessCount: 1, pollFailureCount: 0 },
      },
    );

    rerender({ pollSuccessCount: 1, pollFailureCount: 1 });
    rerender({ pollSuccessCount: 1, pollFailureCount: 2 });
    rerender({ pollSuccessCount: 1, pollFailureCount: 3 });

    expect(result.current).toBe(true);
  });

  it("clears as soon as a poll succeeds again", () => {
    const { result, rerender } = renderHook(
      ({ pollSuccessCount, pollFailureCount }) => useClockSyncStatus(pollSuccessCount, pollFailureCount),
      {
        initialProps: { pollSuccessCount: 1, pollFailureCount: 0 },
      },
    );

    rerender({ pollSuccessCount: 1, pollFailureCount: 1 });
    rerender({ pollSuccessCount: 1, pollFailureCount: 2 });
    rerender({ pollSuccessCount: 1, pollFailureCount: 3 });
    expect(result.current).toBe(true);

    rerender({ pollSuccessCount: 2, pollFailureCount: 3 });

    expect(result.current).toBe(false);
  });

  it("does not double-count the same settled failure across re-renders with unrelated prop changes", () => {
    const { result, rerender } = renderHook(
      ({ pollSuccessCount, pollFailureCount }) => useClockSyncStatus(pollSuccessCount, pollFailureCount),
      {
        initialProps: { pollSuccessCount: 1, pollFailureCount: 1 },
      },
    );

    // Re-rendering with the same pollFailureCount (e.g. a parent re-render for
    // an unrelated reason) must not advance the failure count.
    rerender({ pollSuccessCount: 1, pollFailureCount: 1 });
    rerender({ pollSuccessCount: 1, pollFailureCount: 1 });
    rerender({ pollSuccessCount: 1, pollFailureCount: 1 });

    expect(result.current).toBe(false);
  });

  it("is not fooled by a local cache write that changes neither counter", () => {
    // Mutation-driven setQueryData calls (optimistic pause, rollback,
    // reconcile) don't touch these counters — only the queryFn does. A
    // re-render triggered by one of those writes must be a no-op here.
    const { result, rerender } = renderHook(
      ({ pollSuccessCount, pollFailureCount }) => useClockSyncStatus(pollSuccessCount, pollFailureCount),
      {
        initialProps: { pollSuccessCount: 1, pollFailureCount: 0 },
      },
    );

    rerender({ pollSuccessCount: 1, pollFailureCount: 1 });
    rerender({ pollSuccessCount: 1, pollFailureCount: 2 });
    rerender({ pollSuccessCount: 1, pollFailureCount: 3 });
    expect(result.current).toBe(true);

    rerender({ pollSuccessCount: 1, pollFailureCount: 3 });

    expect(result.current).toBe(true);
  });
});
