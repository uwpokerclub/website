import { FaExclamationTriangle } from "react-icons/fa";
import { Button } from "@uwpokerclub/components";
import styles from "./CardErrorState.module.css";

type CardErrorStateProps = {
  message: string;
  onRetry?: () => void;
};

export function CardErrorState({ message, onRetry }: CardErrorStateProps) {
  return (
    <div className={styles.container} role="alert" data-qa="dashboard-card-error">
      <FaExclamationTriangle className={styles.icon} aria-hidden="true" />
      <p className={styles.message}>{message}</p>
      {onRetry && (
        <Button variant="secondary" size="small" onClick={onRetry} data-qa="dashboard-card-retry">
          Retry
        </Button>
      )}
    </div>
  );
}
