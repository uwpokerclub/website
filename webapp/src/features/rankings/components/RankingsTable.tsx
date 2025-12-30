import { useEffect, useState, useMemo, useCallback } from "react";
import { Table, TableColumn, Input, Pagination, Button } from "@uwpokerclub/components";
import { Ranking } from "@/types";
import { useAuth } from "@/hooks";
import { FaSearch, FaTimes, FaTrophy, FaDownload } from "react-icons/fa";
import styles from "./RankingsTable.module.css";

const ITEMS_PER_PAGE = 25;

type RankingsTableProps = {
  rankings: Ranking[];
  semesterId: string;
};

export function RankingsTable({ rankings, semesterId }: RankingsTableProps) {
  const { hasPermission } = useAuth();
  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedSearchQuery, setDebouncedSearchQuery] = useState("");
  const [currentPage, setCurrentPage] = useState(1);

  // Debounce search query
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearchQuery(searchQuery);
      setCurrentPage(1);
    }, 300);

    return () => clearTimeout(timer);
  }, [searchQuery]);

  // Reset pagination when rankings change
  useEffect(() => {
    setCurrentPage(1);
    setSearchQuery("");
    setDebouncedSearchQuery("");
  }, [semesterId]);

  // Filter rankings by search query
  const filteredRankings = useMemo(() => {
    if (!debouncedSearchQuery.trim()) {
      return rankings;
    }

    const query = debouncedSearchQuery.toLowerCase();
    return rankings.filter(
      (ranking) =>
        ranking.firstName.toLowerCase().includes(query) ||
        ranking.lastName.toLowerCase().includes(query) ||
        `${ranking.firstName} ${ranking.lastName}`.toLowerCase().includes(query),
    );
  }, [rankings, debouncedSearchQuery]);

  // Paginate rankings
  const paginatedRankings = useMemo(() => {
    const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
    const endIndex = startIndex + ITEMS_PER_PAGE;
    return filteredRankings.slice(startIndex, endIndex);
  }, [filteredRankings, currentPage]);

  // Handle clear search
  const handleClearSearch = useCallback(() => {
    setSearchQuery("");
  }, []);

  // Handle CSV export
  const handleExport = async () => {
    const response = await fetch(`/api/v2/semesters/${semesterId}/rankings/export`, {
      credentials: "include",
    });

    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download =
      response.headers.get("Content-Disposition")?.split("filename=")[1]?.replace(/"/g, "") || "rankings.csv";

    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  };

  // Get rank for a given ranking (accounting for original position)
  const getRank = (ranking: Ranking): number => {
    return rankings.findIndex((r) => r.id === ranking.id) + 1;
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
            onChange={(e) => setSearchQuery(e.target.value)}
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
          Showing {paginatedRankings.length} of {filteredRankings.length} rankings
          {debouncedSearchQuery && ` matching "${debouncedSearchQuery}"`}
        </p>
      </div>

      {/* Table */}
      <div className={styles.tableWrapper}>
        <Table
          data-qa="rankings-table"
          variant="striped"
          headerVariant="primary"
          data={paginatedRankings}
          columns={columns}
          rowProps={(row) => ({ "data-qa": `ranking-row-${row.id}` }) as React.HTMLAttributes<HTMLTableRowElement>}
          emptyState={
            <div className={styles.emptyState}>
              <div className={styles.emptyIllustration}>
                <FaTrophy size={64} />
              </div>
              {rankings.length === 0 ? (
                <div data-qa="rankings-empty">
                  <h3>No rankings yet</h3>
                  <p>No members have earned points this semester yet.</p>
                </div>
              ) : (
                <div data-qa="rankings-no-results">
                  <h3>No results found</h3>
                  <p>No rankings found matching &quot;{debouncedSearchQuery}&quot;</p>
                  <p className={styles.emptyHint}>Try adjusting your search terms</p>
                </div>
              )}
            </div>
          }
        />
      </div>

      {/* Pagination */}
      {filteredRankings.length > ITEMS_PER_PAGE && (
        <div className={styles.paginationContainer} data-qa="rankings-pagination">
          <Pagination
            variant="compact"
            totalItems={filteredRankings.length}
            pageSize={ITEMS_PER_PAGE}
            currentPage={currentPage}
            onPageChange={setCurrentPage}
          />
        </div>
      )}
    </div>
  );
}
