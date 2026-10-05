import { CARD_LANES, DashboardCardId } from "./dashboardLayout";

export type LaneAssignment = {
  lead: DashboardCardId;
  wide: DashboardCardId[];
  rail: DashboardCardId[];
};

/**
 * Turns a role preset into a lead card plus two ordered lanes.
 *
 * The preset's first card is promoted to lead regardless of its own lane
 * preference — that promotion is what makes role ordering visible rather than a
 * reshuffle of identical boxes. Everything after it is routed by preference,
 * relative order intact, so a reader scanning either lane still reads the preset
 * in order.
 */
export function assignLanes(preset: readonly DashboardCardId[]): LaneAssignment {
  const [lead, ...rest] = preset;

  return {
    lead,
    wide: rest.filter((id) => CARD_LANES[id] === "wide"),
    rail: rest.filter((id) => CARD_LANES[id] === "rail"),
  };
}
