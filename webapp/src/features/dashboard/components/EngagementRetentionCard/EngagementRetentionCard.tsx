import { DashboardCard } from "../DashboardCard";
import { StatBar } from "../StatBar";
import { RangeTrack } from "../RangeTrack";
import { CARD_TITLES } from "../../dashboardLayout";
import { useEngagementDashboard, useMembershipsDashboard } from "../../hooks/useDashboardQueries";
import type { EngagementDashboardResponse, MembershipsDashboardResponse } from "../../api/dashboardApi";
import styles from "./EngagementRetentionCard.module.css";

/**
 * Historical ranges from the six-term points-system analysis, recorded in
 * docs/superpowers/specs/2026-10-01-dashboard-page-design.md.
 *
 * These are constants because the API returns this term and one comparison term
 * only — six terms of range does not exist in the response. They will go stale and
 * nothing will warn anyone; see the spec's "Hard-coded constants" section for why
 * that trade was made anyway. Briefly: a president reading "median 2 events, down 1"
 * with no baseline concludes the club is collapsing during a completely normal term.
 */
const MEDIAN_EVENTS_BAND = { low: 2, high: 3, min: 0, max: 6 };
const PLAYED_ONCE_BAND = { low: 33, high: 44, min: 0, max: 100 };

type Props = { semesterId: string };

export function EngagementRetentionCard({ semesterId }: Props) {
  const engagement = useEngagementDashboard(semesterId);
  const memberships = useMembershipsDashboard(semesterId);

  // `data` guards "ready": isLoading is false between mount and first fetch on a
  // disabled query, and the render function must never see undefined data.
  const status = engagement.isError
    ? "error"
    : engagement.isLoading || !engagement.data
      ? "loading"
      : engagement.data.current.players === 0
        ? "empty"
        : "ready";

  return (
    <DashboardCard
      title={CARD_TITLES.engagement}
      status={status}
      onRetry={() => engagement.refetch()}
      emptyMessage="No event entries recorded yet this term."
      data-qa="engagement-retention-card"
    >
      {() => (
        <Body
          data={engagement.data as EngagementDashboardResponse}
          memberships={memberships.data as MembershipsDashboardResponse | undefined}
        />
      )}
    </DashboardCard>
  );
}

function Body({
  data,
  memberships,
}: {
  data: EngagementDashboardResponse;
  memberships?: MembershipsDashboardResponse;
}) {
  const { current, comparison } = data;

  // The middle bucket is arithmetic: the endpoint exposes the total, the played-once
  // cohort and the 10+ cohort, so 2–9 is what is left. Deriving it here also removes
  // the old card's redundancy, where "came back" and "played once" asserted the same
  // fact twice with opposite sentiment colouring.
  const cameBack = current.players - current.playedOnceCount - current.tenPlusCount;
  const total = memberships?.current.total;
  const playedOncePercent = Math.round(current.playedOnceShare * 100);

  return (
    <div className={styles.container}>
      <div className={styles.columns}>
        <div className={styles.summary}>
          <p className={styles.lead}>
            <span className={styles.figure}>{current.players.toLocaleString("en-CA")}</span>
            <span className={styles.caption}>
              {total
                ? `of ${total.toLocaleString("en-CA")} members played at least one event`
                : "members played at least one event"}
            </span>
          </p>

          <StatBar
            ariaLabel="Players by events attended"
            segments={[
              { key: "once", label: "played once", value: current.playedOnceCount, color: "var(--dash-ord-1)" },
              { key: "some", label: "played 2–9", value: cameBack, color: "var(--dash-ord-2)" },
              { key: "regular", label: "regulars, 10+", value: current.tenPlusCount, color: "var(--dash-ord-3)" },
            ]}
          />
        </div>

        <div className={styles.tracks}>
          <RangeTrack
            label="Median events played"
            value={current.medianEventsAttended}
            display={`${current.medianEventsAttended}`}
            min={MEDIAN_EVENTS_BAND.min}
            max={MEDIAN_EVENTS_BAND.max}
            bandLow={MEDIAN_EVENTS_BAND.low}
            bandHigh={MEDIAN_EVENTS_BAND.high}
            comparison={
              comparison ? { value: comparison.stats.medianEventsAttended, label: comparison.semester.name } : undefined
            }
          />
          <RangeTrack
            label="Played exactly once"
            value={playedOncePercent}
            display={`${playedOncePercent}%`}
            min={PLAYED_ONCE_BAND.min}
            max={PLAYED_ONCE_BAND.max}
            bandLow={PLAYED_ONCE_BAND.low}
            bandHigh={PLAYED_ONCE_BAND.high}
            comparison={
              comparison
                ? { value: Math.round(comparison.stats.playedOnceShare * 100), label: comparison.semester.name }
                : undefined
            }
          />
        </div>
      </div>

      <p className={styles.foot}>
        Members who entered at least one event this term. The shaded band is the range across six analysed terms
        {comparison ? `; the grey tick is ${comparison.semester.name}.` : "."}
      </p>
    </div>
  );
}
