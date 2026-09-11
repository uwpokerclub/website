import { ReactNode } from "react";
import styles from "./CardEmptyState.module.css";

type CardEmptyStateProps = {
  message: ReactNode;
};

export function CardEmptyState({ message }: CardEmptyStateProps) {
  return (
    <div className={styles.container} data-qa="dashboard-card-empty">
      <p className={styles.message}>{message}</p>
    </div>
  );
}
