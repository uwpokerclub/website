import { useMemo, useCallback } from "react";
import { Table, TableColumn, Input, Pagination, Spinner, Button } from "@uwpokerclub/components";
import { FaSearch, FaTimes, FaUsers, FaSignInAlt, FaSignOutAlt, FaTrash } from "react-icons/fa";
import { Entry, Event, EventState } from "@/types";
import { useAuth } from "@/hooks";
import { useSignInEntry, useSignOutEntry, useUnregisterEntry } from "@/features/entries/hooks/useEntryQueries";

import styles from "./EntriesTable.module.css";

type EntriesTableProps = {
  entries: Entry[];
  event: Event;
  semesterId: string;
  isLoading: boolean;
  totalItems: number;
  currentPage: number;
  pageSize: number;
  onPageChange: (page: number) => void;
  searchQuery: string;
  onSearchChange: (query: string) => void;
};

// Rows sharing a signed_out_at are one tie group that EndEvent scores identically.
const signedOutKey = (entry: Entry): number | null =>
  entry.signedOutAt ? new Date(entry.signedOutAt).getTime() : null;

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

export function EntriesTable({
  entries,
  event,
  semesterId,
  isLoading,
  totalItems,
  currentPage,
  pageSize,
  onPageChange,
  searchQuery,
  onSearchChange,
}: EntriesTableProps) {
  const { hasPermission } = useAuth();

  const signInMutation = useSignInEntry();
  const signOutMutation = useSignOutEntry();
  const unregisterMutation = useUnregisterEntry();

  // Track which specific entry is being processed (mutations may fire in parallel)
  const processingEntry =
    (signInMutation.isPending && signInMutation.variables?.membershipId) ||
    (signOutMutation.isPending && signOutMutation.variables?.membershipId) ||
    (unregisterMutation.isPending && unregisterMutation.variables?.membershipId) ||
    null;

  // Handle clear search
  const handleClearSearch = useCallback(() => {
    onSearchChange("");
  }, [onSearchChange]);

  const handleSignOut = useCallback(
    (membershipId: string) => {
      signOutMutation.mutate({ semesterId, eventId: event.id, membershipId });
    },
    [semesterId, event.id, signOutMutation],
  );

  const handleSignIn = useCallback(
    (membershipId: string) => {
      signInMutation.mutate({ semesterId, eventId: event.id, membershipId });
    },
    [semesterId, event.id, signInMutation],
  );

  const handleRemove = useCallback(
    (membershipId: string) => {
      unregisterMutation.mutate({ semesterId, eventId: event.id, membershipId });
    },
    [semesterId, event.id, unregisterMutation],
  );

  // Action buttons component
  const ActionButtons = useCallback(
    ({ entry }: { entry: Entry }) => {
      const isProcessing = processingEntry === entry.membershipId;
      const isEnded = event.state === EventState.Ended;

      if (isEnded || !entry.membershipId) {
        return null;
      }

      return (
        <div className={styles.actionButtons}>
          {hasPermission("signin", "event", "participant") && entry.signedOutAt && (
            <Button
              variant="primary"
              size="small"
              onClick={() => handleSignIn(entry.membershipId!)}
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
              onClick={() => handleSignOut(entry.membershipId!)}
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
              onClick={() => handleRemove(entry.membershipId!)}
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

  const rowOrdinalByEntryId = useMemo(() => {
    const offset = (currentPage - 1) * pageSize;
    return new Map(entries.map((entry, index) => [entry.entryId, offset + index + 1]));
  }, [entries, currentPage, pageSize]);

  // Placement is derived from row order rather than a server field (issue #425), so it only
  // holds for a finished event listed in full - under search the index is a position within a
  // filtered subset. Tied rows share the group's first position, as EndEvent scores them. A tie
  // group split across pages restarts numbering; the neighbouring rows aren't loaded.
  const placementByEntryId = useMemo(() => {
    const places = new Map<number, number>();
    if (event.state !== EventState.Ended || searchQuery) {
      return places;
    }
    const offset = (currentPage - 1) * pageSize;
    for (let i = 0; i < entries.length; ) {
      let j = i;
      while (j + 1 < entries.length && signedOutKey(entries[j + 1]) === signedOutKey(entries[i])) {
        j++;
      }
      for (let k = i; k <= j; k++) {
        places.set(entries[k].entryId, offset + i + 1);
      }
      i = j + 1;
    }
    return places;
  }, [entries, event.state, searchQuery, currentPage, pageSize]);

  // Define table columns
  const columns: TableColumn<Entry>[] = useMemo(() => {
    const cols: TableColumn<Entry>[] = [
      {
        key: "index",
        header: "#",
        accessor: () => "",
        sortable: false,
        render: (_, row) => rowOrdinalByEntryId.get(row.entryId) ?? "",
      },
      {
        key: "firstName",
        header: "First Name",
        accessor: (row) => row.firstName || (row.membershipId === null ? "Unknown" : ""),
        sortable: false,
      },
      {
        key: "lastName",
        header: "Last Name",
        accessor: (row) => row.lastName || (row.membershipId === null ? "Member" : ""),
        sortable: false,
      },
      {
        key: "studentNumber",
        header: "Student Number",
        accessor: (row) => row.id || "--",
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
        accessor: () => "",
        sortable: false,
        render: (_, row) => {
          const place = placementByEntryId.get(row.entryId);
          if (place === undefined) {
            return "--";
          }
          const badgeClass =
            place === 1
              ? styles.placementFirst
              : place === 2
                ? styles.placementSecond
                : place === 3
                  ? styles.placementThird
                  : styles.placementOther;
          return <span className={`${styles.placementBadge} ${badgeClass}`}>{place}</span>;
        },
      },
      {
        key: "points",
        header: "Points",
        accessor: (row) => (row.points !== undefined && row.points !== null ? String(row.points) : "--"),
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
  }, [event.state, hasPermission, ActionButtons, rowOrdinalByEntryId, placementByEntryId]);

  // Empty state component
  const emptyState = useMemo(
    () => (
      <div className={styles.emptyState}>
        <div className={styles.emptyIllustration}>
          <FaUsers size={64} />
        </div>
        {entries.length === 0 && !searchQuery ? (
          <>
            <h3>No entries yet</h3>
            <p>No participants have been registered for this event yet.</p>
          </>
        ) : (
          <>
            <h3>No results found</h3>
            <p>No entries found matching &quot;{searchQuery}&quot;</p>
            <p className={styles.emptyHint}>Try adjusting your search terms</p>
          </>
        )}
      </div>
    ),
    [entries.length, searchQuery],
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
        <strong>{totalItems + event.rebuys} Entries</strong>
        <span>
          ({totalItems} Players, {event.rebuys} Rebuys)
        </span>
      </div>

      <div className={styles.searchContainer}>
        <div className={styles.searchInputWrapper}>
          <Input
            type="search"
            placeholder="Search entries..."
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
          Showing {entries.length} of {totalItems} entries
          {searchQuery && ` matching "${searchQuery}"`}
        </p>
      </div>

      <div className={styles.tableWrapper}>
        <Table
          variant="striped"
          headerVariant="primary"
          data={entries}
          columns={columns}
          loading={isLoading}
          emptyState={emptyState}
          getRowKey={(row) => row.membershipId ?? `entry-${row.entryId}`}
          data-qa="entries-table"
        />
      </div>

      {totalItems > pageSize && (
        <div className={styles.paginationContainer}>
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
