import { DashboardCard } from "../DashboardCard";
import { CardActionLink } from "../CardActionLink";
import { CARD_TITLES } from "../../dashboardLayout";
import { useTopRankings } from "../../hooks/useDashboardQueries";
import type { RankingEntry } from "../../api/dashboardApi";
import styles from "./LeaderboardCard.module.css";

type Props = { semesterId: string };

export function LeaderboardCard({ semesterId }: Props) {
  const { data, isLoading, isError, refetch } = useTopRankings(semesterId);
  const status = isError ? "error" : isLoading || !data ? "loading" : data.length === 0 ? "empty" : "ready";

  return (
    <DashboardCard
      title={CARD_TITLES.leaderboard}
      status={status}
      onRetry={() => refetch()}
      emptyMessage="No points awarded yet this term."
      action={<CardActionLink to="/admin/rankings">Full rankings</CardActionLink>}
      data-qa="leaderboard-card"
    >
      {() => <Body rows={data as RankingEntry[]} />}
    </DashboardCard>
  );
}

/**
 * The gap between first and fifth is the interesting part and a bare list hides it,
 * so each row carries a bar scaled to the leader's points. Emphasis rather than
 * categorical: only the leader is coloured, everyone else is the de-emphasis grey.
 *
 * Rank comes from the API's `position`, which semester_rankings_view computes with
 * RANK() — so a tie shows two firsts and no second, which the array index would get
 * wrong.
 */
function Body({ rows }: { rows: RankingEntry[] }) {
  const leaderPoints = Math.max(...rows.map((row) => row.points), 1);

  return (
    <ol className={styles.list}>
      {rows.map((row) => (
        <li key={row.id} className={styles.row} data-leader={row.position === 1 || undefined}>
          <span className={styles.rank}>{row.position}</span>
          <span className={styles.name}>
            {row.firstName} {row.lastName}
          </span>
          <span className={styles.bar} aria-hidden="true">
            <i style={{ width: `${(row.points / leaderPoints) * 100}%` }} />
          </span>
          <span className={styles.points}>{row.points}</span>
        </li>
      ))}
    </ol>
  );
}
