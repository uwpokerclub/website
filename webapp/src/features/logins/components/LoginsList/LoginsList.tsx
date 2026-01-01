import { useEffect, useState, useMemo, useCallback } from "react";
import { Table, TableColumn, Button, Input, Pagination, Spinner } from "@uwpokerclub/components";
import { useAuth } from "@/hooks";
import { FaEdit, FaTrash, FaSearch, FaPlus, FaTimes, FaKey } from "react-icons/fa";
import { LoginResponse } from "../../types";
import { fetchLogins } from "../../api/loginsApi";
import { CreateLoginModal } from "../CreateLoginModal";
import { EditPasswordModal } from "../EditPasswordModal";
import { DeleteLoginModal } from "../DeleteLoginModal";
import styles from "./LoginsList.module.css";

const ITEMS_PER_PAGE = 25;

export function LoginsList() {
  const { hasPermission } = useAuth();
  const [logins, setLogins] = useState<LoginResponse[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const [sortKey, setSortKey] = useState<string | undefined>(undefined);
  const [sortDirection, setSortDirection] = useState<"asc" | "desc">("asc");

  // Modal states
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [selectedLogin, setSelectedLogin] = useState<LoginResponse | null>(null);
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  // Debounced search query
  const [debouncedSearchQuery, setDebouncedSearchQuery] = useState("");

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearchQuery(searchQuery);
      setCurrentPage(1);
    }, 300);

    return () => clearTimeout(timer);
  }, [searchQuery]);

  // Fetch logins from API
  useEffect(() => {
    const loadLogins = async () => {
      setIsLoading(true);
      setError(null);

      const result = await fetchLogins();

      if (result.success) {
        setLogins(result.data);
      } else {
        setError(result.error);
      }

      setIsLoading(false);
    };

    loadLogins();
  }, [refreshTrigger]);

  // Filter logins by search query
  const filteredLogins = useMemo(() => {
    if (!debouncedSearchQuery.trim()) {
      return logins;
    }

    const query = debouncedSearchQuery.toLowerCase();
    return logins.filter(
      (login) =>
        login.username.toLowerCase().includes(query) ||
        login.role.toLowerCase().includes(query) ||
        (login.linkedMember &&
          `${login.linkedMember.firstName} ${login.linkedMember.lastName}`.toLowerCase().includes(query)),
    );
  }, [logins, debouncedSearchQuery]);

  // Sort logins
  const sortedLogins = useMemo(() => {
    if (!sortKey) {
      return filteredLogins;
    }

    const sorted = [...filteredLogins].sort((a, b) => {
      let aValue: string = "";
      let bValue: string = "";

      switch (sortKey) {
        case "username":
          aValue = a.username.toLowerCase();
          bValue = b.username.toLowerCase();
          break;
        case "role":
          aValue = a.role.toLowerCase();
          bValue = b.role.toLowerCase();
          break;
        case "linkedMember":
          aValue = a.linkedMember ? `${a.linkedMember.firstName} ${a.linkedMember.lastName}`.toLowerCase() : "";
          bValue = b.linkedMember ? `${b.linkedMember.firstName} ${b.linkedMember.lastName}`.toLowerCase() : "";
          break;
        default:
          return 0;
      }

      if (aValue < bValue) return sortDirection === "asc" ? -1 : 1;
      if (aValue > bValue) return sortDirection === "asc" ? 1 : -1;
      return 0;
    });

    return sorted;
  }, [filteredLogins, sortKey, sortDirection]);

  // Paginate logins
  const paginatedLogins = useMemo(() => {
    const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
    const endIndex = startIndex + ITEMS_PER_PAGE;
    return sortedLogins.slice(startIndex, endIndex);
  }, [sortedLogins, currentPage]);

  // Handle sort
  const handleSort = useCallback((key: string, direction: "asc" | "desc") => {
    setSortKey(key);
    setSortDirection(direction);
  }, []);

  // Handle clear search
  const handleClearSearch = () => {
    setSearchQuery("");
  };

  // Handle create login
  const handleCreateLogin = () => {
    setIsCreateModalOpen(true);
  };

  // Handle edit password
  const handleEditPassword = (login: LoginResponse) => {
    setSelectedLogin(login);
    setIsEditModalOpen(true);
  };

  // Handle delete
  const handleDelete = (login: LoginResponse) => {
    setSelectedLogin(login);
    setIsDeleteModalOpen(true);
  };

  // Handle success callbacks
  const handleSuccess = () => {
    setRefreshTrigger((prev) => prev + 1);
  };

  // Render role badge
  const renderRoleBadge = (role: string) => {
    const roleClass = `role-${role.replace("_", "-")}`;
    const displayRole = role
      .split("_")
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
      .join(" ");
    return <span className={`${styles.roleBadge} ${styles[roleClass]}`}>{displayRole}</span>;
  };

  // Render linked member
  const renderLinkedMember = (linkedMember: LoginResponse["linkedMember"]) => {
    if (!linkedMember) {
      return <span className={styles.noLinkedMember}>No linked member</span>;
    }
    return (
      <span className={styles.linkedMember}>
        {linkedMember.firstName} {linkedMember.lastName}
        <span className={styles.memberId}>({linkedMember.id})</span>
      </span>
    );
  };

  // Define table columns
  const columns: TableColumn<LoginResponse>[] = [
    {
      key: "username",
      header: "Username",
      accessor: "username",
      sortable: true,
      headerProps: { "data-qa": "sort-username-header" } as React.ThHTMLAttributes<HTMLTableCellElement>,
      cellProps: (row) =>
        ({ "data-qa": `login-username-${row.username}` }) as React.TdHTMLAttributes<HTMLTableCellElement>,
    },
    {
      key: "role",
      header: "Role",
      accessor: "role",
      sortable: true,
      render: (_value, row) => renderRoleBadge(row.role),
      headerProps: { "data-qa": "sort-role-header" } as React.ThHTMLAttributes<HTMLTableCellElement>,
      cellProps: (row) => ({ "data-qa": `login-role-${row.username}` }) as React.TdHTMLAttributes<HTMLTableCellElement>,
    },
    {
      key: "linkedMember",
      header: "Linked Member",
      accessor: (row) => (row.linkedMember ? `${row.linkedMember.firstName} ${row.linkedMember.lastName}` : ""),
      sortable: true,
      render: (_value, row) => renderLinkedMember(row.linkedMember),
      headerProps: { "data-qa": "sort-linkedMember-header" } as React.ThHTMLAttributes<HTMLTableCellElement>,
      cellProps: (row) =>
        ({ "data-qa": `login-linkedMember-${row.username}` }) as React.TdHTMLAttributes<HTMLTableCellElement>,
    },
    {
      key: "actions",
      header: "Actions",
      accessor: () => "",
      sortable: false,
      render: (_value, row) => (
        <div className={styles.actions}>
          {hasPermission("edit", "login") && (
            <button
              className={styles.iconButton}
              onClick={() => handleEditPassword(row)}
              title="Edit Password"
              aria-label="Edit password"
              data-qa={`edit-password-btn-${row.username}`}
            >
              <FaEdit />
            </button>
          )}
          {hasPermission("delete", "login") && (
            <button
              className={`${styles.iconButton} ${styles.danger}`}
              onClick={() => handleDelete(row)}
              title="Delete Login"
              aria-label="Delete login"
              data-qa={`delete-login-btn-${row.username}`}
            >
              <FaTrash />
            </button>
          )}
        </div>
      ),
    },
  ];

  // Show loading state
  if (isLoading) {
    return (
      <div className={styles.container}>
        <div className={styles.centerContent} data-qa="logins-loading">
          <Spinner size="lg" />
          <p>Loading logins...</p>
        </div>
      </div>
    );
  }

  // Show error state
  if (error) {
    return (
      <div className={styles.container}>
        <div className={styles.errorState} data-qa="logins-error">
          <p>Error: {error}</p>
          <Button data-qa="retry-btn" onClick={() => window.location.reload()}>
            Retry
          </Button>
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
            data-qa="input-logins-search"
            type="search"
            placeholder="Search by username or role..."
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
          />
        </div>
        {hasPermission("create", "login") && (
          <Button data-qa="create-login-btn" onClick={handleCreateLogin} iconBefore={<FaPlus />}>
            Create Login
          </Button>
        )}
      </div>

      <div className={styles.resultsInfo} data-qa="logins-results-info">
        <p>
          Showing {paginatedLogins.length} of {sortedLogins.length} logins
          {debouncedSearchQuery && ` matching "${debouncedSearchQuery}"`}
        </p>
      </div>

      {/* Table */}
      <div className={styles.tableWrapper}>
        <Table
          data-qa="logins-table"
          variant="striped"
          headerVariant="primary"
          data={paginatedLogins}
          columns={columns}
          sortKey={sortKey}
          sortDirection={sortDirection}
          onSort={handleSort}
          rowProps={(row) => ({ "data-qa": `login-row-${row.username}` }) as React.HTMLAttributes<HTMLTableRowElement>}
          emptyState={
            <div className={styles.emptyState}>
              <div className={styles.emptyIllustration}>
                <FaKey size={64} />
              </div>
              {logins.length === 0 ? (
                <>
                  <h3 data-qa="logins-empty">No logins yet</h3>
                  <p>No login accounts have been created yet.</p>
                </>
              ) : (
                <>
                  <h3 data-qa="logins-no-results">No results found</h3>
                  <p>No logins found matching &quot;{debouncedSearchQuery}&quot;</p>
                  <p className={styles.emptyHint}>Try adjusting your search terms</p>
                </>
              )}
            </div>
          }
        />
      </div>

      {/* Pagination */}
      {sortedLogins.length > ITEMS_PER_PAGE && (
        <div className={styles.paginationContainer} data-qa="logins-pagination">
          <Pagination
            variant="compact"
            totalItems={sortedLogins.length}
            pageSize={ITEMS_PER_PAGE}
            currentPage={currentPage}
            onPageChange={setCurrentPage}
          />
        </div>
      )}

      {/* Modals */}
      <CreateLoginModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onSuccess={handleSuccess}
      />

      <EditPasswordModal
        isOpen={isEditModalOpen}
        login={selectedLogin}
        onClose={() => {
          setIsEditModalOpen(false);
          setSelectedLogin(null);
        }}
        onSuccess={handleSuccess}
      />

      <DeleteLoginModal
        isOpen={isDeleteModalOpen}
        login={selectedLogin}
        onClose={() => {
          setIsDeleteModalOpen(false);
          setSelectedLogin(null);
        }}
        onSuccess={handleSuccess}
      />
    </div>
  );
}
