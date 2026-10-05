import { AXIS_PROPS, CHART_COLORS, GRID_PROPS } from "./chartTheme";

describe("chartTheme", () => {
  it("uses the validated categorical slots", () => {
    expect(CHART_COLORS.slot1).toBe("#6f4fae");
    expect(CHART_COLORS.slot2).toBe("#c98f00");
    expect(CHART_COLORS.slot3).toBe("#2a78d6");
    expect(CHART_COLORS.slot4).toBe("#d5578a");
  });

  it("never reuses a status colour as a series colour", () => {
    const reserved = ["#d93025", "#1e8e3e", "#a50e0e", "#0d652d"];
    expect(Object.values(CHART_COLORS).filter((hex) => reserved.includes(hex))).toHaveLength(0);
  });

  it("keeps axes and grid recessive", () => {
    expect(AXIS_PROPS.tickLine).toBe(false);
    expect(AXIS_PROPS.axisLine).toBe(false);
    expect(GRID_PROPS.vertical).toBe(false);
  });
});
