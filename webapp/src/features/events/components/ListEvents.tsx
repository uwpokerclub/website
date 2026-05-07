import { useContext, useEffect, useState, useMemo, useCallback, useRef } from "react";
import { Link } from "react-router-dom";
import { Table, TableColumn, Button, Input, Pagination, Spinner, useToast, Modal } from "@uwpokerclub/components";
import { SemesterContext } from "@/contexts";
import { useAuth } from "@/hooks";
import { FaSearch, FaTimes, FaPlus, FaCalendarAlt, FaPencilAlt, FaEllipsisV, FaStop, FaRedo } from "react-icons/fa";
import { EventState } from "@/sdk/events";
import { CreateEventModal } from "./CreateEventModal";
import { EditEventModal, type EventData } from "./EditEventModal";
import { Event } from "@/types";
import { useEvents, useEndEvent, useRestartEvent } from "../hooks/useEventQueries";
import styles from "./ListEvents.module.css";

const ITEMS_PER_PAGE = 25;

type EventActionsProps = {
  event: Event;
  onEditClick: (event: Event) => void;
};

function EventActions({ event, onEditClick }: EventActionsProps) {
  const { hasPermission } = useAuth();
  const semesterContext = useContext(SemesterContext);
  const { showToast } = useToast();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isEndConfirmOpen, setIsEndConfirmOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const endEventMutation = useEndEvent();
  const restartEventMutation = useRestartEvent();
  const isProcessing = endEventMutation.isPending || restartEventMutation.isPending;

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

  const handleEndEventClick = () => {
    setIsMenuOpen(false);
    setIsEndConfirmOpen(true);
  };

  const handleEndEventConfirm = async () => {
    if (!semesterContext?.currentSemester) return;

    try {
      await endEventMutation.mutateAsync({ semesterId: semesterContext.currentSemester.id, eventId: event.id });
      showToast({
        message: `"${event.name}" has been ended successfully`,
        variant: "success",
        duration: 3000,
      });
    } catch (err) {
      showToast({
        message: err instanceof Error ? err.message : "Failed to end event",
        variant: "error",
        duration: 5000,
      });
    } finally {
      setIsEndConfirmOpen(false);
    }
  };

  const handleRestartEvent = async () => {
    if (!semesterContext?.currentSemester) return;
    try {
      await restartEventMutation.mutateAsync({ semesterId: semesterContext.currentSemester.id, eventId: event.id });
      showToast({
        message: `"${event.name}" has been restarted successfully`,
        variant: "success",
        duration: 3000,
      });
    } catch (err) {
      showToast({
        message: err instanceof Error ? err.message : "Failed to restart event",
        variant: "error",
        duration: 5000,
      });
    } finally {
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
    <>
      <div className={styles.actions} ref={menuRef}>
        <div className={styles.menuWrapper}>
          <button
            type="button"
            className={styles.iconButton}
            onClick={() => setIsMenuOpen(!isMenuOpen)}
            title="Actions"
            aria-label="Actions"
            disabled={isProcessing}
            data-qa={`actions-menu-btn-${event.id}`}
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
                  <button
                    type="button"
                    className={styles.menuItem}
                    onClick={() => {
                      setIsMenuOpen(false);
                      onEditClick(event);
                    }}
                    data-qa={`edit-event-btn-${event.id}`}
                  >
                    <FaPencilAlt /> Edit Event
                  </button>
                ))}
              {showEndEvent && (
                <button
                  className={styles.menuItem}
                  onClick={handleEndEventClick}
                  disabled={isProcessing}
                  data-qa={`end-event-btn-${event.id}`}
                >
                  <FaStop /> End Event
                </button>
              )}
              {showRestartEvent && (
                <button
                  className={styles.menuItem}
                  onClick={handleRestartEvent}
                  disabled={isProcessing}
                  data-qa={`restart-event-btn-${event.id}`}
                >
                  <FaRedo /> Restart Event
                </button>
              )}
            </div>
          )}
        </div>
      </div>

      <Modal
        isOpen={isEndConfirmOpen}
        onClose={() => setIsEndConfirmOpen(false)}
        title="End Event"
        size="sm"
        footer={
          <div className={styles.confirmFooter}>
            <Button
              variant="tertiary"
              onClick={() => setIsEndConfirmOpen(false)}
              disabled={isProcessing}
              data-qa={`end-confirm-cancel-btn-${event.id}`}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleEndEventConfirm}
              disabled={isProcessing}
              data-qa={`end-confirm-btn-${event.id}`}
            >
              {isProcessing ? "Ending..." : "End Event"}
            </Button>
          </div>
        }
        data-qa={`end-confirm-modal-${event.id}`}
      >
        <p>
          Are you sure you want to end <strong>&quot;{event.name}&quot;</strong>? This will finalize the event results.
        </p>
      </Modal>
    </>
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
  const [searchQuery, setSearchQuery] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [editingEvent, setEditingEvent] = useState<EventData | null>(null);

  // Debounce search query
  const [debouncedSearchQuery, setDebouncedSearchQuery] = useState("");

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearchQuery(searchQuery);
      setCurrentPage(1);
    }, 300);

    return () => clearTimeout(timer);
  }, [searchQuery]);

  const offset = (currentPage - 1) * ITEMS_PER_PAGE;
  const {
    data: eventsResponse,
    isLoading,
    error,
  } = useEvents(semesterContext?.currentSemester?.id, {
    limit: ITEMS_PER_PAGE,
    offset,
    search: debouncedSearchQuery || undefined,
  });
  const events = eventsResponse?.data ?? [];
  const totalItems = eventsResponse?.total ?? 0;

  // Reset pagination when semester changes
  useEffect(() => {
    setCurrentPage(1);
    setSearchQuery("");
    setDebouncedSearchQuery("");
  }, [semesterContext?.currentSemester?.id]);

  // Handle clear search
  const handleClearSearch = useCallback(() => {
    setSearchQuery("");
  }, []);

  // Handle edit click
  const handleEditClick = useCallback((event: Event) => {
    setEditingEvent({
      id: event.id,
      name: event.name,
      format: event.format,
      notes: event.notes,
      startDate: event.startDate,
      state: event.state,
    });
    setIsEditModalOpen(true);
  }, []);

  // Handle edit modal close
  const handleEditModalClose = useCallback(() => {
    setIsEditModalOpen(false);
    setEditingEvent(null);
  }, []);

  // Check if user has any action permissions
  const hasAnyActionPermission = useMemo(
    () => hasPermission("edit", "event") || hasPermission("end", "event") || hasPermission("restart", "event"),
    [hasPermission],
  );

  // Define table columns
  const columns: TableColumn<Event>[] = [
    {
      key: "name",
      header: "Name",
      accessor: "name",
      sortable: false,
      render: (_value, row) =>
        hasPermission("get", "event") ? (
          <Link to={`${row.id}`} className={styles.eventLink} data-qa={`event-name-${row.id}`}>
            {row.name}
          </Link>
        ) : (
          <span data-qa={`event-name-${row.id}`}>{row.name}</span>
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
        <span
          className={row.state === EventState.Started ? styles.statusActive : styles.statusEnded}
          data-qa={`event-status-${row.id}`}
        >
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
            render: (_value: unknown, row: Event) => <EventActions event={row} onEditClick={handleEditClick} />,
          },
        ]
      : []),
  ];

  // Show loading state
  if (isLoading) {
    return (
      <div className={styles.container} data-qa="events-loading">
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
      <div className={styles.container} data-qa="events-error">
        <div className={styles.errorState}>
          <p>Error: {error.message}</p>
          <Button onClick={() => window.location.reload()} data-qa="events-retry-btn">
            Retry
          </Button>
        </div>
      </div>
    );
  }

  // Show empty state when no semester is selected
  if (!semesterContext?.currentSemester) {
    return (
      <div className={styles.container} data-qa="events-no-semester">
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
                  data-qa="clear-search-btn"
                >
                  <FaTimes />
                </button>
              ) : null
            }
            fullWidth
            data-qa="input-events-search"
          />
        </div>
        {hasPermission("create", "event") && (
          <Button onClick={() => setIsCreateModalOpen(true)} iconBefore={<FaPlus />} data-qa="create-event-btn">
            Create Event
          </Button>
        )}
      </div>

      <div className={styles.resultsInfo} data-qa="events-results-info">
        <p>
          Showing {events.length} of {totalItems} events
          {debouncedSearchQuery && ` matching "${debouncedSearchQuery}"`}
        </p>
      </div>

      {/* Table */}
      <div className={styles.tableWrapper}>
        <Table
          variant="striped"
          headerVariant="primary"
          data={events}
          columns={columns}
          emptyState={
            <div className={styles.emptyState} data-qa={debouncedSearchQuery ? "events-no-results" : "events-empty"}>
              <div className={styles.emptyIllustration}>
                <FaCalendarAlt size={64} />
              </div>
              {debouncedSearchQuery ? (
                <>
                  <h3>No results found</h3>
                  <p>No events found matching &quot;{debouncedSearchQuery}&quot;</p>
                  <p className={styles.emptyHint}>Try adjusting your search terms</p>
                </>
              ) : (
                <>
                  <h3>No events yet</h3>
                  <p>No events have been created for this semester yet.</p>
                </>
              )}
            </div>
          }
          data-qa="events-table"
        />
      </div>

      {/* Pagination */}
      {totalItems > ITEMS_PER_PAGE && (
        <div className={styles.paginationContainer} data-qa="events-pagination">
          <Pagination
            variant="compact"
            totalItems={totalItems}
            pageSize={ITEMS_PER_PAGE}
            currentPage={currentPage}
            onPageChange={setCurrentPage}
          />
        </div>
      )}

      {/* Create Event Modal */}
      <CreateEventModal isOpen={isCreateModalOpen} onClose={() => setIsCreateModalOpen(false)} />

      {/* Edit Event Modal */}
      <EditEventModal isOpen={isEditModalOpen} event={editingEvent} onClose={handleEditModalClose} />
    </div>
  );
}
