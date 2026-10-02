# Admin Dashboard — page design

**Date:** 2026-10-01
**Scope:** the whole of `/admin/dashboard` — grid, shared system, and all eight cards
**Supersedes:** the card-level design implied by #434, #438 and #441 as currently implemented
**Epic:** #426

## Why this exists

Two cards of seven are built. #438 (Memberships) shipped a hero-number-plus-mini-stat-grid;
#441 (Engagement) grew into a sectioned comparison table. Side by side they read as two
different products, and neither answers the question an officer actually has. Five cards
remain unbuilt, so settling the page now costs one spec and saves five retrofits.

The page is the default landing route for every executive. It has to be worth landing on.

## Governing rule

**Each card leads with whatever its data actually is.**

A card's form is chosen from the shape of its data — a name, a distribution, a ratio, a
series, a ranked list, or nothing at all — not from a shared template. Seven cards forced
into one anatomy is what produced the spreadsheet feel; eight cards each shaped to their
own question is the fix.

What is shared is chrome and grammar, never layout. See *The shared system*.

## Layout

### Two lanes

```
┌─────────────────────────────────┬──────────────┐
│ LEAD CARD (hero scale)          │ rail card    │
│                                 ├──────────────┤
├─────────────────────────────────┤ rail card    │
│ wide card (chart)               ├──────────────┤
├─────────────────────────────────┤ rail card    │
│ wide card (chart)               ├──────────────┤
└─────────────────────────────────┴──────────────┘
```

A wide lane (`2fr`) and a rail (`1fr`). Cards are assigned by this algorithm, which runs
client-side against the role preset:

1. `preset[0]` is promoted to the **lead card**: first in the wide lane, rendered at hero
   scale (larger lead figure, more vertical room) regardless of its natural width.
2. Every remaining card is routed by its declared `lane` preference — `wide` for cards
   carrying a chart, `rail` for everything else.
3. Relative preset order is preserved within each lane.

This makes role ordering **perceptible**. Epic goal #4 is "order the page by role", but
reordering seven visually identical boxes is nearly invisible; promoting the role's top
card to hero scale is the ordering you can actually see.

### Responsiveness

The grid wrapper is a **CSS container** (`container-type: inline-size`), not a viewport
media query consumer. This is load-bearing rather than stylistic: the dashboard sits in a
flex lane beside the admin sidenav (`AdminLayout.module.css`), so viewport width
systematically overstates the room available. Lanes collapse on real measured space.

- Container ≥ `64rem`: two lanes as above.
- Container < `64rem`: one lane, **exact preset order restored**, lead card keeps hero scale.

Every card additionally declares `container-type: inline-size` on its own root, so card
internals reflow against their own tile rather than the page. A card does not need to know
whether it is in the wide lane, the rail, or a collapsed single column.

**Remove** `grid-auto-flow: dense` from `DashboardGrid.module.css`. Backfill reorders cards
unpredictably, which directly contradicts the role-ordering goal.

**Remove** `CARD_TILES` as currently written. `rows: 3` on engagement is already dead
config — `DashboardGrid.module.css` only implements a rule for `[data-rows="2"]`. It is
replaced by a `lane: "wide" | "rail"` preference per card.

## The shared system

Everything here is common to all eight cards. Nothing else is.

**Chrome.** White surface, `0.5rem` radius, gold top edge (`inset 0 3px 0 var(--color-primary)`),
`1rem 1.125rem` padding. One shadow, `--shadow-sm`.

**Type.** Montserrat throughout, from `@uwpokerclub/components/tokens.css`. Card titles are
**sentence case at `--font-size-sm` / weight 600** — changed from the current tracked-out
uppercase. Uppercase micro-labels above every heading is the commonest tell of a generated
interface, and the page stacks eight of them.

**Tokens over fallbacks.** Every dashboard CSS file currently writes
`var(--color-gray-900, #212121)`. The tokens exist and are loaded in `main.tsx`; the
fallbacks are dead weight that silently permits drift. Drop them.

