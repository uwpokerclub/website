import type { Structure } from "./structures";

/**
 * Event is the JSON object returned by the v2 events API.
 * `entries` and `structure` are populated on some endpoints (list, detail) but not others.
 */
export type Event = {
  id: number;
  name: string;
  format: string;
  notes: string;
  semesterId: string;
  startDate: string;
  state: number;
  rebuys: number;
  pointsMultiplier: number;
  structureId: number;
  structure?: Structure;
  entries?: { membershipId: string }[];
};
