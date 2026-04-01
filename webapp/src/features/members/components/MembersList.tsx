import { useContext, useEffect, useState, useMemo, useCallback, useRef } from "react";
import { useSearchParams } from "react-router-dom";
import { Table, TableColumn, Button, Pagination, Spinner } from "@uwpokerclub/components";
import { SemesterContext } from "@/contexts";
import { Membership } from "@/types";
import { useAuth } from "@/hooks";
import { FaEdit, FaTrash, FaPlus, FaUsers, FaFilter } from "react-icons/fa";
import { RegisterMemberModal } from "./RegisterMemberModal";
import { EditMemberModal } from "./EditMemberModal";
import { DeleteMembershipModal } from "./DeleteMembershipModal";
import { MemberFilters, type MemberFilterValues } from "./MemberFilters";
import styles from "./MembersList.module.css";

const ITEMS_PER_PAGE = 25;
const DEBOUNCE_MS = 300;

const EMPTY_FILTERS: MemberFilterValues = { studentId: "", name: "", email: "", faculty: "", paid: "", discounted: "" };
const FILTER_KEYS = Object.keys(EMPTY_FILTERS) as (keyof MemberFilterValues)[];

function filtersFromParams(params: URLSearchParams): MemberFilterValues {
  return {
    studentId: params.get("studentId") ?? "",
    name: params.get("name") ?? "",
    email: params.get("email") ?? "",
    faculty: params.get("faculty") ?? "",
    paid: params.get("paid") ?? "",
    discounted: params.get("discounted") ?? "",
  };
}

