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
  const { data, isLoading, isError, refetch } = useSpotlight(semesterId);
  const status = isError ? "error" : isLoading ? "loading" : !data ? "empty" : "ready";

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
  const live = event.state === EVENT_STATE_STARTED && start <= new Date();

  return (
    <div className={styles.container}>
      <p className={styles.state} data-live={live || undefined}>
        {live ? (
          <>
            <span className={styles.pulse} aria-hidden="true" />
            Live · started {start.toLocaleTimeString("en-CA", { hour: "numeric", minute: "2-digit" })}
          </>
        ) : (
          <>Scheduled · {start.toLocaleDateString("en-CA", { weekday: "long", month: "short", day: "numeric" })}</>
        )}
      </p>
      <h4 className={styles.name}>{event.name}</h4>
      <p className={styles.format}>{event.format}</p>

      <dl className={styles.stats}>
        <div>
          <dd>{event.entries}</dd>
          <dt>entries</dt>
        </div>
        <div>
          <dd>{event.rebuys}</dd>
          <dt>rebuys</dt>
        </div>
      </dl>

      <Link className={styles.cta} to={`/admin/events/${event.id}`}>
        Open event
      </Link>
    </div>
  );
}
