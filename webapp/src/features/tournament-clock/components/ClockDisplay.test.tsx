/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";
import { ClockDisplay } from "./ClockDisplay";

jest.mock("../utils/playSound", () => ({ playSound: jest.fn() }));
// The real barrel transitively imports modules that read `import.meta.env`,
// which Jest's CommonJS transform can't parse (see useClockQueries.test.ts).
jest.mock("../../../components", () => ({
  Icon: ({ iconType }: { iconType: string }) => <span data-qa={`icon-${iconType}`} />,
}));

import { playSound } from "../utils/playSound";

class FakeAudioContext {
  currentTime = 0;
  createOscillator = jest.fn();
  createGain = jest.fn();
  destination = {};
}

function baseProps(overrides: Partial<React.ComponentProps<typeof ClockDisplay>> = {}) {
  return {
    remainingMs: 5 * 60_000,
    totalMs: 10 * 60_000,
    isPaused: true,
    onResume: jest.fn(),
    onPause: jest.fn(),
    onPreviousLevel: jest.fn(),
    onNextLevel: jest.fn(),
    onSubtractTime: jest.fn(),
    onAddTime: jest.fn(),
    ...overrides,
  };
}

describe("ClockDisplay", () => {
  beforeEach(() => {
    (global as unknown as { AudioContext: unknown }).AudioContext = FakeAudioContext;
    (playSound as jest.Mock).mockClear();
  });

  it("renders the remaining time it is given, in minutes and seconds", () => {
    const { container } = render(<ClockDisplay {...baseProps({ remainingMs: 2 * 60_000 + 5_000 })} />);

    expect(container.querySelector('[data-qa="timer"]')).toHaveTextContent("2:05");
  });

  it("renders the progress bar as the fraction of the level remaining", () => {
    render(<ClockDisplay {...baseProps({ remainingMs: 3 * 60_000, totalMs: 10 * 60_000 })} />);

    expect(screen.getByRole("progressbar")).toHaveValue(0.3);
  });

  it("creates an AudioContext and calls onResume when the start button is clicked while paused", async () => {
    const user = userEvent.setup();
    const onResume = jest.fn();
    render(<ClockDisplay {...baseProps({ isPaused: true, onResume })} />);

    await user.click(document.querySelector('[data-qa="toggle-timer-btn"]')!);

    expect(onResume).toHaveBeenCalledTimes(1);
  });

  it("calls onPause when the pause button is clicked while running", async () => {
    const user = userEvent.setup();
    const onPause = jest.fn();
    render(<ClockDisplay {...baseProps({ isPaused: false, onPause })} />);

    await user.click(document.querySelector('[data-qa="toggle-timer-btn"]')!);

    expect(onPause).toHaveBeenCalledTimes(1);
  });

  // The clock must have been started at least once (a real user gesture) before
  // beeps can play, since that's what creates the AudioContext.
  async function renderRunning(remainingMs: number) {
    const user = userEvent.setup();
    const utils = render(<ClockDisplay {...baseProps({ isPaused: true, remainingMs })} />);
    await user.click(document.querySelector('[data-qa="toggle-timer-btn"]')!);
    utils.rerender(<ClockDisplay {...baseProps({ isPaused: false, remainingMs })} />);
    return utils;
  }

  it("plays the low pitch warning beep once time is down to the last 5 seconds of a minute", async () => {
    const { rerender } = await renderRunning(10_000);
    rerender(<ClockDisplay {...baseProps({ isPaused: false, remainingMs: 3_000 })} />);

    expect(playSound).toHaveBeenCalledWith(expect.anything(), 493.883, expect.anything(), 0.15);
  });

  it("plays the high pitch end beep once remaining time reaches zero", async () => {
    const { rerender } = await renderRunning(3_000);
    rerender(<ClockDisplay {...baseProps({ isPaused: false, remainingMs: 0 })} />);

    expect(playSound).toHaveBeenCalledWith(expect.anything(), 659.255, expect.anything(), 0.25);
  });

  it("does not play a beep while paused", async () => {
    const { rerender } = await renderRunning(3_000);
    (playSound as jest.Mock).mockClear();

    rerender(<ClockDisplay {...baseProps({ isPaused: true, remainingMs: 0 })} />);

    expect(playSound).not.toHaveBeenCalled();
  });
});
