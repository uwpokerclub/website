import type { Semester } from "@/types";

export type SemesterRosterStatus = "current" | "past" | "unknown";

function calendarDay(value: string | null | undefined): string | null {
  if (!value) return null;
  const match = /^(\d{4}-\d{2}-\d{2})(?:$|T)/.exec(value);
  if (!match) return null;
  if (value.length > 10 && !Number.isFinite(Date.parse(value))) return null;
  const day = match[1];
  const [year, month, date] = day.split("-").map(Number);
  const parsed = new Date(Date.UTC(year, month - 1, date));
  if (parsed.toISOString().slice(0, 10) !== day) return null;
  return day;
}

export function semesterRosterStatus(
  semester: Pick<Semester, "startDate" | "endDate">,
  now: Date,
): SemesterRosterStatus {
  const start = calendarDay(semester.startDate);
  const end = calendarDay(semester.endDate);
  if (!start || !end || start > end || !Number.isFinite(now.getTime())) return "unknown";

  const parts = new Intl.DateTimeFormat("en-CA", {
    timeZone: "America/Toronto",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(now);
  const values = Object.fromEntries(parts.map(({ type, value }) => [type, value]));
  const today = `${values.year}-${values.month}-${values.day}`;
  return today > end ? "past" : "current";
}