export function MembersList() {
  const semesterContext = useContext(SemesterContext);
  const { hasPermission } = useAuth();
  const [searchParams, setSearchParams] = useSearchParams();

  // Filter state: immediate (for inputs) and debounced (for API calls)
  const [filters, setFilters] = useState<MemberFilterValues>(() => filtersFromParams(searchParams));
  const [debouncedFilters, setDebouncedFilters] = useState<MemberFilterValues>(filters);
  const debounceTimer = useRef<ReturnType<typeof setTimeout>>(undefined);

  const [members, setMembers] = useState<Membership[]>([]);
  const [totalItems, setTotalItems] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(() => {
    const page = parseInt(searchParams.get("page") ?? "1", 10);
    return isNaN(page) || page < 1 ? 1 : page;
  });
  const [sortKey, setSortKey] = useState<string | undefined>(undefined);
  const [sortDirection, setSortDirection] = useState<"asc" | "desc">("asc");
  const [isRegisterModalOpen, setIsRegisterModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [selectedMembership, setSelectedMembership] = useState<Membership | null>(null);
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [isFilterOpen, setIsFilterOpen] = useState(false);

  // Debounce text filter changes, apply dropdown immediately
  const handleFilterChange = useCallback((key: keyof MemberFilterValues, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }));

    if (key === "faculty" || key === "paid" || key === "discounted") {
      // Dropdown: apply immediately
      setDebouncedFilters((prev) => ({ ...prev, [key]: value }));
      setCurrentPage(1);
    } else {
      // Text input: debounce
      clearTimeout(debounceTimer.current);
      debounceTimer.current = setTimeout(() => {
        setDebouncedFilters((prev) => ({ ...prev, [key]: value }));
        setCurrentPage(1);
      }, DEBOUNCE_MS);
    }
  }, []);

  const handleClearFilters = useCallback(() => {
    setFilters(EMPTY_FILTERS);
    setDebouncedFilters(EMPTY_FILTERS);
    setCurrentPage(1);
    clearTimeout(debounceTimer.current);
  }, []);

  // Sync debounced filters + page to URL (skip initial mount to avoid spurious re-render)
  const isInitialMount = useRef(true);
  useEffect(() => {
    if (isInitialMount.current) {
      isInitialMount.current = false;
      return;
    }
    const params = new URLSearchParams();
    for (const key of FILTER_KEYS) {
      if (debouncedFilters[key]) {
        params.set(key, debouncedFilters[key]);
      }
    }
    if (currentPage > 1) {
      params.set("page", String(currentPage));
    }
    setSearchParams(params, { replace: true });
  }, [debouncedFilters, currentPage, setSearchParams]);

  // Clean up debounce timer on unmount
  useEffect(() => {
    return () => clearTimeout(debounceTimer.current);
  }, []);

  // Reset filters when semester changes (skip initial mount to preserve URL-initialized filters)
  const prevSemesterId = useRef(semesterContext?.currentSemester?.id);
  useEffect(() => {
    if (prevSemesterId.current === semesterContext?.currentSemester?.id) return;
    prevSemesterId.current = semesterContext?.currentSemester?.id;
    setCurrentPage(1);
    setFilters(EMPTY_FILTERS);
    setDebouncedFilters(EMPTY_FILTERS);
    clearTimeout(debounceTimer.current);
  }, [semesterContext?.currentSemester?.id]);

  // Fetch members from API
  useEffect(() => {
    if (!semesterContext?.currentSemester) {
      setIsLoading(false);
      return;
    }

    const fetchMembers = async () => {
      setIsLoading(true);
      setError(null);

      try {
        const offset = (currentPage - 1) * ITEMS_PER_PAGE;
        const params = new URLSearchParams({
          limit: String(ITEMS_PER_PAGE),
          offset: String(offset),
        });

        if (debouncedFilters.name) params.set("name", debouncedFilters.name);
        if (debouncedFilters.email) params.set("email", debouncedFilters.email);
        if (debouncedFilters.faculty) params.set("faculty", debouncedFilters.faculty);
        if (debouncedFilters.studentId) params.set("studentId", debouncedFilters.studentId);
        if (debouncedFilters.paid) params.set("paid", debouncedFilters.paid);
        if (debouncedFilters.discounted) params.set("discounted", debouncedFilters.discounted);

        const url = `/api/v2/semesters/${semesterContext.currentSemester!.id}/memberships?${params.toString()}`;
        const response = await fetch(url, { credentials: "include" });

        if (!response.ok) {
          throw new Error(`Failed to fetch members: ${response.statusText}`);
        }

        const resp: { data: Membership[]; total: number } = await response.json();
        setMembers(resp.data);
        setTotalItems(resp.total);
      } catch (err) {
        setError(err instanceof Error ? err.message : "An error occurred while fetching members");
      } finally {
        setIsLoading(false);
      }
    };

    fetchMembers();
  }, [semesterContext?.currentSemester, refreshTrigger, currentPage, debouncedFilters]);

  // Sort members
  const sortedMembers = useMemo(() => {
    if (!sortKey) {
      return members;
    }

    const sorted = [...members].sort((a, b) => {
      let aValue: string | number = "";
      let bValue: string | number = "";

      switch (sortKey) {
        case "userId":
          aValue = a.userId;
          bValue = b.userId;
          break;
        case "name":
          aValue = `${a.user.firstName} ${a.user.lastName}`.toLowerCase();
          bValue = `${b.user.firstName} ${b.user.lastName}`.toLowerCase();
          break;
        case "email":
          aValue = a.user.email.toLowerCase();
          bValue = b.user.email.toLowerCase();
          break;
        case "status":
          aValue = a.paid ? (a.discounted ? "Discounted" : "Paid") : "Unpaid";
          bValue = b.paid ? (b.discounted ? "Discounted" : "Paid") : "Unpaid";
          break;
        default:
          return 0;
      }

      if (aValue < bValue) return sortDirection === "asc" ? -1 : 1;
      if (aValue > bValue) return sortDirection === "asc" ? 1 : -1;
      return 0;
    });

    return sorted;
  }, [members, sortKey, sortDirection]);

  const handleSort = useCallback((key: string, direction: "asc" | "desc") => {
    setSortKey(key);
    setSortDirection(direction);
  }, []);

  const handleRegisterMember = () => {
    setIsRegisterModalOpen(true);
  };

  const handleRegistrationSuccess = () => {
    setRefreshTrigger((prev) => prev + 1);
  };

  const handleViewEdit = (membership: Membership) => {
    setSelectedMembership(membership);
    setIsEditModalOpen(true);
  };

  const handleEditModalClose = () => {
    setIsEditModalOpen(false);
    setSelectedMembership(null);
  };

  const handleEditSuccess = () => {
    setRefreshTrigger((prev) => prev + 1);
  };

  const handleDelete = (membership: Membership) => {
    setSelectedMembership(membership);
    setIsDeleteModalOpen(true);
  };

  const handleDeleteModalClose = () => {
    setIsDeleteModalOpen(false);
    setSelectedMembership(null);
  };

  const handleDeleteSuccess = () => {
    setRefreshTrigger((prev) => prev + 1);
  };

  const activeFilterCount = Object.values(debouncedFilters).filter((v) => v !== "").length;
  const hasActiveFilters = activeFilterCount > 0;

  const columns: TableColumn<Membership>[] = [
    {
      key: "userId",
      header: "Student ID",
      accessor: "userId",
      sortable: true,
      headerProps: { "data-qa": "sort-userId-header" } as React.ThHTMLAttributes<HTMLTableCellElement>,
      cellProps: (row) => ({ "data-qa": `member-userId-${row.id}` }) as React.TdHTMLAttributes<HTMLTableCellElement>,
    },
    {
      key: "name",
      header: "Name",
      accessor: (row) => `${row.user.firstName} ${row.user.lastName}`,
      sortable: true,
      headerProps: { "data-qa": "sort-name-header" } as React.ThHTMLAttributes<HTMLTableCellElement>,
      cellProps: (row) => ({ "data-qa": `member-name-${row.id}` }) as React.TdHTMLAttributes<HTMLTableCellElement>,
    },
    {
      key: "email",
      header: "Email",
      accessor: (row) => row.user.email,
      sortable: true,
      headerProps: { "data-qa": "sort-email-header" } as React.ThHTMLAttributes<HTMLTableCellElement>,
      cellProps: (row) => ({ "data-qa": `member-email-${row.id}` }) as React.TdHTMLAttributes<HTMLTableCellElement>,
    },
    {
      key: "status",
      header: "Membership Status",
      accessor: (row) => {
        if (row.paid) {
          return row.discounted ? "Paid (Discounted)" : "Paid";
        }
        return "Unpaid";
      },
      sortable: true,
      headerProps: { "data-qa": "sort-status-header" } as React.ThHTMLAttributes<HTMLTableCellElement>,
      cellProps: (row) => ({ "data-qa": `member-status-${row.id}` }) as React.TdHTMLAttributes<HTMLTableCellElement>,
    },
    {
      key: "actions",
      header: "Actions",
      accessor: () => "",
      sortable: false,
      render: (_value, row) => (
        <div className={styles.actions}>
          {hasPermission("edit", "membership") && (
            <button
              className={styles.iconButton}
              onClick={() => handleViewEdit(row)}
              title="Edit member"
              aria-label="Edit member"
              data-qa={`edit-member-btn-${row.id}`}
            >
              <FaEdit />
            </button>
          )}
          {hasPermission("delete", "membership") && (
            <button
              className={`${styles.iconButton} ${styles.danger}`}
              onClick={() => handleDelete(row)}
              title="Delete membership"
              aria-label="Delete membership"
              data-qa={`delete-member-btn-${row.id}`}
            >
              <FaTrash />
            </button>
          )}
        </div>
      ),
    },
  ];

  const renderContent = () => {
    if (isLoading && members.length === 0) {
      return (
        <div className={styles.centerContent} data-qa="members-loading">
          <Spinner size="lg" />
          <p>Loading members...</p>
        </div>
      );
    }

    if (error) {
      return (
        <div className={styles.errorState} data-qa="members-error">
          <p>Error: {error}</p>
          <Button data-qa="retry-btn" onClick={() => window.location.reload()}>
            Retry
          </Button>
        </div>
      );
    }

    if (!semesterContext?.currentSemester) {
      return (
        <div className={styles.emptyState} data-qa="members-no-semester">
          <p>Please select a semester to view members.</p>
        </div>
      );
    }

    return (
      <>
        {/* Action bar */}
        <div className={styles.actionBar}>
          <div className={styles.resultsInfo} data-qa="members-results-info">
            <p>
              Showing {sortedMembers.length} of {totalItems} members
              {hasActiveFilters && " (filtered)"}
            </p>
          </div>
          <div className={styles.actionButtons}>
            {hasPermission("create", "membership") && (
              <Button data-qa="register-member-btn" onClick={handleRegisterMember} iconBefore={<FaPlus />}>
                Register New Member
              </Button>
            )}
            <button
              type="button"
              className={styles.filterToggle}
              onClick={() => setIsFilterOpen((prev) => !prev)}
              aria-label="Toggle filters"
              data-qa="filter-toggle-btn"
            >
              <FaFilter />
              {activeFilterCount > 0 && (
                <span className={styles.filterBadge} data-qa="filter-active-count">
                  {activeFilterCount}
                </span>
              )}
            </button>
          </div>
        </div>

        {/* Table */}
        <div className={styles.tableWrapper}>
          <Table
            data-qa="members-table"
            variant="striped"
            headerVariant="primary"
            data={sortedMembers}
            columns={columns}
            sortKey={sortKey}
            sortDirection={sortDirection}
            onSort={handleSort}
            rowProps={(row) => ({ "data-qa": `member-row-${row.id}` }) as React.HTMLAttributes<HTMLTableRowElement>}
            emptyState={
              <div className={styles.emptyState}>
                <div className={styles.emptyIllustration}>
                  <FaUsers size={64} />
                </div>
                {hasActiveFilters ? (
                  <>
                    <h3 data-qa="members-no-results">No results found</h3>
                    <p>No members found matching your filters</p>
                    <p className={styles.emptyHint}>Try adjusting your filter criteria</p>
                  </>
                ) : (
                  <>
                    <h3 data-qa="members-empty">No members yet</h3>
                    <p>No members have been registered for this semester yet.</p>
                  </>
                )}
              </div>
            }
          />
        </div>

        {/* Pagination */}
        {totalItems > ITEMS_PER_PAGE && (
          <div className={styles.paginationContainer} data-qa="members-pagination">
            <Pagination
              variant="compact"
              totalItems={totalItems}
              pageSize={ITEMS_PER_PAGE}
              currentPage={currentPage}
              onPageChange={setCurrentPage}
            />
          </div>
        )}

        {/* Register Member Modal */}
        <RegisterMemberModal
          isOpen={isRegisterModalOpen}
          onClose={() => setIsRegisterModalOpen(false)}
          onSuccess={handleRegistrationSuccess}
        />

        {/* Edit Member Modal */}
        <EditMemberModal
          isOpen={isEditModalOpen}
          membership={selectedMembership}
          onClose={handleEditModalClose}
          onSuccess={handleEditSuccess}
        />

        {/* Delete Membership Modal */}
        <DeleteMembershipModal
          isOpen={isDeleteModalOpen}
          membership={selectedMembership}
          semesterId={semesterContext.currentSemester.id}
          onClose={handleDeleteModalClose}
          onSuccess={handleDeleteSuccess}
        />
      </>
    );
  };

  return (
    <div className={styles.container}>
      {renderContent()}
      <MemberFilters
        isOpen={isFilterOpen}
        onClose={() => setIsFilterOpen(false)}
        filters={filters}
        onFilterChange={handleFilterChange}
        onClear={handleClearFilters}
      />
    </div>
  );
}
