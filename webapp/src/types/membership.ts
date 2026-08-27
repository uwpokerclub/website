import type { User } from "./user";

export type Membership = {
  id: string;
  userId: number;
  user: User;
  semesterId: string;
  paid: boolean;
  discounted: boolean;
  /**
   * Computed by the API (models.MembershipWithAttendance), not a stored column.
   * No longer read by any component — the free-trial warning now reads
   * freeTrialAvailable instead. Kept because the list endpoint still returns it.
   */
  attendance: number;
  /**
   * False when this membership has used up its semester's free trial events.
   * A display-only cache maintained by the API; it can go stale (deleting an
   * entry never restores it), and it is only meaningful for unpaid memberships,
   * so never read it without also checking `paid` — use hasExhaustedFreeTrial.
   */
  freeTrialAvailable: boolean;
};
