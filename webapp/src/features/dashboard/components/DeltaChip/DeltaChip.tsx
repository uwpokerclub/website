import styles from "./DeltaChip.module.css";

export type DeltaSentiment = "positive-is-good" | "negative-is-good" | "neutral";
export type DeltaChipMode = "count" | "percentage-points";

type DeltaChipProps = {
  current: number;
  comparison: number;
  comparisonLabel: string;
  sentiment: DeltaSentiment;
  compact?: boolean;
  mode?: DeltaChipMode;
};

type Direction = "up" | "down" | "none";
type Tone = "positive" | "negative" | "neutral";

function resolveDirection(current: number, comparison: number): Direction {
  if (current === comparison) return "none";
  return current > comparison ? "up" : "down";
}

function resolveTone(direction: Direction, sentiment: DeltaSentiment): Tone {
  if (direction === "none" || sentiment === "neutral") return "neutral";
  const upIsGood = sentiment === "positive-is-good";
  const isGood = direction === "up" ? upIsGood : !upIsGood;
  return isGood ? "positive" : "negative";
}

function rawPercentChange(current: number, comparison: number): number {
  return (Math.abs(current - comparison) / Math.abs(comparison)) * 100;
}

// A real (non-zero) change can still round to 0%, which would misleadingly
// read as "no change" — show "<1%" (or "less than 1%" spoken) instead.
function percentDisplay(current: number, comparison: number): string {
  const rounded = Math.round(rawPercentChange(current, comparison));
  return rounded === 0 ? "<1%" : `${rounded}%`;
}

function percentSpoken(current: number, comparison: number): string {
  const rounded = Math.round(rawPercentChange(current, comparison));
  return rounded === 0 ? "less than 1%" : `${rounded}%`;
}

function formatPercentagePoints(value: number): string {
  const percentagePoints = Math.abs(value) * 100;

  if (percentagePoints > 0 && percentagePoints < 0.1) return "<0.1";

  return new Intl.NumberFormat("en-CA", { maximumFractionDigits: 1 }).format(percentagePoints);
}

function formatVisible(current: number, comparison: number, direction: Direction, mode: DeltaChipMode): string {
  if (direction === "none") return "No change";

  const arrow = direction === "up" ? "▲" : "▼";
  const sign = direction === "up" ? "+" : "−";
  const delta = Math.abs(current - comparison);

  if (mode === "percentage-points") return `${arrow} ${sign}${formatPercentagePoints(delta)} pp`;

  if (comparison === 0) return `${arrow} ${sign}${delta}`;

  return `${arrow} ${sign}${delta} (${sign}${percentDisplay(current, comparison)})`;
}

function formatAriaLabel(
  current: number,
  comparison: number,
  direction: Direction,
  comparisonLabel: string,
  mode: DeltaChipMode,
): string {
  if (direction === "none") return `No change from ${comparisonLabel}`;

  const verb = direction === "up" ? "Up" : "Down";
  const delta = Math.abs(current - comparison);

  if (mode === "percentage-points") {
    return `${verb} ${formatPercentagePoints(delta)} percentage points from ${comparisonLabel}`;
  }

  if (comparison === 0) return `${verb} ${delta} from ${comparisonLabel}`;

  return `${verb} ${delta} (${percentSpoken(current, comparison)}) from ${comparisonLabel}`;
}

export function DeltaChip({
  current,
  comparison,
  comparisonLabel,
  sentiment,
  compact = false,
  mode = "count",
}: DeltaChipProps) {
  const direction = resolveDirection(current, comparison);
  const tone = resolveTone(direction, sentiment);
  const visible = formatVisible(current, comparison, direction, mode);
  const ariaLabel = formatAriaLabel(current, comparison, direction, comparisonLabel, mode);
  const visibleText = compact ? visible : `${visible} vs ${comparisonLabel}`;

  return (
    <span className={`${styles.chip} ${styles[tone]}`} data-qa="delta-chip" data-tone={tone}>
      <span aria-hidden="true">{visibleText}</span>
      <span className={styles.srOnly}>{ariaLabel}</span>
    </span>
  );
}
