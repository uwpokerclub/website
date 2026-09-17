export const OFFICER_ROLES = ["president", "vice_president", "secretary", "treasurer"] as const;

export type OfficerRole = (typeof OFFICER_ROLES)[number];

export interface OfficerTransitionFormData {
  presidentQuestId: string;
  vicePresidentQuestId: string;
  secretaryQuestId: string;
  treasurerQuestId: string;
}

export interface OfficerTransitionNominee {
  firstName: string;
  lastName: string;
}

export interface OfficerTransition {
  id: string;
  status: "pending" | "completed" | "cancelled";
  nominees: Partial<Record<OfficerRole, OfficerTransitionNominee>>;
}

export interface StageOfficerTransitionResponse {
  transition: OfficerTransition;
  activationTokens: Partial<Record<OfficerRole, string>>;
}

export interface ResolvedQuestId {
  questId: string;
  nominee: OfficerTransitionNominee;
}
