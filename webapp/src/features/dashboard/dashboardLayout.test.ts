import { resolveDashboardLayout } from "./dashboardLayout";
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
    ];

    expect(resolveDashboardLayout(ROLES.EXECUTIVE)).toEqual(expected);
    expect(resolveDashboardLayout(ROLES.TOURNAMENT_DIRECTOR)).toEqual(expected);
  });

  it("orders cards for the Records preset (secretary, treasurer)", () => {
    const expected = [
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
