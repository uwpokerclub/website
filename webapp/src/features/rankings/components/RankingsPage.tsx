import { useEffect, useState } from "react";
import { Spinner, Button } from "@uwpokerclub/components";
import { useCurrentSemester } from "@/hooks";
import { Ranking } from "@/types";
import { RankingsPodium } from "./RankingsPodium";
import { RankingsTable } from "./RankingsTable";
import styles from "./RankingsPage.module.css";

export function RankingsPage() {
  const { currentSemester, loading: semesterLoading } = useCurrentSemester();
  const [rankings, setRankings] = useState<Ranking[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!currentSemester) {
      setIsLoading(false);
      return;
    }

    const fetchRankings = async () => {
      setIsLoading(true);
      setError(null);

      try {
        const response = await fetch(`/api/v2/semesters/${currentSemester.id}/rankings`, {
          credentials: "include",
        });

        if (!response.ok) {
          throw new Error(`Failed to fetch rankings: ${response.statusText}`);
        }

        const resp: { data: Ranking[] } = await response.json();
        setRankings(resp.data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "An error occurred while fetching rankings");
      } finally {
        setIsLoading(false);
      }
    };

    fetchRankings();
  }, [currentSemester]);

  // Show loading state
  if (isLoading || semesterLoading) {
    return (
      <div className={styles.container}>
        <div className={styles.centerContent} data-qa="rankings-loading">
          <Spinner size="lg" />
          <p>Loading rankings...</p>
        </div>
      </div>
    );
  }

  // Show error state
  if (error) {
    return (
      <div className={styles.container}>
        <div className={styles.errorState} data-qa="rankings-error">
          <p>Error: {error}</p>
          <Button data-qa="retry-btn" onClick={() => window.location.reload()}>
            Retry
          </Button>
        </div>
      </div>
    );
  }

  // Show empty state when no semester is selected
  if (!currentSemester) {
    return (
      <div className={styles.container}>
        <div className={styles.emptyState} data-qa="rankings-no-semester">
          <p>Please select a semester to view rankings.</p>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      {/* Podium - only shown when 3+ members have points */}
      <RankingsPodium rankings={rankings} />

      {/* Table with search and pagination */}
      <RankingsTable rankings={rankings} semesterId={currentSemester.id} />
    </div>
  );
}
