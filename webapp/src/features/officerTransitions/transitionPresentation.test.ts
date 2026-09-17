import { activationLabel, roleLabels, usernameForRole } from "./transitionPresentation";
import type { OfficerTransition } from "./types";

const transition: OfficerTransition = {
  id: "transition-id",
  status: "pending",
  presidentUsername: "president-id",
  vicePresidentUsername: "vp-id",
  secretaryUsername: "secretary-id",
  treasurerUsername: "treasurer-id",
  nominees: {},
};

describe("pending transition presentation", () => {
  it("maps each API Quest ID to its officer role", () => {
    expect(usernameForRole(transition, "president")).toBe("president-id");
    expect(usernameForRole(transition, "vice_president")).toBe("vp-id");
    expect(usernameForRole(transition, "secretary")).toBe("secretary-id");
    expect(usernameForRole(transition, "treasurer")).toBe("treasurer-id");
    expect(roleLabels.vice_president).toBe("Vice President");
  });

  it("makes inactive and active nominee progress unambiguous", () => {
    expect(activationLabel(false)).toBe("Activation required");
    expect(activationLabel(undefined)).toBe("Activation required");
    expect(activationLabel(true)).toBe("Account already active");
  });
});
