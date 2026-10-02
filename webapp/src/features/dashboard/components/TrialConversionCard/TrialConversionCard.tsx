import { DashboardCard } from "../DashboardCard";
import { StatBar } from "../StatBar";
import { CARD_TITLES } from "../../dashboardLayout";
import { useTrialConversion } from "../../hooks/useDashboardQueries";
import { SAMPLE_TERM } from "../../fixtures/sampleTerm";
import type { ConversionResponse } from "../../api/dashboardApi";
import styles from "./TrialConversionCard.module.css";

/**
 * GET …/dashboard/conversion does not exist yet, so the card falls back to the sample
 * term behind a visible marker rather than showing a permanent error — the page is
 * designed and reviewable before its endpoint lands.
 *
 * Real data always wins when it is present, so this needs no flipping when the
 * endpoint ships: the response arrives, the fallback stops being used and the marker
 * disappears on its own. Set to false to see the real loading and error states.
 */
const FALL_BACK_TO_SAMPLE = true;

type Props = { semesterId: string };

export function TrialConversionCard({ semesterId }: Props) {
  const { data, isLoading, isError, refetch } = useTrialConversion(semesterId);
  const usingSample = FALL_BACK_TO_SAMPLE && !data;
  const resolved = data ?? (FALL_BACK_TO_SAMPLE ? SAMPLE_TERM.conversion : undefined);

  const status = usingSample
    ? "ready"
    : isError
      ? "error"
      : isLoading || !resolved
        ? "loading"
        : resolved.current.players === 0
          ? "empty"
          : "ready";

  return (
    <DashboardCard
      title={CARD_TITLES.trialConversion}
      status={status}
      sampleData={usingSample}
      onRetry={() => refetch()}
      emptyMessage="No event entries recorded yet this term."
      data-qa="trial-conversion-card"
    >
      {() => <Body data={resolved as ConversionResponse} />}
    </DashboardCard>
  );
}

function Body({ data }: { data: ConversionResponse }) {
  const { current, freeTrialLimit } = data;

  // free_trial_limit defaults to 0, which disables the trial entirely. A term with no
  // trial has no funnel, and four zeroes would read as a catastrophic one.
  if (freeTrialLimit === 0) {
    return <p className={styles.disabled}>Free trial is not enabled this term.</p>;
  }

  return (
    <div className={styles.container}>
      <p className={styles.lead}>
        <span className={styles.figure} data-testid="trial-conversion-figure">
          {current.trialSpent.toLocaleString("en-CA")}
        </span>
      </p>
      <p className={styles.caption}>spent all {freeTrialLimit} free entries and never bought a membership</p>

      <StatBar
        ariaLabel="Players by membership status"
        segments={[
          { key: "paid", label: "paid", value: current.paid, color: "var(--dash-ord-3)" },
          { key: "spent", label: "trial spent, unpaid", value: current.trialSpent, color: "var(--dash-cat-2)" },
          { key: "open", label: "trial still open", value: current.trialOpen, color: "var(--dash-muted)" },
          { key: "exec", label: "executive, comped", value: current.executive, color: "var(--dash-muted-soft)" },
        ]}
      />

      <p className={styles.foot}>
        A snapshot of where players stand now, not a conversion rate — a membership that converts loses the record that
        it ever trialled.
      </p>
    </div>
  );
}
