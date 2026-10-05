import { Link } from "react-router-dom";
import { DashboardCard } from "../DashboardCard";
import { CARD_TITLES } from "../../dashboardLayout";
import { useSpotlight } from "../../hooks/useDashboardQueries";
import type { SpotlightEvent } from "../../api/dashboardApi";
import styles from "./EventSpotlightCard.module.css";

/** models.EventStateStarted — an event that has not yet been ended. */
const EVENT_STATE_STARTED = 0;

type Props = { semesterId: string };

export function EventSpotlightCard({ semesterId }: Props) {
  // null is a legitimate response here — "no event scheduled" — so absent data
  // cannot stand in for "still loading". Between a failed attempt and its retry,
  // isLoading is false and isError is not yet true; isPending covers that window,
  // which is what stops the card briefly claiming there are no events.
  const { data, isPending, isError, refetch } = useSpotlight(semesterId);
  const status = isPending ? "loading" : isError ? "error" : !data ? "empty" : "ready";

  return (
    <DashboardCard
      title={CARD_TITLES.spotlight}
      status={status}
      onRetry={() => refetch()}
      emptyMessage="No events scheduled this term."
      data-qa="event-spotlight-card"
    >
      {() => <Body event={data as SpotlightEvent} />}
    </DashboardCard>
  );
}

/**
 * The only card whose hero is text. At the door you need to know which event is
 * running, not how many of anything — so the name carries the weight and the counts
 * sit underneath.
 *
 * This card also owns the page's only motion: a slow pulse on the live dot, off
 * under prefers-reduced-motion. Nothing else on the dashboard animates.
 */
function Body({ event }: { event: SpotlightEvent }) {
  const start = new Date(event.startDate);
  const now = new Date();
  const live = event.state === EVENT_STATE_STARTED && start <= now;
  const priorTorontoDay = live && torontoDateKey(start) < torontoDateKey(now);
  const priorStartDate = priorTorontoDay
    ? `${new Intl.DateTimeFormat("en-CA", {
        timeZone: "America/Toronto",
        weekday: "short",
        month: "short",
        day: "numeric",
      }).format(start)} at `
    : "";

  return (
    <div className={styles.container}>
      <p className={styles.state} data-live={live || undefined}>
        {live ? (
          <>
            <span className={styles.pulse} aria-hidden="true" />
            Live · started {priorStartDate}
            {start.toLocaleTimeString("en-CA", { hour: "numeric", minute: "2-digit", timeZone: "America/Toronto" })}
          </>
        ) : (
          <>Scheduled · {start.toLocaleDateString("en-CA", { weekday: "long", month: "short", day: "numeric" })}</>
        )}
      </p>
      <h3 className={styles.name}>{event.name}</h3>
      <p className={styles.format}>{event.format}</p>

      <dl className={styles.stats}>
        <div>
          <dt>entries</dt>
          <dd>{event.entries}</dd>
        </div>
        <div>
          <dt>rebuys</dt>
          <dd>{event.rebuys}</dd>
        </div>
      </dl>

      <Link className={styles.cta} to={`/admin/events/${event.id}`}>
        Open event
      </Link>
    </div>
  );
}

function torontoDateKey(date: Date): string {
  const parts = new Intl.DateTimeFormat("en-CA", {
    timeZone: "America/Toronto",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(date);
  const values = Object.fromEntries(parts.map((part) => [part.type, part.value]));
  return `${values.year}-${values.month}-${values.day}`;
}
