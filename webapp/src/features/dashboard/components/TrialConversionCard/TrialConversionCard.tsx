import { DashboardCard } from "../DashboardCard";
import { StatBar } from "../StatBar";
import { CARD_TITLES } from "../../dashboardLayout";
import { useTrialConversion } from "../../hooks/useDashboardQueries";
import type { ConversionResponse } from "../../api/dashboardApi";
import styles from "./TrialConversionCard.module.css";

/**
 * The endpoint returns the current term's status snapshot. Loading, errors, and
 * empty results remain card-local states rather than substituting fabricated data.
 */
type Props = { semesterId: string };

export function TrialConversionCard({ semesterId }: Props) {
  const { data, isPending, isError, refetch } = useTrialConversion(semesterId);
  const status = isPending
    ? "loading"
    : isError
      ? "error"
      : !data
        ? "loading"
        : data.current.players === 0
          ? "empty"
          : "ready";

  return (
    <DashboardCard
      title={CARD_TITLES.trialConversion}
      status={status}
      onRetry={() => refetch()}
      emptyMessage="No event entries recorded yet this term."
      data-qa="trial-conversion-card"
    >
      {() => <Body data={data as ConversionResponse} />}
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
