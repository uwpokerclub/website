import { useCallback, useEffect, useState } from "react";
import { Spinner, Button } from "@uwpokerclub/components";
import { useCurrentSemester } from "@/hooks";
import { Ranking } from "@/types";
import { RankingsPodium } from "./RankingsPodium";
import { RankingsTable } from "./RankingsTable";
import styles from "./RankingsPage.module.css";

const ITEMS_PER_PAGE = 25;

export function RankingsPage() {
  const { currentSemester, loading: semesterLoading } = useCurrentSemester();
  const [rankings, setRankings] = useState<Ranking[]>([]);
  const [totalItems, setTotalItems] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Reset pagination when semester changes
  useEffect(() => {
    setCurrentPage(1);
  }, [currentSemester?.id]);

  useEffect(() => {
    if (!currentSemester) {
      setIsLoading(false);
      return;
    }

    const fetchRankings = async () => {
      setIsLoading(true);
      setError(null);

      try {
        const offset = (currentPage - 1) * ITEMS_PER_PAGE;
        const response = await fetch(
          `/api/v2/semesters/${currentSemester.id}/rankings?limit=${ITEMS_PER_PAGE}&offset=${offset}`,
          { credentials: "include" },
        );

        if (!response.ok) {
          throw new Error(`Failed to fetch rankings: ${response.statusText}`);
        }

        const resp: { data: Ranking[]; total: number } = await response.json();
        setRankings(resp.data);
        setTotalItems(resp.total);
      } catch (err) {
        setError(err instanceof Error ? err.message : "An error occurred while fetching rankings");
      } finally {
        setIsLoading(false);
      }
    };

    fetchRankings();
  }, [currentSemester, currentPage]);

  const handlePageChange = useCallback((page: number) => {
    setCurrentPage(page);
  }, []);

  // Show loading state
  if ((isLoading || semesterLoading) && rankings.length === 0) {
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
      {/* Podium - only shown on page 1 where top-3 are present */}
      {currentPage === 1 && <RankingsPodium rankings={rankings} />}

      {/* Table with search and pagination */}
      <RankingsTable
        rankings={rankings}
        semesterId={currentSemester.id}
        totalItems={totalItems}
        currentPage={currentPage}
        pageSize={ITEMS_PER_PAGE}
        onPageChange={handlePageChange}
      />
    </div>
  );
}
