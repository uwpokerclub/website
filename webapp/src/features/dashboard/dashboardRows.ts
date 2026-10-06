import { CARD_SIZES, CardSize, ColumnSpan, DashboardCardId } from "./dashboardLayout";

export type MediumSpan = 6 | 12;

export type PlacedCard = {
  id: DashboardCardId;
  lead: boolean;
  /** Spans in the 12-column grid: `medium` is the two-up layout, `wide` the full desktop layout. */
  spans: { medium: MediumSpan; wide: ColumnSpan };
};

const COLUMNS = 12;

// Every row is one of these, so a row always spans the full grid and the next row
// starts clean. Two-card rows try an even split first.
const ROW_PATTERNS: Record<number, readonly (readonly ColumnSpan[])[]> = {
  1: [[12]],
  2: [
    [6, 6],
    [8, 4],
    [4, 8],
  ],
  3: [[4, 4, 4]],
};

type RowCard = { id: DashboardCardId; span: ColumnSpan; size: CardSize };

// The lead is promoted to at least half a row, and prefers two thirds, whatever its
// own size — that promotion is what makes role ordering visible.
function sizeFor(id: DashboardCardId, lead: boolean): CardSize {
  const size = CARD_SIZES[id];
  if (!lead) return size;

  return {
    min: Math.max(size.min, 6) as ColumnSpan,
    preferred: Math.max(size.preferred, 8) as ColumnSpan,
  };
}

/** Widens a closed row to the pattern closest to its cards' preferred spans that respects every minimum. */
function fitRow(row: RowCard[]): ColumnSpan[] {
  let best: readonly ColumnSpan[] | undefined;
  let bestCost = Infinity;

  for (const pattern of ROW_PATTERNS[row.length]) {
    if (pattern.some((span, index) => span < row[index].size.min)) continue;

    const cost = pattern.reduce((sum, span, index) => sum + Math.abs(span - row[index].size.preferred), 0);
    if (cost < bestCost) {
      best = pattern;
      bestCost = cost;
    }
  }

  if (!best) {
    throw new Error(`No row pattern fits ${row.map((card) => card.id).join(", ")}`);
  }

  return [...best];
}

/**
 * Packs a role preset into full-width rows of the 12-column desktop grid, in preset
 * order. A card takes its preferred span when the row has room, or its minimum if
 * only that fits; otherwise the row closes and its cards widen to fill it.
 *
 * Cards are never reordered to fill a gap: preset order is reading order, for
 * assistive tech and for the eye alike.
 */
function packWide(preset: readonly DashboardCardId[]): ColumnSpan[] {
  const spans: ColumnSpan[] = [];
  let row: RowCard[] = [];
  let used = 0;

  const close = () => {
    if (row.length > 0) spans.push(...fitRow(row));
    row = [];
    used = 0;
  };

  preset.forEach((id, index) => {
    const size = sizeFor(id, index === 0);
    let span: ColumnSpan | undefined =
      used + size.preferred <= COLUMNS ? size.preferred : used + size.min <= COLUMNS ? size.min : undefined;

    if (span === undefined) {
      close();
      span = size.preferred;
    }

    row.push({ id, span, size });
    used += span;
    if (used === COLUMNS) close();
  });
  close();

  return spans;
}

/**
 * The two-up layout: the lead and any card that prefers a wide row take the full
 * width; everything else pairs at half width, and a card left without a partner
 * takes the row to itself.
 */
function packMedium(preset: readonly DashboardCardId[]): MediumSpan[] {
  const spans: MediumSpan[] = [];
  let pending: number | undefined;

  const closePending = () => {
    if (pending !== undefined) spans[pending] = 12;
    pending = undefined;
  };

  preset.forEach((id, index) => {
    const full = index === 0 || sizeFor(id, false).preferred >= 8;

    if (full) {
      closePending();
      spans[index] = 12;
    } else if (pending === undefined) {
      pending = index;
      spans[index] = 6;
    } else {
      spans[index] = 6;
      pending = undefined;
    }
  });
  closePending();

  return spans;
}

/** Places a role preset on the dashboard grid: preset order, the first card as lead, and its span at each width. */
export function placeDashboardCards(preset: readonly DashboardCardId[]): PlacedCard[] {
  const wide = packWide(preset);
  const medium = packMedium(preset);

  return preset.map((id, index) => ({
    id,
    lead: index === 0,
    spans: { medium: medium[index], wide: wide[index] },
  }));
}
