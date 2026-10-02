import { CARD_LANES, CARD_TITLES, resolveDashboardLayout } from "./dashboardLayout";
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

  it("routes only the charting cards to the wide lane", () => {
    expect(CARD_LANES.signupTimeline).toBe("wide");
    expect(CARD_LANES.eventActivity).toBe("wide");
    expect(CARD_LANES.engagement).toBe("wide");
    expect(CARD_LANES.leaderboard).toBe("rail");
    expect(CARD_LANES.quickActions).toBe("rail");
    expect(CARD_TITLES.trialConversion).toBe("Trial conversion");
  });
});
