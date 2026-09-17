import type { OfficerRole, OfficerTransition } from "./types";

export const roleLabels: Record<OfficerRole, string> = {
  president: "President",
  vice_president: "Vice President",
  secretary: "Secretary",
  treasurer: "Treasurer",
};

export function usernameForRole(transition: OfficerTransition, role: OfficerRole): string | undefined {
  const usernames: Record<OfficerRole, string | undefined> = {
    president: transition.presidentUsername,
    vice_president: transition.vicePresidentUsername,
    secretary: transition.secretaryUsername,
    treasurer: transition.treasurerUsername,
  };
  return usernames[role];
}

export function activationLabel(activated: boolean | undefined): string {
  return activated ? "Account activated" : "Awaiting account activation";
}
