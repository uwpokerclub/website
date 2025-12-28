import { Route, Routes } from "react-router-dom";
import { EventDetails, ListEvents, NewEvent } from "../features/events";
import { EditEventPage } from "./events/EditEventPage";
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
        path="/new"
        element={
          <RequirePermission resource="event" action="create">
            <NewEvent />
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
      <Route
        path="/:eventId/edit"
        element={
          <RequirePermission resource="event" action="edit">
            <EditEventPage />
          </RequirePermission>
        }
      />
    </Routes>
  );
}
