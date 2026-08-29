export type Semester = {
  id: string;
  name: string;
  meta: string;
  startDate: string;
  endDate: string;
  startingBudget: number;
  currentBudget: number;
  membershipFee: number;
  membershipDiscountFee: number;
  rebuyFee: number;
  /**
   * Number of events an unpaid member may attend before their free trial runs
   * out. 0 disables the check — no member is ever flagged or blocked.
   */
  freeTrialLimit: number;
};
