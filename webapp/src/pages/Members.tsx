import { Route, Routes } from "react-router-dom";
import { RequirePermission } from "@/components";
import { MembersList } from "@/features/members";

export function Members() {
  return (
    <Routes>
      <Route
        path="/"
        element={
          <RequirePermission resource="membership" action="list">
            <MembersList />
          </RequirePermission>
        }
      />
    </Routes>
  );
}
