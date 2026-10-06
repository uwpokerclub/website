import { useId } from "react";
import { Bar, BarChart, CartesianGrid, ReferenceLine, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { DashboardCard } from "../DashboardCard";
import { CardActionLink } from "../CardActionLink";
import { CARD_TITLES } from "../../dashboardLayout";
import { useEventActivity } from "../../hooks/useDashboardQueries";
import { ActiveTooltipContent } from "../../charts/ActiveTooltipContent";
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
 * Average field sizes are reference lines rather than a second event series. Two
 * measures on one chart would need two y-scales, and a dual-axis chart is the single
 * most misread form there is.
 */
function Body({ data }: { data: EventActivityResponse }) {
  const captionId = useId();
  const { eventsRun, eventsScheduled, totalEntries, averageFieldSize, series } = data.current;

  // The two counts are disjoint — ended and not-yet-ended — so the term's total is
  // their sum. Reading eventsScheduled as the total rendered "3 of 2 events run".
  const totalEvents = eventsRun + eventsScheduled;
  const average = Math.round(averageFieldSize * 10) / 10;
  const comparison = data.comparison;
  const comparisonAverage =
    comparison?.averageFieldSize == null ? null : Math.round(comparison.averageFieldSize * 10) / 10;

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
        {eventsScheduled} still scheduled · {totalEntries.toLocaleString("en-CA")} entries
        {eventsRun > 0 && <> · average field {average}</>}
      </p>

      <div className={styles.averageKey} role="group" aria-label="Average field size reference lines">
        {eventsRun > 0 ? (
          <p className={styles.averageItem}>
            <i className={styles.currentAverageSwatch} aria-hidden="true" />
            <span>Current average: {average.toFixed(1)} players per completed event</span>
          </p>
        ) : null}
        {eventsRun === 0 && <p className={styles.averageItem}>No completed events yet.</p>}
        {comparison &&
          (comparisonAverage === null ? (
            <p className={styles.averageItem}>{comparison.semester.name}: no completed events by this point.</p>
          ) : (
            <p className={styles.averageItem}>
              <i className={styles.comparisonAverageSwatch} aria-hidden="true" />
              <span>
                {comparison.semester.name}: {comparisonAverage.toFixed(1)} players per completed event at the same
                elapsed-term span
              </span>
            </p>
          ))}
      </div>

      <div className={styles.chart} data-qa="event-activity-chart">
        <ResponsiveContainer width="100%" height={160}>
          <BarChart data={series} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
            <CartesianGrid {...GRID_PROPS} />
            <XAxis dataKey="name" {...AXIS_PROPS} interval="preserveStartEnd" />
            <YAxis
              {...AXIS_PROPS}
              width={48}
              allowDecimals={false}
              domain={[
                0,
                Math.ceil(
                  Math.max(
                    eventsRun > 0 ? average : 0,
                    comparisonAverage ?? 0,
                    ...series.map((point) => point.entries),
                  ),
                ),
              ]}
              tickFormatter={(value: number) => Math.round(value).toLocaleString("en-CA")}
            />
            <Tooltip {...TOOLTIP_STYLE} content={ActiveTooltipContent} />
            {eventsRun > 0 && (
              <ReferenceLine
                y={average}
                stroke={CHART_COLORS.slot4}
                strokeDasharray="4 4"
                strokeWidth={2}
                ifOverflow="extendDomain"
              />
            )}
            {comparisonAverage !== null && comparison && (
              <ReferenceLine
                y={comparisonAverage}
                stroke={CHART_COLORS.slot2}
                strokeDasharray="2 3"
                strokeWidth={2}
                ifOverflow="extendDomain"
              />
            )}
            <Bar
              dataKey="entries"
              name="Entries"
              fill={CHART_COLORS.slot1}
              radius={[4, 4, 0, 0]}
              isAnimationActive={false}
            />
          </BarChart>
        </ResponsiveContainer>
      </div>

      <details className={styles.dataDetails}>
        <summary>View event entry data</summary>
        {/* Wide tables scroll inside the card rather than widening it; the region is
            focusable so the scroll is reachable from the keyboard. */}
        <div
          className={styles.tableScroll}
          role="region"
          aria-labelledby={captionId}
          tabIndex={0}
          data-qa="event-activity-table"
        >
          <table>
            <caption id={captionId}>Event entries for {data.current.series.length} events</caption>
            <thead>
              <tr>
                <th scope="col">Event</th>
                <th scope="col">Date</th>
                <th scope="col">Entries</th>
              </tr>
            </thead>
            <tbody>
              {series.map((event) => (
                <tr key={event.id}>
                  <th scope="row">{event.name}</th>
                  <td>{event.startDate}</td>
                  <td>{event.entries}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </details>

      <p className={styles.foot}>
        An event counts as run once it has ended. Average field size covers run events only.
      </p>
    </div>
  );
}
