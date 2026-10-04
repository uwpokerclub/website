import type { Semester } from "@/types";
import { hasFutureSemester } from "./hasFutureSemester";

const nowMs = Date.parse("2026-09-18T12:00:00Z");

function semester(overrides: Partial<Semester>): Semester {
  return {
    id: "semester-id",
    name: "Fall 2026",
    meta: "",
    startDate: "2026-09-01T00:00:00Z",
    endDate: "2026-12-31T00:00:00Z",
    startingBudget: 0,
    currentBudget: 0,
    membershipFee: 10,
    membershipDiscountFee: 5,
    rebuyFee: 2,
    freeTrialLimit: 0,
    ...overrides,
  };
}

describe("hasFutureSemester", () => {
  it("returns false when no semesters are present or all end dates have passed", () => {
    expect(hasFutureSemester([], nowMs)).toBe(false);
    expect(hasFutureSemester([semester({ endDate: "2026-09-18T12:00:00Z" })], nowMs)).toBe(false);
  });

  it("returns true when any parseable end date is strictly in the future", () => {
    expect(hasFutureSemester([semester({ endDate: "2026-09-18T12:00:00.001Z" })], nowMs)).toBe(true);
  });

  it("ignores unparseable end dates", () => {
    expect(hasFutureSemester([semester({ endDate: "not-a-date" })], nowMs)).toBe(false);
  });

  it("uses only the end date for reversed legacy ranges", () => {
    expect(
      hasFutureSemester([semester({ startDate: "2027-01-01T00:00:00Z", endDate: "2026-12-31T00:00:00Z" })], nowMs),
    ).toBe(true);
  });
});
