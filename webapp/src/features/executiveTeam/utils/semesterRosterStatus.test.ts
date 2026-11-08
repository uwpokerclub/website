import { semesterRosterStatus } from "./semesterRosterStatus";

describe("semesterRosterStatus", () => {
  const now = new Date("2026-09-01T12:00:00.000Z");

  it("treats the end date as inclusive in the Toronto calendar", () => {
    expect(semesterRosterStatus({ startDate: "2026-01-01", endDate: "2026-09-01" }, now)).toBe("current");
    expect(semesterRosterStatus({ startDate: "2026-01-01", endDate: "2026-08-31" }, now)).toBe("past");
  });

  it("uses the Toronto day even when UTC is already the next day", () => {
    expect(
      semesterRosterStatus({ startDate: "2026-01-01", endDate: "2026-08-31" }, new Date("2026-09-01T02:30:00.000Z")),
    ).toBe("current");
  });

  it.each([
    [{ startDate: "", endDate: "2026-08-31" }],
    [{ startDate: "2026-01-01", endDate: "invalid" }],
    [{ startDate: "2026-01-01Tbad", endDate: "2026-08-31" }],
    [{ startDate: "2026-09-01", endDate: "2026-08-31" }],
    [{ startDate: "2026-02-30", endDate: "2026-08-31" }],
  ])("marks missing, malformed, or reversed dates unknown", (semester) => {
    expect(semesterRosterStatus(semester, now)).toBe("unknown");
  });
});
