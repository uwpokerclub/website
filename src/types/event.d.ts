export type EventState = 0 | 1;

export type Event = {
  id: string;
  semester_id: string;
  name: string;
  start_date: Date;
  format: string;
  notes: string;
  state: EventState;
};
