import { Semester } from "@/types";
import { createContext } from "react";

interface ISemesterContext {
  currentSemester: Semester | null;
  loading: boolean;
  error: string;
  setCurrentSemester: (semester: Semester | null) => void;
}

export const SemesterContext = createContext<ISemesterContext | null>(null);
