import { useEffect, useMemo, useRef } from "react";
import { Blind } from "@/types";
import { useDerivedClock } from "../hooks/useDerivedClock";
import { useClockSyncStatus } from "../hooks/useClockSyncStatus";
import {
  useAdjustClock,
  useEventClock,
  usePauseClock,
  useResumeClock,
  useSetClockLevel,
} from "../hooks/useClockQueries";
import { ClockDisplay } from "./ClockDisplay";
import { LevelInfo } from "./LevelInfo";
import { MS_IN_MINUTE } from "../utils/time";

import styles from "./TournamentClock.module.css";
import { Icon } from "../../../components";

const ADJUST_STEP_SECONDS = 60;

type Props = {
  semesterId: string;
  eventId: number;
  levels: Blind[];
};

/**
 * Formats a number for display, abbreviating values >= 10000 (5+ digits)
 * Examples: 1000 → "1000", 5000 → "5000", 10000 → "10K", 50000 → "50K", 100000 → "100K"
 */
function formatChipValue(value: number): string {
  if (value >= 10000) {
    const kValue = value / 1000;
    // Remove decimal if it's a whole number
    return kValue % 1 === 0 ? `${kValue}K` : `${kValue.toFixed(1)}K`;
  }
  return String(value);
}

export function TournamentClock({ semesterId, eventId, levels }: Props) {
  const levelDurationsMs = useMemo(() => levels.map((level) => level.time * MS_IN_MINUTE), [levels]);

  const {
    data: clockData,
    pollSuccessCount,
    pollFailureCount,
  } = useEventClock(levels.length > 0 ? semesterId : undefined, eventId);
  const derived = useDerivedClock(clockData, levelDurationsMs);
  const isOffline = useClockSyncStatus(pollSuccessCount, pollFailureCount);

  const pauseMutation = usePauseClock();
  const resumeMutation = useResumeClock();
  const adjustMutation = useAdjustClock();
  const setLevelMutation = useSetClockLevel();

  // Ref to hold the entire clock HTML elements used to enable fullscreen
  const clockElementRef = useRef<HTMLDivElement | null>(null);
  // Ref used to enable/disable screen lock
  const wakeLockRef = useRef<WakeLockSentinel | null>(null);

  // Handles acquiring a screen clock so the device does not go to sleep when the clock is running
  useEffect(() => {
    const requestWakeLock = async () => {
      if ("wakeLock" in navigator) {
        try {
          wakeLockRef.current = await navigator.wakeLock.request("screen");
        } catch (err) {
          console.error("Failed to acquire a screen lock: ", err);
        }
      } else {
        console.error("The Screen Wake Lock API is not supported");
      }
    };

    requestWakeLock();

    return () => {
      if (wakeLockRef.current) {
        wakeLockRef.current.release().then(() => {
          wakeLockRef.current = null;
        });
      }
    };
  }, []);

  const handleResume = () => resumeMutation.mutate({ semesterId, eventId });
  const handlePause = () => pauseMutation.mutate({ semesterId, eventId });

  const handlePreviousLevel = () => {
    if (!derived || derived.levelIndex === 0) return;
    setLevelMutation.mutate({ semesterId, eventId, index: derived.levelIndex - 1 });
  };

  const handleNextLevel = () => {
    if (!derived || derived.levelIndex === levels.length - 1) return;
    setLevelMutation.mutate({ semesterId, eventId, index: derived.levelIndex + 1 });
  };

  const handleAddTime = () => adjustMutation.mutate({ semesterId, eventId, deltaSeconds: ADJUST_STEP_SECONDS });
  const handleSubtractTime = () => adjustMutation.mutate({ semesterId, eventId, deltaSeconds: -ADJUST_STEP_SECONDS });

  // Handles enabling/disabling fullscreening the clock
  const handleFullscreen = () => {
    if (document.fullscreenElement !== null) {
      document.exitFullscreen();
    } else {
      clockElementRef.current!.requestFullscreen();
    }
  };

  // Render an empty state when there are no blinds (structure not loaded or empty)
  if (levels.length === 0) {
    return (
      <div className={`${styles.grid}`}>
        <header className={styles.timerHeader}>No blind structure available</header>
      </div>
    );
  }

  // The clock row is materialised lazily by the server on first read; until
  // then (or while the first poll is in flight) there's nothing to derive.
  if (!derived) {
    return (
      <div className={`${styles.grid}`}>
        <header className={styles.timerHeader}>Loading clock…</header>
      </div>
    );
  }

  const { remainingMs, pausedAt } = derived;
  // The structure can be edited out from under an in-progress event (e.g. an
  // admin shrinks the blind levels), leaving a stored levelIndex derive() has
  // no bound for beyond levels.length; clamp it here so indexing never throws.
  const levelIndex = Math.min(derived.levelIndex, levels.length - 1);

  return (
    <div ref={clockElementRef} className={`${styles.grid}`}>
      <span onClick={handleFullscreen} className={styles.fullscreen}>
        <Icon scale={2} iconType="expand" />
      </span>

      <header data-qa="level" className={styles.timerHeader}>
        Level {levelIndex + 1}
      </header>

      {isOffline && (
        // Not a claim the displayed time is wrong (deriveClock is exact) — only that
        // a control action from another room may not have been heard yet. See #455.
        <span data-qa="offline-badge" className={styles.offlineBadge} title="Not receiving updates from the server">
          Not synced
        </span>
      )}

      <ClockDisplay
        remainingMs={remainingMs}
        totalMs={levelDurationsMs[levelIndex]}
        isPaused={pausedAt !== null}
        onResume={handleResume}
        onPause={handlePause}
        onPreviousLevel={handlePreviousLevel}
        onNextLevel={handleNextLevel}
        onSubtractTime={handleSubtractTime}
        onAddTime={handleAddTime}
      />

      <LevelInfo
        type="blinds"
        title="Blinds"
        current={`${formatChipValue(levels[levelIndex].small)} / ${formatChipValue(levels[levelIndex].big)}`}
        next={
          levelIndex < levels.length - 1
            ? `${formatChipValue(levels[levelIndex + 1].small)} / ${formatChipValue(levels[levelIndex + 1].big)}`
            : undefined
        }
      />

      <LevelInfo
        type="ante"
        title="Ante"
        current={formatChipValue(levels[levelIndex].ante)}
        next={levelIndex < levels.length - 1 ? formatChipValue(levels[levelIndex + 1].ante) : undefined}
      />
    </div>
  );
}
