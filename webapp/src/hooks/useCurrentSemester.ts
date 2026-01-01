import { useContext } from "react";
import { SemesterContext } from "@/contexts";

export const useCurrentSemester = () => {
  const context = useContext(SemesterContext);

  if (!context) {
    throw new Error("useCurrentSemester must be used within a SemesterProvider");
  }

  return context;
};
