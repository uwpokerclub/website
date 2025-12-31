/**
 * Shared helper functions for Cypress E2E tests
 */

import { MEMBERS, USERS } from "../seed";
import { Membership, User } from "../types";

/**
 * Get the user associated with a membership
 * @param memberOrId - Either a Membership object or a membership ID string
 * @returns The User object associated with the membership
 * @throws Error if member or user is not found
 */
export function getUserForMember(memberOrId: Membership | string): User {
  const member =
    typeof memberOrId === "string"
      ? MEMBERS.find((m) => m.id === memberOrId)
      : memberOrId;

  if (!member) {
    throw new Error(
      `Member not found: ${typeof memberOrId === "string" ? memberOrId : memberOrId.id}`
    );
  }

  const user = USERS.find((u) => u.id === member.userId);
  if (!user) {
    throw new Error(`User not found for member userId: ${member.userId}`);
  }

  return user;
}

/**
 * Format a member's full name
 * @param member - The membership to get the name for
 * @returns Full name in "FirstName LastName" format
 */
export function getMemberFullName(member: Membership): string {
  const user = getUserForMember(member);
  return `${user.firstName} ${user.lastName}`;
}

/**
 * Semantic aliases for common member states in tests
 */
export const UNPAID_MEMBER = MEMBERS[0]; // Heinrik Drust - paid: false, discounted: false
export const PAID_MEMBER = MEMBERS[3]; // Khalil Duckham - paid: true, discounted: false
export const DISCOUNTED_MEMBER = MEMBERS[4]; // Amandie Libbis - paid: true, discounted: true

/**
 * Default blind level values used in structure creation
 */
export const DEFAULT_BLIND_LEVEL = {
  small: 25,
  big: 50,
  ante: 0,
  time: 15,
} as const;
