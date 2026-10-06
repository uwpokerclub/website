import { useId } from "react";
import {
  Area,
  ComposedChart,
  CartesianGrid,
  Legend,
  Line,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { DashboardCard } from "../DashboardCard";
import { CardActionLink } from "../CardActionLink";
import { CARD_TITLES } from "../../dashboardLayout";
import { useSignups } from "../../hooks/useDashboardQueries";
import { ActiveTooltipContent } from "../../charts/ActiveTooltipContent";
import { AXIS_PROPS, CHART_COLORS, GRID_PROPS, TOOLTIP_STYLE } from "../../charts/chartTheme";
import type { SignupsResponse } from "../../api/dashboardApi";
import styles from "./SignupTimelineCard.module.css";

type Props = { semesterId: string };

export function SignupTimelineCard({ semesterId }: Props) {
  const { data, isPending, isError, refetch } = useSignups(semesterId);

  const status = isPending
    ? "loading"
    : isError
      ? "error"
      : !data
        ? "loading"
        : data.series.length === 0 || (data.total === 0 && !data.comparison?.dailyTotals.length)
          ? "empty"
          : "ready";

  return (
    <DashboardCard
      title={CARD_TITLES.signupTimeline}
      status={status}
      onRetry={() => refetch()}
      emptyMessage="No signups recorded for this term."
      action={<CardActionLink to="/admin/members">Memberships</CardActionLink>}
      data-qa="signup-timeline-card"
    >
      {() => <Body data={data as SignupsResponse} />}
    </DashboardCard>
  );
}

function Body({ data }: { data: SignupsResponse }) {
  const captionId = useId();
  const busiest = data.series.reduce((max, point) => Math.max(max, point.admin + point.discord + point.unknown), 0);
  const comparisonByDay = new Map(data.comparison?.dailyTotals.map((point) => [point.elapsedDay, point.total]) ?? []);
  const firstDate = data.series[0]?.date;
  const chartData = data.series.map((point) => {
    const elapsedDay = firstDate
      ? (Date.parse(`${point.date}T00:00:00Z`) - Date.parse(`${firstDate}T00:00:00Z`)) / 86_400_000
      : -1;

    return {
      ...point,
      comparisonTotal: comparisonByDay.has(elapsedDay) ? comparisonByDay.get(elapsedDay) : undefined,
    };
  });
  const hasComparison = Boolean(data.comparison && data.comparison.dailyTotals.length > 0);
  const eventDates = new Set(data.eventDates);

  return (
    <div className={styles.container}>
      {data.total === 0 && hasComparison && (
        <p role="status">No signups recorded for this term; the prior-term comparison is shown below.</p>
      )}
      <p className={styles.lead}>
        <span className={styles.figure}>{data.total.toLocaleString("en-CA")}</span>
        <span className={styles.caption}>memberships created, {busiest} on the busiest day</span>
      </p>

      <div className={styles.chart} data-qa="signup-timeline-chart">
        <ResponsiveContainer width="100%" height={170}>
          <ComposedChart data={chartData} margin={{ top: 8, right: 4, bottom: 0, left: 0 }}>
            <CartesianGrid {...GRID_PROPS} />
            <XAxis dataKey="date" {...AXIS_PROPS} interval="preserveStartEnd" />
            <YAxis
              {...AXIS_PROPS}
              width={48}
              allowDecimals={false}
              tickFormatter={(value: number) => Math.round(value).toLocaleString("en-CA")}
            />
            <Tooltip {...TOOLTIP_STYLE} content={ActiveTooltipContent} />
            <Legend wrapperStyle={{ fontSize: 11, fontFamily: "Montserrat, sans-serif" }} />
            {/* The spikes are the event days; without them the series is unreadable. */}
            {[...eventDates].map((date) => (
              <ReferenceLine key={date} x={date} stroke="#616161" strokeDasharray="4 3" strokeWidth={2} />
            ))}
            <Area
              type="monotone"
              dataKey="admin"
              stackId="1"
              name="Admin"
              stroke={CHART_COLORS.slot1}
              fill={CHART_COLORS.slot1}
              fillOpacity={0.3}
              strokeWidth={2}
              isAnimationActive={false}
            />
            <Area
              type="monotone"
              dataKey="discord"
              stackId="1"
              name="Discord"
              stroke={CHART_COLORS.slot2}
              fill={CHART_COLORS.slot2}
              fillOpacity={0.3}
              strokeWidth={2}
              isAnimationActive={false}
            />
            <Area
              type="monotone"
              dataKey="unknown"
              stackId="1"
              name="Unknown"
              stroke={CHART_COLORS.slot3}
              fill={CHART_COLORS.slot3}
              fillOpacity={0.3}
              strokeWidth={2}
              isAnimationActive={false}
            />
            {hasComparison && data.comparison && (
              <Line
                type="monotone"
                dataKey="comparisonTotal"
                name={`${data.comparison.semester.name} daily total`}
                stroke={CHART_COLORS.slot4}
                strokeWidth={2}
                dot={false}
                connectNulls={false}
                isAnimationActive={false}
              />
            )}
          </ComposedChart>
        </ResponsiveContainer>
      </div>

      <details className={styles.dataDetails}>
        <summary>View daily signup data</summary>
        {/* Wide tables scroll inside the card rather than widening it; the region is
            focusable so the scroll is reachable from the keyboard. */}
        <div
          className={styles.tableScroll}
          role="region"
          aria-labelledby={captionId}
          tabIndex={0}
          data-qa="signup-timeline-table"
        >
          <table>
            <caption id={captionId}>
              Daily memberships created by source
              {data.comparison ? ` and ${data.comparison.semester.name} comparison` : ""}
            </caption>
            <thead>
              <tr>
                <th scope="col">Date</th>
                <th scope="col">Admin</th>
                <th scope="col">Discord</th>
                <th scope="col">Unknown</th>
                <th scope="col">Event day</th>
                {data.comparison && <th scope="col">{data.comparison.semester.name}</th>}
              </tr>
            </thead>
            <tbody>
              {chartData.map((point) => (
                <tr key={point.date}>
                  <th scope="row">{point.date}</th>
                  <td>{point.admin}</td>
                  <td>{point.discord}</td>
                  <td>{point.unknown}</td>
                  <td>{eventDates.has(point.date) ? "Yes" : "No"}</td>
                  {data.comparison && <td>{point.comparisonTotal ?? "—"}</td>}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </details>

      <p className={styles.foot}>
        Dashed grey lines mark event days. Memberships from before signup tracking are dated from the member&apos;s
        first event where possible; members without a reliable creation date are omitted. Historical comparison dates
        may be reconstructed from first participation, and the comparison line can omit undated memberships.
      </p>
    </div>
  );
}
