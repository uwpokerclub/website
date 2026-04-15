import { SemesterContext } from "@/contexts";
import { useSemesters } from "@/features/semesters/hooks/useSemesterQueries";
import { Semester } from "@/types";
import { ReactNode, useEffect, useRef, useState } from "react";

interface SemesterProviderProps {
  children: ReactNode;
}

export default function SemesterProvider({ children }: SemesterProviderProps) {
  const [currentSemester, setCurrentSemester] = useState<Semester | null>(null);
  const { data: semesters, isLoading, error } = useSemesters();
  const hasAutoSelected = useRef(false);

  // Auto-select the first semester (latest) on initial load only
  useEffect(() => {
    if (!hasAutoSelected.current && semesters && semesters.length > 0) {
      hasAutoSelected.current = true;
      setCurrentSemester(semesters[0]);
    }
  }, [semesters]);

  const value = {
    currentSemester,
    loading: isLoading,
    error: error?.message ?? "",
    setCurrentSemester,
  };

  return <SemesterContext.Provider value={value}>{children}</SemesterContext.Provider>;
}
