import { ReactNode } from "react";
import { FaInbox } from "react-icons/fa";
import styles from "./CardEmptyState.module.css";

type CardEmptyStateProps = {
  message: ReactNode;
};

export function CardEmptyState({ message }: CardEmptyStateProps) {
  return (
    <div className={styles.container} data-qa="dashboard-card-empty">
      <FaInbox className={styles.icon} aria-hidden="true" />
      <p className={styles.message}>{message}</p>
    </div>
  );
}
