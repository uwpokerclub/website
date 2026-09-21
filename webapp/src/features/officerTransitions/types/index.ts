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
  presidentUsername?: string;
  vicePresidentUsername?: string;
  secretaryUsername?: string;
  treasurerUsername?: string;
  nominees?: Partial<Record<OfficerRole, OfficerTransitionNominee>>;
  activated?: Partial<Record<OfficerRole, boolean>>;
}

export interface StageOfficerTransitionResponse {
  transition: OfficerTransition;
  activationTokens: Partial<Record<OfficerRole, string>>;
}

export interface ResolvedQuestId {
  questId: string;
  nominee: OfficerTransitionNominee;
}
