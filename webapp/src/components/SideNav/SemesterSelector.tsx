import { FaGraduationCap, FaPlus, FaChevronDown } from "react-icons/fa";

import styles from "./SemesterSelector.module.css";
import { useCurrentSemester, useAuth } from "@/hooks";
import { useNavigate } from "react-router-dom";
import { useEffect, useState, useRef } from "react";
import { Semester } from "@/types";
import { CreateSemesterModal } from "@/features/semesters";
import { useSemesters } from "@/features/semesters/hooks/useSemesterQueries";

type SemesterSelectorProps = {
  isExpanded: boolean;
  onIconClick: () => void;
};

function SemesterSelector({ isExpanded, onIconClick }: SemesterSelectorProps) {
  const { currentSemester, setCurrentSemester, error } = useCurrentSemester();
  const { hasPermission } = useAuth();
  const { data: semesters = [] } = useSemesters();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();

  const canCreateSemester = hasPermission("create", "semester");

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
    // Auto-select the new semester — the query cache is invalidated by the mutation
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
