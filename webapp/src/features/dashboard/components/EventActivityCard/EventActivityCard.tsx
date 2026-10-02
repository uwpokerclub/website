import { Bar, BarChart, CartesianGrid, ReferenceLine, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { DashboardCard } from "../DashboardCard";
import { CardActionLink } from "../CardActionLink";
import { CARD_TITLES } from "../../dashboardLayout";
import { useEventActivity } from "../../hooks/useDashboardQueries";
import { AXIS_PROPS, CHART_COLORS, GRID_PROPS, TOOLTIP_STYLE } from "../../charts/chartTheme";
import type { EventActivityResponse } from "../../api/dashboardApi";
import styles from "./EventActivityCard.module.css";

type Props = { semesterId: string };

export function EventActivityCard({ semesterId }: Props) {
  const { data, isLoading, isError, refetch } = useEventActivity(semesterId);

  const status = isError
    ? "error"
    : isLoading || !data
      ? "loading"
      : data.current.eventsScheduled === 0 && data.current.eventsRun === 0
        ? "empty"
        : "ready";

  return (
    <DashboardCard
      title={CARD_TITLES.eventActivity}
      status={status}
      onRetry={() => refetch()}
      emptyMessage="No events scheduled yet this term."
      action={<CardActionLink to="/admin/events">All events</CardActionLink>}
      data-qa="event-activity-card"
    >
      {() => <Body data={data as EventActivityResponse} />}
    </DashboardCard>
  );
}

/**
 * A ratio against a limit, then a series — so: a meter, then columns.
 *
 * Average field size is drawn as a reference line rather than a second series. Two
 * measures on one chart would need two y-scales, and a dual-axis chart is the single
 * most misread form there is.
 */
function Body({ data }: { data: EventActivityResponse }) {
  const { eventsRun, eventsScheduled, totalEntries, averageFieldSize, series } = data.current;

  // The two counts are disjoint — ended and not-yet-ended — so the term's total is
  // their sum. Reading eventsScheduled as the total rendered "3 of 2 events run".
  const totalEvents = eventsRun + eventsScheduled;
  const average = Math.round(averageFieldSize * 10) / 10;

  return (
    <div className={styles.container}>
      <p className={styles.lead}>
        <span className={styles.figure}>{eventsRun}</span>
        <span className={styles.caption}>of {totalEvents} events run this term</span>
      </p>

      <div className={styles.meter} aria-hidden="true">
        <i style={{ flexGrow: eventsRun }} />
        {eventsScheduled > 0 && <u style={{ flexGrow: eventsScheduled }} />}
      </div>
      <p className={styles.meta}>
        {eventsScheduled} still scheduled · {totalEntries.toLocaleString("en-CA")} entries · average field {average}
      </p>

      <div className={styles.chart}>
        <ResponsiveContainer width="100%" height={160}>
          <BarChart data={series} margin={{ top: 8, right: 8, bottom: 0, left: -20 }}>
            <CartesianGrid {...GRID_PROPS} />
            <XAxis dataKey="name" {...AXIS_PROPS} interval="preserveStartEnd" />
            <YAxis {...AXIS_PROPS} width={34} />
            <Tooltip {...TOOLTIP_STYLE} />
            <ReferenceLine
              y={average}
              stroke={CHART_COLORS.slot4}
              strokeDasharray="4 4"
              strokeWidth={2}
              label={{ value: `avg ${average}`, position: "right", fontSize: 10, fill: CHART_COLORS.slot4 }}
            />
            <Bar dataKey="entries" name="Entries" fill={CHART_COLORS.slot1} radius={[4, 4, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </div>

      <p className={styles.foot}>
        An event counts as run once it has ended. Average field size covers run events only.
      </p>
    </div>
  );
}
