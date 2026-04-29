import { Link, useParams } from "react-router-dom";
import { useAuth, useCurrentSemester } from "@/hooks";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { EntriesTable } from "./EntriesTable";
import { Spinner, Button, useToast } from "@uwpokerclub/components";
import {
  FaPencilAlt,
  FaTrash,
  FaCalendarAlt,
  FaGamepad,
  FaClock,
  FaUsers,
  FaUserPlus,
  FaRedo,
  FaStop,
  FaExclamationTriangle,
  FaChartLine,
  FaListOl,
} from "react-icons/fa";
import { EventState } from "@/sdk/events";

import styles from "./EventDetails.module.css";
import { TournamentClock } from "../../tournament-clock";
import { EndEventModal } from "./EndEventModal";
import { EditEventModal, type EventData } from "./EditEventModal";
import { EventRegistrationModal } from "./EventRegistrationModal";
import { DropdownMenu, type DropdownMenuItem } from "./DropdownMenu";
import { useEvent, useRebuyEvent, useRestartEvent } from "../hooks/useEventQueries";
import { useEntries } from "@/features/entries/hooks/useEntryQueries";
import { participantToEntry } from "@/features/entries/api/entriesApi";
import { useStructure } from "@/features/structures/hooks/useStructureQueries";

const ENTRIES_PER_PAGE = 25;

