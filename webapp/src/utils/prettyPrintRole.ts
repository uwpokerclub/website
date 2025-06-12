import { Role, ROLES } from "@/types/roles";

export default function prettyPrintRole(role: Role): string {
  switch (role) {
    case ROLES.EXECUTIVE:
      return "Executive";
    case ROLES.TOURNAMENT_DIRECTOR:
      return "Tournament Director";
    case ROLES.SECRETARY:
      return "Secretary";
    case ROLES.TREASURER:
      return "Treasurer";
    case ROLES.VICE_PRESIDENT:
      return "Vice President";
    case ROLES.PRESIDENT:
      return "President";
    case ROLES.WEBMASTER:
      return "Webmaster";
    case ROLES.BOT:
      return "Bot";
    default:
      throw new Error(`Unknown role: ${role}`);
  }
}
