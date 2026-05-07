import { Route, Routes } from "react-router-dom";
import { EventDetails, ListEvents } from "../features/events";
import { RequirePermission } from "@/components";

export function Events() {
  return (
    <Routes>
      <Route
        path="/"
        element={
          <RequirePermission resource="event" action="list">
            <ListEvents />
          </RequirePermission>
        }
      />
      <Route
        path="/:eventId"
        element={
          <RequirePermission resource="event" action="get">
            <EventDetails />
          </RequirePermission>
        }
      />
    </Routes>
  );
}
