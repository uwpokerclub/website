import { Route, Routes } from "react-router-dom";
import { RequirePermission } from "@/components";
import { LoginsPage } from "@/features/logins/pages/LoginsPage";

export function Logins() {
  return (
    <Routes>
      <Route
        path="/"
        element={
          <RequirePermission resource="login" action="list">
            <LoginsPage />
          </RequirePermission>
        }
      />
    </Routes>
  );
}
