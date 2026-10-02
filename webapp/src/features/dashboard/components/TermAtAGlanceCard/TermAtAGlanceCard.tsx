import { DashboardCard } from "../DashboardCard";
import { DeltaChip } from "../DeltaChip";
import { StatBar } from "../StatBar";
import { CARD_TITLES } from "../../dashboardLayout";
import { useMembershipsDashboard } from "../../hooks/useDashboardQueries";
import type { MembershipsDashboardResponse } from "../../api/dashboardApi";
import styles from "./TermAtAGlanceCard.module.css";

type TermAtAGlanceCardProps = {
  semesterId: string;
};

export function TermAtAGlanceCard({ semesterId }: TermAtAGlanceCardProps) {
  const { data, isLoading, isError, refetch } = useMembershipsDashboard(semesterId);

  // `data` guards "ready": isLoading is false between mount and first fetch on
  // a disabled query, and the render function must never see undefined data.
  const status = isError ? "error" : isLoading || !data ? "loading" : data.current.total === 0 ? "empty" : "ready";

  return (
    <DashboardCard
      title={CARD_TITLES.termAtAGlance}
      status={status}
      onRetry={() => refetch()}
      emptyMessage="No memberships recorded yet this term."
      data-qa="term-at-a-glance-card"
    >
      {() => <TermAtAGlanceBody data={data as MembershipsDashboardResponse} />}
    </DashboardCard>
  );
}

/**
 * Two part-to-whole bars rather than a grid of independent numbers. The buckets are
 * an exclusive partition of one total — a 2x2 of bare figures cannot say that, and
 * reading four numbers to work out the paid/unpaid balance is work the bar does.
 */
function TermAtAGlanceBody({ data }: { data: MembershipsDashboardResponse }) {
  const { current, comparison } = data;

  return (
    <div className={styles.container}>
      <div className={styles.headline}>
        <span className={styles.total}>{current.total.toLocaleString("en-CA")}</span>
        {comparison ? (
          <DeltaChip
            current={current.total}
            comparison={comparison.stats.total}
            comparisonLabel={comparison.semester.name}
            sentiment="positive-is-good"
          />
        ) : (
          <p className={styles.noComparisonNote}>No comparable term to compare against yet.</p>
        )}
      </div>

      <StatBar
        ariaLabel="Memberships by payment status"
        segments={[
          { key: "paid", label: "paid", value: current.paid, color: "var(--dash-cat-1)" },
          { key: "unpaid", label: "unpaid", value: current.unpaid, color: "var(--dash-cat-2)" },
          { key: "discounted", label: "discounted", value: current.discounted, color: "var(--dash-cat-3)" },
          { key: "executive", label: "exec", value: current.executive, color: "var(--dash-cat-4)" },
        ]}
      />

      <StatBar
        ariaLabel="Memberships by prior membership"
        segments={[
          { key: "new", label: "first-ever term", value: current.new, color: "var(--dash-ord-3)" },
          { key: "returning", label: "returning", value: current.returning, color: "var(--dash-ord-1)" },
        ]}
      />
    </div>
  );
}
