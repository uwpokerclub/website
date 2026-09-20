import { useAuth } from "@/hooks/useAuth";
import { Icon } from "../../../components";

import styles from "./ClockActions.module.css";

type Props = {
  isPaused: boolean;
  onStart: () => void;
  onPause: () => void;
  onStepBack: () => void;
  onStepForward: () => void;
  onSubtractTime: () => void;
  onAddTime: () => void;
};

export function ClockActions({
  isPaused,
  onStart,
  onPause,
  onStepBack,
  onStepForward,
  onSubtractTime,
  onAddTime,
}: Props) {
  const { hasPermission } = useAuth();

  // A mirror screen run by an executive has no control permission; the bar
  // is absent there rather than shown-but-disabled, per #455.
  if (!hasPermission("control", "event", "clock")) {
    return null;
  }

  return (
    <div className={styles.row}>
      <span data-qa="prev-level-btn" onClick={onStepBack}>
        <Icon iconType="backward_step" scale={1.5} />
      </span>
      <span data-qa="sub-btn" onClick={onSubtractTime}>
        <Icon iconType="minus" scale={1.5} />
      </span>
      <span data-qa="toggle-timer-btn" onClick={isPaused ? onStart : onPause}>
        {isPaused ? <Icon iconType="circle-play" scale={2} /> : <Icon iconType="circle-pause" scale={2} />}
      </span>
      <span data-qa="add-btn" onClick={onAddTime}>
        <Icon iconType="plus" scale={1.5} />
      </span>
      <span data-qa="advance-level-btn" onClick={onStepForward}>
        <Icon iconType="forward_step" scale={1.5} />
      </span>
    </div>
  );
}
