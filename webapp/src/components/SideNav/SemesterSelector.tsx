import { FaGraduationCap } from "react-icons/fa";

import styles from "./SemesterSelector.module.css";
import { useCurrentSemester } from "@/hooks";
import { useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import { Semester } from "@/types";

type SemesterSelectorProps = {
  isExpanded: boolean;
  onIconClick: () => void;
};

function SemesterSelector({ isExpanded, onIconClick }: SemesterSelectorProps) {
  const { currentSemester, setCurrentSemester, error } = useCurrentSemester();
  const [semesters, setSemesters] = useState<Semester[]>([]);
  const navigate = useNavigate();

  useEffect(() => {
    const abortController = new AbortController();
    const fetchSemesters = async () => {
      const response = await fetch("/api/v2/semesters", {
        credentials: "include",
        signal: abortController.signal,
      });

      if (response.ok) {
        const data: Semester[] = await response.json();
        setSemesters(data);
      } else if (response.status === 401) {
        navigate("/admin/login");
      }
    };

    fetchSemesters();

    return () => abortController.abort();
  }, [navigate]);

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const selectedSemesterId = e.target.value;
    const selectedSemester = semesters.find((s) => s.id === selectedSemesterId);
    setCurrentSemester(selectedSemester || null);
  };

  // Handle potential 401 errors by redirecting to login
  if (error && error.includes("401")) {
    navigate("/admin/login");
  }

  if (!isExpanded) {
    return (
      <button className={styles.semesterSelectorCollapsed} title="Change the current semester" onClick={onIconClick}>
        <FaGraduationCap />
      </button>
    );
  }

  return (
    <div className={styles.semesterSelector}>
      <div className={styles.semesterLabel}>
        <span>Viewing Semester</span>
      </div>
      <select value={currentSemester?.id || ""} onChange={handleChange} className={styles.semesterDropdown}>
        {semesters.map((semester) => (
          <option key={semester.id} value={semester.id}>
            {semester.name}
          </option>
        ))}
      </select>
    </div>
  );
}

export default SemesterSelector;
