import { useState, useEffect, useMemo, useCallback } from "react";
import { Table, TableColumn, Input, Pagination, Spinner, Button } from "@uwpokerclub/components";
import { FaSearch, FaTimes, FaUsers, FaSignInAlt, FaSignOutAlt, FaTrash } from "react-icons/fa";
import { Entry } from "../../../types";
import { useAuth } from "@/hooks";
import { EventState } from "@/sdk/events";
import { EventResponse } from "../api/eventApi";

import styles from "./EntriesTable.module.css";

const ITEMS_PER_PAGE = 25;

type EntriesTableProps = {
  entries: Entry[];
  event: EventResponse;
  semesterId: string;
  isLoading: boolean;
  updateParticipants: () => void;
};

// Format signed out at date
const formatSignedOutAt = (entry: Entry): string => {
  if (!entry.signedOutAt) return "Not Signed Out";
  return new Date(entry.signedOutAt).toLocaleString("en-US", {
    hour12: true,
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
};

export function EntriesTable({ entries, event, semesterId, isLoading, updateParticipants }: EntriesTableProps) {
  const { hasPermission } = useAuth();

  // Search state
  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedSearchQuery, setDebouncedSearchQuery] = useState("");

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);

  // Processing state for action buttons
  const [processingEntry, setProcessingEntry] = useState<string | null>(null);

  // Debounce search query
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearchQuery(searchQuery);
      setCurrentPage(1);
    }, 300);

    return () => clearTimeout(timer);
  }, [searchQuery]);

  // Filter entries by search query
  const filteredEntries = useMemo(() => {
    if (!debouncedSearchQuery.trim()) {
      return entries;
    }

    const query = debouncedSearchQuery.toLowerCase();
    return entries.filter((entry) => `${entry.firstName} ${entry.lastName}`.toLowerCase().includes(query));
  }, [entries, debouncedSearchQuery]);

  // Paginate entries
  const paginatedEntries = useMemo(() => {
    const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
    const endIndex = startIndex + ITEMS_PER_PAGE;
    return filteredEntries.slice(startIndex, endIndex);
  }, [filteredEntries, currentPage]);

  // Handle clear search
  const handleClearSearch = useCallback(() => {
    setSearchQuery("");
  }, []);

  // V2 API actions
  const handleSignOut = useCallback(
    async (membershipId: string) => {
      setProcessingEntry(membershipId);
      try {
        const response = await fetch(
          `/api/v2/semesters/${semesterId}/events/${event.id}/entries/${membershipId}/sign-out`,
          {
            method: "POST",
            credentials: "include",
          },
        );

        if (response.ok) {
          updateParticipants();
        }
      } finally {
        setProcessingEntry(null);
      }
    },
    [semesterId, event.id, updateParticipants],
  );

  const handleSignIn = useCallback(
    async (membershipId: string) => {
      setProcessingEntry(membershipId);
      try {
        const response = await fetch(
          `/api/v2/semesters/${semesterId}/events/${event.id}/entries/${membershipId}/sign-in`,
          {
            method: "POST",
            credentials: "include",
          },
        );

        if (response.ok) {
          updateParticipants();
        }
      } finally {
        setProcessingEntry(null);
      }
    },
    [semesterId, event.id, updateParticipants],
  );

  const handleRemove = useCallback(
    async (membershipId: string) => {
      setProcessingEntry(membershipId);
      try {
        const response = await fetch(`/api/v2/semesters/${semesterId}/events/${event.id}/entries/${membershipId}`, {
          method: "DELETE",
          credentials: "include",
        });

        if (response.ok || response.status === 204) {
          updateParticipants();
        }
      } finally {
        setProcessingEntry(null);
      }
    },
    [semesterId, event.id, updateParticipants],
  );

  // Action buttons component
  const ActionButtons = useCallback(
    ({ entry }: { entry: Entry }) => {
      const isProcessing = processingEntry === entry.membershipId;
      const isEnded = event.state === EventState.Ended;

      if (isEnded) {
        return null;
      }

      return (
        <div className={styles.actionButtons}>
          {hasPermission("signin", "event", "participant") && entry.signedOutAt && (
            <Button
              variant="primary"
              size="small"
              onClick={() => handleSignIn(entry.membershipId)}
              loading={isProcessing}
              aria-label="Sign Back In"
              data-qa="sign-in-btn"
              iconBefore={<FaSignInAlt />}
            />
          )}
          {hasPermission("signout", "event", "participant") && !entry.signedOutAt && (
            <Button
              variant="secondary"
              size="small"
              onClick={() => handleSignOut(entry.membershipId)}
              loading={isProcessing}
              aria-label="Sign Out"
              data-qa="sign-out-btn"
              iconBefore={<FaSignOutAlt />}
            />
          )}
          {hasPermission("delete", "event", "participant") && (
            <Button
              variant="destructive"
              size="small"
              onClick={() => handleRemove(entry.membershipId)}
              loading={isProcessing}
              aria-label="Remove"
              data-qa="remove-btn"
              iconBefore={<FaTrash />}
            />
          )}
        </div>
      );
    },
    [processingEntry, event.state, hasPermission, handleSignIn, handleSignOut, handleRemove],
  );

  // Define table columns
  const columns: TableColumn<Entry>[] = useMemo(() => {
    const cols: TableColumn<Entry>[] = [
      {
        key: "index",
        header: "#",
        accessor: () => "",
        sortable: false,
        render: (_, row) => {
          const index = paginatedEntries.findIndex((e) => e.membershipId === row.membershipId);
          return (currentPage - 1) * ITEMS_PER_PAGE + index + 1;
        },
      },
      {
        key: "firstName",
        header: "First Name",
        accessor: "firstName",
        sortable: false,
      },
      {
        key: "lastName",
        header: "Last Name",
        accessor: "lastName",
        sortable: false,
      },
      {
        key: "studentNumber",
        header: "Student Number",
        accessor: "id",
        sortable: false,
      },
      {
        key: "signedOutAt",
        header: "Signed Out At",
        accessor: (row) => formatSignedOutAt(row),
        sortable: false,
      },
      {
        key: "placement",
        header: "Place",
        accessor: (row) => (row.placement ? String(row.placement) : "--"),
        sortable: false,
      },
    ];

    // Only add actions column if event is not ended and user has any action permission
    if (
      event.state !== EventState.Ended &&
      (hasPermission("signin", "event", "participant") ||
        hasPermission("signout", "event", "participant") ||
        hasPermission("delete", "event", "participant"))
    ) {
      cols.push({
        key: "actions",
        header: "Actions",
        accessor: () => "",
        sortable: false,
        render: (_, row) => <ActionButtons entry={row} />,
      });
    }

    return cols;
  }, [currentPage, paginatedEntries, event.state, hasPermission, ActionButtons]);

  // Empty state component
  const emptyState = useMemo(
    () => (
      <div className={styles.emptyState}>
        <div className={styles.emptyIllustration}>
          <FaUsers size={64} />
        </div>
        {entries.length === 0 ? (
          <>
            <h3>No entries yet</h3>
            <p>No participants have been registered for this event yet.</p>
          </>
        ) : (
          <>
            <h3>No results found</h3>
            <p>No entries found matching &quot;{debouncedSearchQuery}&quot;</p>
            <p className={styles.emptyHint}>Try adjusting your search terms</p>
          </>
        )}
      </div>
    ),
    [entries.length, debouncedSearchQuery],
  );

  // Loading state
  if (isLoading && entries.length === 0) {
    return (
      <div className={styles.centerContent}>
        <Spinner size="lg" />
        <p>Loading entries...</p>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <strong>{entries.length + event.rebuys} Entries</strong>
        <span>
          ({entries.length} Players, {event.rebuys} Rebuys)
        </span>
      </div>

      <div className={styles.searchContainer}>
        <div className={styles.searchInputWrapper}>
          <Input
            type="search"
            placeholder="Search entries..."
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
                >
                  <FaTimes />
                </button>
              ) : null
            }
            fullWidth
            data-qa="input-search"
          />
        </div>
      </div>

      <div className={styles.resultsInfo}>
        <p>
          Showing {paginatedEntries.length} of {filteredEntries.length} entries
          {debouncedSearchQuery && ` matching "${debouncedSearchQuery}"`}
        </p>
      </div>

      <div className={styles.tableWrapper}>
        <Table
          variant="striped"
          headerVariant="primary"
          data={paginatedEntries}
          columns={columns}
          loading={isLoading}
          emptyState={emptyState}
          getRowKey={(row) => row.membershipId}
        />
      </div>

      {filteredEntries.length > ITEMS_PER_PAGE && (
        <div className={styles.paginationContainer}>
          <Pagination
            variant="compact"
            totalItems={filteredEntries.length}
            pageSize={ITEMS_PER_PAGE}
            currentPage={currentPage}
            onPageChange={setCurrentPage}
          />
        </div>
      )}
    </div>
  );
}
