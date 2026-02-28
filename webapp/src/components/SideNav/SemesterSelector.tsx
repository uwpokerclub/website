import { FaGraduationCap, FaPlus, FaChevronDown } from "react-icons/fa";

import styles from "./SemesterSelector.module.css";
import { useCurrentSemester, useAuth } from "@/hooks";
import { useNavigate } from "react-router-dom";
import { useEffect, useState, useCallback, useRef } from "react";
import { Semester } from "@/types";
import { CreateSemesterModal } from "@/features/semesters";

type SemesterSelectorProps = {
  isExpanded: boolean;
  onIconClick: () => void;
};

function SemesterSelector({ isExpanded, onIconClick }: SemesterSelectorProps) {
  const { currentSemester, setCurrentSemester, error } = useCurrentSemester();
  const { hasPermission } = useAuth();
  const [semesters, setSemesters] = useState<Semester[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();

  const canCreateSemester = hasPermission("create", "semester");

  const fetchSemesters = useCallback(
    async (signal?: AbortSignal) => {
      const response = await fetch("/api/v2/semesters", {
        credentials: "include",
        signal,
      });

      if (response.ok) {
        const resp: { data: Semester[] } = await response.json();
        setSemesters(resp.data);
      } else if (response.status === 401) {
        navigate("/admin/login");
      }
    },
    [navigate],
  );

  useEffect(() => {
    const abortController = new AbortController();
    fetchSemesters(abortController.signal);
    return () => abortController.abort();
  }, [fetchSemesters]);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsDropdownOpen(false);
      }
    };

    if (isDropdownOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isDropdownOpen]);

  const handleSemesterSelect = (semester: Semester) => {
    setCurrentSemester(semester);
    setIsDropdownOpen(false);
  };

  const handleCreateClick = () => {
    setIsDropdownOpen(false);
    setIsModalOpen(true);
  };

  const handleModalClose = () => {
    setIsModalOpen(false);
  };

  const handleCreateSuccess = (newSemester: Semester) => {
    // Add new semester to the list (at the beginning since it's newest)
    setSemesters((prev) => [newSemester, ...prev]);
    // Auto-select the new semester
    setCurrentSemester(newSemester);
    setIsModalOpen(false);
  };

  // Handle potential 401 errors by redirecting to login
  if (error && error.includes("401")) {
    navigate("/admin/login");
  }

  if (!isExpanded) {
    return (
      <>
        <button
          className={styles.semesterSelectorCollapsed}
          title="Change the current semester"
          onClick={onIconClick}
          data-qa="semester-selector"
        >
          <FaGraduationCap />
        </button>
        {canCreateSemester && (
          <CreateSemesterModal isOpen={isModalOpen} onClose={handleModalClose} onSuccess={handleCreateSuccess} />
        )}
      </>
    );
  }

  return (
    <>
      <div className={styles.semesterSelector} ref={dropdownRef} data-qa="semester-selector">
        <div className={styles.semesterLabel}>
          <span>Viewing Semester</span>
        </div>
        <button
          className={styles.dropdownTrigger}
          onClick={() => setIsDropdownOpen(!isDropdownOpen)}
          data-qa="semester-dropdown"
          aria-expanded={isDropdownOpen}
          aria-haspopup="listbox"
        >
          <span className={styles.selectedValue}>{currentSemester?.name || "Select semester"}</span>
          <FaChevronDown className={`${styles.dropdownArrow} ${isDropdownOpen ? styles.open : ""}`} />
        </button>

        {isDropdownOpen && (
          <div className={styles.dropdownMenu} role="listbox">
            <div className={styles.dropdownOptions}>
              {semesters.map((semester) => (
                <button
                  key={semester.id}
                  className={`${styles.dropdownOption} ${semester.id === currentSemester?.id ? styles.selected : ""}`}
                  onClick={() => handleSemesterSelect(semester)}
                  role="option"
                  aria-selected={semester.id === currentSemester?.id}
                  data-qa={`semester-option-${semester.id}`}
                >
                  {semester.name}
                </button>
              ))}
            </div>

            {canCreateSemester && (
              <>
                <div className={styles.dropdownDivider} />
                <button
                  className={styles.createSemesterOption}
                  onClick={handleCreateClick}
                  data-qa="create-semester-btn"
                >
                  <FaPlus className={styles.createIcon} />
                  <span>Create Semester</span>
                </button>
              </>
            )}
          </div>
        )}
      </div>

      {canCreateSemester && (
        <CreateSemesterModal isOpen={isModalOpen} onClose={handleModalClose} onSuccess={handleCreateSuccess} />
      )}
    </>
  );
}

export default SemesterSelector;
