# Dashboard Page Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild `/admin/dashboard` as a two-lane, container-query layout with eight cards, each shaped to its own data, every card rendering real or fixture data.

**Architecture:** A pure lane-assignment function turns a role preset into a lead card plus two ordered lanes. The grid is a CSS container, so lanes collapse on measured space beside the sidenav rather than viewport width. Cards share chrome, a validated palette and delta grammar, but no layout — each picks its own form. Cards whose endpoint does not exist yet render from a shared fixture module behind a visible "sample data" marker.

**Tech Stack:** React 19, TypeScript, Vite, CSS Modules, TanStack Query v5, Recharts (new), Jest + Testing Library.

**Spec:** `docs/superpowers/specs/2026-10-01-dashboard-page-design.md`

## Global Constraints

- **Design tokens.** Use `@uwpokerclub/components/tokens.css` variables with **no hex fallbacks**. Write `var(--color-gray-900)`, never `var(--color-gray-900, #212121)`.
- **Categorical palette** (identity): slot 1 `#6f4fae`, slot 2 `#c98f00`, slot 3 `#2a78d6`, slot 4 `#d5578a`. Gold (slot 2) is below 3:1 contrast, so **every gold segment must carry a visible text label**.
- **Ordinal ramp** (magnitude): `#ab97d2` → `#7c5eb6` → `#4b2e83`.
- **Red and green are reserved for `DeltaChip`.** Never a series colour.
- **Card titles are sentence case**, `--font-size-sm`, weight 600. No `text-transform: uppercase` anywhere in dashboard CSS.
- **Light mode only.** `tokens.css` defines no dark palette. Do not add `prefers-color-scheme` blocks.
- **Motion:** only `EventSpotlightCard`'s live pulse, and it must be disabled under `@media (prefers-reduced-motion: reduce)`.
- **Stacked bar segments** are separated by a 2px gap and have `border-radius: 3px`.
- Every card keeps its own query, so one failing endpoint renders one failed card.
- Run `npm run lint` and `npm run build` in `webapp/` before any commit that touches `webapp/`.

---

## Precondition — read before Task 1

The working tree has **uncommitted exploratory changes** from a prior iteration, across
`server/internal/store/dashboard.go`, `server/internal/store/postgres/dashboard.go`,
`webapp/src/features/dashboard/api/dashboardApi.ts`,
`webapp/src/features/dashboard/dashboardLayout.ts`, `webapp/src/pages/Dashboard.tsx`,
`webapp/src/pages/Dashboard.module.css`, the `DashboardGrid` CSS and the
`EngagementRetentionCard` files.

**This plan replaces all of it.** Task 1 discards those uncommitted changes and builds
forward from commit `e114284`. **This is destructive and irreversible — confirm with the
repository owner before running it.**

## File Structure

**Created**

| File | Responsibility |
|---|---|
| `webapp/src/features/dashboard/lanes.ts` | Pure lane-assignment algorithm |
| `webapp/src/features/dashboard/lanes.test.ts` | Its tests |
| `webapp/src/features/dashboard/fixtures/sampleTerm.ts` | One internally-consistent sample term for all eight cards |
| `webapp/src/features/dashboard/charts/chartTheme.ts` | Shared Recharts theme (#439) |
| `webapp/src/features/dashboard/components/StatBar/` | Stacked composition bar + legend |
| `webapp/src/features/dashboard/components/RangeTrack/` | Value-against-historical-band track |
| `webapp/src/features/dashboard/components/EventSpotlightCard/` | Card 1 |
| `webapp/src/features/dashboard/components/QuickActionsCard/` | Card 2 |
| `webapp/src/features/dashboard/components/SignupTimelineCard/` | Card 4 |
| `webapp/src/features/dashboard/components/EventActivityCard/` | Card 5 |
| `webapp/src/features/dashboard/components/LeaderboardCard/` | Card 7 |
| `webapp/src/features/dashboard/components/TrialConversionCard/` | Card 8 |

**Modified**

| File | Change |
|---|---|
| `dashboardLayout.ts` | Eight ids, lane preferences, eight-card presets |
| `dashboardCards.ts` | Register all eight components |
| `api/dashboardApi.ts` | Add spotlight / event-activity / signups / conversion types and fetchers; drop `paidPlayers` |
| `hooks/useDashboardQueries.ts` | One hook per endpoint |
| `pages/Dashboard.tsx` + `.module.css` | Lane rendering, palette custom properties |
| `components/DashboardGrid/` | Two-lane container-query layout |
| `components/DashboardCard/` | Sentence-case title, `sampleData` badge, token cleanup |
| `components/TermAtAGlanceCard/` | Mini-stat grid → two `StatBar`s |
| `components/EngagementRetentionCard/` | Comparison table → `StatBar` + `RangeTrack`s |

---

### Task 1: Reset to committed state and establish the palette

**Files:**
- Modify: `webapp/src/pages/Dashboard.module.css`
- Modify: `webapp/src/features/dashboard/components/DashboardCard/DashboardCard.tsx`
- Modify: `webapp/src/features/dashboard/components/DashboardCard/DashboardCard.module.css`
- Test: `webapp/src/features/dashboard/components/DashboardCard/DashboardCard.test.tsx`

**Interfaces:**
- Produces: `DashboardCardProps` gains `sampleData?: boolean`. Palette custom properties `--dash-cat-1..4` and `--dash-ord-1..3` available to every descendant of `.page`.

- [ ] **Step 1: Discard the uncommitted exploratory changes**

Confirm with the repository owner first — this is irreversible.

```bash
cd /home/adam/projects/uwpokerclub/website-codex/.claude/worktrees/issue-441-engagement-card
git restore server/internal/store/dashboard.go \
  server/internal/store/postgres/dashboard.go \
  webapp/src/features/dashboard/api/dashboardApi.ts \
  webapp/src/features/dashboard/components/DashboardGrid/DashboardGrid.module.css \
  webapp/src/features/dashboard/components/EngagementRetentionCard/ \
  webapp/src/features/dashboard/dashboardLayout.ts \
  webapp/src/pages/Dashboard.module.css \
  webapp/src/pages/Dashboard.tsx
git status --short
```

Expected: no modified files remain.

- [ ] **Step 2: Write the failing test for the sample-data badge**

Append to `webapp/src/features/dashboard/components/DashboardCard/DashboardCard.test.tsx`:

```tsx
it("marks a card rendering fixture data so it cannot be mistaken for real figures", () => {
  render(
    <DashboardCard title="Trial conversion" status="ready" sampleData>
      {() => <p>78</p>}
    </DashboardCard>,
  );

  expect(screen.getByText("Sample data")).toBeInTheDocument();
});

it("shows no sample-data marker by default", () => {
  render(
    <DashboardCard title="Memberships" status="ready">
      {() => <p>967</p>}
    </DashboardCard>,
  );

  expect(screen.queryByText("Sample data")).not.toBeInTheDocument();
});
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `cd webapp && npx jest src/features/dashboard/components/DashboardCard -t "sample"`
Expected: FAIL — `sampleData` is not a known prop and no "Sample data" text renders.

- [ ] **Step 4: Add the prop and badge**

In `DashboardCard.tsx`, add `sampleData?: boolean;` to `DashboardCardProps`, accept it in the destructured parameters with a default of `false`, and render it in the header after the title:

```tsx
<header className={styles.header}>
  <h3 className={styles.title}>{title}</h3>
  {sampleData && <span className={styles.sample}>Sample data</span>}
  {action && <div className={styles.action}>{action}</div>}
</header>
```

- [ ] **Step 5: Restyle the card chrome**

Replace the `.title` rule in `DashboardCard.module.css` and add `.sample`:

```css
.title {
  margin: 0;
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-900);
}

