import styles from "./StatBar.module.css";

export type StatBarSegment = {
  key: string;
  label: string;
  value: number;
  color: string;
};

type StatBarProps = {
  segments: StatBarSegment[];
  ariaLabel: string;
};

/**
 * A horizontal part-to-whole bar with a labelled legend. Segments are sized by
 * flex-grow, so the bar is fluid at any width with no measurement and no library.
 *
 * Every segment carries a visible value in the legend — identity is never colour
 * alone, and the gold categorical slot sits below 3:1 against the card surface, so
 * its label is mandatory rather than a nicety.
 */
export function StatBar({ segments, ariaLabel }: StatBarProps) {
  const visible = segments.filter((segment) => segment.value > 0);

  return (
    <div className={styles.container}>
      <div className={styles.track} data-testid="statbar-track" aria-hidden="true">
        {visible.map((segment) => (
          <span key={segment.key} style={{ flexGrow: segment.value, background: segment.color }} />
        ))}
      </div>
      <ul className={styles.legend} aria-label={ariaLabel}>
        {segments.map((segment) => (
          <li key={segment.key}>
            <span className={styles.swatch} style={{ background: segment.color }} aria-hidden="true" />
            <b>{segment.value.toLocaleString("en-CA")}</b> {segment.label}
          </li>
        ))}
      </ul>
    </div>
  );
}
