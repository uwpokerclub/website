import { Routes, Route } from "react-router-dom";
import { SemestersList, SemesterInfo } from "../features/semesters";
import { RequirePermission } from "@/components";

export function Semesters() {
  return (
    <Routes>
      <Route
        path="/"
        element={
          <RequirePermission resource="semester" action="list">
            <SemestersList />
          </RequirePermission>
        }
      />
      <Route
        path="/:semesterId"
        element={
          <RequirePermission resource="semester" action="get">
            <SemesterInfo />
          </RequirePermission>
        }
      />
    </Routes>
  );
}
