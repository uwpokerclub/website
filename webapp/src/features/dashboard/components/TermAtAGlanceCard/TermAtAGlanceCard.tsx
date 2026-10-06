import { useCurrentSemester } from "@/hooks";
import { DashboardCard } from "../DashboardCard";
import { DeltaChip } from "../DeltaChip";
import { StatBar } from "../StatBar";
import { CARD_TITLES } from "../../dashboardLayout";
import { useMembershipsDashboard } from "../../hooks/useDashboardQueries";
import type { MembershipsDashboardResponse } from "../../api/dashboardApi";
import { termProgress } from "./termProgress";
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
      <div className={styles.columns}>
        <div className={styles.summary}>
          <div className={styles.headline}>
            <span className={styles.total}>{current.total.toLocaleString("en-CA")}</span>
            <Comparison current={current.total} comparison={comparison} />
          </div>

          <PaceTrack current={current.total} comparison={comparison} />
        </div>

        <div className={styles.breakdown}>
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
      </div>
    </div>
  );
}

/**
 * A membership total only means something against the same point of the earlier
 * term. Comparing a term three weeks old against a finished one reads as a collapse
 * every single time, which is what this card used to do.
 *
 * totalAsOf carries that like-for-like figure, and is null for terms whose
 * memberships predate created_at. In that case the card shows pace toward the final
 * total instead — honest about being a different question — rather than a delta it
 * cannot compute. Once those terms are dated, the delta appears on its own.
 */
function Comparison({
  current,
  comparison,
}: {
  current: number;
  comparison: MembershipsDashboardResponse["comparison"];
}) {
  if (!comparison) {
    return <p className={styles.noComparisonNote}>No comparable term to compare against yet.</p>;
  }

  if (comparison.totalAsOf === null) {
    // Nothing honest to put in a delta chip; the track below carries the pace.
    return null;
  }

  if (comparison.stats.total === 0) {
    return (
      <p className={styles.noComparisonNote}>No membership baseline is available for {comparison.semester.name}.</p>
    );
  }

  return (
    <DeltaChip
      current={current}
      comparison={comparison.totalAsOf}
      comparisonLabel={`memberships by this point in ${comparison.semester.name}`}
      sentiment="positive-is-good"
    />
  );
}

/**
 * Progress toward a full term's worth of memberships, with a marker for where the
 * comparison term stood at this same point.
 *
 * The delta chip answers "are we ahead or behind right now"; this answers "how much
 * of a term is still to come", which is the question that makes a mid-term figure
 * actionable. They are different questions, so both earn their place — unlike the
 * engagement card's old pair, which stated one fact twice.
 */
function PaceTrack({
  current,
  comparison,
}: {
  current: number;
  comparison: MembershipsDashboardResponse["comparison"];
}) {
  const { currentSemester } = useCurrentSemester();

  if (!comparison) {
    return null;
  }

  if (comparison.stats.total === 0) {
    return null;
  }
  const target = comparison.stats.total;
  const share = Math.round((current / target) * 100);
  const fill = Math.min(share, 100);
  const markerAt = comparison.totalAsOf === null ? null : Math.min((comparison.totalAsOf / target) * 100, 100);

  const label =
    `${current.toLocaleString("en-CA")} of ${comparison.stats.total.toLocaleString("en-CA")} ` +
    `— ${share}% of ${comparison.semester.name}'s final total` +
    (comparison.totalAsOf === null
      ? "."
      : `, with ${comparison.totalAsOf.toLocaleString("en-CA")} memberships recorded by this point in ${comparison.semester.name}.`);

  return (
    <div className={styles.paceBlock}>
      <div className={styles.paceTrack} role="img" aria-label={`Progress toward a full term: ${label}`}>
        <i className={styles.paceFill} style={{ width: `${fill}%` }} />
        {markerAt !== null && (
          <i className={styles.paceMarker} data-qa="pace-marker" style={{ left: `${markerAt}%` }} />
        )}
      </div>
      <p className={styles.pace}>
        {share}% of {comparison.semester.name}&apos;s final {comparison.stats.total.toLocaleString("en-CA")}
        {termProgress(currentSemester) && <> · {termProgress(currentSemester)}</>}
      </p>
      {comparison.totalAsOf !== null && (
        <p className={styles.pace}>
          Marker: {comparison.semester.name} had {comparison.totalAsOf.toLocaleString("en-CA")} memberships by this
          point.
        </p>
      )}
    </div>
  );
}
