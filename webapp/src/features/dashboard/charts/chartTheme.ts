/**
 * Shared Recharts theming (#439).
 *
 * The palette is validated rather than chosen by eye — see the Colour section of
 * docs/superpowers/specs/2026-10-01-dashboard-page-design.md. Slots 1 and 2 are the
 * club's purple and gold stepped into the passing lightness band; the raw brand
 * hexes fail it.
 *
 * Red and green are reserved for DeltaChip and must never appear here: a series
 * that happens to be red would read as a bad number rather than as a category.
 *
 * These are literal hexes rather than var(--dash-*) because Recharts passes them
 * into SVG attributes that do not resolve custom properties in every context.
 */
export const CHART_COLORS = {
  slot1: "#6f4fae",
  slot2: "#c98f00",
  slot3: "#2a78d6",
  slot4: "#d5578a",
} as const;

export const AXIS_PROPS = {
  tickLine: false,
  axisLine: false,
  tick: { fontSize: 11, fill: "#9e9e9e", fontFamily: "Montserrat, sans-serif" },
} as const;

export const GRID_PROPS = {
  stroke: "#eeeeee",
  vertical: false,
} as const;

export const TOOLTIP_STYLE = {
  contentStyle: {
    fontSize: 12,
    fontFamily: "Montserrat, sans-serif",
    borderRadius: 6,
    border: "1px solid #e0e0e0",
    boxShadow: "0 1px 3px rgb(0 0 0 / 0.1)",
  },
} as const;
