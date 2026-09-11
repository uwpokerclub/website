import styles from "./CardSkeleton.module.css";

type CardSkeletonProps = {
  label: string;
};

export function CardSkeleton({ label }: CardSkeletonProps) {
  return (
    <div
      className={styles.skeleton}
      role="status"
      aria-busy="true"
      aria-label={label}
      data-qa="dashboard-card-skeleton"
    >
      <div className={styles.line} />
      <div className={styles.line} />
      <div className={styles.lineShort} />
    </div>
  );
}
