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
    <div className={styles.podiumContainer}>
      {/* 2nd place - left */}
      <div className={`${styles.podiumPosition} ${styles.second}`}>
        <div className={styles.memberInfo}>
          <span className={styles.memberName}>
            {second.firstName} {second.lastName}
          </span>
          <span className={styles.points}>{second.points} pts</span>
        </div>
        <div className={styles.podiumBase}>
          <span className={styles.rank}>2</span>
        </div>
      </div>

      {/* 1st place - center (tallest) */}
      <div className={`${styles.podiumPosition} ${styles.first}`}>
        <div className={styles.memberInfo}>
          <span className={styles.memberName}>
            {first.firstName} {first.lastName}
          </span>
          <span className={styles.points}>{first.points} pts</span>
        </div>
        <div className={styles.podiumBase}>
          <span className={styles.rank}>1</span>
        </div>
      </div>

      {/* 3rd place - right */}
      <div className={`${styles.podiumPosition} ${styles.third}`}>
        <div className={styles.memberInfo}>
          <span className={styles.memberName}>
            {third.firstName} {third.lastName}
          </span>
          <span className={styles.points}>{third.points} pts</span>
        </div>
        <div className={styles.podiumBase}>
          <span className={styles.rank}>3</span>
        </div>
      </div>
    </div>
  );
}
