import { Ranking } from "@/types";
import styles from "./RankingsPodium.module.css";

type RankingsPodiumProps = {
  rankings: Ranking[];
};

export function RankingsPodium({ rankings }: RankingsPodiumProps) {
  if (rankings.length < 3) {
    return null;
  }

  const [first, second, third] = rankings.slice(0, 3);

  return (
    <div className={styles.podiumContainer} data-qa="rankings-podium">
      {/* 2nd place - left */}
      <div className={`${styles.podiumPosition} ${styles.second}`} data-qa="podium-position-2">
        <div className={styles.memberInfo}>
          <span className={styles.memberName} data-qa="podium-name-2">
            {second.firstName} {second.lastName}
          </span>
          <span className={styles.points} data-qa="podium-points-2">
            {second.points} pts
          </span>
        </div>
        <div className={styles.podiumBase}>
          <span className={styles.rank}>2</span>
        </div>
      </div>

      {/* 1st place - center (tallest) */}
      <div className={`${styles.podiumPosition} ${styles.first}`} data-qa="podium-position-1">
        <div className={styles.memberInfo}>
          <span className={styles.memberName} data-qa="podium-name-1">
            {first.firstName} {first.lastName}
          </span>
          <span className={styles.points} data-qa="podium-points-1">
            {first.points} pts
          </span>
        </div>
        <div className={styles.podiumBase}>
          <span className={styles.rank}>1</span>
        </div>
      </div>

      {/* 3rd place - right */}
      <div className={`${styles.podiumPosition} ${styles.third}`} data-qa="podium-position-3">
        <div className={styles.memberInfo}>
          <span className={styles.memberName} data-qa="podium-name-3">
            {third.firstName} {third.lastName}
          </span>
          <span className={styles.points} data-qa="podium-points-3">
            {third.points} pts
          </span>
        </div>
        <div className={styles.podiumBase}>
          <span className={styles.rank}>3</span>
        </div>
      </div>
    </div>
  );
}
