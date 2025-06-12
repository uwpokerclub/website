import { SemesterContext } from "@/contexts";
import { Semester } from "@/types";
import { ReactNode, useEffect, useState } from "react";

interface SemesterProviderProps {
  children: ReactNode;
}

export default function SemesterProvider({ children }: SemesterProviderProps) {
  const [currentSemester, setCurrentSemester] = useState<Semester | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const fetchSemestersForInitialSelection = async (abortSignal: AbortSignal) => {
    setLoading(true);

    try {
      const response = await fetch("/api/semesters", {
        credentials: "include",
        signal: abortSignal,
      });

      if (response.ok) {
        const data: Semester[] = await response.json();

        // Set the first semester (latest) as current if we don't have one selected
        if (data.length > 0) {
          setCurrentSemester(data[0]);
        }
      } else {
        setError("Failed to fetch semesters");
      }
    } catch (err) {
      if (!(err instanceof DOMException && err.name === "AbortError")) {
        setError("Network error while fetching semesters");
      }
    } finally {
      if (!abortSignal.aborted) {
        setLoading(false);
      }
    }
  };

  const value = {
    currentSemester,
    loading,
    error,
    setCurrentSemester,
  };

  useEffect(() => {
    const abortController = new AbortController();

    fetchSemestersForInitialSelection(abortController.signal);

    return () => abortController.abort();
  }, []);

  return <SemesterContext.Provider value={value}>{children}</SemesterContext.Provider>;
}
