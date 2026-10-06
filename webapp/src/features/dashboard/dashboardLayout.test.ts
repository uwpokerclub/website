import { CARD_SIZES, CARD_TITLES, resolveDashboardLayout } from "./dashboardLayout";
import { ROLES } from "@/types/roles";

describe("resolveDashboardLayout", () => {
  it("orders cards for the Ops preset (executive, tournament_director)", () => {
    const expected = [
      "spotlight",
      "quickActions",
      "eventActivity",
      "leaderboard",
      "termAtAGlance",
      "engagement",
      "signupTimeline",
      "trialConversion",
    ];

    expect(resolveDashboardLayout(ROLES.EXECUTIVE)).toEqual(expected);
    expect(resolveDashboardLayout(ROLES.TOURNAMENT_DIRECTOR)).toEqual(expected);
  });

  it("orders cards for the Records preset (secretary, treasurer)", () => {
    const expected = [
      "trialConversion",
      "termAtAGlance",
      "signupTimeline",
      "spotlight",
      "eventActivity",
      "engagement",
      "leaderboard",
      "quickActions",
    ];

    expect(resolveDashboardLayout(ROLES.SECRETARY)).toEqual(expected);
    expect(resolveDashboardLayout(ROLES.TREASURER)).toEqual(expected);
  });

  it("orders cards for the Leadership preset (vice_president, president, webmaster)", () => {
    const expected = [
      "engagement",
      "eventActivity",
      "termAtAGlance",
      "trialConversion",
      "spotlight",
      "signupTimeline",
      "leaderboard",
      "quickActions",
    ];

    expect(resolveDashboardLayout(ROLES.VICE_PRESIDENT)).toEqual(expected);
    expect(resolveDashboardLayout(ROLES.PRESIDENT)).toEqual(expected);
    expect(resolveDashboardLayout(ROLES.WEBMASTER)).toEqual(expected);
  });

  it("falls back to the Ops preset for an unknown role", () => {
    expect(resolveDashboardLayout("bot")).toEqual(resolveDashboardLayout(ROLES.EXECUTIVE));
  });

  it("falls back to the Ops preset for a null or undefined role", () => {
    expect(resolveDashboardLayout(null)).toEqual(resolveDashboardLayout(ROLES.EXECUTIVE));
    expect(resolveDashboardLayout(undefined)).toEqual(resolveDashboardLayout(ROLES.EXECUTIVE));
  });
});

describe("eight-card presets", () => {
  it("gives every role all eight cards exactly once", () => {
    for (const role of [ROLES.PRESIDENT, ROLES.TREASURER, ROLES.EXECUTIVE]) {
      const layout = resolveDashboardLayout(role);
      expect(layout).toHaveLength(8);
      expect(new Set(layout).size).toBe(8);
    }
  });

  it("leads each preset with the card that role opens the page for", () => {
    expect(resolveDashboardLayout(ROLES.TREASURER)[0]).toBe("trialConversion");
    expect(resolveDashboardLayout(ROLES.PRESIDENT)[0]).toBe("engagement");
    expect(resolveDashboardLayout(ROLES.TOURNAMENT_DIRECTOR)[0]).toBe("spotlight");
  });

  it("asks for wide rows for the charting cards and the two-term conversion card", () => {
    expect(CARD_SIZES.signupTimeline.min).toBe(8);
    expect(CARD_SIZES.eventActivity.min).toBe(8);
    expect(CARD_SIZES.engagement.preferred).toBe(8);
    expect(CARD_SIZES.leaderboard).toEqual({ min: 4, preferred: 4 });
    expect(CARD_SIZES.quickActions).toEqual({ min: 4, preferred: 4 });
    expect(CARD_SIZES.trialConversion).toEqual({ min: 4, preferred: 8 });
    expect(CARD_TITLES.trialConversion).toBe("Trial conversion");
  });
});
