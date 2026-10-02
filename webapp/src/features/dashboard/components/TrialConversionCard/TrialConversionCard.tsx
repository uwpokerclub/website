import { DashboardCard } from "../DashboardCard";
import { StatBar } from "../StatBar";
import { CARD_TITLES } from "../../dashboardLayout";
import { useTrialConversion } from "../../hooks/useDashboardQueries";
import type { ConversionResponse } from "../../api/dashboardApi";
import styles from "./TrialConversionCard.module.css";

/**
 * The endpoint pairs the current status snapshot with a separately tracked
 * observed trial cohort. Empty snapshots can still have retained cohort history.
 */
type Props = { semesterId: string };

export function TrialConversionCard({ semesterId }: Props) {
  const { data, isPending, isError, refetch } = useTrialConversion(semesterId);
  const status = isPending ? "loading" : isError ? "error" : !data ? "loading" : "ready";

  return (
    <DashboardCard
      title={CARD_TITLES.trialConversion}
      status={status}
      onRetry={() => refetch()}
      emptyMessage="No trial conversion data recorded yet."
      data-qa="trial-conversion-card"
    >
      {() => <Body data={data as ConversionResponse} />}
    </DashboardCard>
  );
}

function Body({ data }: { data: ConversionResponse }) {
  const { current, conversion, freeTrialLimit, comparison } = data;

  // free_trial_limit defaults to 0, which disables the trial entirely. A term with no
  // trial has no funnel, and four zeroes would read as a catastrophic one.
  return (
    <div className={styles.container}>
      {freeTrialLimit === 0 ? (
        <p className={styles.disabled}>Free trial is not enabled this term.</p>
      ) : current.players === 0 ? (
        <p className={styles.disabled}>No current event entries to show in the status snapshot.</p>
      ) : (
        <>
          <p className={styles.lead}>
            <span className={styles.figure} data-testid="trial-conversion-figure">
              {current.trialSpent.toLocaleString("en-CA")}
            </span>
          </p>
          <p className={styles.caption}>spent all {freeTrialLimit} free entries and are currently unpaid</p>

          <StatBar ariaLabel="Players by membership status" segments={trialSegments(current)} />
        </>
      )}

      <ConversionSummary conversion={conversion} />

      <Comparison comparison={comparison} />

      <p className={styles.foot}>
        Status is a current snapshot. The observed-ever conversion rate covers only trial starts tracked from rollout;
        untracked entrants include upfront buyers and legacy unknowns, so this is not a whole-term rate. Paid status is
        the purchase proxy; later refunds or status reversals do not erase a recorded conversion.
      </p>
    </div>
  );
}

function Comparison({ comparison }: { comparison: ConversionResponse["comparison"] }) {
  if (!comparison) {
    return <p className={styles.noComparison}>No comparable term to compare against yet.</p>;
  }

  const { semester, stats, conversion, freeTrialLimit } = comparison;
  const trialDisabled = freeTrialLimit === 0;

  return (
    <div className={styles.comparison}>
      <p className={styles.comparisonTitle}>{semester.name}</p>
      <p className={styles.comparisonContext}>
        {trialDisabled
          ? "Free trial was not enabled; unpaid players are not shown as having exhausted a trial."
          : `Free trial limit: ${freeTrialLimit} entries.`}
      </p>
      <StatBar
        ariaLabel={`Players by membership status in ${semester.name}`}
        segments={
          trialDisabled
            ? [
                { key: "paid", label: "paid", value: stats.paid, color: "var(--dash-ord-3)" },
                {
                  key: "unpaid-no-trial",
                  label: "unpaid, no trial",
                  value: stats.trialSpent,
                  color: "var(--dash-cat-2)",
                },
                { key: "exec", label: "executive, comped", value: stats.executive, color: "var(--dash-muted-soft)" },
              ]
            : trialSegments(stats)
        }
      />
      <ConversionSummary conversion={conversion} />
    </div>
  );
}

function ConversionSummary({ conversion }: { conversion: ConversionResponse["conversion"] }) {
  const rate =
    conversion.rate === null
      ? "Not yet measurable"
      : new Intl.NumberFormat("en-CA", { style: "percent", maximumFractionDigits: 1 }).format(conversion.rate);

  return (
    <div className={styles.conversion} data-qa="trial-conversion-rate">
      <p className={styles.conversionRate} data-testid="trial-conversion-rate-value">
        {rate}
      </p>
      <p className={styles.comparisonContext}>
        Observed-ever conversion: {conversion.numerator} of {conversion.denominator} tracked trial starters
      </p>
      <p className={styles.comparisonContext}>
        {conversion.untrackedEntrants} untracked entrants (upfront buyers and legacy unknowns; not in the rate)
      </p>
    </div>
  );
}

function trialSegments(stats: ConversionResponse["current"]) {
  return [
    { key: "paid", label: "paid", value: stats.paid, color: "var(--dash-ord-3)" },
    { key: "spent", label: "trial spent, unpaid", value: stats.trialSpent, color: "var(--dash-cat-2)" },
    { key: "open", label: "trial still open", value: stats.trialOpen, color: "var(--dash-muted)" },
    { key: "exec", label: "executive, comped", value: stats.executive, color: "var(--dash-muted-soft)" },
  ];
}
