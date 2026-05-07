import { useCallback } from "react";
import { Table, TableColumn, Input, Pagination, Button, useToast } from "@uwpokerclub/components";
import { Ranking } from "@/types";
import { useAuth } from "@/hooks";
import { FaSearch, FaTimes, FaTrophy, FaDownload } from "react-icons/fa";
import { exportRankings } from "../api/rankingsApi";
import styles from "./RankingsTable.module.css";

type RankingsTableProps = {
  rankings: Ranking[];
  semesterId: string;
  totalItems: number;
  currentPage: number;
  pageSize: number;
  onPageChange: (page: number) => void;
  searchQuery: string;
  activeSearchQuery: string;
  onSearchChange: (query: string) => void;
};

export function RankingsTable({
  rankings,
  semesterId,
  totalItems,
  currentPage,
  pageSize,
  onPageChange,
  searchQuery,
  activeSearchQuery,
  onSearchChange,
}: RankingsTableProps) {
  const { hasPermission } = useAuth();
  const { showToast } = useToast();

  const handleClearSearch = useCallback(() => {
    onSearchChange("");
  }, [onSearchChange]);

  const handleExport = async () => {
    try {
      const { blob, filename } = await exportRankings(semesterId);
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = filename;

      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (err) {
      showToast({
        message: err instanceof Error ? err.message : "Failed to export rankings",
        variant: "error",
        duration: 5000,
      });
    }
  };

  // Get rank for a given ranking (uses server-provided position)
  const getRank = (ranking: Ranking): number => {
    return ranking.position;
  };

  // Define table columns
  const columns: TableColumn<Ranking>[] = [
    {
      key: "rank",
      header: "Rank",
      accessor: (row) => getRank(row),
      sortable: false,
      headerProps: { "data-qa": "rank-header" } as React.ThHTMLAttributes<HTMLTableCellElement>,
      cellProps: (row) => ({ "data-qa": `ranking-rank-${row.id}` }) as React.TdHTMLAttributes<HTMLTableCellElement>,
    },
    {
      key: "name",
      header: "Name",
      accessor: (row) => `${row.firstName} ${row.lastName}`,
      sortable: false,
      headerProps: { "data-qa": "name-header" } as React.ThHTMLAttributes<HTMLTableCellElement>,
      cellProps: (row) => ({ "data-qa": `ranking-name-${row.id}` }) as React.TdHTMLAttributes<HTMLTableCellElement>,
    },
    {
      key: "points",
      header: "Points",
      accessor: "points",
      sortable: false,
      headerProps: { "data-qa": "points-header" } as React.ThHTMLAttributes<HTMLTableCellElement>,
      cellProps: (row) => ({ "data-qa": `ranking-points-${row.id}` }) as React.TdHTMLAttributes<HTMLTableCellElement>,
    },
  ];

  return (
    <div className={styles.container}>
      {/* Search and action bar */}
      <div className={styles.searchContainer}>
        <div className={styles.searchInputWrapper}>
          <Input
            data-qa="input-rankings-search"
            type="search"
            placeholder="Search by name..."
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
            prefix={<FaSearch />}
            suffix={
              searchQuery ? (
                <button
                  type="button"
                  onClick={handleClearSearch}
                  className={styles.clearButton}
                  aria-label="Clear search"
                  data-qa="clear-search-btn"
                >
                  <FaTimes />
                </button>
              ) : null
            }
            fullWidth
          />
        </div>
        {hasPermission("export", "semester", "rankings") && (
          <Button data-qa="export-rankings-btn" onClick={handleExport} iconBefore={<FaDownload />}>
            Export CSV
          </Button>
        )}
      </div>

      <div className={styles.resultsInfo} data-qa="rankings-results-info">
        <p>
          Showing {rankings.length} of {totalItems} rankings
          {activeSearchQuery && ` matching "${activeSearchQuery}"`}
        </p>
      </div>

      {/* Table */}
      <div className={styles.tableWrapper}>
        <Table
          data-qa="rankings-table"
          variant="striped"
          headerVariant="primary"
          data={rankings}
          columns={columns}
          rowProps={(row) => ({ "data-qa": `ranking-row-${row.id}` }) as React.HTMLAttributes<HTMLTableRowElement>}
          emptyState={
            <div className={styles.emptyState}>
              <div className={styles.emptyIllustration}>
                <FaTrophy size={64} />
              </div>
              {activeSearchQuery ? (
                <>
                  <h3 data-qa="rankings-no-results">No results found</h3>
                  <p>No rankings found matching &quot;{activeSearchQuery}&quot;</p>
                  <p className={styles.emptyHint}>Try adjusting your search terms</p>
                </>
              ) : (
                <>
                  <h3 data-qa="rankings-empty">No rankings yet</h3>
                  <p>No members have earned points this semester yet.</p>
                </>
              )}
            </div>
          }
        />
      </div>

      {/* Pagination */}
      {totalItems > pageSize && (
        <div className={styles.paginationContainer} data-qa="rankings-pagination">
          <Pagination
            variant="compact"
            totalItems={totalItems}
            pageSize={pageSize}
            currentPage={currentPage}
            onPageChange={onPageChange}
          />
        </div>
      )}
    </div>
  );
}
