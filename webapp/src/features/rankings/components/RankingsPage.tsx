import { useCallback, useEffect, useMemo, useState } from "react";
import { Spinner, Button } from "@uwpokerclub/components";
import { useCurrentSemester } from "@/hooks";
import { useRankings } from "../hooks/useRankingQueries";
import { RankingsPodium } from "./RankingsPodium";
import { RankingsTable } from "./RankingsTable";
import styles from "./RankingsPage.module.css";

const ITEMS_PER_PAGE = 25;

export function RankingsPage() {
  const { currentSemester, loading: semesterLoading } = useCurrentSemester();
  const [currentPage, setCurrentPage] = useState(1);
  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedSearchQuery, setDebouncedSearchQuery] = useState("");

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearchQuery(searchQuery);
      setCurrentPage(1);
    }, 300);

    return () => clearTimeout(timer);
  }, [searchQuery]);

  // Reset pagination and search when semester changes
  useEffect(() => {
    setCurrentPage(1);
    setSearchQuery("");
    setDebouncedSearchQuery("");
  }, [currentSemester?.id]);

  const queryParams = useMemo(
    () => ({
      limit: ITEMS_PER_PAGE,
      offset: (currentPage - 1) * ITEMS_PER_PAGE,
      ...(debouncedSearchQuery && { search: debouncedSearchQuery }),
    }),
    [currentPage, debouncedSearchQuery],
  );

  const { data, isLoading, error: queryError } = useRankings(currentSemester?.id, queryParams);
  const rankings = useMemo(() => data?.data ?? [], [data]);
  const totalItems = data?.total ?? 0;
  const error = queryError?.message ?? null;

  const handlePageChange = useCallback((page: number) => {
    setCurrentPage(page);
  }, []);

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
      {currentPage === 1 && <RankingsPodium rankings={rankings} />}

      <RankingsTable
        rankings={rankings}
        semesterId={currentSemester.id}
        totalItems={totalItems}
        currentPage={currentPage}
        pageSize={ITEMS_PER_PAGE}
        onPageChange={handlePageChange}
        searchQuery={searchQuery}
        activeSearchQuery={debouncedSearchQuery}
        onSearchChange={setSearchQuery}
      />
    </div>
  );
}
