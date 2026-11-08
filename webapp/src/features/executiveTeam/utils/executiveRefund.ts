import type { Semester } from "@/types";
import type { Membership } from "@/types/membership";

export function executiveMembershipRefund(
  membership: Pick<Membership, "paid" | "discounted"> | undefined,
  semester: Pick<Semester, "membershipFee" | "membershipDiscountFee">,
): number {
  if (!membership?.paid) return 0;
  return membership.discounted ? semester.membershipDiscountFee : semester.membershipFee;
}
