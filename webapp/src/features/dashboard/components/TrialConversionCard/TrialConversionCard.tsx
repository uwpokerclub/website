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
        <p className={styles.disabled}>Free trials were not offered this term.</p>
      ) : current.players === 0 ? (
        <p className={styles.disabled}>No current event entries to show in the status snapshot.</p>
      ) : (
        <>
          <p className={styles.lead}>
            <span className={styles.figure} data-testid="trial-conversion-figure">
              {current.trialSpent.toLocaleString("en-CA")} {current.trialSpent === 1 ? "player" : "players"}
            </span>
          </p>
          <p className={styles.caption}>
            used all {freeTrialLimit} free {freeTrialLimit === 1 ? "entry" : "entries"} and{" "}
            {current.trialSpent === 1 ? "is" : "are"} still unpaid
          </p>

          <StatBar ariaLabel="Players by membership status" segments={trialSegments(current)} />
        </>
      )}

      <p className={styles.comparisonContext}>
        Full-term observed trial cohort · paid status in the current historical snapshot
      </p>
      <ConversionSummary conversion={conversion} noRateMessage="No tracked trial players yet" />

      <Comparison comparison={comparison} />

      <p className={styles.foot}>
        Conversion uses all trials observed since tracking began, including players who later changed membership status.
        Paid and executive counts reflect the current historical snapshot, including refunds; they do not reconstruct
        payment status at term end. Earlier untracked trial and payment history is unavailable.
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
          ? "Full-term snapshot · Free trials were not offered this term."
          : `Full-term snapshot · Free trial limit: ${freeTrialLimit} ${freeTrialLimit === 1 ? "entry" : "entries"}.`}
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
      <ConversionSummary conversion={conversion} noRateMessage="Conversion history unavailable" />
    </div>
  );
}

function ConversionSummary({
  conversion,
  noRateMessage,
}: {
  conversion: ConversionResponse["conversion"];
  noRateMessage: string;
}) {
  const rate =
    conversion.rate === null
      ? noRateMessage
      : new Intl.NumberFormat("en-CA", { style: "percent", maximumFractionDigits: 1 }).format(conversion.rate);

  return (
    <div className={styles.conversion} data-qa="trial-conversion-rate">
      <p className={styles.conversionRate} data-testid="trial-conversion-rate-value">
        {rate}
      </p>
      {conversion.rate !== null && (
        <>
          <p className={styles.comparisonContext}>
            of the full-term observed trial cohort, paid in the current snapshot
          </p>
          <p className={styles.comparisonContext}>
            {conversion.numerator} of {conversion.denominator} tracked trial players became paid members
          </p>
        </>
      )}
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
