import { useContext, useEffect, useState, useMemo, useCallback } from "react";
import { Table, TableColumn, Button, Input, Pagination, Spinner } from "@uwpokerclub/components";
import { SemesterContext } from "@/contexts";
import { Membership } from "@/types";
import { useAuth } from "@/hooks";
import { FaEdit, FaTrash, FaSearch, FaPlus, FaTimes, FaUsers } from "react-icons/fa";
import styles from "./MembersList.module.css";

const ITEMS_PER_PAGE = 25;

export function MembersList() {
  const semesterContext = useContext(SemesterContext);
  const { hasPermission } = useAuth();
  const [members, setMembers] = useState<Membership[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const [sortKey, setSortKey] = useState<string | undefined>(undefined);
  const [sortDirection, setSortDirection] = useState<"asc" | "desc">("asc");

  // Debounce search query
  const [debouncedSearchQuery, setDebouncedSearchQuery] = useState("");

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearchQuery(searchQuery);
      setCurrentPage(1); // Reset to first page on search
    }, 300);

    return () => clearTimeout(timer);
  }, [searchQuery]);

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
        const response = await fetch(`/api/v2/semesters/${semesterContext.currentSemester!.id}/memberships`, {
          credentials: "include",
        });

        if (!response.ok) {
          throw new Error(`Failed to fetch members: ${response.statusText}`);
        }

        const data: Membership[] = await response.json();
        setMembers(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "An error occurred while fetching members");
      } finally {
        setIsLoading(false);
      }
    };

    fetchMembers();
  }, [semesterContext?.currentSemester]);

  // Filter members by search query
  const filteredMembers = useMemo(() => {
    if (!debouncedSearchQuery.trim()) {
      return members;
    }

    const query = debouncedSearchQuery.toLowerCase();
    return members.filter(
      (member) =>
        member.firstName.toLowerCase().includes(query) ||
        member.lastName.toLowerCase().includes(query) ||
        member.email.toLowerCase().includes(query) ||
        `${member.firstName} ${member.lastName}`.toLowerCase().includes(query),
    );
  }, [members, debouncedSearchQuery]);

  // Sort members
  const sortedMembers = useMemo(() => {
    if (!sortKey) {
      return filteredMembers;
    }

    const sorted = [...filteredMembers].sort((a, b) => {
      let aValue: string | number = "";
      let bValue: string | number = "";

      switch (sortKey) {
        case "userId":
          aValue = a.userId;
          bValue = b.userId;
          break;
        case "name":
          aValue = `${a.firstName} ${a.lastName}`.toLowerCase();
          bValue = `${b.firstName} ${b.lastName}`.toLowerCase();
          break;
        case "email":
          aValue = a.email.toLowerCase();
          bValue = b.email.toLowerCase();
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
  }, [filteredMembers, sortKey, sortDirection]);

  // Paginate members
  const paginatedMembers = useMemo(() => {
    const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
    const endIndex = startIndex + ITEMS_PER_PAGE;
    return sortedMembers.slice(startIndex, endIndex);
  }, [sortedMembers, currentPage]);

  // Handle sort
  const handleSort = useCallback((key: string, direction: "asc" | "desc") => {
    setSortKey(key);
    setSortDirection(direction);
  }, []);

  // Handle clear search
  const handleClearSearch = () => {
    setSearchQuery("");
  };

  // Handle register new member (placeholder)
  const handleRegisterMember = () => {
    alert("Member registration functionality coming soon!");
  };

  // Handle member actions (placeholders)
  const handleViewEdit = (memberId: string) => {
    console.log("View/Edit member:", memberId);
    alert("View/Edit functionality coming soon!");
  };

  const handleDelete = (memberId: string) => {
    console.log("Delete member:", memberId);
    alert("Delete functionality coming soon!");
  };

  // Define table columns
  const columns: TableColumn<Membership>[] = [
    {
      key: "userId",
      header: "Student ID",
      accessor: "userId",
      sortable: true,
    },
    {
      key: "name",
      header: "Name",
      accessor: (row) => `${row.firstName} ${row.lastName}`,
      sortable: true,
    },
    {
      key: "email",
      header: "Email",
      accessor: "email",
      sortable: true,
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
    },
    {
      key: "actions",
      header: "Actions",
      accessor: () => "",
      sortable: false,
      render: (_value, row) => (
        <div className={styles.actions}>
          <button
            className={styles.iconButton}
            onClick={() => handleViewEdit(row.id)}
            disabled
            title="View/Edit member"
            aria-label="View/Edit member"
          >
            <FaEdit />
          </button>
          <button
            className={`${styles.iconButton} ${styles.danger}`}
            onClick={() => handleDelete(row.id)}
            disabled
            title="Delete member"
            aria-label="Delete member"
          >
            <FaTrash />
          </button>
        </div>
      ),
    },
  ];

  // Show loading state
  if (isLoading) {
    return (
      <div className={styles.container}>
        <div className={styles.header}>
          <h1>Members</h1>
        </div>
        <div className={styles.centerContent}>
          <Spinner size="lg" />
          <p>Loading members...</p>
        </div>
      </div>
    );
  }

  // Show error state
  if (error) {
    return (
      <div className={styles.container}>
        <div className={styles.header}>
          <h1>Members</h1>
        </div>
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
        <div className={styles.header}>
          <h1>Members</h1>
        </div>
        <div className={styles.emptyState}>
          <p>Please select a semester to view members.</p>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h1>Members</h1>
      </div>

      {/* Search and action bar */}
      <div className={styles.searchContainer}>
        <div className={styles.searchInputWrapper}>
          <Input
            type="search"
            placeholder="Search by name or email..."
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
        {hasPermission("create", "membership") && (
          <Button onClick={handleRegisterMember} iconBefore={<FaPlus />}>
            Register New Member
          </Button>
        )}
      </div>

      <div className={styles.resultsInfo}>
        <p>
          Showing {paginatedMembers.length} of {sortedMembers.length} members
          {debouncedSearchQuery && ` matching "${debouncedSearchQuery}"`}
        </p>
      </div>

      {/* Table */}
      <div className={styles.tableWrapper}>
        <Table
          variant="striped"
          headerVariant="primary"
          data={paginatedMembers}
          columns={columns}
          sortKey={sortKey}
          sortDirection={sortDirection}
          onSort={handleSort}
          emptyState={
            <div className={styles.emptyState}>
              <div className={styles.emptyIllustration}>
                <FaUsers size={64} />
              </div>
              {members.length === 0 ? (
                <>
                  <h3>No members yet</h3>
                  <p>No members have been registered for this semester yet.</p>
                  {hasPermission("create", "membership") && (
                    <Button onClick={handleRegisterMember} iconBefore={<FaPlus />}>
                      Register First Member
                    </Button>
                  )}
                </>
              ) : (
                <>
                  <h3>No results found</h3>
                  <p>No members found matching &quot;{debouncedSearchQuery}&quot;</p>
                  <p className={styles.emptyHint}>Try adjusting your search terms</p>
                </>
              )}
            </div>
          }
        />
      </div>

      {/* Pagination */}
      {sortedMembers.length > ITEMS_PER_PAGE && (
        <div className={styles.paginationContainer}>
          <Pagination
            variant="compact"
            totalItems={sortedMembers.length}
            pageSize={ITEMS_PER_PAGE}
            currentPage={currentPage}
            onPageChange={setCurrentPage}
          />
        </div>
      )}
    </div>
  );
}
