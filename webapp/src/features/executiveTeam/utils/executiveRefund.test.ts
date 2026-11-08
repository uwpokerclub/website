import { executiveMembershipRefund } from "./executiveRefund";

describe("executiveMembershipRefund", () => {
  const semester = { membershipFee: 20, membershipDiscountFee: 8 };

  it("returns the paid membership fee", () => {
    expect(executiveMembershipRefund({ paid: true, discounted: false }, semester)).toBe(20);
  });

  it("returns the discounted fee for discounted members", () => {
    expect(executiveMembershipRefund({ paid: true, discounted: true }, semester)).toBe(8);
  });

  it("does not refund an unpaid or missing membership", () => {
    expect(executiveMembershipRefund({ paid: false, discounted: false }, semester)).toBe(0);
    expect(executiveMembershipRefund(undefined, semester)).toBe(0);
  });
});
