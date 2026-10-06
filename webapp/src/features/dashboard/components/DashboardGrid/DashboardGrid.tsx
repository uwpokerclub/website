import { CSSProperties, ReactNode } from "react";
import type { PlacedCard } from "../../dashboardRows";
import styles from "./DashboardGrid.module.css";

export type DashboardGridItem = PlacedCard & { content: ReactNode };

type DashboardGridProps = {
  items: DashboardGridItem[];
  "data-qa"?: string;
};

/**
 * One grid for every card, in the order given. Each cell carries its spans as custom
 * properties and the stylesheet's container queries choose which applies, so the
 * layout follows the room the dashboard actually has rather than the viewport.
 */
export function DashboardGrid({ items, "data-qa": dataQa }: DashboardGridProps) {
  return (
    <div className={styles.container} data-qa={dataQa}>
      <div className={styles.grid}>
        {items.map((item) => (
          <div
            key={item.id}
            className={item.lead ? `${styles.cell} ${styles.lead}` : styles.cell}
            style={
              {
                "--dashboard-span-medium": item.spans.medium,
                "--dashboard-span-wide": item.spans.wide,
              } as CSSProperties
            }
            data-qa={item.lead ? "dashboard-lead" : "dashboard-cell"}
            data-testid={item.lead ? "dashboard-lead" : undefined}
            data-card={item.id}
            data-span-medium={item.spans.medium}
            data-span-wide={item.spans.wide}
          >
            {item.content}
          </div>
        ))}
      </div>
    </div>
  );
}
