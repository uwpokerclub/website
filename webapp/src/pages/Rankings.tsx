import { Route, Routes } from "react-router-dom";
import { RankingsTable, SemesterList } from "../features/rankings";
import { RequirePermission } from "@/components";

export function Rankings() {
  return (
    <Routes>
      <Route
        path="/"
        element={
          <RequirePermission resource="semester" action="list">
            <SemesterList />
          </RequirePermission>
        }
      />
      <Route
        path="/:semesterId"
        element={
          <RequirePermission resource="semester" subResource="rankings" action="list">
            <RankingsTable />
          </RequirePermission>
        }
      />
    </Routes>
  );
}
