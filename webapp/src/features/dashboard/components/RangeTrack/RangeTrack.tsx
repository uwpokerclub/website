import styles from "./RangeTrack.module.css";

type RangeTrackProps = {
  label: string;
  value: number;
  display: string;
  min: number;
  max: number;
  bandLow: number;
  bandHigh: number;
  comparison?: { value: number; label: string };
};

type Verdict = "below usual" | "normal" | "above usual";

function verdictFor(value: number, bandLow: number, bandHigh: number): Verdict {
  if (value < bandLow) return "below usual";
  if (value > bandHigh) return "above usual";
  return "normal";
}

const pct = (value: number, min: number, max: number) =>
  `${Math.min(100, Math.max(0, ((value - min) / (max - min)) * 100))}%`;

/**
 * One figure placed against the range it normally falls in.
 *
 * This exists because a bare "median 2 events, down 1" reads as a collapse during an
 * entirely ordinary term — median has been 2–3 in all six analysed terms. Position
 * against the band is what makes the figure legible without the interface having to
 * editorialise in a sentence.
 */
export function RangeTrack({ label, value, display, min, max, bandLow, bandHigh, comparison }: RangeTrackProps) {
  const verdict = verdictFor(value, bandLow, bandHigh);
  const spoken =
    `${display}, ${verdict === "normal" ? "normal for this club" : verdict} — ` +
    `the usual range is ${bandLow} to ${bandHigh}.` +
    (comparison ? ` ${comparison.label} was ${comparison.value}.` : "");

  return (
    <div className={styles.row}>
      <span className={styles.label}>{label}</span>
      <span className={styles.value}>{display}</span>
      <span className={styles.track} aria-hidden="true">
        <i
          className={styles.band}
          style={{ left: pct(bandLow, min, max), right: `calc(100% - ${pct(bandHigh, min, max)})` }}
        />
        {comparison && <i className={styles.previous} style={{ left: pct(comparison.value, min, max) }} />}
        <i className={styles.mark} style={{ left: pct(value, min, max) }} />
      </span>
      <span className={styles.verdict} data-verdict={verdict} aria-hidden="true">
        {verdict}
      </span>
      <span className={styles.srOnly}>{spoken}</span>
    </div>
  );
}
