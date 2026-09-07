/** @jest-environment jsdom */
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";
import { ReactNode } from "react";
import { TournamentClock } from "./TournamentClock";
import { Blind } from "@/types";

// clockApi.ts imports the real apiClient, which reads `import.meta.env` — not
// valid under Jest's CommonJS transform (see useClockQueries.test.ts). Mocked
// out here for the same reason.
jest.mock("../api/clockApi", () => ({
  fetchClock: jest.fn(),
  pauseClock: jest.fn(),
  resumeClock: jest.fn(),
  adjustClock: jest.fn(),
  setClockLevel: jest.fn(),
}));
// The real barrel transitively imports modules that also read `import.meta.env`.
jest.mock("../../../components", () => ({
  Icon: ({ iconType }: { iconType: string }) => <span data-qa={`icon-${iconType}`} />,
}));
jest.mock("../utils/playSound", () => ({ playSound: jest.fn() }));
// ClockActions checks clock control permission directly (#455); grant it by default here.
const mockHasPermission = jest.fn().mockReturnValue(true);
jest.mock("@/hooks/useAuth", () => ({ useAuth: () => ({ hasPermission: mockHasPermission }) }));

import { fetchClock, pauseClock, resumeClock, adjustClock, setClockLevel } from "../api/clockApi";
import { ClockState } from "@/types";

const LEVELS: Blind[] = [
  { small: 25, big: 50, ante: 0, time: 10 },
  { small: 50, big: 100, ante: 0, time: 10 },
  { small: 100, big: 200, ante: 0, time: 10 },
];

function clockState(overrides: Partial<ClockState> = {}): ClockState {
  return {
    levelIndex: 0,
    levelEndsAt: new Date(Date.now() + 10 * 60_000).toISOString(),
    pausedAt: null,
    version: 1,
    serverTime: new Date().toISOString(),
    ...overrides,
  };
}

function renderClock(levels: Blind[] = LEVELS) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  return render(<TournamentClock semesterId="s1" eventId={1} levels={levels} />, { wrapper });
}

