export interface ExecutiveTeamMember {
  memberId: number;
  firstName: string;
  lastName: string;
  questId: string;
  username: string | null;
  role: string | null;
  accountStatus: "active" | "pending_activation" | "disabled" | null;
  activationOutstanding: boolean;
  hasLogin: boolean;
  hasExecutiveMembership: boolean;
}

export interface ExecutiveTeamResponse {
  members: ExecutiveTeamMember[];
}

export interface ExecutiveTeamActivationLinkResponse {
  activationToken: string;
}
