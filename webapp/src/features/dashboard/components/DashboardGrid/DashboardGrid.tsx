import { ReactNode } from "react";
import styles from "./DashboardGrid.module.css";

type DashboardGridProps = {
  children: ReactNode;
  "data-qa"?: string;
};

export function DashboardGrid({ children, "data-qa": dataQa }: DashboardGridProps) {
  return (
    <div className={styles.grid} data-qa={dataQa}>
      {children}
    </div>
  );
}
