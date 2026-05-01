import { useState, useEffect, useMemo, useCallback, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Modal, Spinner, useToast } from "@uwpokerclub/components";
import { FaSearch, FaChevronRight, FaChevronLeft, FaUsers, FaUserCheck, FaUserPlus, FaPlus } from "react-icons/fa";
import { RegisterMemberModal, type RegistrationSuccessData } from "@/features/members/components/RegisterMemberModal";
import { fetchMemberships } from "@/features/members/api/memberRegistrationApi";
import { fetchEntries, registerEntries, unregisterEntry, ParticipantResponse } from "@/features/entries/api/entriesApi";
import { entryKeys } from "@/features/entries/hooks/useEntryQueries";
import { Membership } from "@/types";
import styles from "./EventRegistrationModal.module.css";

const PAGE_SIZE = 100;

export interface EventRegistrationModalProps {
  isOpen: boolean;
  onClose: () => void;
  semesterId: string;
  eventId: number;
}

export function EventRegistrationModal({ isOpen, onClose, semesterId, eventId }: EventRegistrationModalProps) {
  const { showToast } = useToast();
  const queryClient = useQueryClient();
  const searchInputRef = useRef<HTMLInputElement>(null);

  // Track whether any registration changes happened so we can invalidate caches on close
  const hasChangesRef = useRef(false);

  // Memberships (available panel) state
  const [memberships, setMemberships] = useState<Membership[]>([]);
  const [membershipsTotal, setMembershipsTotal] = useState(0);
  const [membershipsOffset, setMembershipsOffset] = useState(0);
  const [membershipsLoading, setMembershipsLoading] = useState(false);

  // Entries (registered panel) state
  const [entries, setEntries] = useState<ParticipantResponse[]>([]);
  const [entriesTotal, setEntriesTotal] = useState(0);
  const [entriesOffset, setEntriesOffset] = useState(0);
  const [entriesLoading, setEntriesLoading] = useState(false);

  // Set of registered membership IDs (built from entries)
  const [registeredIds, setRegisteredIds] = useState<Set<string>>(new Set());

  // UI state
  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");
  const [isInitialLoading, setIsInitialLoading] = useState(true);
  const [loadingMemberIds, setLoadingMemberIds] = useState<Set<string>>(new Set());
  const [loadError, setLoadError] = useState<string | null>(null);
  const [isCreateMemberModalOpen, setIsCreateMemberModalOpen] = useState(false);

  // Sentinel refs for IntersectionObserver
  const availableSentinelRef = useRef<HTMLDivElement>(null);
  const registeredSentinelRef = useRef<HTMLDivElement>(null);

  // Refs to track latest values in async callbacks and stable observers
  const membershipsLoadingRef = useRef(false);
  const entriesLoadingRef = useRef(false);
  const membershipsOffsetRef = useRef(0);
  const entriesOffsetRef = useRef(0);
  const hasMoreMembershipsRef = useRef(false);
  const hasMoreEntriesRef = useRef(false);
  const debouncedQueryRef = useRef("");

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

  // Fetch a page of memberships
  const fetchMembershipsPage = useCallback(
    async (offset: number, search: string, reset: boolean) => {
      if (membershipsLoadingRef.current) return;
      membershipsLoadingRef.current = true;
      setMembershipsLoading(true);

      try {
        const resp = await fetchMemberships(semesterId, { limit: PAGE_SIZE, offset, search: search || undefined });

        setMemberships((prev) => (reset ? resp.data : [...prev, ...resp.data]));
        setMembershipsTotal(resp.total);
        const newOffset = offset + resp.data.length;
        setMembershipsOffset(newOffset);
        membershipsOffsetRef.current = newOffset;
        hasMoreMembershipsRef.current = newOffset < resp.total;
      } catch (error) {
        const message = error instanceof Error ? error.message : "Failed to load members";
        showToast({ message, variant: "error", duration: 5000 });
      } finally {
        membershipsLoadingRef.current = false;
        setMembershipsLoading(false);
      }
    },
    [semesterId, showToast],
  );

  // Fetch a page of entries
  const fetchEntriesPage = useCallback(
    async (offset: number, search: string, reset: boolean) => {
      if (entriesLoadingRef.current) return;
      entriesLoadingRef.current = true;
      setEntriesLoading(true);

      try {
        const resp = await fetchEntries(semesterId, eventId, { limit: PAGE_SIZE, offset, search: search || undefined });

        setEntries((prev) => (reset ? resp.data : [...prev, ...resp.data]));
        setEntriesTotal(resp.total);
        const newOffset = offset + resp.data.length;
        setEntriesOffset(newOffset);
        entriesOffsetRef.current = newOffset;
        hasMoreEntriesRef.current = newOffset < resp.total;

        // Update registeredIds
        if (reset) {
          setRegisteredIds(new Set(resp.data.map((e) => e.membershipId)));
        } else {
          setRegisteredIds((prev) => {
            const next = new Set(prev);
            for (const e of resp.data) {
              next.add(e.membershipId);
            }
            return next;
          });
        }
      } catch (error) {
        const message = error instanceof Error ? error.message : "Failed to load entries";
        showToast({ message, variant: "error", duration: 5000 });
      } finally {
        entriesLoadingRef.current = false;
        setEntriesLoading(false);
      }
    },
    [semesterId, eventId, showToast],
  );

  // Initial load when modal opens
  useEffect(() => {
    if (!isOpen || !semesterId || !eventId) return;

    let mounted = true;

    const loadInitial = async () => {
      setIsInitialLoading(true);
      setLoadError(null);

      try {
        const [membersResp, entriesResp] = await Promise.all([
          fetchMemberships(semesterId, { limit: PAGE_SIZE, offset: 0 }),
          fetchEntries(semesterId, eventId, { limit: PAGE_SIZE, offset: 0 }),
        ]);

        if (mounted) {
          setMemberships(membersResp.data);
          setMembershipsTotal(membersResp.total);
          const mOffset = membersResp.data.length;
          setMembershipsOffset(mOffset);
          membershipsOffsetRef.current = mOffset;
          hasMoreMembershipsRef.current = mOffset < membersResp.total;

          setEntries(entriesResp.data);
          setEntriesTotal(entriesResp.total);
          const eOffset = entriesResp.data.length;
          setEntriesOffset(eOffset);
          entriesOffsetRef.current = eOffset;
          hasMoreEntriesRef.current = eOffset < entriesResp.total;

          setRegisteredIds(new Set(entriesResp.data.map((e) => e.membershipId)));
        }
      } catch (error) {
        if (mounted) {
          const message = error instanceof Error ? error.message : "Failed to load data";
          setLoadError(message);
          showToast({ message, variant: "error", duration: 5000 });
        }
      } finally {
        if (mounted) {
          setIsInitialLoading(false);
        }
      }
    };

    loadInitial();

    return () => {
      mounted = false;
    };
  }, [isOpen, semesterId, eventId, showToast]);

  // Reset and re-fetch when search changes
  useEffect(() => {
    if (isInitialLoading || !isOpen) return;

    setMemberships([]);
    setMembershipsOffset(0);
    membershipsOffsetRef.current = 0;
    setMembershipsTotal(0);
    hasMoreMembershipsRef.current = false;
    setEntries([]);
    setEntriesOffset(0);
    entriesOffsetRef.current = 0;
    setEntriesTotal(0);
    hasMoreEntriesRef.current = false;
    setRegisteredIds(new Set());
    debouncedQueryRef.current = debouncedQuery;

    fetchMembershipsPage(0, debouncedQuery, true);
    fetchEntriesPage(0, debouncedQuery, true);
  }, [debouncedQuery, isInitialLoading, isOpen, fetchMembershipsPage, fetchEntriesPage]);

  // Compute derived values
  const hasMoreMemberships = membershipsOffset < membershipsTotal;
  const hasMoreEntries = entriesOffset < entriesTotal;

  // Keep refs in sync for stable observer closures
  useEffect(() => {
    hasMoreMembershipsRef.current = hasMoreMemberships;
  }, [hasMoreMemberships]);
  useEffect(() => {
    hasMoreEntriesRef.current = hasMoreEntries;
  }, [hasMoreEntries]);
  useEffect(() => {
    debouncedQueryRef.current = debouncedQuery;
  }, [debouncedQuery]);

  // Available members = loaded memberships minus registered ones
  const availableMembers = useMemo(
    () => memberships.filter((m) => !registeredIds.has(m.id)),
    [memberships, registeredIds],
  );

  // Build a lookup from entries to membership data for display
  const membershipsMap = useMemo(() => {
    const map = new Map<string, Membership>();
    for (const m of memberships) {
      map.set(m.id, m);
    }
    return map;
  }, [memberships]);

  const registeredMembers = useMemo(
    () =>
      entries.map((entry) => ({
        entry,
        membership: membershipsMap.get(entry.membershipId),
      })),
    [entries, membershipsMap],
  );

  // Stable IntersectionObserver for available panel (uses refs, not re-created on state changes)
  useEffect(() => {
    const sentinel = availableSentinelRef.current;
    if (!sentinel || isInitialLoading) return;

    const observer = new IntersectionObserver(
      (observerEntries) => {
        if (observerEntries[0].isIntersecting && hasMoreMembershipsRef.current && !membershipsLoadingRef.current) {
          fetchMembershipsPage(membershipsOffsetRef.current, debouncedQueryRef.current, false);
        }
      },
      { threshold: 0.1 },
    );

    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [isInitialLoading, fetchMembershipsPage]);

  // Stable IntersectionObserver for registered panel
  useEffect(() => {
    const sentinel = registeredSentinelRef.current;
    if (!sentinel || isInitialLoading) return;

    const observer = new IntersectionObserver(
      (observerEntries) => {
        if (observerEntries[0].isIntersecting && hasMoreEntriesRef.current && !entriesLoadingRef.current) {
          fetchEntriesPage(entriesOffsetRef.current, debouncedQueryRef.current, false);
        }
      },
      { threshold: 0.1 },
    );

    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [isInitialLoading, fetchEntriesPage]);

  // Reset state when modal closes
  const handleClose = useCallback(() => {
    setSearchQuery("");
    setDebouncedQuery("");
    setLoadError(null);
    setIsInitialLoading(true);
    setMemberships([]);
    setMembershipsTotal(0);
    setMembershipsOffset(0);
    membershipsOffsetRef.current = 0;
    hasMoreMembershipsRef.current = false;
    setEntries([]);
    setEntriesTotal(0);
    setEntriesOffset(0);
    entriesOffsetRef.current = 0;
    hasMoreEntriesRef.current = false;
    setRegisteredIds(new Set());
    debouncedQueryRef.current = "";

    // The modal manages its own entries state imperatively; reconcile the shared cache
    // (used by EntriesTable in EventDetails) once on close if any registrations changed.
    if (hasChangesRef.current) {
      hasChangesRef.current = false;
      queryClient.invalidateQueries({ queryKey: entryKeys.byEvent(semesterId, eventId) });
    }

    onClose();
  }, [onClose, queryClient, semesterId, eventId]);

  // Check if member should be highlighted as danger (unpaid with 3+ attendance)
  const isDangerMember = useCallback((member: Membership): boolean => {
    return member.attendance >= 3 && !member.paid;
  }, []);

  // Register a member
  const handleRegister = useCallback(
    async (membershipId: string) => {
      setLoadingMemberIds((prev) => new Set(prev).add(membershipId));

      try {
        const results = await registerEntries(semesterId, eventId, [membershipId]);
        const result = results.find((r) => r.membershipId === membershipId);

        if (result?.status === "created") {
          hasChangesRef.current = true;
          setRegisteredIds((prev) => new Set(prev).add(membershipId));

          // Add to entries list with membership data for display
          setEntries((prev) => {
            const membership = membershipsMap.get(membershipId);
            const newEntry: ParticipantResponse = {
              membershipId,
              eventId: String(eventId),
              membership: membership
                ? {
                    id: membership.id,
                    user: {
                      id: membership.user.id,
                      firstName: membership.user.firstName,
                      lastName: membership.user.lastName,
                    },
                  }
                : undefined,
              signedOutAt: null as unknown as Date,
            };
            return [...prev, newEntry];
          });
          setEntriesTotal((prev) => prev + 1);
          setEntriesOffset((prev) => {
            const newOffset = prev + 1;
            entriesOffsetRef.current = newOffset;
            return newOffset;
          });
        } else {
          showToast({
            message: result?.error || "Failed to register member",
            variant: "error",
            duration: 4000,
          });
        }
      } catch (err) {
        showToast({
          message: err instanceof Error ? err.message : "Failed to register member",
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
    [semesterId, eventId, membershipsMap, showToast],
  );

  // Unregister a member
  const handleUnregister = useCallback(
    async (membershipId: string) => {
      setLoadingMemberIds((prev) => new Set(prev).add(membershipId));

      try {
        await unregisterEntry(semesterId, eventId, membershipId);
        hasChangesRef.current = true;

        setRegisteredIds((prev) => {
          const next = new Set(prev);
          next.delete(membershipId);
          return next;
        });

        setEntries((prev) => prev.filter((e) => e.membershipId !== membershipId));
        setEntriesTotal((prev) => prev - 1);
        setEntriesOffset((prev) => {
          const newOffset = prev - 1;
          entriesOffsetRef.current = newOffset;
          return newOffset;
        });
      } catch (err) {
        showToast({
          message: err instanceof Error ? err.message : "Failed to unregister member",
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
    [semesterId, eventId, showToast],
  );

  // Handle new member creation - auto-register them for the event
  const handleCreateMemberSuccess = useCallback(
    async (data?: RegistrationSuccessData) => {
      setIsCreateMemberModalOpen(false);

      if (!data) {
        fetchMembershipsPage(0, debouncedQueryRef.current, true);
        return;
      }

      const newMember: Membership = {
        id: data.membershipId,
        userId: data.userId,
        user: {
          id: data.userId.toString(),
          firstName: data.firstName,
          lastName: data.lastName,
          email: "",
          faculty: "",
          questId: "",
          createdAt: "",
        },
        semesterId,
        paid: false,
        discounted: false,
        attendance: 0,
      };
      setMemberships((prev) => [...prev, newMember]);
      setMembershipsTotal((prev) => prev + 1);
      setMembershipsOffset((prev) => {
        const newOffset = prev + 1;
        membershipsOffsetRef.current = newOffset;
        return newOffset;
      });

      showToast({
        message: `${data.firstName} ${data.lastName} added. Registering for event...`,
        variant: "info",
        duration: 2000,
      });

      await handleRegister(data.membershipId);
    },
    [semesterId, showToast, handleRegister, fetchMembershipsPage],
  );

  // Render a member row for the available panel
  const renderAvailableMemberRow = (member: Membership) => {
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
          className={`${styles.actionButton} ${styles.actionButtonRegister}`}
          onClick={() => handleRegister(member.id)}
          disabled={isLoading}
          aria-label={`Register ${member.user.firstName} ${member.user.lastName}`}
          data-qa={`register-member-btn-${member.id}`}
        >
          {isLoading ? <Spinner size="sm" className={styles.actionButtonSpinner} /> : <FaChevronRight />}
        </button>
      </div>
    );
  };

  // Render a member row for the registered panel
  const renderRegisteredMemberRow = (entry: ParticipantResponse, membership?: Membership) => {
    const membershipId = entry.membershipId;
    const isLoading = loadingMemberIds.has(membershipId);
    const firstName = membership?.user.firstName ?? entry.membership?.user?.firstName ?? "";
    const lastName = membership?.user.lastName ?? entry.membership?.user?.lastName ?? "";
    // Danger highlighting is best-effort: only applies when full membership data is locally loaded
    const isDanger = membership ? isDangerMember(membership) : false;

    return (
      <div
        key={membershipId}
        className={`${styles.memberRow} ${isDanger ? styles.memberRowDanger : ""}`}
        data-qa={`member-row-${membershipId}`}
      >
        <div className={styles.memberInfo}>
          <span className={styles.memberName} data-qa={`member-name-${membershipId}`}>
            {firstName} {lastName}
          </span>
          {membership && (
            <span className={styles.memberStudentId} data-qa={`member-studentId-${membershipId}`}>
              {membership.userId}
            </span>
          )}
        </div>
        <button
          type="button"
          className={`${styles.actionButton} ${styles.actionButtonUnregister}`}
          onClick={() => handleUnregister(membershipId)}
          disabled={isLoading}
          aria-label={`Unregister ${firstName} ${lastName}`}
          data-qa={`unregister-member-btn-${membershipId}`}
        >
          {isLoading ? <Spinner size="sm" className={styles.actionButtonSpinner} /> : <FaChevronLeft />}
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

  // Loading sentinel component
  const renderSentinel = (ref: React.RefObject<HTMLDivElement | null>, isLoading: boolean) => (
    <div ref={ref} className={styles.sentinel}>
      {isLoading && (
        <div className={styles.sentinelLoading}>
          <Spinner size="sm" />
        </div>
      )}
    </div>
  );

  const availableCount = debouncedQuery ? availableMembers.length : membershipsTotal - entriesTotal;

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
                      {availableCount < 0 ? 0 : availableCount}
                    </span>
                  </div>
                </div>
                <div className={styles.panelList}>
                  {availableMembers.length === 0 && !membershipsLoading ? (
                    debouncedQuery ? (
                      renderEmptyState("no-results", true)
                    ) : (
                      renderEmptyState("available", true)
                    )
                  ) : (
                    <>
                      {availableMembers.map((member) => renderAvailableMemberRow(member))}
                      {renderSentinel(availableSentinelRef, membershipsLoading)}
                    </>
                  )}
                </div>
              </div>

              {/* Registered Members Panel */}
              <div className={styles.panel} data-qa="panel-registered">
                <div className={styles.panelHeader}>
                  <span className={styles.panelTitle}>Registered</span>
                  <span className={styles.panelCount} data-qa="panel-registered-count">
                    {entriesTotal}
                  </span>
                </div>
                <div className={styles.panelList}>
                  {registeredMembers.length === 0 && !entriesLoading ? (
                    debouncedQuery ? (
                      renderEmptyState("no-results")
                    ) : (
                      renderEmptyState("registered")
                    )
                  ) : (
                    <>
                      {registeredMembers.map(({ entry, membership }) => renderRegisteredMemberRow(entry, membership))}
                      {renderSentinel(registeredSentinelRef, entriesLoading)}
                    </>
                  )}
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
