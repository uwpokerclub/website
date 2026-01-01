import { useState, useEffect, useMemo, useCallback, useRef } from "react";
import { Modal, Spinner, useToast } from "@uwpokerclub/components";
import { FaSearch, FaChevronRight, FaChevronLeft, FaUsers, FaUserCheck, FaUserPlus, FaPlus } from "react-icons/fa";
import { RegisterMemberModal, type RegistrationSuccessData } from "../../../members/components/RegisterMemberModal";
import styles from "./EventRegistrationModal.module.css";

/**
 * Membership data from the API
 */
interface Membership {
  id: string;
  userId: number;
  user: {
    id: string;
    firstName: string;
    lastName: string;
    email: string;
  };
  semesterId: string;
  paid: boolean;
  discounted: boolean;
  attendance: number;
}

/**
 * Entry/Participant data from the API
 */
interface Entry {
  membershipId: string;
  eventId: number;
  membership?: {
    id: string;
    user: {
      firstName: string;
      lastName: string;
    };
  };
}

/**
 * API response for batch entry creation
 */
interface CreateEntryResult {
  membershipId: string;
  status: "created" | "error";
  error?: string;
}

export interface EventRegistrationModalProps {
  isOpen: boolean;
  onClose: () => void;
  semesterId: string;
  eventId: number;
  onRegistrationChange?: () => void;
}

/**
 * EventRegistrationModal - Dual-list transfer interface for event registration
 *
 * Displays available and registered members side-by-side with immediate
 * registration/unregistration via action buttons.
 */
