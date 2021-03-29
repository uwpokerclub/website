export type Entry = {
  id: number;
  user_id: string;
  event_id: string;
  placement: number;
  signed_out_at: Date | null;
};
