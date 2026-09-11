import { DashboardCard } from "../DashboardCard";
import { DeltaChip, type DeltaSentiment } from "../DeltaChip";
import { CARD_TITLES } from "../../dashboardLayout";
import { useMembershipsDashboard } from "../../hooks/useDashboardQueries";
import type { MembershipStats, MembershipsDashboardResponse } from "../../api/dashboardApi";
import styles from "./TermAtAGlanceCard.module.css";

type TermAtAGlanceCardProps = {
  semesterId: string;
};

type MiniStatKey = "paid" | "unpaid" | "discounted" | "executive" | "new" | "returning";

type MiniStatConfig = {
  key: MiniStatKey;
  label: string;
  sentiment: DeltaSentiment;
};

const BUCKET_STATS: MiniStatConfig[] = [
  { key: "paid", label: "Paid", sentiment: "positive-is-good" },
  { key: "unpaid", label: "Unpaid", sentiment: "negative-is-good" },
  { key: "discounted", label: "Discounted", sentiment: "neutral" },
  { key: "executive", label: "Executive", sentiment: "neutral" },
];

const RETENTION_STATS: MiniStatConfig[] = [
  { key: "new", label: "New", sentiment: "positive-is-good" },
  { key: "returning", label: "Returning", sentiment: "positive-is-good" },
];

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

function TermAtAGlanceBody({ data }: { data: MembershipsDashboardResponse }) {
  const { current, comparison } = data;

  return (
    <div className={styles.container}>
      <div className={styles.headline}>
        <span className={styles.eyebrow}>TOTAL</span>
        <span className={styles.total}>{current.total}</span>
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

      <hr className={styles.divider} />
      <div className={styles.grid}>
        {BUCKET_STATS.map((stat) => (
          <MiniStat
            key={stat.key}
            config={stat}
            current={current}
            comparisonStats={comparison?.stats}
            comparisonLabel={comparison?.semester.name}
          />
        ))}
      </div>

      <hr className={styles.divider} />
      <div className={styles.grid}>
        {RETENTION_STATS.map((stat) => (
          <MiniStat
            key={stat.key}
            config={stat}
            current={current}
            comparisonStats={comparison?.stats}
            comparisonLabel={comparison?.semester.name}
          />
        ))}
      </div>
    </div>
  );
}

function MiniStat({
  config,
  current,
  comparisonStats,
  comparisonLabel,
}: {
  config: MiniStatConfig;
  current: MembershipStats;
  comparisonStats?: MembershipStats;
  comparisonLabel?: string;
}) {
  return (
    <div className={styles.stat}>
      <span className={styles.eyebrow}>{config.label}</span>
      <span className={styles.value}>{current[config.key]}</span>
      {comparisonStats && comparisonLabel && (
        <DeltaChip
          current={current[config.key]}
          comparison={comparisonStats[config.key]}
          comparisonLabel={comparisonLabel}
          sentiment={config.sentiment}
          compact
        />
      )}
    </div>
  );
}