**Delta grammar.** `DeltaChip` is unchanged in behaviour and used identically everywhere:
`positive-is-good` / `negative-is-good` / `neutral`, count or percentage-point mode, a
`<1%` floor so a real change never renders as "no change", and a screen-reader label
spelling the comparison term. Red and green belong to deltas and are **never** used as a
series colour.

**Card states.** `DashboardCard`'s loading / error / empty / ready contract is unchanged.
Each card owns its own query, so one failing endpoint renders one failed card. This is the
whole payoff of per-card endpoints and must survive the redesign.

### Colour

Derived from the club's own brand and **validated**, not eyeballed
(`dataviz/scripts/validate_palette.js`, light mode, surface `#fcfcfb`).

Raw brand hexes fail the lightness band — `#4b2e83` at L 0.385, `#f4b80f` at L 0.816 — so
both are stepped into the passing band rather than abandoned.

**Categorical** (identity: membership buckets, signup source):

| Slot | Role | Hex |
|---|---|---|
| 1 | UW purple, stepped | `#6f4fae` |
| 2 | UWPSC gold, stepped | `#c98f00` |
| 3 | blue | `#2a78d6` |
| 4 | rose | `#d5578a` |

Passes lightness, chroma, CVD separation (worst adjacent ΔE 12.1, protan) and
normal-vision floor (worst adjacent ΔE 26.2). Slot 2 carries a contrast WARN at 2.76:1,
which obligates **visible direct labels on gold segments** — satisfied by the legend
pattern below, and not dismissable.

**Ordinal** (magnitude: attendance buckets, commitment tiers):

`#ab97d2` → `#7c5eb6` → `#4b2e83`

Passes monotone lightness, adjacent ΔL ≥ 0.06, light-end contrast 2.53:1, single hue
(spread 5°). The dark end is the exact brand purple.

**Light mode only.** `tokens.css` defines no dark palette and the admin UI has no theme
toggle. Do not invent dark steps.

**Mark specs.** 2px gap between stacked segments, 4px rounded data-ends, 2px lines,
recessive grid and axes, selective direct labels — never a number on every point.

## The cards

Eight cards. Each entry gives the data's job, the form that job selects, and the states.

### 1. Event Spotlight — *a name, not a number*

**Job:** identify one thing happening now. **Form:** status panel; the hero is **text**.

The only card whose lead is an event name, because at the door you need to know *which*
event, not how many. Live state shows a pulsing dot, the event name at `--font-size-xl`,
format beneath, then a three-up readout (entries / rebuys / elapsed) and a primary action.

This card owns **the page's only motion**: a 2.4s pulse on the live indicator, suppressed
under `prefers-reduced-motion`. No other card animates.

**States:** live → as above. Next scheduled → countdown replaces elapsed, action becomes
"View event". None scheduled → "Create event" for roles holding `event.create`, a plain
statement for roles that do not.

