import type { Semester } from "@/types";

/**
 * A semester covers the upcoming term when its end instant is still ahead.
 * Start dates are deliberately not considered: legacy data can contain an
 * end date before its start date, and the product rule is end-date-only.
 */
export function hasFutureSemester(semesters: Semester[], nowMs: number): boolean {
  return semesters.some((semester) => {
    const endMs = Date.parse(semester.endDate);
    return Number.isFinite(endMs) && endMs > nowMs;
  });
}
