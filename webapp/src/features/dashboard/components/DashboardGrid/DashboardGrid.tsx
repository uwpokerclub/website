import { ReactNode } from "react";
import styles from "./DashboardGrid.module.css";

type DashboardGridProps = {
  lead: ReactNode;
  wide: ReactNode;
  rail: ReactNode;
  "data-qa"?: string;
};

export function DashboardGrid({ lead, wide, rail, "data-qa": dataQa }: DashboardGridProps) {
  return (
    <div className={styles.container} data-qa={dataQa}>
      <div className={styles.grid}>
        <div className={styles.wideLane}>
          <div className={styles.lead} data-testid="dashboard-lead">
            {lead}
          </div>
          <div className={styles.stack} data-testid="dashboard-wide">
            {wide}
          </div>
        </div>
        <div className={styles.rail} data-testid="dashboard-rail">
          {rail}
        </div>
      </div>
    </div>
  );
}
