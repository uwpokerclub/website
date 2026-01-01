export const ROLES = {
  BOT: "bot",
  EXECUTIVE: "executive",
  TOURNAMENT_DIRECTOR: "tournament_director",
  SECRETARY: "secretary",
  TREASURER: "treasurer",
  VICE_PRESIDENT: "vice_president",
  PRESIDENT: "president",
  WEBMASTER: "webmaster",
} as const;

export type Role = (typeof ROLES)[keyof typeof ROLES];
