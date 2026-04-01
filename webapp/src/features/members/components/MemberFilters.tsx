import { useEffect, useRef } from "react";
import { Input, Select, Button } from "@uwpokerclub/components";
import { FaTimes } from "react-icons/fa";
import { FACULTIES } from "@/data/constants";
import styles from "./MemberFilters.module.css";

export interface MemberFilterValues {
  studentId: string;
  name: string;
  email: string;
  faculty: string;
  paid: string;
  discounted: string;
}

interface MemberFiltersProps {
  isOpen: boolean;
  onClose: () => void;
  filters: MemberFilterValues;
  onFilterChange: (key: keyof MemberFilterValues, value: string) => void;
  onClear: () => void;
}

const FACULTY_OPTIONS = [{ value: "", label: "All" }, ...FACULTIES.map((f) => ({ value: f, label: f }))];

const BOOL_OPTIONS = [
  { value: "", label: "All" },
  { value: "true", label: "Yes" },
  { value: "false", label: "No" },
];

export function MemberFilters({ isOpen, onClose, filters, onFilterChange, onClear }: MemberFiltersProps) {
  const activeCount = Object.values(filters).filter((v) => v !== "").length;
  const panelRef = useRef<HTMLDivElement>(null);
  const onCloseRef = useRef(onClose);
  onCloseRef.current = onClose;

  // Close on Escape key or click outside
  useEffect(() => {
    if (!isOpen) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onCloseRef.current();
    };
    const handleClick = (e: MouseEvent) => {
      if (panelRef.current && !panelRef.current.contains(e.target as Node)) {
        onCloseRef.current();
      }
    };
    document.addEventListener("keydown", handleKeyDown);
    document.addEventListener("mousedown", handleClick);
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      document.removeEventListener("mousedown", handleClick);
    };
  }, [isOpen]);

  return (
    <aside ref={panelRef} className={`${styles.drawer} ${isOpen ? styles.open : ""}`} data-qa="filter-sidebar">
      <div className={styles.header}>
        <span className={styles.headerTitle}>Filters</span>
        <button type="button" className={styles.closeButton} onClick={onClose} aria-label="Close filters">
          <FaTimes />
        </button>
      </div>

      <div className={styles.fields}>
        <div className={styles.field}>
          <label className={styles.label} htmlFor="filter-studentId">
            Student Number
          </label>
          <Input
            id="filter-studentId"
            data-qa="filter-studentId"
            type="text"
            placeholder="e.g. 20780648"
            value={filters.studentId}
            onChange={(e) => onFilterChange("studentId", e.target.value)}
            fullWidth
          />
        </div>

        <div className={styles.field}>
          <label className={styles.label} htmlFor="filter-name">
            Name
          </label>
          <Input
            id="filter-name"
            data-qa="filter-name"
            type="text"
            placeholder="Search by name..."
            value={filters.name}
            onChange={(e) => onFilterChange("name", e.target.value)}
            fullWidth
          />
        </div>

        <div className={styles.field}>
          <label className={styles.label} htmlFor="filter-email">
            Email
          </label>
          <Input
            id="filter-email"
            data-qa="filter-email"
            type="text"
            placeholder="Search by email..."
            value={filters.email}
            onChange={(e) => onFilterChange("email", e.target.value)}
            fullWidth
          />
        </div>

        <div className={styles.field}>
          <label className={styles.label} htmlFor="filter-faculty">
            Faculty
          </label>
          <Select
            id="filter-faculty"
            data-qa="filter-faculty"
            options={FACULTY_OPTIONS}
            value={filters.faculty}
            onChange={(e) => onFilterChange("faculty", e.target.value)}
            fullWidth
          />
        </div>

        <div className={styles.field}>
          <label className={styles.label} htmlFor="filter-paid">
            Paid
          </label>
          <Select
            id="filter-paid"
            data-qa="filter-paid"
            options={BOOL_OPTIONS}
            value={filters.paid}
            onChange={(e) => onFilterChange("paid", e.target.value)}
            fullWidth
          />
        </div>

        <div className={styles.field}>
          <label className={styles.label} htmlFor="filter-discounted">
            Discounted
          </label>
          <Select
            id="filter-discounted"
            data-qa="filter-discounted"
            options={BOOL_OPTIONS}
            value={filters.discounted}
            onChange={(e) => onFilterChange("discounted", e.target.value)}
            fullWidth
          />
        </div>
      </div>

      {activeCount > 0 && (
        <Button data-qa="filter-clear-btn" variant="tertiary" onClick={onClear} iconBefore={<FaTimes />} fullWidth>
          Clear All Filters
        </Button>
      )}
    </aside>
  );
}
