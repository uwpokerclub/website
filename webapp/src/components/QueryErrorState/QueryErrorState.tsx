import { FaExclamationTriangle } from "react-icons/fa";
import { Button } from "@uwpokerclub/components";
import styles from "./QueryErrorState.module.css";

interface QueryErrorStateProps {
  title: string;
  message: string;
  onRetry: () => void;
  "data-qa"?: string;
  retryDataQa?: string;
}

export function QueryErrorState({
  title,
  message,
  onRetry,
  "data-qa": dataQa,
  retryDataQa = "retry-btn",
}: QueryErrorStateProps) {
  return (
    <div className={styles.container} data-qa={dataQa}>
      <div className={styles.icon}>
        <FaExclamationTriangle />
      </div>
      <h3 className={styles.title}>{title}</h3>
      <p className={styles.message}>{message}</p>
      <Button onClick={onRetry} data-qa={retryDataQa}>
        Retry
      </Button>
    </div>
  );
}
