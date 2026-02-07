import { Link, useParams } from "react-router-dom";
import { useAuth, useCurrentSemester } from "../../../hooks";
import { Entry, StructureWithBlinds } from "../../../types";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { sendAPIRequest } from "../../../lib";
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
import { EventResponse, fetchEvent } from "../api/eventApi";

export function EventDetails() {
  const { eventId: eventIdParam = "" } = useParams<{ eventId: string }>();
  const eventId = parseInt(eventIdParam, 10);
  const { currentSemester } = useCurrentSemester();
  const { hasPermission } = useAuth();
  const { showToast } = useToast();

  // Data state
  const [event, setEvent] = useState<EventResponse | null>(null);
  const [entries, setEntries] = useState<Entry[]>([]);
  const [structure, setStructure] = useState<StructureWithBlinds | null>(null);

  // Loading/error state
  const [isLoading, setIsLoading] = useState(true);
  const [isEntriesLoading, setIsEntriesLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // UI state
  const [showEndModal, setShowEndModal] = useState(false);
  const [activeTab, setActiveTab] = useState<"info" | "clock">("info");
  const [activeSubTab, setActiveSubTab] = useState<"entries" | "structure">("entries");
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isRegistrationModalOpen, setIsRegistrationModalOpen] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);

  const endEventBtnRef = useRef<HTMLButtonElement | null>(null);
  const restartEventBtnRef = useRef<HTMLButtonElement | null>(null);

  // Fetch event data
  useEffect(() => {
    if (!currentSemester || isNaN(eventId)) {
      setIsLoading(false);
      return;
    }

    const loadEvent = async () => {
      setIsLoading(true);
      setError(null);

      const result = await fetchEvent(currentSemester.id, eventId);

      if (result.success) {
        setEvent(result.data);
      } else {
        setError(result.error);
      }

      setIsLoading(false);
    };

    loadEvent();
  }, [currentSemester, eventId]);

  // Fetch entries - only depends on IDs, not the full event object
  const fetchEntries = useCallback(async () => {
    if (!currentSemester || !eventId) return;

    setIsEntriesLoading(true);
    try {
      const response = await fetch(`/api/v2/semesters/${currentSemester.id}/events/${eventId}/entries`, {
        credentials: "include",
      });

      if (response.ok) {
        const resp = await response.json();
        // Transform API response to match Entry type
        // API returns: { membershipId, membership: { user: { firstName, lastName, id } }, ... }
        // Entry expects: { membershipId, firstName, lastName, id, ... }
        const transformedEntries: Entry[] = resp.data.map(
          (participant: {
            membershipId: string;
            membership?: { user?: { firstName?: string; lastName?: string; id?: string } };
            signedOutAt: Date;
            placement?: number;
            eventId: string;
          }) => ({
            id: participant.membership?.user?.id ?? "",
            membershipId: participant.membershipId,
            eventId: participant.eventId,
            firstName: participant.membership?.user?.firstName ?? "",
            lastName: participant.membership?.user?.lastName ?? "",
            signedOutAt: participant.signedOutAt,
            placement: participant.placement,
          }),
        );
        setEntries(transformedEntries);
      }
    } catch {
      // Silently handle entries fetch error
    } finally {
      setIsEntriesLoading(false);
    }
  }, [currentSemester, eventId]);

  // Fetch entries only when event is first loaded (event.id changes)
  useEffect(() => {
    if (event?.id) {
      fetchEntries();
    }
  }, [event?.id, fetchEntries]);

  // Fetch structure
  useEffect(() => {
    if (!event?.structureId) return;

    sendAPIRequest<StructureWithBlinds>(`structures/${event.structureId}`).then(({ data }) => {
      if (data) {
        setStructure(data);
      }
    });
  }, [event?.structureId]);

  // Event handlers
  const handleRebuy = useCallback(async () => {
    if (!currentSemester || !event) return;

    setIsProcessing(true);
    try {
      const response = await fetch(`/api/v2/semesters/${currentSemester.id}/events/${event.id}/rebuy`, {
        method: "POST",
        credentials: "include",
      });

      if (response.ok) {
        setEvent((prev) => (prev ? { ...prev, rebuys: prev.rebuys + 1 } : prev));
        showToast({
          message: "Rebuy recorded",
          variant: "success",
          duration: 2000,
        });
      } else {
        showToast({
          message: "Failed to record rebuy",
          variant: "error",
          duration: 3000,
        });
      }
    } catch {
      showToast({
        message: "Failed to record rebuy",
        variant: "error",
        duration: 3000,
      });
    } finally {
      setIsProcessing(false);
    }
  }, [currentSemester, event, showToast]);

  const handleEndEventClick = useCallback(() => {
    setShowEndModal(true);
    if (endEventBtnRef.current) {
      endEventBtnRef.current.disabled = true;
    }
  }, []);

  const handleEndModalSuccess = useCallback(() => {
    setEvent((prev) => (prev ? { ...prev, state: EventState.Ended } : prev));
    fetchEntries();
    if (endEventBtnRef.current) {
      endEventBtnRef.current.disabled = false;
    }
  }, [fetchEntries]);

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
      const response = await fetch(`/api/v2/semesters/${currentSemester.id}/events/${event.id}/restart`, {
        method: "POST",
        credentials: "include",
      });

      if (response.ok) {
        setEvent((prev) => (prev ? { ...prev, state: EventState.Started } : prev));
        fetchEntries();
        showToast({
          message: "Event restarted",
          variant: "success",
          duration: 3000,
        });
      } else {
        showToast({
          message: "Failed to restart event",
          variant: "error",
          duration: 3000,
        });
      }
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
  }, [currentSemester, event, fetchEntries, showToast]);

  // Edit modal handlers
  const handleEditClick = useCallback(() => {
    setIsEditModalOpen(true);
  }, []);

  const handleEditModalClose = useCallback(() => {
    setIsEditModalOpen(false);
  }, []);

  const handleEditSuccess = useCallback(() => {
    if (currentSemester && event) {
      fetchEvent(currentSemester.id, event.id).then((result) => {
        if (result.success) {
          setEvent(result.data);
        }
      });
    }
  }, [currentSemester, event]);

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
        <p className={styles.errorText}>{error}</p>
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
          <div className={styles.statCard}>
            <span className={styles.statValue}>{entries.length + event.rebuys}</span>
            <span className={styles.statLabel}>Total Entries</span>
          </div>
          <div className={styles.statCard}>
            <span className={styles.statValue}>{entries.length}</span>
            <span className={styles.statLabel}>Players</span>
          </div>
          <div className={styles.statCard}>
            <span className={styles.statValue}>{event.rebuys}</span>
            <span className={styles.statLabel}>Rebuys</span>
          </div>
          <div className={styles.statCard}>
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
                  updateParticipants={fetchEntries}
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

      <EditEventModal
        isOpen={isEditModalOpen}
        event={editEventData}
        onClose={handleEditModalClose}
        onSuccess={handleEditSuccess}
      />

      <EventRegistrationModal
        isOpen={isRegistrationModalOpen}
        onClose={() => setIsRegistrationModalOpen(false)}
        semesterId={currentSemester.id}
        eventId={event.id}
        onRegistrationChange={fetchEntries}
      />
    </>
  );
}