export function EventRegistrationModal({
  isOpen,
  onClose,
  semesterId,
  eventId,
  onRegistrationChange,
}: EventRegistrationModalProps) {
  const { showToast } = useToast();
  const searchInputRef = useRef<HTMLInputElement>(null);

  // Data state
  const [allMembers, setAllMembers] = useState<Membership[]>([]);
  const [registeredIds, setRegisteredIds] = useState<Set<string>>(new Set());

  // UI state
  const [searchQuery, setSearchQuery] = useState("");
  const [isInitialLoading, setIsInitialLoading] = useState(true);
  const [loadingMemberIds, setLoadingMemberIds] = useState<Set<string>>(new Set());
  const [loadError, setLoadError] = useState<string | null>(null);
  const [isCreateMemberModalOpen, setIsCreateMemberModalOpen] = useState(false);

  // Debounced search query
  const [debouncedQuery, setDebouncedQuery] = useState("");

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedQuery(searchQuery);
    }, 200);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  // Focus search input when modal opens
  useEffect(() => {
    if (isOpen && searchInputRef.current) {
      const timer = setTimeout(() => {
        searchInputRef.current?.focus();
      }, 100);
      return () => clearTimeout(timer);
    }
  }, [isOpen]);

  // Load data when modal opens
  useEffect(() => {
    if (!isOpen || !semesterId || !eventId) {
      return;
    }

    let mounted = true;

    const loadData = async () => {
      setIsInitialLoading(true);
      setLoadError(null);

      try {
        // Fetch memberships and entries in parallel
        const [membersResponse, entriesResponse] = await Promise.all([
          fetch(`/api/v2/semesters/${semesterId}/memberships`, {
            credentials: "include",
          }),
          fetch(`/api/v2/semesters/${semesterId}/events/${eventId}/entries`, {
            credentials: "include",
          }),
        ]);

        if (!membersResponse.ok) {
          throw new Error("Failed to load members");
        }
        if (!entriesResponse.ok) {
          throw new Error("Failed to load entries");
        }

        const members: Membership[] = await membersResponse.json();
        const entries: Entry[] = await entriesResponse.json();

        if (mounted) {
          // Build set of registered membership IDs
          const registeredSet = new Set(entries.map((e) => e.membershipId));

          setAllMembers(members);
          setRegisteredIds(registeredSet);
        }
      } catch (error) {
        if (mounted) {
          const message = error instanceof Error ? error.message : "Failed to load data";
          setLoadError(message);
          showToast({
            message,
            variant: "error",
            duration: 5000,
          });
        }
      } finally {
        if (mounted) {
          setIsInitialLoading(false);
        }
      }
    };

    loadData();

    return () => {
      mounted = false;
    };
  }, [isOpen, semesterId, eventId, showToast]);

  // Reset state when modal closes
  const handleClose = useCallback(() => {
    setSearchQuery("");
    setDebouncedQuery("");
    setLoadError(null);
    onClose();
  }, [onClose]);

  // Filter function for search
  const matchesSearch = useCallback((member: Membership, query: string): boolean => {
    if (!query.trim()) return true;
    const searchLower = query.toLowerCase();
    const fullName = `${member.user.firstName} ${member.user.lastName}`.toLowerCase();
    const studentId = member.userId.toString();
    return fullName.includes(searchLower) || studentId.includes(searchLower);
  }, []);

  // Derived filtered lists
  const filteredAvailable = useMemo(() => {
    return allMembers
      .filter((m) => !registeredIds.has(m.id))
      .filter((m) => matchesSearch(m, debouncedQuery))
      .sort((a, b) => {
        const nameA = `${a.user.lastName} ${a.user.firstName}`.toLowerCase();
        const nameB = `${b.user.lastName} ${b.user.firstName}`.toLowerCase();
        return nameA.localeCompare(nameB);
      });
  }, [allMembers, registeredIds, debouncedQuery, matchesSearch]);

  const filteredRegistered = useMemo(() => {
    return allMembers
      .filter((m) => registeredIds.has(m.id))
      .filter((m) => matchesSearch(m, debouncedQuery))
      .sort((a, b) => {
        const nameA = `${a.user.lastName} ${a.user.firstName}`.toLowerCase();
        const nameB = `${b.user.lastName} ${b.user.firstName}`.toLowerCase();
        return nameA.localeCompare(nameB);
      });
  }, [allMembers, registeredIds, debouncedQuery, matchesSearch]);

  // Check if member should be highlighted as danger (unpaid with 3+ attendance)
  const isDangerMember = useCallback((member: Membership): boolean => {
    return member.attendance >= 3 && !member.paid;
  }, []);

  // Register a member
  const handleRegister = useCallback(
    async (membershipId: string) => {
      setLoadingMemberIds((prev) => new Set(prev).add(membershipId));

      try {
        const response = await fetch(`/api/v2/semesters/${semesterId}/events/${eventId}/entries`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify([membershipId]),
        });

        if (response.ok || response.status === 207) {
          const results: CreateEntryResult[] = await response.json();
          const result = results.find((r) => r.membershipId === membershipId);

          if (result?.status === "created") {
            setRegisteredIds((prev) => new Set(prev).add(membershipId));
            onRegistrationChange?.();
          } else {
            showToast({
              message: result?.error || "Failed to register member",
              variant: "error",
              duration: 4000,
            });
          }
        } else {
          const errorData = await response.json().catch(() => null);
          showToast({
            message: errorData?.message || "Failed to register member",
            variant: "error",
            duration: 4000,
          });
        }
      } catch {
        showToast({
          message: "Network error - please try again",
          variant: "error",
          duration: 4000,
        });
      } finally {
        setLoadingMemberIds((prev) => {
          const next = new Set(prev);
          next.delete(membershipId);
          return next;
        });
      }
    },
    [semesterId, eventId, showToast, onRegistrationChange],
  );

  // Unregister a member
  const handleUnregister = useCallback(
    async (membershipId: string) => {
      setLoadingMemberIds((prev) => new Set(prev).add(membershipId));

      try {
        const response = await fetch(`/api/v2/semesters/${semesterId}/events/${eventId}/entries/${membershipId}`, {
          method: "DELETE",
          credentials: "include",
        });

        if (response.ok || response.status === 204) {
          setRegisteredIds((prev) => {
            const next = new Set(prev);
            next.delete(membershipId);
            return next;
          });
          onRegistrationChange?.();
        } else {
          const errorData = await response.json().catch(() => null);
          showToast({
            message: errorData?.message || "Failed to unregister member",
            variant: "error",
            duration: 4000,
          });
        }
      } catch {
        showToast({
          message: "Network error - please try again",
          variant: "error",
          duration: 4000,
        });
      } finally {
        setLoadingMemberIds((prev) => {
          const next = new Set(prev);
          next.delete(membershipId);
          return next;
        });
      }
    },
    [semesterId, eventId, showToast, onRegistrationChange],
  );

  // Handle new member creation - auto-register them for the event
  const handleCreateMemberSuccess = useCallback(
    async (data?: RegistrationSuccessData) => {
      setIsCreateMemberModalOpen(false);

      if (!data) {
        // No data returned, just refresh the members list
        // Re-fetch to get the updated list
        const membersResponse = await fetch(`/api/v2/semesters/${semesterId}/memberships`, {
          credentials: "include",
        });
        if (membersResponse.ok) {
          const members: Membership[] = await membersResponse.json();
          setAllMembers(members);
        }
        return;
      }

      // Add the new member to the local state
      const newMember: Membership = {
        id: data.membershipId,
        userId: data.userId,
        user: {
          id: data.userId.toString(),
          firstName: data.firstName,
          lastName: data.lastName,
          email: "",
        },
        semesterId,
        paid: false,
        discounted: false,
        attendance: 0,
      };
      setAllMembers((prev) => [...prev, newMember]);

      // Auto-register the new member for the event
      showToast({
        message: `${data.firstName} ${data.lastName} added. Registering for event...`,
        variant: "info",
        duration: 2000,
      });

      // Register them for the event
      await handleRegister(data.membershipId);
    },
    [semesterId, showToast, handleRegister],
  );

  // Render a member row
  const renderMemberRow = (member: Membership, isRegistered: boolean) => {
    const isLoading = loadingMemberIds.has(member.id);
    const isDanger = isDangerMember(member);

    return (
      <div
        key={member.id}
        className={`${styles.memberRow} ${isDanger ? styles.memberRowDanger : ""}`}
        data-qa={`member-row-${member.id}`}
      >
        <div className={styles.memberInfo}>
          <span className={styles.memberName} data-qa={`member-name-${member.id}`}>
            {member.user.firstName} {member.user.lastName}
          </span>
          <span className={styles.memberStudentId} data-qa={`member-studentId-${member.id}`}>
            {member.userId}
          </span>
        </div>
        <button
          type="button"
          className={`${styles.actionButton} ${
            isRegistered ? styles.actionButtonUnregister : styles.actionButtonRegister
          }`}
          onClick={() => (isRegistered ? handleUnregister(member.id) : handleRegister(member.id))}
          disabled={isLoading}
          aria-label={
            isRegistered
              ? `Unregister ${member.user.firstName} ${member.user.lastName}`
              : `Register ${member.user.firstName} ${member.user.lastName}`
          }
          data-qa={isRegistered ? `unregister-member-btn-${member.id}` : `register-member-btn-${member.id}`}
        >
          {isLoading ? (
            <Spinner size="sm" className={styles.actionButtonSpinner} />
          ) : isRegistered ? (
            <FaChevronLeft />
          ) : (
            <FaChevronRight />
          )}
        </button>
      </div>
    );
  };

  // Render empty state
  const renderEmptyState = (type: "available" | "registered" | "no-results", showCreateButton = false) => {
    const configs = {
      available: {
        icon: <FaUserCheck />,
        text: "All members are registered for this event",
        qaAttr: "event-registration-all-registered",
      },
      registered: {
        icon: <FaUsers />,
        text: "No members registered yet",
        qaAttr: "event-registration-no-members",
      },
      "no-results": {
        icon: <FaSearch />,
        text: `No members match "${debouncedQuery}"`,
        qaAttr: "event-registration-no-results",
      },
    };

    const config = configs[type];

    return (
      <div className={styles.emptyState} data-qa={config.qaAttr}>
        <div className={styles.emptyIcon}>{config.icon}</div>
        <p className={styles.emptyText}>{config.text}</p>
        {showCreateButton && (
          <button type="button" className={styles.createMemberButton} onClick={() => setIsCreateMemberModalOpen(true)}>
            <FaUserPlus />
            Create New Member
          </button>
        )}
      </div>
    );
  };

  // Determine empty state type for each panel
  const getEmptyStateType = (isAvailable: boolean): "available" | "registered" | "no-results" | null => {
    const list = isAvailable ? filteredAvailable : filteredRegistered;
    const fullList = isAvailable
      ? allMembers.filter((m) => !registeredIds.has(m.id))
      : allMembers.filter((m) => registeredIds.has(m.id));

    if (list.length > 0) return null;
    if (debouncedQuery && fullList.length > 0) return "no-results";
    return isAvailable ? "available" : "registered";
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Manage Event Registration"
      size="xl"
      closeOnEscape={!isCreateMemberModalOpen}
      data-qa="event-registration-modal"
    >
      <div className={styles.modalContent}>
        {loadError && (
          <div className={styles.errorAlert} data-qa="event-registration-error-alert">
            {loadError}
          </div>
        )}

        {isInitialLoading ? (
          <div className={styles.loadingContainer} data-qa="event-registration-loading">
            <Spinner size="lg" />
            <p className={styles.loadingText}>Loading members...</p>
          </div>
        ) : (
          <>
            {/* Search Input */}
            <div className={styles.searchSection}>
              <FaSearch className={styles.searchIcon} />
              <input
                ref={searchInputRef}
                type="text"
                className={styles.searchInput}
                placeholder="Search by name or student ID..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                aria-label="Search members"
                data-qa="input-event-registration-search"
              />
            </div>

            {/* Dual Panel Layout */}
            <div className={styles.panelsContainer}>
              {/* Available Members Panel */}
              <div className={styles.panel} data-qa="panel-available">
                <div className={styles.panelHeader}>
                  <span className={styles.panelTitle}>Available</span>
                  <div className={styles.panelHeaderActions}>
                    <button
                      type="button"
                      className={styles.addMemberButton}
                      onClick={() => setIsCreateMemberModalOpen(true)}
                      aria-label="Create new member"
                      title="Create new member"
                      data-qa="add-member-btn"
                    >
                      <FaPlus />
                    </button>
                    <span className={styles.panelCount} data-qa="panel-available-count">
                      {filteredAvailable.length}
                    </span>
                  </div>
                </div>
                <div className={styles.panelList}>
                  {(() => {
                    const emptyType = getEmptyStateType(true);
                    if (emptyType) return renderEmptyState(emptyType, true);
                    return filteredAvailable.map((member) => renderMemberRow(member, false));
                  })()}
                </div>
              </div>

              {/* Registered Members Panel */}
              <div className={styles.panel} data-qa="panel-registered">
                <div className={styles.panelHeader}>
                  <span className={styles.panelTitle}>Registered</span>
                  <span className={styles.panelCount} data-qa="panel-registered-count">
                    {filteredRegistered.length}
                  </span>
                </div>
                <div className={styles.panelList}>
                  {(() => {
                    const emptyType = getEmptyStateType(false);
                    if (emptyType) return renderEmptyState(emptyType);
                    return filteredRegistered.map((member) => renderMemberRow(member, true));
                  })()}
                </div>
              </div>
            </div>
          </>
        )}
      </div>

      {/* Nested modal for creating new members */}
      <RegisterMemberModal
        isOpen={isCreateMemberModalOpen}
        onClose={() => setIsCreateMemberModalOpen(false)}
        onSuccess={handleCreateMemberSuccess}
      />
    </Modal>
  );
}
