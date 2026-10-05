-- atlas:txmode none

-- Create index "idx_events_semester_state_start_date" to table: "events"
CREATE INDEX CONCURRENTLY "idx_events_semester_state_start_date" ON "events" ("semester_id", "state", "start_date");
-- Create index "idx_participants_event_id" to table: "participants"
CREATE INDEX CONCURRENTLY "idx_participants_event_id" ON "participants" ("event_id");
