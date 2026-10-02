import {
  Area,
  AreaChart,
  CartesianGrid,
  Legend,
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
import { SAMPLE_TERM } from "../../fixtures/sampleTerm";
import type { SignupsResponse } from "../../api/dashboardApi";
import styles from "./SignupTimelineCard.module.css";

/**
 * GET …/dashboard/signups (#433) does not exist yet, so a failed request falls back
 * to the sample term behind a visible marker.
 *
 * Gated on isError, not on absent data: substituting fixtures whenever data is
 * missing would flash a fabricated 967-signup chart during every normal load.
 * Delete this and its branch when #433 ships.
 */
const FALL_BACK_TO_SAMPLE = true;

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
  const usingSample = FALL_BACK_TO_SAMPLE && isError;
  const resolved = data ?? (usingSample ? SAMPLE_TERM.signups : undefined);

  const status = isPending
    ? "loading"
    : isError && !usingSample
      ? "error"
      : !resolved
        ? "loading"
        : resolved.series.length === 0
          ? "empty"
          : "ready";

  return (
    <DashboardCard
      title={CARD_TITLES.signupTimeline}
      status={status}
      sampleData={usingSample}
      onRetry={() => refetch()}
      emptyMessage={emptyMessageFor(resolved)}
      action={<CardActionLink to="/admin/members">Memberships</CardActionLink>}
      data-qa="signup-timeline-card"
    >
      {() => <Body data={resolved as SignupsResponse} />}
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
  const busiest = data.series.reduce((max, point) => Math.max(max, point.admin + point.discord), 0);

  return (
    <div className={styles.container}>
      <p className={styles.lead}>
        <span className={styles.figure}>{data.total.toLocaleString("en-CA")}</span>
        <span className={styles.caption}>memberships created, {busiest} on the busiest day</span>
      </p>

      <div className={styles.chart}>
        <ResponsiveContainer width="100%" height={170}>
          <AreaChart data={data.series} margin={{ top: 8, right: 8, bottom: 0, left: -20 }}>
            <CartesianGrid {...GRID_PROPS} />
            <XAxis dataKey="date" {...AXIS_PROPS} interval="preserveStartEnd" />
            <YAxis {...AXIS_PROPS} width={34} />
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
          </AreaChart>
        </ResponsiveContainer>
      </div>

      <p className={styles.foot}>
        Gold lines mark event days. Memberships from before signup tracking are dated from the member&apos;s first
        event, so members who never played are not counted.
      </p>
    </div>
  );
}
