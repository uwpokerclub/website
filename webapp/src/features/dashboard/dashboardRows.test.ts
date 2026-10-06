import { ROLES } from "@/types/roles";
import { CARD_SIZES, resolveDashboardLayout } from "./dashboardLayout";
import { PlacedCard, placeDashboardCards } from "./dashboardRows";

const ROLE_PRESETS = [ROLES.PRESIDENT, ROLES.TREASURER, ROLES.EXECUTIVE].map((role) => resolveDashboardLayout(role));

/** Groups placed cards into rows of 12 columns at the given width, as the grid flows them. */
function rowsAt(cards: PlacedCard[], width: "medium" | "wide") {
  const rows: string[][] = [];
  let used = 12;

  for (const card of cards) {
    const span = card.spans[width];
    if (used + span > 12) {
      rows.push([]);
      used = 0;
    }
    rows[rows.length - 1].push(`${card.id}:${span}`);
    used += span;
  }

  return rows;
}

describe("placeDashboardCards", () => {
  it("keeps preset order and promotes only the first card to lead", () => {
    for (const preset of ROLE_PRESETS) {
      const placed = placeDashboardCards(preset);

      expect(placed.map((card) => card.id)).toEqual(preset);
      expect(placed.filter((card) => card.lead).map((card) => card.id)).toEqual([preset[0]]);
    }
  });

  it("fills every row exactly, at both widths", () => {
    for (const preset of ROLE_PRESETS) {
      const placed = placeDashboardCards(preset);

      for (const width of ["medium", "wide"] as const) {
        const totals = rowsAt(placed, width).map((row) =>
          row.reduce((sum, cell) => sum + Number(cell.split(":")[1]), 0),
        );
        expect(totals.every((total) => total === 12)).toBe(true);
      }
    }
  });

  it("never narrows a card below its minimum", () => {
    for (const preset of ROLE_PRESETS) {
      for (const card of placeDashboardCards(preset)) {
        expect(card.spans.wide).toBeGreaterThanOrEqual(CARD_SIZES[card.id].min);
      }
    }
  });

  it("gives the lead at least two thirds of a desktop row and a full two-up row", () => {
    const leads = ROLE_PRESETS.map((preset) => placeDashboardCards(preset)[0]);

    expect(leads.map((lead) => lead.spans.wide)).toEqual([12, 8, 8]);
    expect(leads.map((lead) => lead.spans.medium)).toEqual([12, 12, 12]);
  });

  it("composes the Leadership desktop rows", () => {
    expect(rowsAt(placeDashboardCards(resolveDashboardLayout(ROLES.PRESIDENT)), "wide")).toEqual([
      ["engagement:12"],
      ["eventActivity:8", "termAtAGlance:4"],
      ["trialConversion:8", "spotlight:4"],
      ["signupTimeline:12"],
      ["leaderboard:6", "quickActions:6"],
    ]);
  });

  it("composes the Ops desktop rows", () => {
    expect(rowsAt(placeDashboardCards(resolveDashboardLayout(ROLES.EXECUTIVE)), "wide")).toEqual([
      ["spotlight:8", "quickActions:4"],
      ["eventActivity:8", "leaderboard:4"],
      ["termAtAGlance:4", "engagement:8"],
      ["signupTimeline:12"],
      ["trialConversion:12"],
    ]);
  });

  it("composes the Records desktop rows", () => {
    expect(rowsAt(placeDashboardCards(resolveDashboardLayout(ROLES.TREASURER)), "wide")).toEqual([
      ["trialConversion:8", "termAtAGlance:4"],
      ["signupTimeline:12"],
      ["spotlight:4", "eventActivity:8"],
      ["engagement:8", "leaderboard:4"],
      ["quickActions:12"],
    ]);
  });

  it("gives the lead and wide-preferring cards full rows in the two-up layout, pairing the rest", () => {
    expect(rowsAt(placeDashboardCards(resolveDashboardLayout(ROLES.PRESIDENT)), "medium")).toEqual([
      ["engagement:12"],
      ["eventActivity:12"],
      ["termAtAGlance:12"],
      ["trialConversion:12"],
      ["spotlight:12"],
      ["signupTimeline:12"],
      ["leaderboard:6", "quickActions:6"],
    ]);
  });
});
