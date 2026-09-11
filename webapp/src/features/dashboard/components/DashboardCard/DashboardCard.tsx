import { ReactNode } from "react";
import { CardSkeleton } from "./CardSkeleton";
import { CardErrorState } from "./CardErrorState";
import { CardEmptyState } from "./CardEmptyState";
import styles from "./DashboardCard.module.css";

export type DashboardCardStatus = "loading" | "error" | "empty" | "ready";

type DashboardCardProps = {
  title: string;
  action?: ReactNode;
  status: DashboardCardStatus;
  errorMessage?: string;
  onRetry?: () => void;
  emptyMessage?: ReactNode;
  "data-qa"?: string;
  /**
   * Rendered only when status is "ready" - kept as a function so callers can
   * reference data that only exists once loading has finished.
   */
  children: () => ReactNode;
};

export function DashboardCard({
  title,
  action,
  status,
  errorMessage = "Something went wrong.",
  onRetry,
  emptyMessage = "Nothing to show yet.",
  "data-qa": dataQa,
  children,
}: DashboardCardProps) {
  return (
    <section className={styles.container} data-qa={dataQa}>
      <header className={styles.header}>
        <h3 className={styles.title}>{title}</h3>
        {action && <div className={styles.action}>{action}</div>}
      </header>
      <div className={styles.body}>
        {status === "loading" && <CardSkeleton label={`Loading ${title}`} />}
        {status === "error" && <CardErrorState message={errorMessage} onRetry={onRetry} />}
        {status === "empty" && <CardEmptyState message={emptyMessage} />}
        {status === "ready" && children()}
      </div>
    </section>
  );
}
