import { QueryClient } from "@tanstack/query-core";

// clockApi.ts imports the real apiClient, which reads `import.meta.env` —
// not valid under Jest's CommonJS transform. These tests exercise cache
// wiring only and supply their own queryFn/mutationFn per call, so the real
// api module is never invoked; stub it out to avoid loading it at all.
jest.mock("../api/clockApi", () => ({
  fetchClock: jest.fn(),
  pauseClock: jest.fn(),
  resumeClock: jest.fn(),
  adjustClock: jest.fn(),
  setClockLevel: jest.fn(),
}));

import {
  acceptClockUpdate,
  clockKeys,
  clockQueryOptions,
  computeOffsetMs,
  fetchClockWithOffset,
  reconcileClockResponse,
  rollback,
  withOptimisticPause,
  ClockQueryData,
} from "./useClockQueries";
import { fetchClock } from "../api/clockApi";
import { ClockState } from "@/types";

function clockData(version: number, overrides: Partial<ClockQueryData> = {}): ClockQueryData {
  return {
    levelIndex: 0,
    levelEndsAt: "2026-01-01T00:10:00.000Z",
    pausedAt: null,
    version,
    serverTime: "2026-01-01T00:00:00.000Z",
    offsetMs: 0,
    ...overrides,
  };
}

describe("acceptClockUpdate", () => {
  it("rejects an incoming value with a lower version than the cached one", () => {
    const cached = { version: 5 };
    const stale = { version: 3 };

    expect(acceptClockUpdate(cached, stale)).toBe(cached);
  });

  it("accepts an incoming value with a higher version than the cached one", () => {
    const cached = { version: 5 };
    const fresh = { version: 6 };

    expect(acceptClockUpdate(cached, fresh)).toBe(fresh);
  });

  it("accepts an incoming value with an equal version (server may return unchanged state after a no-op)", () => {
    const cached = { version: 5, pausedAt: null };
    const unchanged = { version: 5, pausedAt: null };

    expect(acceptClockUpdate(cached, unchanged)).toBe(unchanged);
  });

  it("accepts the incoming value when there is no cached value yet", () => {
    const first = { version: 1 };

    expect(acceptClockUpdate(undefined, first)).toBe(first);
  });
});

describe("computeOffsetMs", () => {
  it("is the difference between the server's clock and the local receive time", () => {
    const serverTime = "2026-01-01T00:00:05.000Z";
    const receivedAt = Date.parse("2026-01-01T00:00:00.000Z");

    expect(computeOffsetMs(serverTime, receivedAt)).toBe(5000);
  });

  it("is negative when the local clock is ahead of the server", () => {
    const serverTime = "2026-01-01T00:00:00.000Z";
    const receivedAt = Date.parse("2026-01-01T00:00:05.000Z");

    expect(computeOffsetMs(serverTime, receivedAt)).toBe(-5000);
  });
});

describe("fetchClockWithOffset", () => {
  afterEach(() => {
    jest.useRealTimers();
  });

  it("computes the offset from the time the response was received, not when the request was sent", async () => {
    jest.useFakeTimers().setSystemTime(new Date("2026-01-01T00:00:00.000Z"));
    (fetchClock as jest.Mock).mockImplementation(async () => {
      jest.advanceTimersByTime(3000); // simulate a slow 3s round trip
      return clockData(5, { serverTime: "2026-01-01T00:00:10.000Z" });
    });

    const result = await fetchClockWithOffset("s1", 1);

    // Received at 00:00:03 (after the simulated round trip); server says
    // 00:00:10 => offset should be 7000ms, not 10000ms.
    expect(result.offsetMs).toBe(7000);
  });
});

describe("clockQueryOptions", () => {
  it("rejects a poll that resolves with a lower version than the cached state", async () => {
    const client = new QueryClient();
    const options = clockQueryOptions("s1", 1);

    await client.fetchQuery({ ...options, queryFn: () => Promise.resolve(clockData(5)) });
    await client.fetchQuery({ ...options, queryFn: () => Promise.resolve(clockData(3)) });

    expect(client.getQueryData(options.queryKey)).toEqual(clockData(5));
  });

  it("accepts a poll with an equal version, replacing the cached value", async () => {
    const client = new QueryClient();
    const options = clockQueryOptions("s1", 1);
    const unchangedButNewer = clockData(5, { serverTime: "2026-01-01T00:00:02.000Z" });

    await client.fetchQuery({ ...options, queryFn: () => Promise.resolve(clockData(5)) });
    await client.fetchQuery({ ...options, queryFn: () => Promise.resolve(unchangedButNewer) });

    expect(client.getQueryData(options.queryKey)).toEqual(unchangedButNewer);
  });

  it("keeps the last good state in the cache when a poll fails", async () => {
    const client = new QueryClient();
    const options = clockQueryOptions("s1", 1);

    await client.fetchQuery({ ...options, queryFn: () => Promise.resolve(clockData(5)) });
    await expect(
      client.fetchQuery({ ...options, queryFn: () => Promise.reject(new Error("network down")), retry: false }),
    ).rejects.toThrow("network down");

    expect(client.getQueryData(options.queryKey)).toEqual(clockData(5));
  });
});

