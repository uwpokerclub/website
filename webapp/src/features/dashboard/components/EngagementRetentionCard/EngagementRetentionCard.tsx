import { DashboardCard } from "../DashboardCard";
import { DeltaChip, type DeltaSentiment } from "../DeltaChip";
import { CARD_TITLES } from "../../dashboardLayout";
import { useEngagementDashboard } from "../../hooks/useDashboardQueries";
import type { EngagementDashboardResponse, EngagementStats } from "../../api/dashboardApi";
import styles from "./EngagementRetentionCard.module.css";

type EngagementRetentionCardProps = {
  semesterId: string;
};

type StatKey = "medianEventsAttended" | "playedOnceShare" | "tenPlusCount";

type StatConfig = {
  key: StatKey;
  label: string;
  sentiment: DeltaSentiment;
  mode?: "percentage-points";
};

const STATS: StatConfig[] = [
  { key: "medianEventsAttended", label: "Median events attended", sentiment: "positive-is-good" },
  { key: "playedOnceShare", label: "Played exactly once", sentiment: "negative-is-good", mode: "percentage-points" },
  { key: "tenPlusCount", label: "Played 10+ events", sentiment: "positive-is-good" },
];

function formatNumber(value: number): string {
  return new Intl.NumberFormat("en-CA", { maximumFractionDigits: 2 }).format(value);
}

function formatValue(key: StatKey | "players", value: number): string {
  return key === "playedOnceShare" ? `${formatNumber(value * 100)}%` : formatNumber(value);
}

export function EngagementRetentionCard({ semesterId }: EngagementRetentionCardProps) {
  const { data, isLoading, isError, refetch } = useEngagementDashboard(semesterId);
  const status = isError ? "error" : isLoading || !data ? "loading" : data.current.players === 0 ? "empty" : "ready";

  return (
    <DashboardCard
      title={CARD_TITLES.engagement}
      status={status}
      onRetry={() => refetch()}
      emptyMessage="No event entries recorded yet this term."
      data-qa="engagement-retention-card"
    >
      {() => <EngagementRetentionBody data={data as EngagementDashboardResponse} />}
    </DashboardCard>
  );
}

function EngagementRetentionBody({ data }: { data: EngagementDashboardResponse }) {
  const { current, comparison } = data;

  return (
    <div className={styles.container}>
      <p className={styles.population}>Members who entered at least one event this term.</p>
      <div className={styles.headline}>
        <span className={styles.eyebrow}>DISTINCT PLAYERS</span>
        <span className={styles.total}>{formatValue("players", current.players)}</span>
        {comparison ? (
          <DeltaChip
            current={current.players}
            comparison={comparison.stats.players}
            comparisonLabel={comparison.semester.name}
            sentiment="positive-is-good"
          />
        ) : (
          <p className={styles.noComparisonNote}>No comparable term to compare against yet.</p>
        )}
      </div>
      <hr className={styles.divider} />
      <div className={styles.grid}>
        {STATS.map((stat) => (
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
  config: StatConfig;
  current: EngagementStats;
  comparisonStats?: EngagementStats;
  comparisonLabel?: string;
}) {
  const currentValue = current[config.key];

  return (
    <div className={styles.stat}>
      <span className={styles.eyebrow}>{config.label}</span>
      <span className={styles.value}>{formatValue(config.key, currentValue)}</span>
      {comparisonStats && comparisonLabel && (
        <DeltaChip
          current={currentValue}
          comparison={comparisonStats[config.key]}
          comparisonLabel={comparisonLabel}
          sentiment={config.sentiment}
          mode={config.mode}
          compact
        />
      )}
    </div>
  );
}