**Source:** `GET …/dashboard/spotlight` (#427, built).

### 2. Quick Actions — *no data at all*

**Job:** none. **Form:** a 2×2 of labelled actions. The quietest card on the page,
distinguished by absence. No figures, no chart, no delta.

Each action hidden when the session lacks its permission. **Source:** client-side session
permissions. No endpoint.

### 3. Memberships — *part-to-whole, twice*

**Job:** composition. **Form:** total as lead, then two horizontal stacked bars —
paid / unpaid / discounted / executive (categorical slots 1–4), and first-ever-term /
returning (ordinal, two steps).

This replaces the current 2×2 mini-stat grid, which cannot say "these are shares of one
total". Delta chip on the total only; per-bucket deltas moved to the legend as compact
chips so the bar stays readable.

**Source:** `GET …/dashboard/memberships` (#430, built).

### 4. Signup Timeline — *a spiky annotated series*

**Job:** change over time, with annotations. **Form:** area chart, one series, with event
days ticked on the axis in gold — because the spikes *are* the event days, and the chart is
unreadable without that.

**Lane:** wide. **States:** terms predating #429 return an empty series permanently. The
card says so using the `dataStartsAt` value the endpoint returns for exactly this purpose —
"Signup tracking began {dataStartsAt}" — never "collecting data since…", which implies a
wait that never ends. The date is read from the response, not hard-coded.

**Source:** `GET /api/v2/semesters/:semesterId/dashboard/signups` (#433, not built).

### 5. Event Activity — *a ratio, then a series*

**Job:** progress against a limit, plus magnitude over time. **Form:** a **meter** for
events run of events scheduled, then columns of entries per event with average field size
as a dashed reference line (categorical slot 4, used as a reference mark not a series).

**Lane:** wide. An event counts as run at `EventStateEnded`; average field size covers run
events only.

**States:** no events yet → start-of-term state, not a meter at zero.

**Source:** `GET …/dashboard/events` (#431, built).

### 6. Engagement & Retention — *a distribution*

**Job:** how a population distributes. **Form:** one horizontal stacked bar on the ordinal
ramp — played once / played 2–9 / regulars 10+ — beneath a lead figure of distinct players
against total memberships.

Three buckets, not four: the API exposes `players`, `playedOnceCount` and `tenPlusCount`,
so the middle bucket is arithmetic and 2–3 cannot be split from 4–9 without a backend
change. Three is sufficient; no follow-up issue.

**This replaces the comparison table.** It also removes a redundancy in the current
implementation: "Came back (2+ events)" and "Played exactly once" are the same fact — they
sum to 100% by construction — stated twice in two sections with opposite sentiment
colouring.

Beneath the bar, two **normal-range tracks**: median events played and played-once share,
each showing this term's value as a mark, the historical band as a shaded region, and last
year as a grey tick.

**Source:** `GET …/dashboard/engagement` (#432, built). `paidPlayers` / `paidPlayerShare`,
added to the working tree during iteration, are **removed** — superseded by card 8's
four-way split.

### 7. Leaderboard — *a ranked list where the gap is the point*

**Job:** rank plus magnitude. **Form:** five rows, each with name, points, and a bar scaled
to the leader. **Emphasis**, not categorical: the leader in purple, the rest in the
de-emphasis grey. A bare list hides whether first place is running away with it.

**Source:** existing `GET /semesters/:semesterId/rankings?limit=5`.

### 8. Trial Conversion — *a funnel snapshot* **(new card)**

**Job:** where a population stands on a commitment ladder. **Form:** lead figure is the
gap — members who spent the trial and never paid — then a stacked bar: paid / trial spent
and unpaid / trial still open / executive.

**What this is honest about.** It is a **snapshot, not a rate**. The free trial is enforced
in `participants_service.go:78-94` against `semesters.free_trial_limit`, so "spent the
trial, still unpaid" is exact. The *conversion rate* is not recoverable: when a membership
flips unpaid → paid, `membership_service.go:155-162` resets `FreeTrialAvailable` to `true`
deliberately, destroying the only record that the person ever trialled. After that, someone
who trialled then paid is indistinguishable from someone who paid up front. See *Backend
changes* for the fix, and note the card ships useful before that fix lands.

**States:** `free_trial_limit == 0` means the trial is disabled for that term — the card
renders "Free trial not enabled this term", never a bar of zeroes.

**Lane:** rail. **Source:** `GET …/dashboard/conversion`, new (see below).

## Role presets

Eight cards now. Trial Conversion is placed by who asks the question.

- **Ops** (`executive`, `tournament_director`) — Spotlight, Quick Actions, Event Activity,
  Leaderboard, Memberships, Engagement, Signup Timeline, Trial Conversion
- **Records** (`secretary`, `treasurer`) — **Trial Conversion**, Memberships, Signup
  Timeline, Spotlight, Event Activity, Engagement, Leaderboard, Quick Actions
- **Leadership** (`vice_president`, `president`, `webmaster`) — Engagement, Event Activity,
  Memberships, Trial Conversion, Spotlight, Signup Timeline, Leaderboard, Quick Actions

Records leads with Trial Conversion because it is the treasurer's question and gives that
preset a genuine lead card rather than Ops' leftovers. Any unmapped role falls back to Ops.

## Backend changes

**1. Conversion endpoint — new issue.** `GET /api/v2/semesters/:semesterId/dashboard/conversion`,
same semester-scoped group and `semester.get` permission as its siblings. Returns, over the
population of members who entered ≥1 event this term: `paid`, `trialSpent`, `trialOpen`,
`executive`, and the term's `freeTrialLimit`. `trialSpent` is unpaid, non-executive, with
entry count ≥ `free_trial_limit`; `trialOpen` is the same cohort below the limit. Follows
`DashboardRepository`; `InMemoryStore` returns `store.ErrNotImplemented` per epic decision 6.

**2. `memberships.converted_at` — new issue.** `ALTER TABLE memberships ADD COLUMN converted_at
timestamp NULL`, stamped at the unpaid → paid transition in `membership_service.go` where
`FreeTrialAvailable` is currently reset. Nullable, **no backfill** — the history genuinely
does not exist and inventing it would be worse. The real conversion rate becomes available
from the following term, and until then the card shows the snapshot. Same honest-empty
treatment as the signup timeline.

> **Migration caution** carries over from the epic: `server/internal/models/transaction.go`
> warns that Atlas diffs every model in the package. The generated migration must be read
> and confirmed to contain only the `ADD COLUMN`.

**3. Revert `paidPlayers` / `paidPlayerShare`** from `store/dashboard.go` and
`store/postgres/dashboard.go` — including the `BOOL_OR(m.paid)` added to the engagement
query — and from `dashboardApi.ts`. Superseded by the conversion endpoint.

## Hard-coded constants

The normal-range bands on card 6 — **median events played 2–3**, **played exactly once
33–44%** — come from the six-term points-system analysis, not from the API, which returns
only this term and one comparison term. They ship as named constants in the engagement card
with a comment citing the analysis and this spec.

This is a deliberate trade with a known cost: the bands will go stale and nothing will
alert anyone. Accepted because a president reading "median 2 events, ▼ −1" without a
baseline concludes the club is collapsing during an entirely normal term, which is the
specific failure this redesign exists to fix.

## Mock data

Every card ships with a fixture module exporting a realistic populated response, so the five
unbuilt cards are fully designed and visible before their endpoints exist, and so later work
has a reference render rather than a blank slate. Fixtures live beside each card as
`<Card>.fixtures.ts` and are used by Jest, by Storybook-less visual inspection, and as the
`placeholderData` for cards whose endpoint is not yet implemented.

Fixtures are internally consistent across cards — 967 memberships, 824 of whom played, 712
paid — so the page reads as one coherent term rather than eight unrelated samples. A card
rendering fixture data shows a visible "sample data" marker; it must never be mistakable for
a real figure.

## Testing

- **Jest, per card:** populated render; start-of-term empty state; error state; delta chips
  including the no-comparable-term path; the population qualifier visible on cards 6 and 8;
  `free_trial_limit == 0` rendering the disabled state on card 8.
- **Jest, layout:** the lane-assignment algorithm — lead promotion, wide/rail routing,
  relative order preserved within each lane, and single-lane collapse restoring exact preset
  order. Unmapped-role fallback to Ops.
- **Cypress (#443):** card order differs between an ops-role and a leadership-role login;
  one card's endpoint failing leaves its neighbours rendered.
- **Backend:** conversion endpoint against real Postgres via testcontainers — a term with
  the trial disabled, a term with no entries, an executive who played, a member who spent
  the trial and later paid.

## What this changes in existing code

| File | Change |
|---|---|
| `DashboardGrid.module.css` | Two-lane container-query layout; `dense` removed |
| `dashboardLayout.ts` | `CARD_TILES` → `lane` preference; eight-card presets |
| `Dashboard.tsx` | Lane assignment, lead-card promotion |
| `DashboardCard.module.css` | Sentence-case titles; fallback hexes dropped |
| `TermAtAGlanceCard` | Mini-stat grid → two composition bars |
| `EngagementRetentionCard` | Comparison table → distribution bar + normal-range tracks |
| `DeltaChip` | Unchanged |
| `store/dashboard.go`, `store/postgres/dashboard.go` | `paidPlayers` reverted |

## Out of scope

- Money. Budget and revenue belong to the Finances page (epic non-goal).
- Access control. Ordering changes only; no new authorizer.
- Events 3.0. Card 1 is built against today's event model.
- User-level customisation: no drag-to-reorder, no dismissable cards, no preferences table.
- Dark mode. No dark tokens exist.
