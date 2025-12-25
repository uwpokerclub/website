import { useContext, useEffect, useState, useMemo, useCallback, useRef } from "react";
import { Link } from "react-router-dom";
import { Table, TableColumn, Button, Input, Pagination, Spinner, useToast } from "@uwpokerclub/components";
import { SemesterContext } from "@/contexts";
import { useAuth } from "@/hooks";
import { FaSearch, FaTimes, FaPlus, FaCalendarAlt, FaPencilAlt, FaEllipsisV, FaStop, FaRedo } from "react-icons/fa";
import { EventState } from "@/sdk/events";
import { Participant } from "@/sdk/participants";
import styles from "./ListEvents.module.css";

const ITEMS_PER_PAGE = 25;

type ListEventsResponse = {
  id: number;
  name: string;
  format: string;
  notes: string;
  semesterId: string;
  startDate: string;
  state: number;
  entries?: Participant[];
};

type EventActionsProps = {
  event: ListEventsResponse;
  onActionComplete: () => void;
};

function EventActions({ event, onActionComplete }: EventActionsProps) {
  const { hasPermission } = useAuth();
  const semesterContext = useContext(SemesterContext);
  const { showToast } = useToast();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  // Close menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setIsMenuOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleEndEvent = async () => {
    if (!semesterContext?.currentSemester) return;

    const confirmed = window.confirm(
      `Are you sure you want to end "${event.name}"? This will finalize the event results.`,
    );
    if (!confirmed) return;

    setIsProcessing(true);
    try {
      const response = await fetch(`/api/v2/semesters/${semesterContext.currentSemester.id}/events/${event.id}/end`, {
        method: "POST",
        credentials: "include",
      });
      if (!response.ok) {
        throw new Error(`Failed to end event: ${response.statusText}`);
      }
      showToast({
        message: `"${event.name}" has been ended successfully`,
        variant: "success",
        duration: 3000,
      });
      onActionComplete();
    } catch (err) {
      showToast({
        message: err instanceof Error ? err.message : "Failed to end event",
        variant: "error",
        duration: 5000,
      });
    } finally {
      setIsProcessing(false);
      setIsMenuOpen(false);
    }
  };

  const handleRestartEvent = async () => {
    if (!semesterContext?.currentSemester) return;
    setIsProcessing(true);
    try {
      const response = await fetch(
        `/api/v2/semesters/${semesterContext.currentSemester.id}/events/${event.id}/restart`,
        { method: "POST", credentials: "include" },
      );
      if (!response.ok) {
        throw new Error(`Failed to restart event: ${response.statusText}`);
      }
      showToast({
        message: `"${event.name}" has been restarted successfully`,
        variant: "success",
        duration: 3000,
      });
      onActionComplete();
    } catch (err) {
      showToast({
        message: err instanceof Error ? err.message : "Failed to restart event",
        variant: "error",
        duration: 5000,
      });
    } finally {
      setIsProcessing(false);
      setIsMenuOpen(false);
    }
  };

  const showEdit = hasPermission("edit", "event");
  const isEditDisabled = event.state === EventState.Ended;
  const showEndEvent = event.state === EventState.Started && hasPermission("end", "event");
  const showRestartEvent = event.state === EventState.Ended && hasPermission("restart", "event");
  const showMenu = showEdit || showEndEvent || showRestartEvent;

  if (!showMenu) {
    return null;
  }

  return (
    <div className={styles.actions} ref={menuRef}>
      <div className={styles.menuWrapper}>
        <button
          className={styles.iconButton}
          onClick={() => setIsMenuOpen(!isMenuOpen)}
          title="Actions"
          aria-label="Actions"
          disabled={isProcessing}
        >
          {isProcessing ? <Spinner size="sm" /> : <FaEllipsisV />}
        </button>

        {isMenuOpen && (
          <div className={styles.dropdownMenu}>
            {showEdit &&
              (isEditDisabled ? (
                <span className={`${styles.menuItem} ${styles.menuItemDisabled}`}>
                  <FaPencilAlt /> Edit Event
                </span>
              ) : (
                <Link to={`${event.id}/edit`} className={styles.menuItem}>
                  <FaPencilAlt /> Edit Event
                </Link>
              ))}
            {showEndEvent && (
              <button className={styles.menuItem} onClick={handleEndEvent} disabled={isProcessing}>
                <FaStop /> End Event
              </button>
            )}
            {showRestartEvent && (
              <button className={styles.menuItem} onClick={handleRestartEvent} disabled={isProcessing}>
                <FaRedo /> Restart Event
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

const formatDate = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
};

export function ListEvents() {
  const semesterContext = useContext(SemesterContext);
  const { hasPermission } = useAuth();
  const [events, setEvents] = useState<ListEventsResponse[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  // Debounce search query
  const [debouncedSearchQuery, setDebouncedSearchQuery] = useState("");

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearchQuery(searchQuery);
      setCurrentPage(1);
    }, 300);

    return () => clearTimeout(timer);
  }, [searchQuery]);

  // Fetch events from API
  useEffect(() => {
    if (!semesterContext?.currentSemester) {
      setIsLoading(false);
      return;
    }

    const fetchEvents = async () => {
      setIsLoading(true);
      setError(null);

      try {
        const response = await fetch(`/api/v2/semesters/${semesterContext.currentSemester!.id}/events`, {
          credentials: "include",
        });

        if (!response.ok) {
          throw new Error(`Failed to fetch events: ${response.statusText}`);
        }

        const data: ListEventsResponse[] = await response.json();
        setEvents(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "An error occurred while fetching events");
      } finally {
        setIsLoading(false);
      }
    };

    fetchEvents();
  }, [semesterContext?.currentSemester, refreshTrigger]);

  // Reset pagination when semester changes
  useEffect(() => {
    setCurrentPage(1);
    setSearchQuery("");
    setDebouncedSearchQuery("");
  }, [semesterContext?.currentSemester?.id]);

  // Filter events by search query
  const filteredEvents = useMemo(() => {
    if (!debouncedSearchQuery.trim()) {
      return events;
    }

    const query = debouncedSearchQuery.toLowerCase();
    return events.filter((event) => event.name.toLowerCase().includes(query));
  }, [events, debouncedSearchQuery]);

  // Paginate events
  const paginatedEvents = useMemo(() => {
    const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
    const endIndex = startIndex + ITEMS_PER_PAGE;
    return filteredEvents.slice(startIndex, endIndex);
  }, [filteredEvents, currentPage]);

  // Handle clear search
  const handleClearSearch = useCallback(() => {
    setSearchQuery("");
  }, []);

  // Handle refresh after actions
  const handleRefresh = useCallback(() => {
    setRefreshTrigger((prev) => prev + 1);
  }, []);

  // Check if user has any action permissions
  const hasAnyActionPermission = useMemo(
    () => hasPermission("edit", "event") || hasPermission("end", "event") || hasPermission("restart", "event"),
    [hasPermission],
  );

  // Define table columns
  const columns: TableColumn<ListEventsResponse>[] = [
    {
      key: "name",
      header: "Name",
      accessor: "name",
      sortable: false,
      render: (_value, row) =>
        hasPermission("get", "event") ? (
          <Link to={`${row.id}`} className={styles.eventLink}>
            {row.name}
          </Link>
        ) : (
          <span>{row.name}</span>
        ),
    },
    {
      key: "startDate",
      header: "Date",
      accessor: (row) => formatDate(row.startDate),
      sortable: false,
    },
    {
      key: "format",
      header: "Format",
      accessor: "format",
      sortable: false,
    },
    {
      key: "entries",
      header: "Entry Count",
      accessor: (row) => row.entries?.length ?? 0,
      sortable: false,
    },
    {
      key: "state",
      header: "Status",
      accessor: (row) => row.state,
      sortable: false,
      render: (_value, row) => (
        <span className={row.state === EventState.Started ? styles.statusActive : styles.statusEnded}>
          {row.state === EventState.Started ? "Active" : "Ended"}
        </span>
      ),
    },
    ...(hasAnyActionPermission
      ? [
          {
            key: "actions",
            header: "Actions",
            accessor: () => "",
            sortable: false,
            render: (_value: unknown, row: ListEventsResponse) => (
              <EventActions event={row} onActionComplete={handleRefresh} />
            ),
          },
        ]
      : []),
  ];

  // Show loading state
  if (isLoading) {
    return (
      <div className={styles.container}>
        <div className={styles.centerContent}>
          <Spinner size="lg" />
          <p>Loading events...</p>
        </div>
      </div>
    );
  }

  // Show error state
  if (error) {
    return (
      <div className={styles.container}>
        <div className={styles.errorState}>
          <p>Error: {error}</p>
          <Button onClick={() => window.location.reload()}>Retry</Button>
        </div>
      </div>
    );
  }

  // Show empty state when no semester is selected
  if (!semesterContext?.currentSemester) {
    return (
      <div className={styles.container}>
        <div className={styles.emptyState}>
          <p>Please select a semester to view events.</p>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      {/* Search and action bar */}
      <div className={styles.searchContainer}>
        <div className={styles.searchInputWrapper}>
          <Input
            type="search"
            placeholder="Search by event name..."
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
          />
        </div>
        {hasPermission("create", "event") && (
          <Button disabled title="Coming soon" iconBefore={<FaPlus />}>
            Create Event
          </Button>
        )}
      </div>

      <div className={styles.resultsInfo}>
        <p>
          Showing {paginatedEvents.length} of {filteredEvents.length} events
          {debouncedSearchQuery && ` matching "${debouncedSearchQuery}"`}
        </p>
      </div>

      {/* Table */}
      <div className={styles.tableWrapper}>
        <Table
          variant="striped"
          headerVariant="primary"
          data={paginatedEvents}
          columns={columns}
          emptyState={
            <div className={styles.emptyState}>
              <div className={styles.emptyIllustration}>
                <FaCalendarAlt size={64} />
              </div>
              {events.length === 0 ? (
                <>
                  <h3>No events yet</h3>
                  <p>No events have been created for this semester yet.</p>
                </>
              ) : (
                <>
                  <h3>No results found</h3>
                  <p>No events found matching &quot;{debouncedSearchQuery}&quot;</p>
                  <p className={styles.emptyHint}>Try adjusting your search terms</p>
                </>
              )}
            </div>
          }
        />
      </div>

      {/* Pagination */}
      {filteredEvents.length > ITEMS_PER_PAGE && (
        <div className={styles.paginationContainer}>
          <Pagination
            variant="compact"
            totalItems={filteredEvents.length}
            pageSize={ITEMS_PER_PAGE}
            currentPage={currentPage}
            onPageChange={setCurrentPage}
          />
        </div>
      )}
    </div>
  );
}