export function EventDetails() {
  const { eventId: eventIdParam = "" } = useParams<{ eventId: string }>();
  const parsedEventId = parseInt(eventIdParam, 10);
  const eventId = Number.isNaN(parsedEventId) ? undefined : parsedEventId;
  const { currentSemester } = useCurrentSemester();
  const { hasPermission } = useAuth();
  const { showToast } = useToast();

  const [entriesPage, setEntriesPage] = useState(1);
  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedSearchQuery, setDebouncedSearchQuery] = useState("");
  const [totalPlayerCount, setTotalPlayerCount] = useState(0);

  // UI state
  const [showEndModal, setShowEndModal] = useState(false);
  const [activeTab, setActiveTab] = useState<"info" | "clock">("info");
  const [activeSubTab, setActiveSubTab] = useState<"entries" | "structure">("entries");
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isRegistrationModalOpen, setIsRegistrationModalOpen] = useState(false);

  const endEventBtnRef = useRef<HTMLButtonElement | null>(null);
  const restartEventBtnRef = useRef<HTMLButtonElement | null>(null);

  // Query hooks
  const { data: event, isLoading, error } = useEvent(currentSemester?.id, eventId);

  // Use the URL-derived eventId so this query fires in parallel with useEvent
  const { data: entriesResponse, isLoading: isEntriesLoading } = useEntries(currentSemester?.id, eventId, {
    limit: ENTRIES_PER_PAGE,
    offset: (entriesPage - 1) * ENTRIES_PER_PAGE,
    search: debouncedSearchQuery || undefined,
  });
  const entries = useMemo(() => (entriesResponse?.data ?? []).map(participantToEntry), [entriesResponse]);
  const totalEntries = entriesResponse?.total ?? 0;

  // Track the unfiltered total separately so the player count stat doesn't change with search
  useEffect(() => {
    if (!debouncedSearchQuery && entriesResponse) {
      setTotalPlayerCount(entriesResponse.total);
    }
  }, [debouncedSearchQuery, entriesResponse]);

  const { data: structure = null } = useStructure(event?.structureId);

  const rebuyMutation = useRebuyEvent();
  const restartMutation = useRestartEvent();
  const isProcessing = rebuyMutation.isPending || restartMutation.isPending;

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearchQuery(searchQuery);
      setEntriesPage(1);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  // Event handlers
  const handleRebuy = useCallback(async () => {
    if (!currentSemester || !event) return;

    try {
      await rebuyMutation.mutateAsync({ semesterId: currentSemester.id, eventId: event.id });
      showToast({
        message: "Rebuy recorded",
        variant: "success",
        duration: 2000,
      });
    } catch {
      showToast({
        message: "Failed to record rebuy",
        variant: "error",
        duration: 3000,
      });
    }
  }, [currentSemester, event, showToast, rebuyMutation]);

  const handleEndEventClick = useCallback(() => {
    setShowEndModal(true);
    if (endEventBtnRef.current) {
      endEventBtnRef.current.disabled = true;
    }
  }, []);

  const handleEndModalSuccess = useCallback(() => {
    if (endEventBtnRef.current) {
      endEventBtnRef.current.disabled = false;
    }
  }, []);

  const handleEndModalClose = useCallback(() => {
    setShowEndModal(false);
    if (endEventBtnRef.current) {
      endEventBtnRef.current.disabled = false;
    }
  }, []);

  const handleRestartEvent = useCallback(async () => {
    if (!currentSemester || !event) return;

    if (restartEventBtnRef.current) {
      restartEventBtnRef.current.disabled = true;
    }

    try {
      await restartMutation.mutateAsync({ semesterId: currentSemester.id, eventId: event.id });
      showToast({
        message: "Event restarted",
        variant: "success",
        duration: 3000,
      });
    } catch {
      showToast({
        message: "Failed to restart event",
        variant: "error",
        duration: 3000,
      });
    } finally {
      if (restartEventBtnRef.current) {
        restartEventBtnRef.current.disabled = false;
      }
    }
  }, [currentSemester, event, showToast, restartMutation]);

  // Edit modal handlers
  const handleEditClick = useCallback(() => {
    setIsEditModalOpen(true);
  }, []);

  const handleEditModalClose = useCallback(() => {
    setIsEditModalOpen(false);
  }, []);

  // Build overflow menu items
  const menuItems: DropdownMenuItem[] = useMemo(() => {
    const items: DropdownMenuItem[] = [];

    if (hasPermission("edit", "event")) {
      items.push({
        key: "edit",
        label: "Edit Event",
        icon: <FaPencilAlt />,
        onClick: handleEditClick,
        disabled: event?.state === EventState.Ended,
      });
    }

    if (hasPermission("delete", "event")) {
      items.push({
        key: "delete",
        label: "Delete Event",
        icon: <FaTrash />,
        onClick: () => {
          /* Placeholder - UWPSC-26 */
        },
        disabled: true,
      });
    }

    return items;
  }, [hasPermission, event?.state, handleEditClick]);

  // Prepare event data for EditEventModal
  const editEventData: EventData | null = useMemo(() => {
    if (!event) return null;
    return {
      id: event.id,
      name: event.name,
      format: event.format,
      notes: event.notes,
      startDate: event.startDate,
      state: event.state,
    };
  }, [event]);

  // Format date helper
  const formatEventDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      weekday: "long",
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "numeric",
      minute: "numeric",
    });
  };

  // Loading state
  if (isLoading) {
    return (
      <div className={styles.loadingContainer}>
        <Spinner size="lg" />
        <p className={styles.loadingText}>Loading event details...</p>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className={styles.errorContainer}>
        <div className={styles.errorIcon}>
          <FaExclamationTriangle />
        </div>
        <p className={styles.errorText}>{error.message}</p>
        <Button onClick={() => window.location.reload()}>Retry</Button>
      </div>
    );
  }

  // No semester selected
  if (!currentSemester) {
    return (
      <div className={styles.errorContainer}>
        <div className={styles.errorIcon}>
          <FaExclamationTriangle />
        </div>
        <p className={styles.errorHint}>Please select a semester to view event details.</p>
      </div>
    );
  }

  // Event not found
  if (!event) {
    return (
      <div className={styles.errorContainer}>
        <div className={styles.errorIcon}>
          <FaExclamationTriangle />
        </div>
        <p className={styles.errorText}>Event not found</p>
        <Link to="/admin/events">
          <Button>Back to Events</Button>
        </Link>
      </div>
    );
  }

  const isEventEnded = event.state === EventState.Ended;

  return (
    <>
      <div className={styles.container}>
        {/* Page Header */}
        <header className={styles.pageHeader}>
          <div className={styles.headerContent}>
            <div className={styles.eventInfo}>
              <h1 className={styles.eventTitle}>{event.name}</h1>

              <div className={styles.eventMeta}>
                <div className={styles.metaItem}>
                  <FaGamepad />
                  <div>
                    <span className={styles.metaLabel}>Format</span>
                    <span className={styles.metaValue}>{event.format}</span>
                  </div>
                </div>
                <div className={styles.metaItem}>
                  <FaCalendarAlt />
                  <div>
                    <span className={styles.metaLabel}>Date</span>
                    <span className={styles.metaValue}>{formatEventDate(event.startDate)}</span>
                  </div>
                </div>
              </div>

              {event.notes && <div className={styles.eventNotes}>{event.notes}</div>}
            </div>

            <div className={styles.headerActions}>
              <span className={`${styles.statusBadge} ${isEventEnded ? styles.statusEnded : styles.statusActive}`}>
                {isEventEnded ? "Ended" : "Active"}
              </span>
              {menuItems.length > 0 && <DropdownMenu items={menuItems} isLoading={isProcessing} />}
            </div>
          </div>
        </header>

        {/* Stats Bar */}
        <div className={styles.statsBar}>
          <div className={styles.statCard} data-qa="stat-total-entries">
            <span className={styles.statValue}>{totalPlayerCount + event.rebuys}</span>
            <span className={styles.statLabel}>Total Entries</span>
          </div>
          <div className={styles.statCard} data-qa="stat-players">
            <span className={styles.statValue}>{totalPlayerCount}</span>
            <span className={styles.statLabel}>Players</span>
          </div>
          <div className={styles.statCard} data-qa="stat-rebuys">
            <span className={styles.statValue}>{event.rebuys}</span>
            <span className={styles.statLabel}>Rebuys</span>
          </div>
          <div className={styles.statCard} data-qa="stat-points-multiplier">
            <span className={styles.statValue}>{event.pointsMultiplier}x</span>
            <span className={styles.statLabel}>Points Multiplier</span>
          </div>
        </div>

        {/* Actions Bar */}
        <div className={styles.actionsBar}>
          {!isEventEnded && (
            <>
              {hasPermission("signin", "event", "participant") && (
                <button
                  data-qa="register-members-btn"
                  type="button"
                  onClick={() => setIsRegistrationModalOpen(true)}
                  className={`${styles.actionButton} ${styles.actionPrimary}`}
                >
                  <FaUserPlus />
                  Register Members
                </button>
              )}

              {hasPermission("rebuy", "event") && (
                <Button
                  data-qa="rebuy-btn"
                  onClick={handleRebuy}
                  variant="secondary"
                  disabled={isProcessing}
                  iconBefore={<FaChartLine />}
                >
                  Add Rebuy
                </Button>
              )}

              {hasPermission("end", "event") && (
                <Button
                  ref={endEventBtnRef}
                  data-qa="end-event-btn"
                  onClick={handleEndEventClick}
                  variant="destructive"
                  iconBefore={<FaStop />}
                >
                  End Event
                </Button>
              )}
            </>
          )}

          {isEventEnded && hasPermission("restart", "event") && (
            <Button
              ref={restartEventBtnRef}
              data-qa="restart-event-btn"
              onClick={handleRestartEvent}
              variant="tertiary"
              iconBefore={<FaRedo />}
            >
              Restart Event
            </Button>
          )}
        </div>

        {/* Tab Navigation */}
        <nav className={styles.tabNavigation}>
          <button
            data-qa="tournament-tab"
            className={`${styles.tabButton} ${activeTab === "info" ? styles.tabButtonActive : ""}`}
            onClick={() => setActiveTab("info")}
          >
            <FaListOl className={styles.tabIcon} />
            Tournament Info
          </button>
          <button
            data-qa="clock-tab"
            className={`${styles.tabButton} ${activeTab === "clock" ? styles.tabButtonActive : ""}`}
            onClick={() => setActiveTab("clock")}
          >
            <FaClock className={styles.tabIcon} />
            Clock
          </button>
        </nav>

        {/* Content Area */}
        <div className={styles.contentArea}>
          {activeTab === "info" ? (
            <>
              {/* Sub-tabs */}
              <div className={styles.subTabs}>
                <button
                  className={`${styles.subTabButton} ${activeSubTab === "entries" ? styles.subTabButtonActive : ""}`}
                  onClick={() => setActiveSubTab("entries")}
                >
                  <FaUsers className={styles.tabIcon} />
                  Entries
                </button>
                {hasPermission("get", "structure") && (
                  <button
                    className={`${styles.subTabButton} ${activeSubTab === "structure" ? styles.subTabButtonActive : ""}`}
                    onClick={() => setActiveSubTab("structure")}
                  >
                    <FaListOl className={styles.tabIcon} />
                    Structure
                  </button>
                )}
              </div>

              {/* Tab Content */}
              {activeSubTab === "entries" && hasPermission("list", "event", "participant") ? (
                <EntriesTable
                  entries={entries}
                  event={event}
                  semesterId={currentSemester.id}
                  isLoading={isEntriesLoading}
                  totalItems={totalEntries}
                  currentPage={entriesPage}
                  pageSize={ENTRIES_PER_PAGE}
                  onPageChange={setEntriesPage}
                  searchQuery={searchQuery}
                  onSearchChange={setSearchQuery}
                />
              ) : activeSubTab === "structure" ? (
                <div className={styles.structureSection}>
                  {structure ? (
                    <>
                      <div className={styles.structureHeader}>
                        <span className={styles.structureName}>{structure.name || "Unknown Structure"}</span>
                      </div>
                      <table className={styles.structureTable}>
                        <thead>
                          <tr>
                            <th>Level</th>
                            <th>Small Blind</th>
                            <th>Big Blind</th>
                            <th>Ante</th>
                            <th>Time (min)</th>
                          </tr>
                        </thead>
                        <tbody>
                          {structure.blinds.map((blind, i) => (
                            <tr key={i}>
                              <td className={styles.levelNumber}>{i + 1}</td>
                              <td className={styles.blindValue}>{blind.small}</td>
                              <td className={styles.blindValue}>{blind.big}</td>
                              <td className={styles.blindValue}>{blind.ante}</td>
                              <td className={styles.timeValue}>{blind.time}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </>
                  ) : (
                    <div className={styles.loadingContainer}>
                      <Spinner size="md" />
                      <p className={styles.loadingText}>Loading structure...</p>
                    </div>
                  )}
                </div>
              ) : null}
            </>
          ) : (
            <TournamentClock levels={structure?.blinds || []} />
          )}
        </div>
      </div>

      <EndEventModal
        show={showEndModal}
        semesterId={currentSemester.id}
        eventId={event.id}
        onClose={handleEndModalClose}
        onSuccess={handleEndModalSuccess}
      />

      <EditEventModal isOpen={isEditModalOpen} event={editEventData} onClose={handleEditModalClose} />

      <EventRegistrationModal
        isOpen={isRegistrationModalOpen}
        onClose={() => setIsRegistrationModalOpen(false)}
        semesterId={currentSemester.id}
        eventId={event.id}
      />
    </>
  );
}
