import { Route, Routes } from "react-router-dom";
import { CreateLogin, LoginPage } from "../features/login";
import { RequireAuth, RequirePermission } from "@/components";

export function Login() {
  return (
    <Routes>
      <Route path="/" element={<LoginPage />} />
      <Route
        path="/new"
        element={
          <RequireAuth>
            <RequirePermission resource="login" action="create">
              <CreateLogin />
            </RequirePermission>
          </RequireAuth>
        }
      />
    </Routes>
  );
}
