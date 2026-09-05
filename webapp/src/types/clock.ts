// Wire shape returned by every clock endpoint (server/internal/models/event_clock.go
// ClockState, #451). Timestamps are RFC3339 strings, not epoch numbers.
export interface ClockState {
  levelIndex: number;
  levelEndsAt: string;
  pausedAt: string | null;
  version: number;
  serverTime: string;
}