describe("TournamentClock", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockHasPermission.mockReturnValue(true);
    (global as unknown as { AudioContext: unknown }).AudioContext = class {
      currentTime = 0;
      createOscillator = jest.fn();
      createGain = jest.fn();
      destination = {};
    };
    // jsdom has no Screen Wake Lock API; stub it so the unrelated effect doesn't log noise.
    Object.defineProperty(navigator, "wakeLock", {
      configurable: true,
      value: { request: jest.fn().mockResolvedValue({ release: jest.fn().mockResolvedValue(undefined) }) },
    });
  });

  it("shows an empty state when there is no blind structure", () => {
    renderClock([]);

    expect(screen.getByText("No blind structure available")).toBeInTheDocument();
  });

  it("renders the level and remaining time derived from server state", async () => {
    (fetchClock as jest.Mock).mockResolvedValue(
      clockState({ levelIndex: 1, levelEndsAt: new Date(Date.now() + 5 * 60_000).toISOString() }),
    );

    renderClock();

    expect(await screen.findByText("Level 2")).toBeInTheDocument();
    await waitFor(() => {
      expect(document.querySelector('[data-qa="timer"]')).toHaveTextContent("4:59");
    });
  });

  it("advances the displayed level once it expires, without issuing any request", async () => {
    (fetchClock as jest.Mock).mockResolvedValue(
      clockState({ levelIndex: 0, levelEndsAt: new Date(Date.now() + 300).toISOString() }),
    );

    renderClock();

    expect(await screen.findByText("Level 1")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByText("Level 2")).toBeInTheDocument(), { timeout: 2000 });

    expect(pauseClock).not.toHaveBeenCalled();
    expect(resumeClock).not.toHaveBeenCalled();
    expect(adjustClock).not.toHaveBeenCalled();
    expect(setClockLevel).not.toHaveBeenCalled();
  });

  it("resumes at the right level and time on remount instead of restarting the level", async () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    (fetchClock as jest.Mock).mockResolvedValue(
      clockState({ levelIndex: 2, levelEndsAt: new Date(Date.now() + 90_000).toISOString() }),
    );
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );

    const first = render(<TournamentClock semesterId="s1" eventId={1} levels={LEVELS} />, { wrapper });
    expect(await screen.findByText("Level 3")).toBeInTheDocument();
    first.unmount();

    render(<TournamentClock semesterId="s1" eventId={1} levels={LEVELS} />, { wrapper });

    expect(await screen.findByText("Level 3")).toBeInTheDocument();
    await waitFor(() => {
      const timer = document.querySelector('[data-qa="timer"]')!;
      expect(timer.textContent).not.toBe("10:00");
    });
  });

  it("clamps a server levelIndex that is out of range for the current structure instead of crashing", async () => {
    // e.g. an admin shrinks the blind structure (PATCH /structures/:id) while
    // this event's clock has already progressed past the new level count.
    (fetchClock as jest.Mock).mockResolvedValue(
      clockState({ levelIndex: 5, levelEndsAt: new Date(Date.now() + 5 * 60_000).toISOString() }),
    );

    renderClock();

    expect(await screen.findByText("Level 3")).toBeInTheDocument();
    expect(screen.getByText("100 / 200")).toBeInTheDocument();
  });

  it("never reads or writes localStorage", async () => {
    const getSpy = jest.spyOn(Storage.prototype, "getItem");
    const setSpy = jest.spyOn(Storage.prototype, "setItem");
    (fetchClock as jest.Mock).mockResolvedValue(clockState());

    renderClock();
    await screen.findByText("Level 1");

    expect(getSpy).not.toHaveBeenCalled();
    expect(setSpy).not.toHaveBeenCalled();
  });

  describe("control buttons", () => {
    async function setup() {
      (fetchClock as jest.Mock).mockResolvedValue(clockState({ levelIndex: 1, pausedAt: new Date().toISOString() }));
      (pauseClock as jest.Mock).mockResolvedValue(clockState({ levelIndex: 1, pausedAt: new Date().toISOString() }));
      (resumeClock as jest.Mock).mockResolvedValue(clockState({ levelIndex: 1, pausedAt: null }));
      (adjustClock as jest.Mock).mockResolvedValue(clockState({ levelIndex: 1 }));
      (setClockLevel as jest.Mock).mockResolvedValue(clockState({ levelIndex: 2 }));

      renderClock();
      await screen.findByText("Level 2");
      return userEvent.setup();
    }

    it("resumes the clock via the resume mutation", async () => {
      const user = await setup();
      await user.click(document.querySelector('[data-qa="toggle-timer-btn"]')!);

      expect(resumeClock).toHaveBeenCalledWith("s1", 1);
    });

    it("adds a minute via the adjust mutation", async () => {
      const user = await setup();
      await user.click(document.querySelector('[data-qa="add-btn"]')!);

      expect(adjustClock).toHaveBeenCalledWith("s1", 1, 60);
    });

    it("subtracts a minute via the adjust mutation", async () => {
      const user = await setup();
      await user.click(document.querySelector('[data-qa="sub-btn"]')!);

      expect(adjustClock).toHaveBeenCalledWith("s1", 1, -60);
    });

    it("advances a level via the set-level mutation", async () => {
      const user = await setup();
      await user.click(document.querySelector('[data-qa="advance-level-btn"]')!);

      expect(setClockLevel).toHaveBeenCalledWith("s1", 1, 2);
    });

    it("returns to the previous level via the set-level mutation", async () => {
      const user = await setup();
      await user.click(document.querySelector('[data-qa="prev-level-btn"]')!);

      expect(setClockLevel).toHaveBeenCalledWith("s1", 1, 0);
    });
  });

  describe("offline indicator", () => {
    it("does not show a not-synced badge while polls are succeeding", async () => {
      (fetchClock as jest.Mock).mockResolvedValue(clockState());

      renderClock();
      await screen.findByText("Level 1");

      expect(document.querySelector('[data-qa="offline-badge"]')).not.toBeInTheDocument();
    });

    it("shows a not-synced badge after three consecutive failed polls, without stopping the clock", async () => {
      (fetchClock as jest.Mock)
        .mockResolvedValueOnce(clockState({ levelEndsAt: new Date(Date.now() + 10 * 60_000).toISOString() }))
        .mockRejectedValue(new Error("network error"));

      renderClock();
      await screen.findByText("Level 1");

      await waitFor(() => expect(document.querySelector('[data-qa="offline-badge"]')).toBeInTheDocument(), {
        timeout: 8000,
      });

      // Ticking locally off last known state the whole time — not a stalled display.
      expect(document.querySelector('[data-qa="timer"]')).not.toHaveTextContent("10:00");
    }, 10000);

    it("clears the badge as soon as a poll succeeds again", async () => {
      (fetchClock as jest.Mock)
        .mockResolvedValueOnce(clockState())
        .mockRejectedValueOnce(new Error("e1"))
        .mockRejectedValueOnce(new Error("e2"))
        .mockRejectedValueOnce(new Error("e3"))
        .mockResolvedValue(clockState({ levelIndex: 1 }));

      renderClock();
      await screen.findByText("Level 1");

      await waitFor(() => expect(document.querySelector('[data-qa="offline-badge"]')).toBeInTheDocument(), {
        timeout: 8000,
      });

      await waitFor(() => expect(document.querySelector('[data-qa="offline-badge"]')).not.toBeInTheDocument(), {
        timeout: 4000,
      });
    }, 15000);

    it("keeps the badge visible through a failed pause action taken while offline", async () => {
      (fetchClock as jest.Mock)
        .mockResolvedValueOnce(clockState({ pausedAt: null }))
        .mockRejectedValue(new Error("network error"));
      (pauseClock as jest.Mock).mockRejectedValue(new Error("network error"));

      renderClock();
      await screen.findByText("Level 1");

      await waitFor(() => expect(document.querySelector('[data-qa="offline-badge"]')).toBeInTheDocument(), {
        timeout: 8000,
      });

      const user = userEvent.setup();
      await user.click(document.querySelector('[data-qa="toggle-timer-btn"]')!);

      // The optimistic pause write (and its rollback) are local cache writes,
      // not contact with the server — they must not clear the badge.
      expect(document.querySelector('[data-qa="offline-badge"]')).toBeInTheDocument();
    }, 10000);
  });
});
