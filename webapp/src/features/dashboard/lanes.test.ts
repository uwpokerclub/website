import { assignLanes } from "./lanes";
import { resolveDashboardLayout } from "./dashboardLayout";
import { ROLES } from "@/types/roles";

describe("assignLanes", () => {
  it("promotes the preset's first card to lead and keeps it out of both lanes", () => {
    const { lead, wide, rail } = assignLanes(resolveDashboardLayout(ROLES.PRESIDENT));

    expect(lead).toBe("engagement");
    expect(wide).not.toContain("engagement");
    expect(rail).not.toContain("engagement");
  });

  it("leads with a rail-preference card when the preset does", () => {
    const { lead, wide } = assignLanes(resolveDashboardLayout(ROLES.TREASURER));

    expect(lead).toBe("trialConversion");
    expect(wide).toEqual(["signupTimeline", "eventActivity", "engagement"]);
  });

  it("preserves relative preset order within each lane", () => {
    const { rail } = assignLanes(resolveDashboardLayout(ROLES.PRESIDENT));

    expect(rail).toEqual(["termAtAGlance", "trialConversion", "spotlight", "leaderboard", "quickActions"]);
  });

  it("accounts for every card exactly once", () => {
    const preset = resolveDashboardLayout(ROLES.EXECUTIVE);
    const { lead, wide, rail } = assignLanes(preset);

    expect([lead, ...wide, ...rail].sort()).toEqual([...preset].sort());
  });
});