.sample {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--color-warning-dark);
  background: var(--color-warning-light);
  padding: 0.125rem 0.5rem;
  border-radius: var(--border-radius-full);
  margin-right: auto;
}
```

Then remove every hex fallback from the remaining rules in this file — `var(--color-background, #ffffff)` becomes `var(--color-background)`, and so on for `--color-primary`, `--shadow-sm`, `--color-gray-700` and `--font-weight-semibold`.

- [ ] **Step 6: Declare the palette on the page root**

In `Dashboard.module.css`, add to the `.page` rule:

```css
.page {
  --dash-cat-1: #6f4fae;
  --dash-cat-2: #c98f00;
  --dash-cat-3: #2a78d6;
  --dash-cat-4: #d5578a;
  --dash-ord-1: #ab97d2;
  --dash-ord-2: #7c5eb6;
  --dash-ord-3: #4b2e83;
  --dash-band: #e3dcf0;
  --dash-muted: #d5d5d5;
  --dash-muted-soft: #ebebeb;
}
```

Remove every hex fallback from the rest of `Dashboard.module.css`.

- [ ] **Step 7: Run the tests and lint**

Run: `cd webapp && npx jest src/features/dashboard && npm run lint`
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add webapp/src/features/dashboard/components/DashboardCard webapp/src/pages/Dashboard.module.css
git commit -m "feat(#426): sentence-case card titles, sample-data badge, validated palette"
```

---

### Task 2: Eight-card layout model

**Files:**
- Modify: `webapp/src/features/dashboard/dashboardLayout.ts`
- Test: `webapp/src/features/dashboard/dashboardLayout.test.ts`

**Interfaces:**
- Produces: `DashboardCardId` gains `"trialConversion"`; `CardLane = "wide" | "rail"`; `CARD_LANES: Record<DashboardCardId, CardLane>`; `resolveDashboardLayout` returns eight ids.
- Removes: `CARD_TILES` and `DashboardCardTile`.

- [ ] **Step 1: Write the failing test**

Append to `dashboardLayout.test.ts`:

```ts
import { CARD_LANES, CARD_TITLES, resolveDashboardLayout } from "./dashboardLayout";
import { ROLES } from "@/types/roles";

describe("eight-card presets", () => {
  it("gives every role all eight cards exactly once", () => {
    for (const role of [ROLES.PRESIDENT, ROLES.TREASURER, ROLES.EXECUTIVE]) {
      const layout = resolveDashboardLayout(role);
      expect(layout).toHaveLength(8);
      expect(new Set(layout).size).toBe(8);
    }
  });

  it("leads Records with trial conversion and Leadership with engagement", () => {
    expect(resolveDashboardLayout(ROLES.TREASURER)[0]).toBe("trialConversion");
    expect(resolveDashboardLayout(ROLES.PRESIDENT)[0]).toBe("engagement");
    expect(resolveDashboardLayout(ROLES.TOURNAMENT_DIRECTOR)[0]).toBe("spotlight");
  });

  it("routes only the charting cards to the wide lane", () => {
    expect(CARD_LANES.signupTimeline).toBe("wide");
    expect(CARD_LANES.eventActivity).toBe("wide");
    expect(CARD_LANES.leaderboard).toBe("rail");
    expect(CARD_TITLES.trialConversion).toBe("Trial conversion");
  });
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd webapp && npx jest src/features/dashboard/dashboardLayout`
Expected: FAIL — `CARD_LANES` is not exported and `trialConversion` is not a `DashboardCardId`.

- [ ] **Step 3: Rewrite the layout model**

Replace the contents of `dashboardLayout.ts` above `LAYOUTS_BY_ROLE` with:

```ts
import { ROLES } from "@/types/roles";

export type DashboardCardId =
  | "spotlight"
  | "quickActions"
  | "termAtAGlance"
  | "signupTimeline"
  | "eventActivity"
  | "engagement"
  | "leaderboard"
  | "trialConversion";

export const CARD_TITLES: Record<DashboardCardId, string> = {
  spotlight: "Event spotlight",
  quickActions: "Quick actions",
  termAtAGlance: "Memberships",
  signupTimeline: "Signup timeline",
  eventActivity: "Event activity",
  engagement: "Engagement & retention",
  leaderboard: "Leaderboard",
  trialConversion: "Trial conversion",
};

export type CardLane = "wide" | "rail";

// Only cards carrying a chart need the wide lane. Everything else reads fine in
// the rail, and routing it there is what keeps the page gap-free.
export const CARD_LANES: Record<DashboardCardId, CardLane> = {
  spotlight: "rail",
  quickActions: "rail",
  termAtAGlance: "rail",
  signupTimeline: "wide",
  eventActivity: "wide",
  engagement: "wide",
  leaderboard: "rail",
  trialConversion: "rail",
};

const OPS_LAYOUT: readonly DashboardCardId[] = [
  "spotlight",
  "quickActions",
  "eventActivity",
  "leaderboard",
  "termAtAGlance",
  "engagement",
  "signupTimeline",
  "trialConversion",
] as const;

const RECORDS_LAYOUT: readonly DashboardCardId[] = [
  "trialConversion",
  "termAtAGlance",
  "signupTimeline",
  "spotlight",
  "eventActivity",
  "engagement",
  "leaderboard",
  "quickActions",
] as const;

const LEADERSHIP_LAYOUT: readonly DashboardCardId[] = [
  "engagement",
  "eventActivity",
  "termAtAGlance",
  "trialConversion",
  "spotlight",
  "signupTimeline",
  "leaderboard",
  "quickActions",
] as const;
```

Leave `LAYOUTS_BY_ROLE` and `resolveDashboardLayout` unchanged. Delete `DashboardCardTile` and `CARD_TILES`.

- [ ] **Step 4: Run the tests**

Run: `cd webapp && npx jest src/features/dashboard/dashboardLayout`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add webapp/src/features/dashboard/dashboardLayout.ts webapp/src/features/dashboard/dashboardLayout.test.ts
git commit -m "feat(#426): add trial conversion card to the layout model"
```

---

### Task 3: Lane assignment

**Files:**
- Create: `webapp/src/features/dashboard/lanes.ts`
- Test: `webapp/src/features/dashboard/lanes.test.ts`

**Interfaces:**
- Consumes: `DashboardCardId`, `CardLane`, `CARD_LANES` from Task 2.
- Produces: `type LaneAssignment = { lead: DashboardCardId; wide: DashboardCardId[]; rail: DashboardCardId[] }` and `assignLanes(preset: readonly DashboardCardId[]): LaneAssignment`.

- [ ] **Step 1: Write the failing test**

Create `webapp/src/features/dashboard/lanes.test.ts`:

```ts
import { assignLanes } from "./lanes";
import { resolveDashboardLayout } from "./dashboardLayout";
import { ROLES } from "@/types/roles";

describe("assignLanes", () => {
  it("promotes the preset's first card to lead and keeps it out of both lanes", () => {
    const { lead, wide, rail } = assignLanes(resolveDashboardLayout(ROLES.PRESIDENT));

    expect(lead).toBe("engagement");
    expect(wide).not.toContain("engagement");
    expect(rail).not.toContain("engagement");
  });

  it("leads with a rail-preference card when the preset does", () => {
    const { lead, wide } = assignLanes(resolveDashboardLayout(ROLES.TREASURER));

    expect(lead).toBe("trialConversion");
    expect(wide).toEqual(["signupTimeline", "eventActivity", "engagement"]);
  });

  it("preserves relative preset order within each lane", () => {
    const { rail } = assignLanes(resolveDashboardLayout(ROLES.PRESIDENT));

    expect(rail).toEqual(["termAtAGlance", "trialConversion", "spotlight", "leaderboard", "quickActions"]);
  });

  it("accounts for every card exactly once", () => {
    const preset = resolveDashboardLayout(ROLES.EXECUTIVE);
    const { lead, wide, rail } = assignLanes(preset);

    expect([lead, ...wide, ...rail].sort()).toEqual([...preset].sort());
  });
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd webapp && npx jest src/features/dashboard/lanes`
Expected: FAIL — cannot find module `./lanes`.

- [ ] **Step 3: Implement**

Create `webapp/src/features/dashboard/lanes.ts`:

```ts
import { CARD_LANES, DashboardCardId } from "./dashboardLayout";

export type LaneAssignment = {
  lead: DashboardCardId;
  wide: DashboardCardId[];
  rail: DashboardCardId[];
};

/**
 * Turns a role preset into a lead card plus two ordered lanes.
 *
 * The preset's first card is promoted to lead regardless of its own lane
 * preference — that promotion is what makes role ordering visible rather than a
 * reshuffle of identical boxes. Everything after it is routed by preference,
 * relative order intact, so a reader scanning either lane still reads the preset
 * in order.
 */
export function assignLanes(preset: readonly DashboardCardId[]): LaneAssignment {
  const [lead, ...rest] = preset;

  return {
    lead,
    wide: rest.filter((id) => CARD_LANES[id] === "wide"),
    rail: rest.filter((id) => CARD_LANES[id] === "rail"),
  };
}
```

- [ ] **Step 4: Run the tests**

Run: `cd webapp && npx jest src/features/dashboard/lanes`
Expected: PASS, all four.

- [ ] **Step 5: Commit**

```bash
git add webapp/src/features/dashboard/lanes.ts webapp/src/features/dashboard/lanes.test.ts
git commit -m "feat(#426): add dashboard lane assignment"
```

---

### Task 4: Two-lane container-query grid

**Files:**
- Modify: `webapp/src/features/dashboard/components/DashboardGrid/DashboardGrid.tsx`
- Modify: `webapp/src/features/dashboard/components/DashboardGrid/DashboardGrid.module.css`
- Modify: `webapp/src/pages/Dashboard.tsx`
- Test: `webapp/src/pages/Dashboard.test.tsx`

**Interfaces:**
- Consumes: `assignLanes` from Task 3, `DASHBOARD_CARDS` from `dashboardCards.ts`.
- Produces: `DashboardGrid` accepts `lead`, `wide` and `rail` as `ReactNode` props instead of `children`.

- [ ] **Step 1: Write the failing test**

Append to `webapp/src/pages/Dashboard.test.tsx`:

```tsx
it("renders the role's first card in the lead slot", () => {
  renderDashboard({ role: ROLES.PRESIDENT });

  const lead = screen.getByTestId("dashboard-lead");
  expect(lead).toHaveTextContent("Engagement & retention");
});

it("puts chart cards in the wide lane and stat cards in the rail", () => {
  renderDashboard({ role: ROLES.PRESIDENT });

  expect(screen.getByTestId("dashboard-wide")).toHaveTextContent("Event activity");
  expect(screen.getByTestId("dashboard-rail")).toHaveTextContent("Leaderboard");
});
```

Use the existing `renderDashboard` helper in that file; extend it to accept a `role` if it does not already.

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd webapp && npx jest src/pages/Dashboard`
Expected: FAIL — no element with test id `dashboard-lead`.

- [ ] **Step 3: Rewrite the grid component**

Replace `DashboardGrid.tsx`:

```tsx
import { ReactNode } from "react";
import styles from "./DashboardGrid.module.css";

type DashboardGridProps = {
  lead: ReactNode;
  wide: ReactNode;
  rail: ReactNode;
  "data-qa"?: string;
};

export function DashboardGrid({ lead, wide, rail, "data-qa": dataQa }: DashboardGridProps) {
  return (
    <div className={styles.grid} data-qa={dataQa}>
      <div className={styles.wideLane}>
        <div className={styles.lead} data-testid="dashboard-lead">
          {lead}
        </div>
        <div data-testid="dashboard-wide" className={styles.stack}>
          {wide}
        </div>
      </div>
      <div className={styles.rail} data-testid="dashboard-rail">
        {rail}
      </div>
    </div>
  );
}
```

- [ ] **Step 4: Rewrite the grid CSS**

Replace `DashboardGrid.module.css` entirely:

```css
/*
 * The grid is a container, not a media-query consumer. The dashboard sits in a
 * flex lane beside the admin sidenav, so viewport width systematically overstates
 * the room actually available; container width does not.
 */
.grid {
  container-type: inline-size;
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.5rem;
  align-items: start;
}

.wideLane,
.rail,
.stack {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  min-width: 0;
}

.lead {
  min-width: 0;
}

@container (min-width: 64rem) {
  .grid {
    grid-template-columns: 2fr 1fr;
  }
}
```

Note: `container-type: inline-size` and `grid-template-columns` on the same element is
deliberate — the element establishes the container and `@container` queries it from the
inside, which is supported. If a browser target disallows it, wrap `.grid` in a
`container-type` parent and move the query there.

- [ ] **Step 5: Rewrite the page**

Replace the `return` block of `Dashboard.tsx` (keeping the no-semester branch above it):

```tsx
const layout = resolveDashboardLayout(user?.role);
const { lead, wide, rail } = assignLanes(layout);

const renderCard = (cardId: DashboardCardId) => {
  const CardComponent = DASHBOARD_CARDS[cardId];

  if (!CardComponent) {
    return (
      <DashboardCard key={cardId} title={CARD_TITLES[cardId]} status="ready" data-qa={`dashboard-card-${cardId}`}>
        {() => <p className={styles.placeholder}>Coming soon.</p>}
      </DashboardCard>
    );
  }

  return <CardComponent key={cardId} semesterId={currentSemester.id} />;
};

return (
  <div className={styles.page} data-qa="dashboard-page">
    <div className={styles.header}>
      <h1>Dashboard</h1>
      <p className={styles.subtitle}>{currentSemester.name}</p>
    </div>
    <DashboardGrid
      data-qa="dashboard-grid"
      lead={renderCard(lead)}
      wide={wide.map((id) => renderCard(id))}
      rail={rail.map((id) => renderCard(id))}
    />
  </div>
);
```

Import `assignLanes` from `../features/dashboard/lanes` and `DashboardCardId` from
`../features/dashboard/dashboardLayout`.

**Hero scale is CSS, not a prop.** The lead card is larger because `.lead` redefines one
custom property that every card's lead figure already reads — no `isLead` boolean threaded
through eight components that would mostly ignore it. Add to `DashboardGrid.module.css`:

```css
/*
 * Every card's lead figure is sized from --card-figure-size. Only this wrapper
 * overrides it, so promotion to lead is a single cascade step and a card never
 * needs to know which lane it landed in.
 */
.lead {
  --card-figure-size: var(--font-size-5xl);
}
```

Cards declare their figure as `font-size: var(--card-figure-size, var(--font-size-3xl));`.

- [ ] **Step 6: Run the tests, lint and build**

Run: `cd webapp && npx jest src/pages/Dashboard && npm run lint && npm run build`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add webapp/src/features/dashboard/components/DashboardGrid webapp/src/pages webapp/src/features/dashboard/dashboardCards.ts
git commit -m "feat(#426): two-lane container-query dashboard grid"
```

---

### Task 5: API types, query hooks and the sample term

**Files:**
- Modify: `webapp/src/features/dashboard/api/dashboardApi.ts`
- Modify: `webapp/src/features/dashboard/hooks/useDashboardQueries.ts`
- Create: `webapp/src/features/dashboard/fixtures/sampleTerm.ts`
- Test: `webapp/src/features/dashboard/fixtures/sampleTerm.test.ts`

**Interfaces:**
- Produces: `SpotlightEvent`, `EventActivityResponse`, `SignupsResponse`, `ConversionResponse` types; `useSpotlight`, `useEventActivity`, `useSignups`, `useTrialConversion` hooks; and `SAMPLE_TERM` with one key per card.
- `EngagementStats` loses `paidPlayers` and `paidPlayerShare`.

- [ ] **Step 1: Write the failing consistency test**

Create `webapp/src/features/dashboard/fixtures/sampleTerm.test.ts`:

```ts
import { SAMPLE_TERM } from "./sampleTerm";

describe("SAMPLE_TERM", () => {
  it("reads as one coherent term across every card", () => {
    const { memberships, engagement, conversion } = SAMPLE_TERM;

    expect(engagement.current.players).toBeLessThanOrEqual(memberships.current.total);
    expect(conversion.current.players).toBe(engagement.current.players);

    const buckets =
      conversion.current.paid +
      conversion.current.trialSpent +
      conversion.current.trialOpen +
      conversion.current.executive;
    expect(buckets).toBe(conversion.current.players);
  });

  it("keeps the engagement buckets within the player total", () => {
    const { players, playedOnceCount, tenPlusCount } = SAMPLE_TERM.engagement.current;

    expect(playedOnceCount + tenPlusCount).toBeLessThanOrEqual(players);
  });
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd webapp && npx jest src/features/dashboard/fixtures`
Expected: FAIL — cannot find module `./sampleTerm`.

- [ ] **Step 3: Extend the API module**

In `api/dashboardApi.ts`, delete `paidPlayers` and `paidPlayerShare` from `EngagementStats`, then append:

```ts
export interface SpotlightEvent {
  id: number;
  name: string;
  format: string;
  startDate: string;
  state: number;
  entries: number;
  rebuys: number;
}

export interface EventActivityStats {
  eventsRun: number;
  eventsScheduled: number;
  totalEntries: number;
  averageFieldSize: number;
}

export interface EventSeriesPoint {
  id: number;
  name: string;
  startDate: string;
  entries: number;
}

export interface EventActivityResponse {
  current: EventActivityStats & { series: EventSeriesPoint[] };
  comparison: { semester: { id: string; name: string }; averageFieldSize: number } | null;
}

export interface SignupPoint {
  date: string;
  admin: number;
  discord: number;
}

export interface SignupsResponse {
  series: SignupPoint[];
  eventDates: string[];
  dataStartsAt: string | null;
  total: number;
}

export interface ConversionStats {
  players: number;
  paid: number;
  trialSpent: number;
  trialOpen: number;
  executive: number;
}

export interface ConversionResponse {
  current: ConversionStats;
  freeTrialLimit: number;
  comparison: { semester: { id: string; name: string }; stats: ConversionStats } | null;
}

export async function fetchSpotlight(semesterId: string): Promise<SpotlightEvent | null> {
  return apiClient<SpotlightEvent | null>(`v2/semesters/${semesterId}/dashboard/spotlight`);
}

export async function fetchEventActivity(semesterId: string): Promise<EventActivityResponse> {
  return apiClient<EventActivityResponse>(`v2/semesters/${semesterId}/dashboard/events`);
}

export async function fetchSignups(semesterId: string): Promise<SignupsResponse> {
  return apiClient<SignupsResponse>(`v2/semesters/${semesterId}/dashboard/signups`);
}

export async function fetchTrialConversion(semesterId: string): Promise<ConversionResponse> {
  return apiClient<ConversionResponse>(`v2/semesters/${semesterId}/dashboard/conversion`);
}
```

- [ ] **Step 4: Add the query hooks**

In `hooks/useDashboardQueries.ts`, add the four key factories and hooks, following the two
that already exist exactly:

```ts
export const dashboardKeys = {
  all: ["dashboard"] as const,
  memberships: (semesterId: string) => [...dashboardKeys.all, "memberships", semesterId] as const,
  engagement: (semesterId: string) => [...dashboardKeys.all, "engagement", semesterId] as const,
  spotlight: (semesterId: string) => [...dashboardKeys.all, "spotlight", semesterId] as const,
  eventActivity: (semesterId: string) => [...dashboardKeys.all, "eventActivity", semesterId] as const,
  signups: (semesterId: string) => [...dashboardKeys.all, "signups", semesterId] as const,
  conversion: (semesterId: string) => [...dashboardKeys.all, "conversion", semesterId] as const,
};

export function useSpotlight(semesterId: string) {
  return useQuery({
    queryKey: dashboardKeys.spotlight(semesterId),
    queryFn: () => fetchSpotlight(semesterId),
    enabled: !!semesterId,
  });
}

export function useEventActivity(semesterId: string) {
  return useQuery({
    queryKey: dashboardKeys.eventActivity(semesterId),
    queryFn: () => fetchEventActivity(semesterId),
    enabled: !!semesterId,
  });
}

export function useSignups(semesterId: string) {
  return useQuery({
    queryKey: dashboardKeys.signups(semesterId),
    queryFn: () => fetchSignups(semesterId),
    enabled: !!semesterId,
  });
}

export function useTrialConversion(semesterId: string) {
  return useQuery({
    queryKey: dashboardKeys.conversion(semesterId),
    queryFn: () => fetchTrialConversion(semesterId),
    enabled: !!semesterId,
  });
}
```

Extend the import from `../api/dashboardApi` to include the four new fetchers.

- [ ] **Step 5: Write the sample term**

Create `webapp/src/features/dashboard/fixtures/sampleTerm.ts`:

```ts
import type {
  ConversionResponse,
  EngagementDashboardResponse,
  EventActivityResponse,
  MembershipsDashboardResponse,
  SignupsResponse,
  SpotlightEvent,
} from "../api/dashboardApi";

/**
 * One internally-consistent sample term, used as placeholder data for cards whose
 * endpoint does not exist yet and as the shared fixture for card tests. The figures
 * agree across cards on purpose — 967 memberships, 824 of whom played, 712 of those
 * paid — so a page rendered entirely from fixtures still reads as one real term.
 *
 * Every card rendering this data must pass `sampleData` to DashboardCard.
 */
const FALL_2025 = { id: "0f2d6a4e-0000-4000-8000-000000000001", name: "Fall 2025" };

export const SAMPLE_TERM = {
  spotlight: {
    id: 412,
    name: "Thursday Night NLH",
    format: "No-limit hold'em · 10k stacks",
    startDate: "2026-11-12T23:04:00Z",
    state: 0,
    entries: 47,
    rebuys: 12,
  } satisfies SpotlightEvent,

  memberships: {
    current: { total: 967, paid: 712, unpaid: 189, discounted: 34, executive: 32, new: 561, returning: 406 },
    comparison: {
      semester: FALL_2025,
      stats: { total: 891, paid: 654, unpaid: 182, discounted: 29, executive: 26, new: 498, returning: 393 },
    },
  } satisfies MembershipsDashboardResponse,

  engagement: {
    current: { players: 824, medianEventsAttended: 2, playedOnceCount: 313, playedOnceShare: 0.38, tenPlusCount: 41 },
    comparison: {
      semester: FALL_2025,
      stats: { players: 891, medianEventsAttended: 3, playedOnceCount: 303, playedOnceShare: 0.34, tenPlusCount: 54 },
    },
  } satisfies EngagementDashboardResponse,

  conversion: {
    current: { players: 824, paid: 712, trialSpent: 78, trialOpen: 14, executive: 20 },
    freeTrialLimit: 4,
    comparison: {
      semester: FALL_2025,
      stats: { players: 891, paid: 788, trialSpent: 69, trialOpen: 11, executive: 23 },
    },
  } satisfies ConversionResponse,

  eventActivity: {
    current: {
      eventsRun: 11,
      eventsScheduled: 14,
      totalEntries: 486,
      averageFieldSize: 44.2,
      series: [
        { id: 1, name: "Week 1", startDate: "2026-09-12T23:00:00Z", entries: 31 },
        { id: 2, name: "Week 2", startDate: "2026-09-19T23:00:00Z", entries: 52 },
        { id: 3, name: "Week 3", startDate: "2026-09-26T23:00:00Z", entries: 41 },
        { id: 4, name: "Week 4", startDate: "2026-10-03T23:00:00Z", entries: 63 },
        { id: 5, name: "Week 5", startDate: "2026-10-10T23:00:00Z", entries: 47 },
        { id: 6, name: "Week 6", startDate: "2026-10-17T23:00:00Z", entries: 68 },
        { id: 7, name: "Week 7", startDate: "2026-10-24T23:00:00Z", entries: 38 },
        { id: 8, name: "Week 8", startDate: "2026-10-31T23:00:00Z", entries: 56 },
        { id: 9, name: "Week 9", startDate: "2026-11-07T23:00:00Z", entries: 27 },
        { id: 10, name: "Week 10", startDate: "2026-11-14T23:00:00Z", entries: 59 },
        { id: 11, name: "Week 11", startDate: "2026-11-21T23:00:00Z", entries: 44 },
      ],
    },
    comparison: { semester: FALL_2025, averageFieldSize: 49.1 },
  } satisfies EventActivityResponse,

  signups: {
    total: 967,
    dataStartsAt: "2026-09-01",
    eventDates: ["2026-09-12", "2026-10-03", "2026-10-17", "2026-11-07", "2026-11-21"],
    series: [
      { date: "2026-09-01", admin: 4, discord: 2 },
      { date: "2026-09-05", admin: 61, discord: 20 },
      { date: "2026-09-12", admin: 38, discord: 31 },
      { date: "2026-09-19", admin: 22, discord: 18 },
      { date: "2026-09-26", admin: 14, discord: 11 },
      { date: "2026-10-03", admin: 29, discord: 24 },
      { date: "2026-10-10", admin: 12, discord: 9 },
      { date: "2026-10-17", admin: 19, discord: 16 },
      { date: "2026-10-24", admin: 8, discord: 6 },
      { date: "2026-11-07", admin: 15, discord: 12 },
      { date: "2026-11-21", admin: 6, discord: 4 },
    ],
  } satisfies SignupsResponse,
};
```

- [ ] **Step 6: Run the tests and lint**

Run: `cd webapp && npx jest src/features/dashboard && npm run lint`
Expected: PASS. The existing `EngagementRetentionCard` tests will fail on the removed
`paidPlayers` field — that card is rewritten in Task 8. Temporarily delete the
`paidPlayers` and `paidPlayerShare` lines from its test's `stats()` helper so the suite is
green at this commit.

- [ ] **Step 7: Commit**

```bash
git add webapp/src/features/dashboard/api webapp/src/features/dashboard/hooks webapp/src/features/dashboard/fixtures webapp/src/features/dashboard/components/EngagementRetentionCard
git commit -m "feat(#426): add dashboard API types, hooks and sample term fixtures"
```

---

### Task 6: StatBar primitive

**Files:**
- Create: `webapp/src/features/dashboard/components/StatBar/StatBar.tsx`
- Create: `webapp/src/features/dashboard/components/StatBar/StatBar.module.css`
- Create: `webapp/src/features/dashboard/components/StatBar/index.ts`
- Test: `webapp/src/features/dashboard/components/StatBar/StatBar.test.tsx`

**Interfaces:**
- Produces: `type StatBarSegment = { key: string; label: string; value: number; color: string }` and `StatBar({ segments, ariaLabel }: { segments: StatBarSegment[]; ariaLabel: string })`.

- [ ] **Step 1: Write the failing test**

```tsx
/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";
import { StatBar } from "./StatBar";

const segments = [
  { key: "paid", label: "paid", value: 712, color: "var(--dash-cat-1)" },
  { key: "unpaid", label: "unpaid", value: 189, color: "var(--dash-cat-2)" },
];

describe("StatBar", () => {
  it("labels every segment with its value, so no segment depends on colour alone", () => {
    render(<StatBar segments={segments} ariaLabel="Membership buckets" />);

    expect(screen.getByText("712")).toBeInTheDocument();
    expect(screen.getByText("paid")).toBeInTheDocument();
    expect(screen.getByText("189")).toBeInTheDocument();
  });

  it("exposes the breakdown to assistive technology as a list", () => {
    render(<StatBar segments={segments} ariaLabel="Membership buckets" />);

    expect(screen.getByRole("list", { name: "Membership buckets" })).toBeInTheDocument();
    expect(screen.getAllByRole("listitem")).toHaveLength(2);
  });

  it("drops zero-value segments from the bar but keeps them in the legend", () => {
    render(
      <StatBar
        segments={[...segments, { key: "exec", label: "exec", value: 0, color: "var(--dash-cat-3)" }]}
        ariaLabel="Membership buckets"
      />,
    );

    expect(screen.getAllByRole("listitem")).toHaveLength(3);
    expect(screen.getByTestId("statbar-track").children).toHaveLength(2);
  });
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd webapp && npx jest src/features/dashboard/components/StatBar`
Expected: FAIL — cannot find module `./StatBar`.

- [ ] **Step 3: Implement**

`StatBar.tsx`:

```tsx
import styles from "./StatBar.module.css";

export type StatBarSegment = {
  key: string;
  label: string;
  value: number;
  color: string;
};

type StatBarProps = {
  segments: StatBarSegment[];
  ariaLabel: string;
};

/**
 * A horizontal part-to-whole bar with a labelled legend. Segments are sized by
 * flex-grow, so the bar is fluid at any width with no measurement. Every segment
 * carries a visible value in the legend — identity is never colour alone, and the
 * gold slot is below 3:1 contrast so its label is mandatory rather than optional.
 */
export function StatBar({ segments, ariaLabel }: StatBarProps) {
  const visible = segments.filter((segment) => segment.value > 0);

  return (
    <div className={styles.container}>
      <div className={styles.track} data-testid="statbar-track" aria-hidden="true">
        {visible.map((segment) => (
          <span key={segment.key} style={{ flexGrow: segment.value, background: segment.color }} />
        ))}
      </div>
      <ul className={styles.legend} aria-label={ariaLabel}>
        {segments.map((segment) => (
          <li key={segment.key}>
            <span className={styles.swatch} style={{ background: segment.color }} aria-hidden="true" />
            <b>{segment.value.toLocaleString("en-CA")}</b> {segment.label}
          </li>
        ))}
      </ul>
    </div>
  );
}
```

`StatBar.module.css`:

```css
.container {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.track {
  display: flex;
  gap: 2px;
  height: 1.375rem;
}

.track > span {
  display: block;
  border-radius: 3px;
  min-width: 2px;
}

.legend {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem 0.875rem;
  margin: 0;
  padding: 0;
  list-style: none;
  font-size: var(--font-size-xs);
  color: var(--color-gray-700);
}

.legend b {
  color: var(--color-gray-900);
  font-weight: var(--font-weight-semibold);
}

.swatch {
  display: inline-block;
  width: 0.5625rem;
  height: 0.5625rem;
  border-radius: 2px;
  margin-right: 0.3125rem;
}

/* In a narrow tile the legend reads better stacked than wrapped mid-item. */
@container (max-width: 18rem) {
  .legend {
    flex-direction: column;
    gap: 0.375rem;
  }
}
```

`index.ts`:

```ts
export { StatBar } from "./StatBar";
export type { StatBarSegment } from "./StatBar";
```

- [ ] **Step 4: Run the tests**

Run: `cd webapp && npx jest src/features/dashboard/components/StatBar`
Expected: PASS, all three.

- [ ] **Step 5: Commit**

```bash
git add webapp/src/features/dashboard/components/StatBar
git commit -m "feat(#426): add StatBar composition bar primitive"
```

---

### Task 7: RangeTrack primitive

**Files:**
- Create: `webapp/src/features/dashboard/components/RangeTrack/RangeTrack.tsx`
- Create: `webapp/src/features/dashboard/components/RangeTrack/RangeTrack.module.css`
- Create: `webapp/src/features/dashboard/components/RangeTrack/index.ts`
- Test: `webapp/src/features/dashboard/components/RangeTrack/RangeTrack.test.tsx`

**Interfaces:**
- Produces: `RangeTrack(props: RangeTrackProps)` where

```ts
type RangeTrackProps = {
  label: string;
  value: number;
  display: string;
  min: number;
  max: number;
  bandLow: number;
  bandHigh: number;
  comparison?: { value: number; label: string };
};
```

- [ ] **Step 1: Write the failing test**

```tsx
/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";
import { RangeTrack } from "./RangeTrack";

const base = {
  label: "Median events played",
  display: "2",
  min: 0,
  max: 5,
  bandLow: 2,
  bandHigh: 3,
};

describe("RangeTrack", () => {
  it("calls a value inside the historical band normal", () => {
    render(<RangeTrack {...base} value={2} />);

    expect(screen.getByText("normal")).toBeInTheDocument();
  });

  it("calls a value under the band low, and over it high", () => {
    const { rerender } = render(<RangeTrack {...base} value={1} />);
    expect(screen.getByText("below usual")).toBeInTheDocument();

    rerender(<RangeTrack {...base} value={4} display="4" />);
    expect(screen.getByText("above usual")).toBeInTheDocument();
  });

  it("spells the band out for assistive technology rather than relying on position", () => {
    render(<RangeTrack {...base} value={2} comparison={{ value: 3, label: "Fall 2025" }} />);

    expect(
      screen.getByText("2, normal for this club — the usual range is 2 to 3. Fall 2025 was 3."),
    ).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd webapp && npx jest src/features/dashboard/components/RangeTrack`
Expected: FAIL — cannot find module `./RangeTrack`.

- [ ] **Step 3: Implement**

`RangeTrack.tsx`:

```tsx
import styles from "./RangeTrack.module.css";

type RangeTrackProps = {
  label: string;
  value: number;
  display: string;
  min: number;
  max: number;
  bandLow: number;
  bandHigh: number;
  comparison?: { value: number; label: string };
};

type Verdict = "below usual" | "normal" | "above usual";

function verdictFor(value: number, bandLow: number, bandHigh: number): Verdict {
  if (value < bandLow) return "below usual";
  if (value > bandHigh) return "above usual";
  return "normal";
}

const pct = (value: number, min: number, max: number) =>
  `${Math.min(100, Math.max(0, ((value - min) / (max - min)) * 100))}%`;

/**
 * One figure placed against the range it normally falls in. This exists because a
 * bare "median 2 events, down 1" reads as a collapse during an entirely ordinary
 * term; position against the band is what makes it legible without a sentence.
 */
export function RangeTrack({ label, value, display, min, max, bandLow, bandHigh, comparison }: RangeTrackProps) {
  const verdict = verdictFor(value, bandLow, bandHigh);
  const spoken =
    `${display}, ${verdict === "normal" ? "normal for this club" : verdict} — ` +
    `the usual range is ${bandLow} to ${bandHigh}.` +
    (comparison ? ` ${comparison.label} was ${comparison.value}.` : "");

  return (
    <div className={styles.row}>
      <span className={styles.label}>{label}</span>
      <span className={styles.value}>{display}</span>
      <span className={styles.track} aria-hidden="true">
        <i className={styles.band} style={{ left: pct(bandLow, min, max), right: `calc(100% - ${pct(bandHigh, min, max)})` }} />
        {comparison && <i className={styles.previous} style={{ left: pct(comparison.value, min, max) }} />}
        <i className={styles.mark} style={{ left: pct(value, min, max) }} />
      </span>
      <span className={styles.verdict} data-verdict={verdict} aria-hidden="true">
        {verdict}
      </span>
      <span className={styles.srOnly}>{spoken}</span>
    </div>
  );
}
```

`RangeTrack.module.css`:

```css
.row {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.5rem 0;
  border-top: 1px solid var(--color-gray-200);
}

.label {
  font-size: var(--font-size-xs);
  color: var(--color-gray-900);
  flex: 1 1 7rem;
  min-width: 0;
}

.value {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-900);
  flex: 0 0 2.75rem;
}

.track {
  position: relative;
  flex: 1 1 4rem;
  height: 0.4375rem;
  background: var(--color-background-tertiary);
  border-radius: var(--border-radius-full);
  min-width: 3.75rem;
}

.band {
  position: absolute;
  top: 0;
  bottom: 0;
  background: var(--dash-band);
  border-radius: var(--border-radius-full);
}

.mark {
  position: absolute;
  top: -0.25rem;
  width: 3px;
  height: 0.9375rem;
  background: var(--dash-ord-3);
  border-radius: 2px;
}

.previous {
  position: absolute;
  top: -0.0625rem;
  width: 3px;
  height: 0.5625rem;
  background: var(--color-gray-400);
  border-radius: 2px;
}

.verdict {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  padding: 0.125rem 0.5rem;
  border-radius: var(--border-radius-full);
  background: var(--color-secondary-light);
  color: var(--color-secondary-dark);
  white-space: nowrap;
  flex-shrink: 0;
}

.srOnly {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}

/* In a narrow tile the track has nowhere useful to go; drop it and keep the words. */
@container (max-width: 20rem) {
  .track {
    display: none;
  }
}
```

`index.ts`:

```ts
export { RangeTrack } from "./RangeTrack";
```

- [ ] **Step 4: Run the tests**

Run: `cd webapp && npx jest src/features/dashboard/components/RangeTrack`
Expected: PASS, all three.

- [ ] **Step 5: Commit**

```bash
git add webapp/src/features/dashboard/components/RangeTrack
git commit -m "feat(#426): add RangeTrack historical-band primitive"
```

---

### Task 8: Rewrite the Engagement & Retention card

**Files:**
- Modify: `webapp/src/features/dashboard/components/EngagementRetentionCard/EngagementRetentionCard.tsx`
- Modify: `webapp/src/features/dashboard/components/EngagementRetentionCard/EngagementRetentionCard.module.css`
- Test: `webapp/src/features/dashboard/components/EngagementRetentionCard/EngagementRetentionCard.test.tsx`

**Interfaces:**
- Consumes: `StatBar` (Task 6), `RangeTrack` (Task 7), `SAMPLE_TERM` (Task 5).
- Produces: `EngagementRetentionCard({ semesterId }: { semesterId: string })`.

- [ ] **Step 1: Replace the test file**

```tsx
/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";
import { useEngagementDashboard } from "../../hooks/useDashboardQueries";
import { useMembershipsDashboard } from "../../hooks/useDashboardQueries";
import { EngagementRetentionCard } from "./EngagementRetentionCard";
import { SAMPLE_TERM } from "../../fixtures/sampleTerm";

jest.mock("../../hooks/useDashboardQueries", () => ({
  useEngagementDashboard: jest.fn(),
  useMembershipsDashboard: jest.fn(),
}));

const mockedEngagement = useEngagementDashboard as jest.Mock;
const mockedMemberships = useMembershipsDashboard as jest.Mock;

function ready() {
  mockedEngagement.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.engagement });
  mockedMemberships.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.memberships });
}

describe("EngagementRetentionCard", () => {
  afterEach(() => jest.clearAllMocks());

  it("leads with distinct players against total memberships", () => {
    ready();
    render(<EngagementRetentionCard semesterId="s1" />);

    expect(screen.getByText("824")).toBeInTheDocument();
    expect(screen.getByText(/of 967 members played at least one event/)).toBeInTheDocument();
  });

  it("splits the population into three attendance buckets that sum to the total", () => {
    ready();
    render(<EngagementRetentionCard semesterId="s1" />);

    expect(screen.getByText("313")).toBeInTheDocument();
    expect(screen.getByText("470")).toBeInTheDocument();
    expect(screen.getByText("41")).toBeInTheDocument();
  });

  it("places both rates against their historical band rather than showing a bare delta", () => {
    ready();
    render(<EngagementRetentionCard semesterId="s1" />);

    expect(screen.getAllByText("normal")).toHaveLength(2);
  });

  it("renders a start-of-term state rather than zeroes when nobody has played", () => {
    mockedEngagement.mockReturnValue({
      isLoading: false,
      isError: false,
      data: { current: { players: 0, medianEventsAttended: 0, playedOnceCount: 0, playedOnceShare: 0, tenPlusCount: 0 }, comparison: null },
    });
    mockedMemberships.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.memberships });
    render(<EngagementRetentionCard semesterId="s1" />);

    expect(screen.getByText("No event entries recorded yet this term.")).toBeInTheDocument();
  });

  it("renders its own error state", () => {
    mockedEngagement.mockReturnValue({ isLoading: false, isError: true, data: undefined, refetch: jest.fn() });
    mockedMemberships.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.memberships });
    render(<EngagementRetentionCard semesterId="s1" />);

    expect(screen.getByRole("button", { name: /retry/i })).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd webapp && npx jest src/features/dashboard/components/EngagementRetentionCard`
Expected: FAIL — the card still renders the comparison table.

- [ ] **Step 3: Rewrite the component**

```tsx
import { DashboardCard } from "../DashboardCard";
import { StatBar } from "../StatBar";
import { RangeTrack } from "../RangeTrack";
import { CARD_TITLES } from "../../dashboardLayout";
import { useEngagementDashboard, useMembershipsDashboard } from "../../hooks/useDashboardQueries";
import type { EngagementDashboardResponse, MembershipsDashboardResponse } from "../../api/dashboardApi";
import styles from "./EngagementRetentionCard.module.css";

/**
 * Historical ranges from the six-term points-system analysis, recorded in
 * docs/superpowers/specs/2026-10-01-dashboard-page-design.md. These are constants
 * because the API returns this term and one comparison term only — six terms of
 * range does not exist in the response. They will go stale; see the spec's
 * "Hard-coded constants" section for why that trade was made anyway.
 */
const MEDIAN_EVENTS_BAND = { low: 2, high: 3, min: 0, max: 6 };
const PLAYED_ONCE_BAND = { low: 33, high: 44, min: 0, max: 100 };

type Props = { semesterId: string };

export function EngagementRetentionCard({ semesterId }: Props) {
  const engagement = useEngagementDashboard(semesterId);
  const memberships = useMembershipsDashboard(semesterId);

  const status = engagement.isError
    ? "error"
    : engagement.isLoading || !engagement.data
      ? "loading"
      : engagement.data.current.players === 0
        ? "empty"
        : "ready";

  return (
    <DashboardCard
      title={CARD_TITLES.engagement}
      status={status}
      onRetry={() => engagement.refetch()}
      emptyMessage="No event entries recorded yet this term."
      data-qa="engagement-retention-card"
    >
      {() => (
        <Body
          data={engagement.data as EngagementDashboardResponse}
          memberships={memberships.data as MembershipsDashboardResponse | undefined}
        />
      )}
    </DashboardCard>
  );
}

function Body({ data, memberships }: { data: EngagementDashboardResponse; memberships?: MembershipsDashboardResponse }) {
  const { current, comparison } = data;
  const cameBack = current.players - current.playedOnceCount - current.tenPlusCount;
  const total = memberships?.current.total;

  return (
    <div className={styles.container}>
      <p className={styles.lead}>
        <span className={styles.figure}>{current.players.toLocaleString("en-CA")}</span>
        <span className={styles.caption}>
          {total ? `of ${total.toLocaleString("en-CA")} members played at least one event` : "members played at least one event"}
        </span>
      </p>

      <StatBar
        ariaLabel="Players by events attended"
        segments={[
          { key: "once", label: "played once", value: current.playedOnceCount, color: "var(--dash-ord-1)" },
          { key: "some", label: "played 2–9", value: cameBack, color: "var(--dash-ord-2)" },
          { key: "regular", label: "regulars, 10+", value: current.tenPlusCount, color: "var(--dash-ord-3)" },
        ]}
      />

      <div className={styles.tracks}>
        <RangeTrack
          label="Median events played"
          value={current.medianEventsAttended}
          display={`${current.medianEventsAttended}`}
          min={MEDIAN_EVENTS_BAND.min}
          max={MEDIAN_EVENTS_BAND.max}
          bandLow={MEDIAN_EVENTS_BAND.low}
          bandHigh={MEDIAN_EVENTS_BAND.high}
          comparison={
            comparison
              ? { value: comparison.stats.medianEventsAttended, label: comparison.semester.name }
              : undefined
          }
        />
        <RangeTrack
          label="Played exactly once"
          value={Math.round(current.playedOnceShare * 100)}
          display={`${Math.round(current.playedOnceShare * 100)}%`}
          min={PLAYED_ONCE_BAND.min}
          max={PLAYED_ONCE_BAND.max}
          bandLow={PLAYED_ONCE_BAND.low}
          bandHigh={PLAYED_ONCE_BAND.high}
          comparison={
            comparison
              ? { value: Math.round(comparison.stats.playedOnceShare * 100), label: comparison.semester.name }
              : undefined
          }
        />
      </div>

      <p className={styles.foot}>
        Members who entered at least one event this term. The shaded band is the range across six analysed terms
        {comparison ? `; the grey tick is ${comparison.semester.name}.` : "."}
      </p>
    </div>
  );
}
```

- [ ] **Step 4: Replace the stylesheet**

```css
.container {
  container-type: inline-size;
  display: flex;
  flex-direction: column;
  gap: 0.875rem;
}

.lead {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 0.6875rem;
  margin: 0;
}

.figure {
  font-size: var(--card-figure-size, var(--font-size-4xl));
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-900);
  line-height: 1;
  letter-spacing: -0.02em;
}

.caption {
  font-size: var(--font-size-sm);
  color: var(--color-gray-700);
}

.tracks {
  display: flex;
  flex-direction: column;
}

.foot {
  margin: 0;
  padding-top: 0.625rem;
  border-top: 1px solid var(--color-gray-200);
  font-size: var(--font-size-xs);
  color: var(--color-gray-700);
  line-height: 1.5;
}
```

- [ ] **Step 5: Run the tests, lint and build**

Run: `cd webapp && npx jest src/features/dashboard && npm run lint && npm run build`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add webapp/src/features/dashboard/components/EngagementRetentionCard
git commit -m "feat(#441): rebuild engagement card as a distribution with historical bands"
```

---

### Task 9: Rewrite the Memberships card

**Files:**
- Modify: `webapp/src/features/dashboard/components/TermAtAGlanceCard/TermAtAGlanceCard.tsx`
- Modify: `webapp/src/features/dashboard/components/TermAtAGlanceCard/TermAtAGlanceCard.module.css`
- Test: `webapp/src/features/dashboard/components/TermAtAGlanceCard/TermAtAGlanceCard.test.tsx`

**Interfaces:**
- Consumes: `StatBar` (Task 6), `DeltaChip` (unchanged), `SAMPLE_TERM` (Task 5).

- [ ] **Step 1: Write the failing test**

Replace the body of the existing describe block with:

```tsx
it("shows the total with its delta and both compositions as bars", () => {
  mockedUseMembershipsDashboard.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.memberships });
  render(<TermAtAGlanceCard semesterId="s1" />);

  expect(screen.getByText("967")).toBeInTheDocument();
  expect(screen.getByRole("list", { name: "Memberships by payment status" })).toBeInTheDocument();
  expect(screen.getByRole("list", { name: "Memberships by prior membership" })).toBeInTheDocument();
  expect(screen.getByText("712")).toBeInTheDocument();
  expect(screen.getByText("561")).toBeInTheDocument();
});

it("states there is nothing to compare against rather than showing a zero delta", () => {
  mockedUseMembershipsDashboard.mockReturnValue({
    isLoading: false,
    isError: false,
    data: { ...SAMPLE_TERM.memberships, comparison: null },
  });
  render(<TermAtAGlanceCard semesterId="s1" />);

  expect(screen.getByText("No comparable term to compare against yet.")).toBeInTheDocument();
  expect(screen.queryByTestId("delta-chip")).not.toBeInTheDocument();
});
```

Import `SAMPLE_TERM` from `../../fixtures/sampleTerm`.

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd webapp && npx jest src/features/dashboard/components/TermAtAGlanceCard`
Expected: FAIL — no element with role `list` named "Memberships by payment status".

- [ ] **Step 3: Replace the body component**

Replace `TermAtAGlanceBody` and delete `MiniStat`, `BUCKET_STATS`, `RETENTION_STATS` and
`MiniStatConfig`:

```tsx
function TermAtAGlanceBody({ data }: { data: MembershipsDashboardResponse }) {
  const { current, comparison } = data;

  return (
    <div className={styles.container}>
      <div className={styles.headline}>
        <span className={styles.total}>{current.total.toLocaleString("en-CA")}</span>
        {comparison ? (
          <DeltaChip
            current={current.total}
            comparison={comparison.stats.total}
            comparisonLabel={comparison.semester.name}
            sentiment="positive-is-good"
          />
        ) : (
          <p className={styles.noComparisonNote}>No comparable term to compare against yet.</p>
        )}
      </div>

      <StatBar
        ariaLabel="Memberships by payment status"
        segments={[
          { key: "paid", label: "paid", value: current.paid, color: "var(--dash-cat-1)" },
          { key: "unpaid", label: "unpaid", value: current.unpaid, color: "var(--dash-cat-2)" },
          { key: "discounted", label: "discounted", value: current.discounted, color: "var(--dash-cat-3)" },
          { key: "executive", label: "exec", value: current.executive, color: "var(--dash-cat-4)" },
        ]}
      />

      <StatBar
        ariaLabel="Memberships by prior membership"
        segments={[
          { key: "new", label: "first-ever term", value: current.new, color: "var(--dash-ord-3)" },
          { key: "returning", label: "returning", value: current.returning, color: "var(--dash-ord-1)" },
        ]}
      />
    </div>
  );
}
```

Add `import { StatBar } from "../StatBar";` and remove the now-unused `DeltaSentiment` and
`MembershipStats` imports if TypeScript reports them unused.

- [ ] **Step 4: Trim the stylesheet**

Replace `TermAtAGlanceCard.module.css`:

```css
.container {
  container-type: inline-size;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.headline {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 0.5625rem;
}

.total {
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-900);
  line-height: 1.1;
}

.noComparisonNote {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--color-gray-700);
}
```

- [ ] **Step 5: Run the tests, lint and build**

Run: `cd webapp && npx jest src/features/dashboard && npm run lint && npm run build`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add webapp/src/features/dashboard/components/TermAtAGlanceCard
git commit -m "feat(#426): show memberships as composition bars"
```

---

### Task 10: Trial Conversion card

**Files:**
- Create: `webapp/src/features/dashboard/components/TrialConversionCard/TrialConversionCard.tsx`
- Create: `webapp/src/features/dashboard/components/TrialConversionCard/TrialConversionCard.module.css`
- Create: `webapp/src/features/dashboard/components/TrialConversionCard/index.ts`
- Test: `webapp/src/features/dashboard/components/TrialConversionCard/TrialConversionCard.test.tsx`
- Modify: `webapp/src/features/dashboard/components/index.ts`, `dashboardCards.ts`

**Interfaces:**
- Consumes: `useTrialConversion` (Task 5), `StatBar` (Task 6), `SAMPLE_TERM` (Task 5).

- [ ] **Step 1: Write the failing test**

```tsx
/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";
import { useTrialConversion } from "../../hooks/useDashboardQueries";
import { TrialConversionCard } from "./TrialConversionCard";
import { SAMPLE_TERM } from "../../fixtures/sampleTerm";

jest.mock("../../hooks/useDashboardQueries", () => ({ useTrialConversion: jest.fn() }));
const mocked = useTrialConversion as jest.Mock;

describe("TrialConversionCard", () => {
  afterEach(() => jest.clearAllMocks());

  it("leads with the number who spent the trial and never paid", () => {
    mocked.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.conversion });
    render(<TrialConversionCard semesterId="s1" />);

    expect(screen.getByText("78")).toBeInTheDocument();
    expect(screen.getByText(/spent all 4 free entries and never bought a membership/)).toBeInTheDocument();
  });

  it("says the trial is off rather than drawing a bar of zeroes", () => {
    mocked.mockReturnValue({
      isLoading: false,
      isError: false,
      data: { ...SAMPLE_TERM.conversion, freeTrialLimit: 0 },
    });
    render(<TrialConversionCard semesterId="s1" />);

    expect(screen.getByText("Free trial is not enabled this term.")).toBeInTheDocument();
    expect(screen.queryByRole("list")).not.toBeInTheDocument();
  });

  it("marks itself as sample data while the endpoint does not exist", () => {
    mocked.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.conversion });
    render(<TrialConversionCard semesterId="s1" />);

    expect(screen.getByText("Sample data")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd webapp && npx jest src/features/dashboard/components/TrialConversionCard`
Expected: FAIL — cannot find module `./TrialConversionCard`.

- [ ] **Step 3: Implement**

```tsx
import { DashboardCard } from "../DashboardCard";
import { StatBar } from "../StatBar";
import { CARD_TITLES } from "../../dashboardLayout";
import { useTrialConversion } from "../../hooks/useDashboardQueries";
import { SAMPLE_TERM } from "../../fixtures/sampleTerm";
import type { ConversionResponse } from "../../api/dashboardApi";
import styles from "./TrialConversionCard.module.css";

/**
 * Flip to false once GET …/dashboard/conversion exists. Until then the card renders
 * the sample term behind a visible marker rather than a permanent error, so the page
 * is designed and reviewable before its endpoint lands.
 */
const USES_SAMPLE_DATA = true;

type Props = { semesterId: string };

export function TrialConversionCard({ semesterId }: Props) {
  const { data, isLoading, isError, refetch } = useTrialConversion(semesterId);
  const resolved = USES_SAMPLE_DATA ? SAMPLE_TERM.conversion : data;

  const status = USES_SAMPLE_DATA
    ? "ready"
    : isError
      ? "error"
      : isLoading || !resolved
        ? "loading"
        : resolved.current.players === 0
          ? "empty"
          : "ready";

  return (
    <DashboardCard
      title={CARD_TITLES.trialConversion}
      status={status}
      sampleData={USES_SAMPLE_DATA}
      onRetry={() => refetch()}
      emptyMessage="No event entries recorded yet this term."
      data-qa="trial-conversion-card"
    >
      {() => <Body data={resolved as ConversionResponse} />}
    </DashboardCard>
  );
}

function Body({ data }: { data: ConversionResponse }) {
  const { current, freeTrialLimit } = data;

  if (freeTrialLimit === 0) {
    return <p className={styles.disabled}>Free trial is not enabled this term.</p>;
  }

  return (
    <div className={styles.container}>
      <p className={styles.lead}>
        <span className={styles.figure}>{current.trialSpent.toLocaleString("en-CA")}</span>
      </p>
      <p className={styles.caption}>
        spent all {freeTrialLimit} free entries and never bought a membership
      </p>

      <StatBar
        ariaLabel="Players by membership status"
        segments={[
          { key: "paid", label: "paid", value: current.paid, color: "var(--dash-ord-3)" },
          { key: "spent", label: "trial spent, unpaid", value: current.trialSpent, color: "var(--dash-cat-2)" },
          { key: "open", label: "trial still open", value: current.trialOpen, color: "var(--dash-muted)" },
          { key: "exec", label: "executive, comped", value: current.executive, color: "var(--dash-muted-soft)" },
        ]}
      />

      <p className={styles.foot}>
        A snapshot of where players stand now, not a conversion rate — a membership that converts loses the record
        that it ever trialled.
      </p>
    </div>
  );
}
```

`TrialConversionCard.module.css`:

```css
.container {
  container-type: inline-size;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.lead {
  margin: 0;
}

.figure {
  font-size: var(--card-figure-size, var(--font-size-4xl));
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-900);
  line-height: 1;
  letter-spacing: -0.02em;
}

.caption {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--color-gray-700);
}

.disabled,
.foot {
  margin: 0;
  font-size: var(--font-size-xs);
  color: var(--color-gray-700);
  line-height: 1.5;
}

.foot {
  padding-top: 0.625rem;
  border-top: 1px solid var(--color-gray-200);
}
```

`index.ts`:

```ts
export { TrialConversionCard } from "./TrialConversionCard";
```

- [ ] **Step 4: Register the card**

Add `export * from "./TrialConversionCard";` to `components/index.ts`, and in
`dashboardCards.ts` add `trialConversion: TrialConversionCard` with its import.

- [ ] **Step 5: Run the tests, lint and build**

Run: `cd webapp && npx jest src/features/dashboard && npm run lint && npm run build`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add webapp/src/features/dashboard
git commit -m "feat(#426): add trial conversion card"
```

---

### Task 11: Leaderboard card

**Files:**
- Create: `webapp/src/features/dashboard/components/LeaderboardCard/` (`.tsx`, `.module.css`, `index.ts`, `.test.tsx`)
- Modify: `components/index.ts`, `dashboardCards.ts`

**Interfaces:**
- Consumes: the existing rankings endpoint. The shape is `models.RankingResponse`
  (`server/internal/models/ranking.go:23-29`) — note it is `id`, not `membershipId`, and it
  already carries a tie-aware `position` computed by `semester_rankings_view`, so the card
  must render `position` rather than the array index. Add to `dashboardApi.ts`:

```ts
export interface RankingEntry {
  id: number;
  firstName: string;
  lastName: string;
  points: number;
  position: number;
}

export async function fetchTopRankings(semesterId: string): Promise<RankingEntry[]> {
  return apiClient<RankingEntry[]>(`v2/semesters/${semesterId}/rankings?limit=5`);
}
```

and a `useTopRankings` hook following the pattern in Task 5.

- [ ] **Step 1: Write the failing test**

```tsx
/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";
import { useTopRankings } from "../../hooks/useDashboardQueries";
import { LeaderboardCard } from "./LeaderboardCard";

jest.mock("../../hooks/useDashboardQueries", () => ({ useTopRankings: jest.fn() }));
const mocked = useTopRankings as jest.Mock;

const rows = [
  { id: 1, firstName: "Robin", lastName: "Chen", points: 412, position: 1 },
  { id: 2, firstName: "Asha", lastName: "Patel", points: 388, position: 2 },
];

describe("LeaderboardCard", () => {
  afterEach(() => jest.clearAllMocks());

  it("ranks players and shows each one's points", () => {
    mocked.mockReturnValue({ isLoading: false, isError: false, data: rows });
    render(<LeaderboardCard semesterId="s1" />);

    expect(screen.getByText("Robin Chen")).toBeInTheDocument();
    expect(screen.getByText("412")).toBeInTheDocument();
    expect(screen.getAllByRole("listitem")).toHaveLength(2);
  });

  it("says so when nobody has scored yet", () => {
    mocked.mockReturnValue({ isLoading: false, isError: false, data: [] });
    render(<LeaderboardCard semesterId="s1" />);

    expect(screen.getByText("No points awarded yet this term.")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd webapp && npx jest src/features/dashboard/components/LeaderboardCard`
Expected: FAIL — cannot find module `./LeaderboardCard`.

- [ ] **Step 3: Implement**

```tsx
import { DashboardCard } from "../DashboardCard";
import { CardActionLink } from "../CardActionLink";
import { CARD_TITLES } from "../../dashboardLayout";
import { useTopRankings } from "../../hooks/useDashboardQueries";
import type { RankingEntry } from "../../api/dashboardApi";
import styles from "./LeaderboardCard.module.css";

type Props = { semesterId: string };

export function LeaderboardCard({ semesterId }: Props) {
  const { data, isLoading, isError, refetch } = useTopRankings(semesterId);
  const status = isError ? "error" : isLoading || !data ? "loading" : data.length === 0 ? "empty" : "ready";

  return (
    <DashboardCard
      title={CARD_TITLES.leaderboard}
      status={status}
      onRetry={() => refetch()}
      emptyMessage="No points awarded yet this term."
      action={<CardActionLink to="/admin/rankings">Full rankings</CardActionLink>}
      data-qa="leaderboard-card"
    >
      {() => <Body rows={data as RankingEntry[]} />}
    </DashboardCard>
  );
}

/**
 * The gap between first and fifth is the interesting part and a bare list hides it,
 * so each row carries a bar scaled to the leader. Emphasis rather than categorical:
 * only the leader is coloured.
 */
function Body({ rows }: { rows: RankingEntry[] }) {
  const leaderPoints = Math.max(...rows.map((row) => row.points), 1);

  return (
    <ol className={styles.list}>
      {rows.map((row) => (
        <li key={row.id} className={styles.row} data-leader={row.position === 1 || undefined}>
          <span className={styles.rank}>{row.position}</span>
          <span className={styles.name}>
            {row.firstName} {row.lastName}
          </span>
          <span className={styles.bar} aria-hidden="true">
            <i style={{ width: `${(row.points / leaderPoints) * 100}%` }} />
          </span>
          <span className={styles.points}>{row.points}</span>
        </li>
      ))}
    </ol>
  );
}
```

```css
.list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.row {
  display: flex;
  align-items: center;
  gap: 0.5625rem;
  padding: 0.375rem 0;
}

.rank {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-500);
  width: 0.8125rem;
  flex-shrink: 0;
}

.name {
  font-size: var(--font-size-xs);
  color: var(--color-gray-900);
  flex: 1 1 5.5rem;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bar {
  flex: 1 1 2rem;
  height: 0.4375rem;
  background: var(--color-gray-100);
  border-radius: var(--border-radius-full);
  min-width: 1.875rem;
}

.bar i {
  display: block;
  height: 100%;
  border-radius: var(--border-radius-full);
  background: var(--color-gray-300);
}

.row[data-leader] .bar i {
  background: var(--dash-ord-3);
}

.points {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-900);
  width: 2.125rem;
  text-align: right;
  flex-shrink: 0;
}
```

- [ ] **Step 4: Register, test, lint and build**

Run: `cd webapp && npx jest src/features/dashboard && npm run lint && npm run build`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add webapp/src/features/dashboard
git commit -m "feat(#436): add leaderboard card"
```

---

### Task 12: Quick Actions card

**Files:**
- Create: `webapp/src/features/dashboard/components/QuickActionsCard/` (`.tsx`, `.module.css`, `index.ts`, `.test.tsx`)
- Modify: `components/index.ts`, `dashboardCards.ts`

**Interfaces:**
- Consumes: `useAuth()` from `@/hooks`, which exposes
  `hasPermission(action: Actions, resource: Resources, subResource?: SubResources): boolean`
  (`src/contexts/AuthContext.ts:10`). There is no `permissions` array — use this function.
  Valid pairs come from `PermissionList` in `src/interfaces/responses/session.ts`.

**A deviation from the epic, decided here.** Epic card 2 names the actions "Create event,
register a membership, add a member, open rankings". Those create flows are **modals on the
list pages** — `Events.tsx` routes only `/` and `/:eventId`, `Members.tsx` only `/` — so no
URL opens a create modal. A link labelled "Create event" that lands on a list is a lie about
what it does, so the actions are labelled by destination instead. Deep-linking the create
modals is listed as follow-up work; when it lands, the labels become the epic's.

- [ ] **Step 1: Write the failing test**

```tsx
/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import "@testing-library/jest-dom";
import { useAuth } from "@/hooks";
import { QuickActionsCard } from "./QuickActionsCard";

jest.mock("@/hooks", () => ({ useAuth: jest.fn() }));
const mocked = useAuth as jest.Mock;

const renderCard = () =>
  render(
    <MemoryRouter>
      <QuickActionsCard />
    </MemoryRouter>,
  );

describe("QuickActionsCard", () => {
  afterEach(() => jest.clearAllMocks());

  it("offers every destination the session is permitted", () => {
    mocked.mockReturnValue({ hasPermission: () => true });
    renderCard();

    expect(screen.getByRole("link", { name: "Events" })).toHaveAttribute("href", "/admin/events");
    expect(screen.getByRole("link", { name: "Members" })).toHaveAttribute("href", "/admin/members");
    expect(screen.getByRole("link", { name: "Rankings" })).toHaveAttribute("href", "/admin/rankings");
  });

  it("hides destinations the session cannot reach", () => {
    mocked.mockReturnValue({
      hasPermission: (action: string, resource: string) => resource === "semester",
    });
    renderCard();

    expect(screen.queryByRole("link", { name: "Events" })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Rankings" })).toBeInTheDocument();
  });

  it("says so when the session can reach none of them", () => {
    mocked.mockReturnValue({ hasPermission: () => false });
    renderCard();

    expect(screen.getByText("No actions available for your role.")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd webapp && npx jest src/features/dashboard/components/QuickActionsCard`
Expected: FAIL — cannot find module `./QuickActionsCard`.

- [ ] **Step 3: Implement**

```tsx
import { Link } from "react-router-dom";
import { useAuth } from "@/hooks";
import { DashboardCard } from "../DashboardCard";
import { CARD_TITLES } from "../../dashboardLayout";
import styles from "./QuickActionsCard.module.css";

import type { Actions, Resources, SubResources } from "@/interfaces/responses";

type Action = {
  key: string;
  label: string;
  to: string;
  action: Actions;
  resource: Resources;
  subResource?: SubResources;
};

/**
 * Labelled by destination, not by verb. Creating an event or a member happens in a
 * modal on the relevant list page and no URL opens one, so "Create event" would name
 * something the link does not do. See the deviation note in the plan.
 */
const ACTIONS: Action[] = [
  { key: "events", label: "Events", to: "/admin/events", action: "list", resource: "event" },
  { key: "members", label: "Members", to: "/admin/members", action: "list", resource: "membership" },
  { key: "rankings", label: "Rankings", to: "/admin/rankings", action: "list", resource: "semester", subResource: "rankings" },
  { key: "logins", label: "Logins", to: "/admin/logins", action: "list", resource: "login" },
];

/**
 * The only card with no data. It shows none — no figure, no chart, no delta — which
 * is what distinguishes it on a page where everything else is a measurement.
 *
 * It takes no props: a parameterless component is assignable to
 * ComponentType<DashboardCardComponentProps>, so it registers like the others
 * without carrying a semesterId it never reads.
 */
export function QuickActionsCard() {
  const { hasPermission } = useAuth();
  const available = ACTIONS.filter((item) => hasPermission(item.action, item.resource, item.subResource));

  return (
    <DashboardCard
      title={CARD_TITLES.quickActions}
      status={available.length === 0 ? "empty" : "ready"}
      emptyMessage="No actions available for your role."
      data-qa="quick-actions-card"
    >
      {() => (
        <div className={styles.grid}>
          {available.map((action) => (
            <Link key={action.key} to={action.to} className={styles.action}>
              {action.label}
            </Link>
          ))}
        </div>
      )}
    </DashboardCard>
  );
}
```

```css
.grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.5rem;
}

.action {
  display: block;
  padding: 0.75rem 0.625rem;
  border: 1px solid var(--color-border);
  border-radius: var(--border-radius-md);
  text-decoration: none;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  color: var(--color-gray-900);
  line-height: 1.3;
  transition: border-color var(--transition-fast);
}

.action:hover,
.action:focus-visible {
  border-color: var(--color-primary);
}

@container (max-width: 15rem) {
  .grid {
    grid-template-columns: 1fr;
  }
}
```

- [ ] **Step 4: Register, test, lint and build**

Run: `cd webapp && npx jest src/features/dashboard && npm run lint && npm run build`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add webapp/src/features/dashboard
git commit -m "feat(#436): add quick actions card"
```

---

### Task 13: Event Spotlight card

**Files:**
- Create: `webapp/src/features/dashboard/components/EventSpotlightCard/` (`.tsx`, `.module.css`, `index.ts`, `.test.tsx`)
- Modify: `components/index.ts`, `dashboardCards.ts`

**Interfaces:**
- Consumes: `useSpotlight` (Task 5), `SpotlightEvent`. `state` is `0` for started and `1` for
  ended, matching `models.EventStateStarted` / `EventStateEnded`.

- [ ] **Step 1: Write the failing test**

```tsx
/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import "@testing-library/jest-dom";
import { useSpotlight } from "../../hooks/useDashboardQueries";
import { EventSpotlightCard } from "./EventSpotlightCard";
import { SAMPLE_TERM } from "../../fixtures/sampleTerm";

jest.mock("../../hooks/useDashboardQueries", () => ({ useSpotlight: jest.fn() }));
const mocked = useSpotlight as jest.Mock;

const renderCard = () =>
  render(
    <MemoryRouter>
      <EventSpotlightCard semesterId="s1" />
    </MemoryRouter>,
  );

describe("EventSpotlightCard", () => {
  afterEach(() => jest.clearAllMocks());

  it("leads with the event's name, not a count", () => {
    mocked.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.spotlight });
    renderCard();

    expect(screen.getByRole("heading", { name: "Thursday Night NLH" })).toBeInTheDocument();
    expect(screen.getByText("47")).toBeInTheDocument();
    expect(screen.getByText("12")).toBeInTheDocument();
  });

  it("offers to create one when nothing is scheduled", () => {
    mocked.mockReturnValue({ isLoading: false, isError: false, data: null });
    renderCard();

    expect(screen.getByText("No events scheduled this term.")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd webapp && npx jest src/features/dashboard/components/EventSpotlightCard`
Expected: FAIL — cannot find module `./EventSpotlightCard`.

- [ ] **Step 3: Implement**

```tsx
import { Link } from "react-router-dom";
import { DashboardCard } from "../DashboardCard";
import { CARD_TITLES } from "../../dashboardLayout";
import { useSpotlight } from "../../hooks/useDashboardQueries";
import type { SpotlightEvent } from "../../api/dashboardApi";
import styles from "./EventSpotlightCard.module.css";

const EVENT_STATE_STARTED = 0;

type Props = { semesterId: string };

export function EventSpotlightCard({ semesterId }: Props) {
  const { data, isLoading, isError, refetch } = useSpotlight(semesterId);
  const status = isError ? "error" : isLoading ? "loading" : !data ? "empty" : "ready";

  return (
    <DashboardCard
      title={CARD_TITLES.spotlight}
      status={status}
      onRetry={() => refetch()}
      emptyMessage="No events scheduled this term."
      data-qa="event-spotlight-card"
    >
      {() => <Body event={data as SpotlightEvent} />}
    </DashboardCard>
  );
}

/**
 * The only card whose hero is text. At the door you need to know which event is
 * running, not how many of anything — so the name carries the weight and the counts
 * sit underneath. This card also owns the page's only motion.
 */
function Body({ event }: { event: SpotlightEvent }) {
  const live = event.state === EVENT_STATE_STARTED && new Date(event.startDate) <= new Date();
  const startTime = new Date(event.startDate).toLocaleTimeString("en-CA", { hour: "numeric", minute: "2-digit" });

  return (
    <div className={styles.container}>
      <p className={styles.state} data-live={live || undefined}>
        {live ? (
          <>
            <span className={styles.pulse} aria-hidden="true" />
            Live · started {startTime}
          </>
        ) : (
          <>Scheduled · {new Date(event.startDate).toLocaleDateString("en-CA", { weekday: "long", month: "short", day: "numeric" })}</>
        )}
      </p>
      <h4 className={styles.name}>{event.name}</h4>
      <p className={styles.format}>{event.format}</p>

      <dl className={styles.stats}>
        <div>
          <dd>{event.entries}</dd>
          <dt>entries</dt>
        </div>
        <div>
          <dd>{event.rebuys}</dd>
          <dt>rebuys</dt>
        </div>
      </dl>

      <Link className={styles.cta} to={`/admin/events/${event.id}`}>
        Open event
      </Link>
    </div>
  );
}
```

```css
.container {
  container-type: inline-size;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.state {
  margin: 0;
  display: flex;
  align-items: center;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-700);
}

.state[data-live] {
  color: var(--color-success-dark);
}

.pulse {
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 50%;
  background: var(--color-success);
  margin-right: 0.4375rem;
  box-shadow: 0 0 0 0 rgb(30 142 62 / 0.6);
  animation: pulse 2.4s ease-out infinite;
}

@keyframes pulse {
  70% {
    box-shadow: 0 0 0 0.5625rem rgb(30 142 62 / 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgb(30 142 62 / 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .pulse {
    animation: none;
  }
}

.name {
  margin: 0.25rem 0 0;
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-900);
  line-height: 1.2;
}

.format {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--color-gray-700);
}

.stats {
  display: flex;
  gap: 1.125rem;
  margin: 0.875rem 0 0;
}

.stats dd {
  margin: 0;
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-900);
  line-height: 1;
}

.stats dt {
  font-size: var(--font-size-xs);
  color: var(--color-gray-500);
}

.cta {
  display: block;
  margin-top: 0.875rem;
  padding: 0.5625rem;
  text-align: center;
  background: var(--dash-ord-3);
  color: var(--color-text-inverse);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  border-radius: var(--border-radius-md);
  text-decoration: none;
}
```

- [ ] **Step 4: Register, test, lint and build**

Run: `cd webapp && npx jest src/features/dashboard && npm run lint && npm run build`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add webapp/src/features/dashboard
git commit -m "feat(#437): add event spotlight card"
```

---

### Task 14: Recharts and the shared chart theme

**Files:**
- Modify: `webapp/package.json`
- Create: `webapp/src/features/dashboard/charts/chartTheme.ts`
- Test: `webapp/src/features/dashboard/charts/chartTheme.test.ts`

**Interfaces:**
- Produces: `CHART_COLORS` (categorical slots keyed by name), `AXIS_PROPS`, `GRID_PROPS`,
  and `TOOLTIP_STYLE`, consumed by Tasks 15 and 16.

This task implements epic issue **#439**.

- [ ] **Step 1: Install Recharts**

```bash
cd webapp && npm install recharts@^2.15.0
```

- [ ] **Step 2: Write the failing test**

```ts
import { CHART_COLORS, AXIS_PROPS } from "./chartTheme";

describe("chartTheme", () => {
  it("uses the validated categorical slots and never a status colour", () => {
    expect(CHART_COLORS.slot1).toBe("#6f4fae");
    expect(CHART_COLORS.slot2).toBe("#c98f00");
    expect(Object.values(CHART_COLORS)).not.toContain("#d93025");
    expect(Object.values(CHART_COLORS)).not.toContain("#1e8e3e");
  });

  it("keeps axes recessive", () => {
    expect(AXIS_PROPS.tickLine).toBe(false);
    expect(AXIS_PROPS.axisLine).toBe(false);
  });
});
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `cd webapp && npx jest src/features/dashboard/charts`
Expected: FAIL — cannot find module `./chartTheme`.

- [ ] **Step 4: Implement**

```ts
/**
 * Shared Recharts theming (#439). The palette is validated — see the spec's Colour
 * section. Red and green are reserved for DeltaChip and must never appear here.
 */
export const CHART_COLORS = {
  slot1: "#6f4fae",
  slot2: "#c98f00",
  slot3: "#2a78d6",
  slot4: "#d5578a",
} as const;

export const AXIS_PROPS = {
  tickLine: false,
  axisLine: false,
  tick: { fontSize: 11, fill: "#9e9e9e", fontFamily: "Montserrat, sans-serif" },
} as const;

export const GRID_PROPS = {
  stroke: "#eeeeee",
  vertical: false,
} as const;

export const TOOLTIP_STYLE = {
  contentStyle: {
    fontSize: 12,
    fontFamily: "Montserrat, sans-serif",
    borderRadius: 6,
    border: "1px solid #e0e0e0",
    boxShadow: "0 1px 3px rgb(0 0 0 / 0.1)",
  },
} as const;
```

- [ ] **Step 5: Run the tests, lint and build**

Run: `cd webapp && npx jest src/features/dashboard/charts && npm run lint && npm run build`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add webapp/package.json webapp/package-lock.json webapp/src/features/dashboard/charts
git commit -m "feat(#439): add Recharts and shared dashboard chart theming"
```

---

### Task 15: Event Activity card

**Files:**
- Create: `webapp/src/features/dashboard/components/EventActivityCard/` (`.tsx`, `.module.css`, `index.ts`, `.test.tsx`)
- Modify: `components/index.ts`, `dashboardCards.ts`

**Interfaces:**
- Consumes: `useEventActivity` (Task 5), `CHART_COLORS` / `AXIS_PROPS` / `GRID_PROPS` /
  `TOOLTIP_STYLE` (Task 14), `SAMPLE_TERM.eventActivity` (Task 5).

Recharts renders nothing in jsdom without a sized container, so the test asserts the
scalar readouts and the accessible chart description, not SVG geometry.

- [ ] **Step 1: Write the failing test**

```tsx
/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";
import { useEventActivity } from "../../hooks/useDashboardQueries";
import { EventActivityCard } from "./EventActivityCard";
import { SAMPLE_TERM } from "../../fixtures/sampleTerm";

jest.mock("../../hooks/useDashboardQueries", () => ({ useEventActivity: jest.fn() }));
const mocked = useEventActivity as jest.Mock;

describe("EventActivityCard", () => {
  afterEach(() => jest.clearAllMocks());

  it("leads with events run against events scheduled", () => {
    mocked.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.eventActivity });
    render(<EventActivityCard semesterId="s1" />);

    expect(screen.getByText("11")).toBeInTheDocument();
    expect(screen.getByText(/of 14 events run this term/)).toBeInTheDocument();
    expect(screen.getByText(/486 entries/)).toBeInTheDocument();
  });

  it("shows a start-of-term state rather than a meter at zero", () => {
    mocked.mockReturnValue({
      isLoading: false,
      isError: false,
      data: { current: { eventsRun: 0, eventsScheduled: 0, totalEntries: 0, averageFieldSize: 0, series: [] }, comparison: null },
    });
    render(<EventActivityCard semesterId="s1" />);

    expect(screen.getByText("No events scheduled yet this term.")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd webapp && npx jest src/features/dashboard/components/EventActivityCard`
Expected: FAIL — cannot find module `./EventActivityCard`.

- [ ] **Step 3: Implement**

```tsx
import { Bar, BarChart, CartesianGrid, ReferenceLine, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { DashboardCard } from "../DashboardCard";
import { CARD_TITLES } from "../../dashboardLayout";
import { useEventActivity } from "../../hooks/useDashboardQueries";
import { AXIS_PROPS, CHART_COLORS, GRID_PROPS, TOOLTIP_STYLE } from "../../charts/chartTheme";
import type { EventActivityResponse } from "../../api/dashboardApi";
import styles from "./EventActivityCard.module.css";

type Props = { semesterId: string };

export function EventActivityCard({ semesterId }: Props) {
  const { data, isLoading, isError, refetch } = useEventActivity(semesterId);
  const status = isError
    ? "error"
    : isLoading || !data
      ? "loading"
      : data.current.eventsScheduled === 0 && data.current.eventsRun === 0
        ? "empty"
        : "ready";

  return (
    <DashboardCard
      title={CARD_TITLES.eventActivity}
      status={status}
      onRetry={() => refetch()}
      emptyMessage="No events scheduled yet this term."
      data-qa="event-activity-card"
    >
      {() => <Body data={data as EventActivityResponse} />}
    </DashboardCard>
  );
}

function Body({ data }: { data: EventActivityResponse }) {
  const { eventsRun, eventsScheduled, totalEntries, averageFieldSize, series } = data.current;
  const remaining = Math.max(0, eventsScheduled - eventsRun);
  const average = Math.round(averageFieldSize * 10) / 10;

  return (
    <div className={styles.container}>
      <p className={styles.lead}>
        <span className={styles.figure}>{eventsRun}</span>
        <span className={styles.caption}>of {eventsScheduled} events run this term</span>
      </p>

      <div className={styles.meter} aria-hidden="true">
        <i style={{ flexGrow: eventsRun }} />
        {remaining > 0 && <u style={{ flexGrow: remaining }} />}
      </div>
      <p className={styles.sub}>
        {remaining} still scheduled · {totalEntries.toLocaleString("en-CA")} entries · average field {average}
      </p>

      <div className={styles.chart}>
        <ResponsiveContainer width="100%" height={160}>
          <BarChart data={series} margin={{ top: 8, right: 8, bottom: 0, left: -20 }}>
            <CartesianGrid {...GRID_PROPS} />
            <XAxis dataKey="name" {...AXIS_PROPS} interval="preserveStartEnd" />
            <YAxis {...AXIS_PROPS} width={34} />
            <Tooltip {...TOOLTIP_STYLE} />
            <ReferenceLine
              y={average}
              stroke={CHART_COLORS.slot4}
              strokeDasharray="4 4"
              strokeWidth={2}
              label={{ value: `avg ${average}`, position: "right", fontSize: 10, fill: CHART_COLORS.slot4 }}
            />
            <Bar dataKey="entries" fill={CHART_COLORS.slot1} radius={[4, 4, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </div>

      <p className={styles.foot}>
        An event counts as run once it has ended. Average field size covers run events only.
      </p>
    </div>
  );
}
```

```css
.container {
  container-type: inline-size;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.lead {
  display: flex;
  align-items: baseline;
  gap: 0.5625rem;
  margin: 0;
}

.figure {
  font-size: var(--card-figure-size, var(--font-size-3xl));
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-900);
  line-height: 1;
}

.caption {
  font-size: var(--font-size-sm);
  color: var(--color-gray-700);
}

.meter {
  display: flex;
  gap: 2px;
  height: 0.625rem;
}

.meter i,
.meter u {
  display: block;
  border-radius: var(--border-radius-full);
}

.meter i {
  background: var(--dash-ord-3);
}

.meter u {
  background: var(--dash-band);
}

.sub,
.foot {
  margin: 0;
  font-size: var(--font-size-xs);
  color: var(--color-gray-700);
}

.chart {
  margin-top: 0.5rem;
}

.foot {
  padding-top: 0.625rem;
  border-top: 1px solid var(--color-gray-200);
  line-height: 1.5;
}
```

- [ ] **Step 4: Register, test, lint and build**

Run: `cd webapp && npx jest src/features/dashboard && npm run lint && npm run build`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add webapp/src/features/dashboard
git commit -m "feat(#440): add event activity card"
```

---

### Task 16: Signup Timeline card

**Files:**
- Create: `webapp/src/features/dashboard/components/SignupTimelineCard/` (`.tsx`, `.module.css`, `index.ts`, `.test.tsx`)
- Modify: `components/index.ts`, `dashboardCards.ts`

**Interfaces:**
- Consumes: `useSignups` (Task 5), chart theme (Task 14), `SAMPLE_TERM.signups` (Task 5).

- [ ] **Step 1: Write the failing test**

```tsx
/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";
import { useSignups } from "../../hooks/useDashboardQueries";
import { SignupTimelineCard } from "./SignupTimelineCard";
import { SAMPLE_TERM } from "../../fixtures/sampleTerm";

jest.mock("../../hooks/useDashboardQueries", () => ({ useSignups: jest.fn() }));
const mocked = useSignups as jest.Mock;

describe("SignupTimelineCard", () => {
  afterEach(() => jest.clearAllMocks());

  it("leads with the term total and marks itself as sample data", () => {
    mocked.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.signups });
    render(<SignupTimelineCard semesterId="s1" />);

    expect(screen.getByText("967")).toBeInTheDocument();
    expect(screen.getByText("Sample data")).toBeInTheDocument();
  });

  it("says tracking has a start date rather than implying a wait that never ends", () => {
    mocked.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.signups });
    render(<SignupTimelineCard semesterId="s1" />);

    expect(screen.getByText(/Signup tracking began/)).toBeInTheDocument();
    expect(screen.queryByText(/Collecting data since/i)).not.toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd webapp && npx jest src/features/dashboard/components/SignupTimelineCard`
Expected: FAIL — cannot find module `./SignupTimelineCard`.

- [ ] **Step 3: Implement**

```tsx
import { Area, AreaChart, CartesianGrid, ReferenceLine, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { DashboardCard } from "../DashboardCard";
import { CARD_TITLES } from "../../dashboardLayout";
import { useSignups } from "../../hooks/useDashboardQueries";
import { AXIS_PROPS, CHART_COLORS, GRID_PROPS, TOOLTIP_STYLE } from "../../charts/chartTheme";
import { SAMPLE_TERM } from "../../fixtures/sampleTerm";
import type { SignupsResponse } from "../../api/dashboardApi";
import styles from "./SignupTimelineCard.module.css";

/** Flip to false once GET …/dashboard/signups (#433) exists. */
const USES_SAMPLE_DATA = true;

type Props = { semesterId: string };

export function SignupTimelineCard({ semesterId }: Props) {
  const { data, isLoading, isError, refetch } = useSignups(semesterId);
  const resolved = USES_SAMPLE_DATA ? SAMPLE_TERM.signups : data;

  const status = USES_SAMPLE_DATA
    ? "ready"
    : isError
      ? "error"
      : isLoading || !resolved
        ? "loading"
        : resolved.series.length === 0
          ? "empty"
          : "ready";

  return (
    <DashboardCard
      title={CARD_TITLES.signupTimeline}
      status={status}
      sampleData={USES_SAMPLE_DATA}
      onRetry={() => refetch()}
      emptyMessage={emptyMessageFor(resolved)}
      data-qa="signup-timeline-card"
    >
      {() => <Body data={resolved as SignupsResponse} />}
    </DashboardCard>
  );
}

/**
 * Terms predating the created_at migration return an empty series permanently, so the
 * message must name the start date rather than implying data is on its way.
 */
function emptyMessageFor(data?: SignupsResponse) {
  return data?.dataStartsAt
    ? `Signup tracking began ${formatDate(data.dataStartsAt)}. Earlier terms have no creation dates.`
    : "Signup tracking had not begun in this term.";
}

const formatDate = (iso: string) =>
  new Date(iso).toLocaleDateString("en-CA", { month: "long", year: "numeric" });

function Body({ data }: { data: SignupsResponse }) {
  const busiest = data.series.reduce((max, point) => Math.max(max, point.admin + point.discord), 0);

  return (
    <div className={styles.container}>
      <p className={styles.lead}>
        <span className={styles.figure}>{data.total.toLocaleString("en-CA")}</span>
        <span className={styles.caption}>memberships created, {busiest} on the busiest day</span>
      </p>

      <div className={styles.chart}>
        <ResponsiveContainer width="100%" height={170}>
          <AreaChart data={data.series} margin={{ top: 8, right: 8, bottom: 0, left: -20 }}>
            <CartesianGrid {...GRID_PROPS} />
            <XAxis dataKey="date" {...AXIS_PROPS} interval="preserveStartEnd" />
            <YAxis {...AXIS_PROPS} width={34} />
            <Tooltip {...TOOLTIP_STYLE} />
            {data.eventDates.map((date) => (
              <ReferenceLine key={date} x={date} stroke={CHART_COLORS.slot2} strokeWidth={2} />
            ))}
            <Area type="monotone" dataKey="admin" stackId="1" stroke={CHART_COLORS.slot1} fill={CHART_COLORS.slot1} fillOpacity={0.3} strokeWidth={2} name="Admin" />
            <Area type="monotone" dataKey="discord" stackId="1" stroke={CHART_COLORS.slot2} fill={CHART_COLORS.slot2} fillOpacity={0.3} strokeWidth={2} name="Discord" />
          </AreaChart>
        </ResponsiveContainer>
      </div>

      <p className={styles.foot}>
        Gold lines mark event days. Signup tracking began {formatDate(data.dataStartsAt ?? data.series[0].date)};
        earlier terms have no creation dates and stay empty.
      </p>
    </div>
  );
}
```

```css
.container {
  container-type: inline-size;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.lead {
  display: flex;
  align-items: baseline;
  gap: 0.5625rem;
  margin: 0;
  flex-wrap: wrap;
}

.figure {
  font-size: var(--card-figure-size, var(--font-size-3xl));
  font-weight: var(--font-weight-semibold);
  color: var(--color-gray-900);
  line-height: 1;
}

.caption {
  font-size: var(--font-size-sm);
  color: var(--color-gray-700);
}

.chart {
  margin-top: 0.25rem;
}

.foot {
  margin: 0;
  padding-top: 0.625rem;
  border-top: 1px solid var(--color-gray-200);
  font-size: var(--font-size-xs);
  color: var(--color-gray-700);
  line-height: 1.5;
}
```

- [ ] **Step 4: Register all eight cards**

`dashboardCards.ts` must now map every id. Change the type from `Partial<Record<…>>` to
`Record<DashboardCardId, ComponentType<DashboardCardComponentProps>>` so a missing card is
a compile error, and delete the placeholder branch in `Dashboard.tsx`.

- [ ] **Step 5: Run the full suite, lint and build**

Run: `cd webapp && npm test && npm run lint && npm run build`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add webapp/src/features/dashboard webapp/src/pages
git commit -m "feat(#442): add signup timeline card and register all eight cards"
```

---

## Follow-up work, not in this plan

Open as separate GitHub issues under epic #426:

1. **Trial conversion endpoint** — `GET /api/v2/semesters/:semesterId/dashboard/conversion`,
   returning `players`, `paid`, `trialSpent`, `trialOpen`, `executive` and `freeTrialLimit`
   over the population of members who entered ≥1 event. Then flip `USES_SAMPLE_DATA` to
   `false` in `TrialConversionCard.tsx`.
2. **`memberships.converted_at`** — nullable, no backfill, stamped at the unpaid → paid
   transition in `membership_service.go` where `FreeTrialAvailable` is currently reset.
   Makes the true conversion rate available from the following term.
3. **#433 signup timeline endpoint** — already an open issue. On completion, flip
   `USES_SAMPLE_DATA` to `false` in `SignupTimelineCard.tsx`.
4. **#443 Cypress coverage** — role ordering differs between logins; one card failing leaves
   its neighbours rendered.
5. **Revert the exploratory `paidPlayers` server changes** if they were committed anywhere
   other than the working tree discarded in Task 1.
6. **Deep-link the create modals** — a route or query parameter that opens the create-event
   and add-member modals directly. Quick Actions then labels its tiles with the epic's verbs
   ("Create event", "Add member") instead of destinations.
