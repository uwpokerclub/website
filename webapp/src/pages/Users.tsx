import { Route, Routes } from "react-router-dom";
import { EditUser, ListUsers, NewUser, ShowUser } from "../features/users";
import { RequirePermission } from "@/components";

export function Users() {
  return (
    <Routes>
      <Route
        path="/"
        element={
          <RequirePermission resource="user" action="list">
            <ListUsers />
          </RequirePermission>
        }
      />
      <Route
        path="/new"
        element={
          <RequirePermission resource="user" action="create">
            <NewUser />
          </RequirePermission>
        }
      />
      <Route
        path="/:userId"
        element={
          <RequirePermission resource="user" action="get">
            <ShowUser />
          </RequirePermission>
        }
      />
      <Route
        path="/:userId/edit"
        element={
          <RequirePermission resource="user" action="edit">
            <EditUser />
          </RequirePermission>
        }
      />
    </Routes>
  );
}