describe("reconcileClockResponse", () => {
  it("replaces the cached value with the server response, without fabricating a version", async () => {
    const client = new QueryClient();
    const options = clockQueryOptions("s1", 1);
    await client.fetchQuery({ ...options, queryFn: () => Promise.resolve(clockData(5)) });

    const response: ClockState = {
      levelIndex: 1,
      levelEndsAt: "2026-01-01T00:20:00.000Z",
      pausedAt: null,
      version: 6,
      serverTime: "2026-01-01T00:00:10.000Z",
    };
    reconcileClockResponse(client, "s1", 1, response);

    const cached = client.getQueryData<ClockQueryData>(options.queryKey);
    expect(cached?.version).toBe(6);
    expect(cached?.levelIndex).toBe(1);
  });

  it("rejects a stale response even when the cache entry was never fetched through clockQueryOptions", () => {
    // The query cache can hold this key before useEventClock's first fetch has run
    // (e.g. a mutation firing first), which is exactly when the query has no
    // structuralSharing registered on it yet. The guard must not depend on that.
    const client = new QueryClient();
    const key = clockKeys.detail("s1", 1);
    client.setQueryData<ClockQueryData>(key, clockData(6));

    const stale: ClockState = {
      levelIndex: 0,
      levelEndsAt: "2026-01-01T00:10:00.000Z",
      pausedAt: null,
      version: 3,
      serverTime: "2026-01-01T00:00:00.000Z",
    };
    reconcileClockResponse(client, "s1", 1, stale);

    expect(client.getQueryData<ClockQueryData>(key)?.version).toBe(6);
  });
});

describe("withOptimisticPause", () => {
  afterEach(() => {
    jest.useRealTimers();
  });

  it("freezes pausedAt locally without changing the version", async () => {
    const client = new QueryClient();
    const options = clockQueryOptions("s1", 1);
    await client.fetchQuery({ ...options, queryFn: () => Promise.resolve(clockData(5, { pausedAt: null })) });

    await withOptimisticPause(client, "s1", 1, true);

    const cached = client.getQueryData<ClockQueryData>(options.queryKey);
    expect(cached?.pausedAt).not.toBeNull();
    expect(cached?.version).toBe(5);
  });

  it("corrects the frozen instant by the cached server offset, not the raw local clock", async () => {
    jest.useFakeTimers().setSystemTime(new Date("2026-01-01T00:00:00.000Z"));
    const client = new QueryClient();
    const options = clockQueryOptions("s1", 1);
    // Local clock reads 5 minutes behind the server.
    await client.fetchQuery({
      ...options,
      queryFn: () => Promise.resolve(clockData(5, { pausedAt: null, offsetMs: 5 * 60 * 1000 })),
    });

    await withOptimisticPause(client, "s1", 1, true);

    const cached = client.getQueryData<ClockQueryData>(options.queryKey);
    expect(cached?.pausedAt).toBe("2026-01-01T00:05:00.000Z");
  });

  it("clears pausedAt when resuming", async () => {
    const client = new QueryClient();
    const options = clockQueryOptions("s1", 1);
    await client.fetchQuery({
      ...options,
      queryFn: () => Promise.resolve(clockData(5, { pausedAt: "2026-01-01T00:00:01.000Z" })),
    });

    await withOptimisticPause(client, "s1", 1, false);

    expect(client.getQueryData<ClockQueryData>(options.queryKey)?.pausedAt).toBeNull();
  });

  it("returns the previous value for rollback", async () => {
    const client = new QueryClient();
    const options = clockQueryOptions("s1", 1);
    const original = clockData(5, { pausedAt: null });
    await client.fetchQuery({ ...options, queryFn: () => Promise.resolve(original) });

    const context = await withOptimisticPause(client, "s1", 1, true);

    expect(context.previous).toEqual(original);
  });
});

describe("rollback", () => {
  it("restores the previous cached value after a failed mutation", async () => {
    const client = new QueryClient();
    const options = clockQueryOptions("s1", 1);
    const original = clockData(5, { pausedAt: null });
    await client.fetchQuery({ ...options, queryFn: () => Promise.resolve(original) });
    await withOptimisticPause(client, "s1", 1, true);

    rollback(client, "s1", 1, { previous: original });

    expect(client.getQueryData(options.queryKey)).toEqual(original);
  });

  it("does not clobber a newer value that landed after the snapshot was taken", () => {
    // e.g. a poll response arrives between a mutation's optimistic patch and
    // its (failed) response — rollback must not stomp that fresher state.
    const client = new QueryClient();
    const key = clockKeys.detail("s1", 1);
    const stale = clockData(4);
    client.setQueryData<ClockQueryData>(key, clockData(6));

    rollback(client, "s1", 1, { previous: stale });

    expect(client.getQueryData<ClockQueryData>(key)?.version).toBe(6);
  });
});
