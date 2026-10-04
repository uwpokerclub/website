/** Inclusive Toronto calendar day within the term, or null before the start/unavailable dates. */
export function termProgress(semester: { startDate: string; endDate: string } | null | undefined, now = new Date()) {
  if (!semester) {
    return null;
  }

  const start = semester.startDate.slice(0, 10);
  const end = semester.endDate.slice(0, 10);
  const today = new Intl.DateTimeFormat("en-CA", {
    timeZone: "America/Toronto",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(now);
  const startInstant = Date.parse(`${start}T00:00:00Z`);
  const endInstant = Date.parse(`${end}T00:00:00Z`);
  const todayInstant = Date.parse(`${today}T00:00:00Z`);
  const total = Math.floor((endInstant - startInstant) / 86_400_000) + 1;
  const elapsed = Math.floor((todayInstant - startInstant) / 86_400_000) + 1;

  if (!Number.isFinite(total) || total <= 0 || elapsed < 1) {
    return null;
  }

  return `day ${Math.min(elapsed, total)} of ${total}`;
}
