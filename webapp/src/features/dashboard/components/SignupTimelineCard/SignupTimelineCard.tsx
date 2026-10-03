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
import { AXIS_PROPS, CHART_COLORS, GRID_PROPS, TOOLTIP_STYLE } from "../../charts/chartTheme";
import type { SignupsResponse } from "../../api/dashboardApi";
import styles from "./SignupTimelineCard.module.css";

/**
 * `dataStartsAt` is a calendar date, not an instant: "2026-09-01" parses as UTC
 * midnight and would render as August 31 for anyone behind UTC. Format it in UTC so
 * the month shown is the month the API meant.
 */
const formatMonth = (iso: string) =>
  new Date(iso).toLocaleDateString("en-CA", { month: "long", year: "numeric", timeZone: "UTC" });

type Props = { semesterId: string };

export function SignupTimelineCard({ semesterId }: Props) {
  const { data, isPending, isError, refetch } = useSignups(semesterId);

  const status = isPending
    ? "loading"
    : isError
      ? "error"
      : !data
        ? "loading"
        : data.series.length === 0
          ? "empty"
          : "ready";

  return (
    <DashboardCard
      title={CARD_TITLES.signupTimeline}
      status={status}
      onRetry={() => refetch()}
      emptyMessage={emptyMessageFor(data)}
      action={<CardActionLink to="/admin/members">Memberships</CardActionLink>}
      data-qa="signup-timeline-card"
    >
      {() => <Body data={data as SignupsResponse} />}
    </DashboardCard>
  );
}

/**
 * An empty series is a fact about the term, not a wait that will end — but it is no
 * longer the normal state for historical terms. The created_at backfill dates 96-100%
 * of every recent term from first participation, so a term reaching this message has
 * genuinely recorded nothing rather than simply predating signup tracking.
 */
function emptyMessageFor(data?: SignupsResponse) {
  return data?.dataStartsAt
    ? `No signups recorded for this term. Dated signups begin ${formatMonth(data.dataStartsAt)}.`
    : "No signups recorded for this term.";
}

function Body({ data }: { data: SignupsResponse }) {
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

  return (
    <div className={styles.container}>
      <p className={styles.lead}>
        <span className={styles.figure}>{data.total.toLocaleString("en-CA")}</span>
        <span className={styles.caption}>memberships created, {busiest} on the busiest day</span>
      </p>

      <div className={styles.chart}>
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
            <Tooltip {...TOOLTIP_STYLE} />
            <Legend wrapperStyle={{ fontSize: 11, fontFamily: "Montserrat, sans-serif" }} />
            {/* The spikes are the event days; without them the series is unreadable. */}
            {data.eventDates.map((date) => (
              <ReferenceLine key={date} x={date} stroke={CHART_COLORS.slot2} strokeWidth={2} />
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
              />
            )}
          </ComposedChart>
        </ResponsiveContainer>
      </div>

      <p className={styles.foot}>
        Gold lines mark event days. Memberships from before signup tracking are dated from the member&apos;s first
        event, so members who never played are not counted. Historical comparison dates may be reconstructed from first
        participation.
      </p>
    </div>
  );
}
