import { useCallback, useEffect, useRef } from "react";
import { ClockActions } from "./ClockActions";
import { playSound } from "../utils/playSound";
import { MS_IN_MINUTE, MS_IN_SECOND, SECONDS_IN_MINUTE } from "../utils/time";

import styles from "./ClockDisplay.module.css";

type Props = {
  // Remaining time in the current level, derived from server state (ms).
  remainingMs: number;
  // The current level's total duration, for the progress bar (ms).
  totalMs: number;
  isPaused: boolean;
  onResume: () => void;
  onPause: () => void;
  onPreviousLevel: () => void;
  onNextLevel: () => void;
  onSubtractTime: () => void;
  onAddTime: () => void;
};

const LOW_PITCH_BEEP = 493.883;
const HIGH_PITCH_BEEP = 659.255;

export function ClockDisplay({
  remainingMs,
  totalMs,
  isPaused,
  onResume,
  onPause,
  onPreviousLevel,
  onNextLevel,
  onSubtractTime,
  onAddTime,
}: Props) {
  // Ref to hold the AudioContext instance, created lazily on the first resume
  // click so it happens within a user gesture (required by browser autoplay policy).
  const audioContextRef = useRef<AudioContext | null>(null);

  const minutes = Math.floor(remainingMs / MS_IN_MINUTE);
  const seconds = Math.floor((remainingMs / MS_IN_SECOND) % SECONDS_IN_MINUTE);

  const handleStart = useCallback(() => {
    if (!audioContextRef.current) {
      audioContextRef.current = new AudioContext();
    }
    onResume();
  }, [onResume]);

  // Plays the level's warning/end beeps off the locally derived remaining time.
  useEffect(() => {
    if (isPaused || !audioContextRef.current) return;

    if (minutes === 0 && seconds >= 1 && seconds <= 5) {
      playSound(audioContextRef.current, LOW_PITCH_BEEP, audioContextRef.current.currentTime, 0.15);
    } else if (minutes <= 0 && seconds <= 0) {
      playSound(audioContextRef.current, HIGH_PITCH_BEEP, audioContextRef.current.currentTime, 0.25);
    }
  }, [isPaused, minutes, seconds]);

  return (
    <section className={styles.container}>
      <div data-qa="timer" className={styles.timer}>
        {minutes}:{seconds < 10 ? `0${seconds}` : seconds}
      </div>

      <ClockActions
        isPaused={isPaused}
        onStart={handleStart}
        onPause={onPause}
        onStepBack={onPreviousLevel}
        onStepForward={onNextLevel}
        onSubtractTime={onSubtractTime}
        onAddTime={onAddTime}
      />

      <progress className={styles.progress} max={1} value={totalMs > 0 ? remainingMs / totalMs : 0} />
    </section>
  );
}
