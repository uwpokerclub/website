/**
 * Real-browser layout checks for the dashboard grid. JSDOM has no layout engine, so
 * overflow, row geometry and chart sizing can only be verified here.
 */

type GridMode = "narrow" | "medium" | "wide";
type SidenavState = "expanded" | "collapsed";

const GAP_PX = 24; // the grid's 1.5rem gap
const TOLERANCE_PX = 1.5;

const DASHBOARD_ENDPOINTS = {
  spotlight: /\/api\/v2\/semesters\/[^/]+\/dashboard\/spotlight$/,
  memberships: /\/api\/v2\/semesters\/[^/]+\/dashboard\/memberships$/,
  engagement: /\/api\/v2\/semesters\/[^/]+\/dashboard\/engagement$/,
  events: /\/api\/v2\/semesters\/[^/]+\/dashboard\/events$/,
  signups: /\/api\/v2\/semesters\/[^/]+\/dashboard\/signups$/,
  conversion: /\/api\/v2\/semesters\/[^/]+\/dashboard\/conversion$/,
  rankings: /\/api\/v2\/semesters\/[^/]+\/rankings\?limit=5$/,
} as const;

type EndpointName = keyof typeof DASHBOARD_ENDPOINTS;

const ROLE_ORDERS = {
  Ops: {
    username: "dashboard_ops",
    order: [
      "spotlight",
      "quickActions",
      "eventActivity",
      "leaderboard",
      "termAtAGlance",
      "engagement",
      "signupTimeline",
      "trialConversion",
    ],
  },
  Records: {
    username: "dashboard_records",
    order: [
      "trialConversion",
      "termAtAGlance",
      "signupTimeline",
      "spotlight",
      "eventActivity",
      "engagement",
      "leaderboard",
      "quickActions",
    ],
  },
  Leadership: {
    username: "dashboard_leadership",
    order: [
      "engagement",
      "eventActivity",
      "termAtAGlance",
      "trialConversion",
      "spotlight",
      "signupTimeline",
      "leaderboard",
      "quickActions",
    ],
  },
} as const;

// Each viewport width, the sidenav states it has, and the grid layout the dashboard
// container should land in. At 768px and below the sidenav is an overlay, so there
// is only one state to check.
const VIEWPORTS: { width: number; modes: Partial<Record<SidenavState, GridMode>> }[] = [
  { width: 1400, modes: { expanded: "wide", collapsed: "wide" } },
  { width: 1280, modes: { expanded: "wide", collapsed: "wide" } },
  { width: 1024, modes: { expanded: "medium", collapsed: "medium" } },
  { width: 768, modes: { expanded: "medium" } },
  { width: 375, modes: { expanded: "narrow" } },
];

const LONG_TERM = "Fall 2025 Extended Intensive Semester with Additional Weekend Series";
const LONG_EVENT = "Fall 2026 Championship Series Event #5 — Pot Limit Omaha High-Low Eight or Better";

const LONG_LABEL_RESPONSES: Record<EndpointName, object> = {
  spotlight: {
    id: 5,
    name: LONG_EVENT,
    format: "Pot Limit Omaha Hi-Lo with a deliberately verbose format description",
    startDate: "2026-10-01T22:00:00Z",
    state: 1,
    entries: 1234,
    rebuys: 567,
  },
  memberships: {
    current: { total: 367, paid: 83, unpaid: 146, discounted: 126, executive: 12, new: 157, returning: 210 },
    comparison: {
      semester: { id: "c1", name: LONG_TERM },
      stats: { total: 917, paid: 400, unpaid: 300, discounted: 200, executive: 17, new: 500, returning: 417 },
      totalAsOf: 595,
    },
  },
  engagement: {
    current: { players: 366, medianEventsAttended: 2, playedOnceCount: 175, playedOnceShare: 0.48, tenPlusCount: 0 },
    comparison: {
      semester: { id: "c1", name: LONG_TERM },
      stats: { players: 500, medianEventsAttended: 3, playedOnceCount: 200, playedOnceShare: 0.4, tenPlusCount: 12 },
    },
  },
  events: {
    current: {
      eventsRun: 5,
      eventsScheduled: 2,
      totalEntries: 731,
      averageFieldSize: 146.2,
      series: [1, 2, 3, 4, 5].map((n) => ({
        id: n,
        name: `${LONG_EVENT} (${n})`,
        startDate: `2026-09-0${n}T22:00:00Z`,
        entries: 100 + n * 10,
      })),
    },
    comparison: { semester: { id: "c1", name: LONG_TERM }, averageFieldSize: 228.2 },
  },
  signups: {
    series: Array.from({ length: 30 }, (_, day) => ({
      date: `2026-09-${String(day + 1).padStart(2, "0")}`,
      admin: day % 7,
      discord: day % 3,
      unknown: day % 2,
    })),
    eventDates: ["2026-09-03", "2026-09-10", "2026-09-17"],
    dataStartsAt: "2026-09-01",
    total: 120,
    comparison: {
      semester: { id: "c1", name: LONG_TERM },
      dailyTotals: Array.from({ length: 30 }, (_, day) => ({ elapsedDay: day, total: day % 5 })),
    },
  },
  conversion: {
    current: { players: 366, paid: 209, trialSpent: 4, trialOpen: 141, executive: 12 },
    conversion: { numerator: 30, denominator: 120, rate: 0.25 },
    freeTrialLimit: 3,
    comparison: {
      semester: { id: "c1", name: LONG_TERM },
      stats: { players: 500, paid: 300, trialSpent: 20, trialOpen: 0, executive: 17 },
      conversion: { numerator: 0, denominator: 0, rate: null },
      freeTrialLimit: 3,
    },
  },
  rankings: {
    data: [1, 2, 3, 4, 5].map((position) => ({
      id: position,
      firstName: `Maximiliano-Alessandro${position}`,
      lastName: "Wolfeschlegelsteinhausenbergerdorff-Featherstonehaugh",
      points: 400 - position * 10,
      position,
    })),
    total: 5,
  },
};

const EMPTY_RESPONSES: Record<EndpointName, object | null> = {
  spotlight: null,
  memberships: {
    current: { total: 0, paid: 0, unpaid: 0, discounted: 0, executive: 0, new: 0, returning: 0 },
    comparison: null,
  },
  engagement: {
    current: { players: 0, medianEventsAttended: 0, playedOnceCount: 0, playedOnceShare: 0, tenPlusCount: 0 },
    comparison: null,
  },
  events: {
    current: { eventsRun: 0, eventsScheduled: 0, totalEntries: 0, averageFieldSize: 0, series: [] },
    comparison: null,
  },
  signups: { series: [], eventDates: [], dataStartsAt: null, total: 0, comparison: null },
  conversion: {
    current: { players: 0, paid: 0, trialSpent: 0, trialOpen: 0, executive: 0 },
    conversion: { numerator: 0, denominator: 0, rate: null },
    freeTrialLimit: 0,
    comparison: null,
  },
  rankings: { data: [], total: 0 },
};

function stubDashboard(responses: Partial<Record<EndpointName, object | null>>) {
  (Object.keys(DASHBOARD_ENDPOINTS) as EndpointName[]).forEach((name) => {
    // A bare null body would be sent empty; the spotlight's "no event" is the JSON literal.
    const body = responses[name] === null ? "null" : responses[name];
    cy.intercept("GET", DASHBOARD_ENDPOINTS[name], {
      statusCode: 200,
      body,
      headers: { "content-type": "application/json" },
    }).as(name);
  });
}

function visitAs(username: string) {
  cy.resetDatabase();
  cy.login(username, "password");
  cy.visit("/admin/dashboard");
  cy.getByData("dashboard-grid").should("exist");
}

function setLayout(width: number, sidenav: SidenavState) {
  cy.viewport(width, 900);
  if (width <= 768) return;

  const expectedWidth = sidenav === "expanded" ? 256 : 72;
  cy.getByData("sidenav").then(($nav) => {
    const collapsed = $nav[0].getBoundingClientRect().width < 160;
    if (collapsed !== (sidenav === "collapsed")) {
      cy.getByData("sidenav-toggle").click();
    }
  });
  // The sidenav animates its width; measure only once it has settled.
  cy.getByData("sidenav").should(($nav) => {
    expect($nav[0].getBoundingClientRect().width).to.be.closeTo(expectedWidth, 1);
  });
}

function describeElement(element: Element) {
  const qa = element.getAttribute("data-qa");
  const text = (element.textContent ?? "").trim().slice(0, 40);
  return `<${element.tagName.toLowerCase()}${qa ? ` data-qa="${qa}"` : ""}> "${text}"`;
}

const near = (actual: number, expected: number) => Math.abs(actual - expected) <= TOLERANCE_PX;
const px = (value: number) => `${Math.round(value * 10) / 10}px`;

/**
 * Every card fills its row exactly, rows run top to bottom in DOM order, and spans
 * match the expected mode. Problems are collected and asserted once, so a failure
 * lists everything wrong and the command log stays readable.
 */
function assertGridFlow(mode: GridMode, order: readonly string[]) {
  cy.getByData("dashboard-grid").should(($container) => {
    const doc = $container[0].ownerDocument;
    const rootFontPx = parseFloat(doc.defaultView!.getComputedStyle(doc.documentElement).fontSize);
    const grid = $container[0].firstElementChild as HTMLElement;
    const gridRect = grid.getBoundingClientRect();
    const widthRem = gridRect.width / rootFontPx;
    const actualMode: GridMode = widthRem >= 58 ? "wide" : widthRem >= 36 ? "medium" : "narrow";
    const cells = Array.from(grid.children) as HTMLElement[];
    const problems: string[] = [];

    if (actualMode !== mode) problems.push(`grid is ${actualMode} at ${px(gridRect.width)}, expected ${mode}`);

    const domOrder = cells.map((cell) => cell.dataset.card).join(", ");
    if (domOrder !== order.join(", ")) problems.push(`DOM order is ${domOrder}`);

    const columnPx = (gridRect.width - 11 * GAP_PX) / 12;
    let rowTop = -Infinity;
    let previousRight = gridRect.left - GAP_PX;

    cells.forEach((cell, index) => {
      const rect = cell.getBoundingClientRect();
      const span = mode === "narrow" ? 12 : Number(mode === "medium" ? cell.dataset.spanMedium : cell.dataset.spanWide);
      const expectedWidth = columnPx * span + GAP_PX * (span - 1);
      const label = cell.dataset.card;

      if (!near(rect.width, expectedWidth)) {
        problems.push(`${label} is ${px(rect.width)} wide, expected ${px(expectedWidth)} for span ${span}`);
      }

      if (near(rect.top, rowTop)) {
        if (!near(rect.left, previousRight + GAP_PX)) problems.push(`${label} leaves a gap after the previous card`);
      } else {
        if (index > 0 && !near(previousRight, gridRect.right)) {
          problems.push(`the row before ${label} stops ${px(gridRect.right - previousRight)} short of the edge`);
        }
        if (rect.top <= rowTop) problems.push(`${label} starts above the row before it`);
        if (!near(rect.left, gridRect.left)) problems.push(`${label} starts a row away from the left edge`);
        rowTop = rect.top;
      }
      previousRight = rect.right;
    });

    if (!near(previousRight, gridRect.right)) problems.push("the last row stops short of the right edge");

    expect(problems, `grid flow (${mode}, ${px(gridRect.width)})`).to.deep.equal([]);
  });
}

/** No horizontal page scroll, and nothing inside a card extends past it. */
function assertNoOverflow() {
  cy.getByData("dashboard-grid").should(($container) => {
    const problems: string[] = [];

    let ancestor: HTMLElement | null = $container[0];
    while (ancestor) {
      if (ancestor.scrollWidth > ancestor.clientWidth + 1) {
        problems.push(`${describeElement(ancestor)} scrolls horizontally by ${ancestor.scrollWidth - ancestor.clientWidth}px`);
      }
      ancestor = ancestor.parentElement;
    }

    $container[0].querySelectorAll("section[data-qa]").forEach((card) => {
      const cardRect = card.getBoundingClientRect();
      const cardQa = card.getAttribute("data-qa");

      if (card.scrollWidth > card.clientWidth + 1) problems.push(`${cardQa} scrolls horizontally`);
      card.querySelectorAll("*").forEach((element) => {
        // A closed <details> still lays out its content in Chrome; it is not on screen.
        if (element.closest("details:not([open])") && element.tagName !== "SUMMARY") return;
        // A data table may be wider than a phone; it scrolls inside its own region,
        // and only that region has to stay within the card.
        const tableRegion = element.closest('[role="region"][data-qa$="-table"]');
        if (tableRegion && tableRegion !== element) return;
        const rect = element.getBoundingClientRect();
        // sr-only text is a 1px box by design.
        if (rect.width <= 1 || rect.height <= 1) return;

        if (rect.left < cardRect.left - 1 || rect.right > cardRect.right + 1) {
          problems.push(`${describeElement(element)} extends past ${cardQa}`);
        }
      });
    });

    expect(problems, "overflow").to.deep.equal([]);
  });
}

/** Charts fill their wrapper's width after a resize rather than keeping a stale size. */
function assertChartsFit() {
  // Loading, empty and error states have no chart; only charts that rendered are checked.
  cy.getByData("dashboard-grid").should(($container) => {
    const problems: string[] = [];

    $container[0].querySelectorAll('[data-qa$="-chart"]').forEach((wrapper) => {
      const chartQa = wrapper.getAttribute("data-qa");
      const svg = wrapper.querySelector("svg");
      const wrapperWidth = wrapper.getBoundingClientRect().width;

      if (!svg) problems.push(`${chartQa} has no svg`);
      else if (Math.abs(svg.getBoundingClientRect().width - wrapperWidth) > 2) {
        problems.push(`${chartQa} is ${px(svg.getBoundingClientRect().width)} in a ${px(wrapperWidth)} wrapper`);
      }
    });

    expect(problems, "chart sizing").to.deep.equal([]);
  });
}

function checkEveryLayout(order: readonly string[]) {
  VIEWPORTS.forEach(({ width, modes }) => {
    (Object.entries(modes) as [SidenavState, GridMode][]).forEach(([sidenav, mode]) => {
      cy.log(`${width}px, sidenav ${sidenav}`);
      setLayout(width, sidenav);
      assertGridFlow(mode, order);
      assertNoOverflow();
      assertChartsFit();
    });
  });
}

describe("Dashboard layout", () => {
  context("role layouts with real data", () => {
    (Object.entries(ROLE_ORDERS) as [string, (typeof ROLE_ORDERS)[keyof typeof ROLE_ORDERS]][]).forEach(
      ([roleName, { username, order }]) => {
        it(`flows the ${roleName} cards into full rows at every width and sidenav state`, () => {
          visitAs(username);
          cy.get('[data-dashboard-status="loading"]').should("not.exist");

          checkEveryLayout(order);
        });
      },
    );
  });

  context("content edge cases", () => {
    const leadership = ROLE_ORDERS.Leadership;

    it("wraps long comparison text, names and labels without overflowing", () => {
      stubDashboard(LONG_LABEL_RESPONSES);
      visitAs(leadership.username);
      cy.get('[data-dashboard-status="loading"]').should("not.exist");

      checkEveryLayout(leadership.order);

      // The whole comparison sentence is on screen: it wraps inside the card rather
      // than running out of it or being cut off.
      cy.viewport(375, 900);
      cy.getByData("term-at-a-glance-card")
        .find('[data-qa="delta-chip"]')
        .should("contain.text", `vs memberships by this point in ${LONG_TERM}`)
        .and(($chip) => {
          const chip = $chip[0].getBoundingClientRect();
          const card = $chip[0].closest("section")!.getBoundingClientRect();
          expect(chip.right).to.be.at.most(card.right);
          expect(chip.height, "chip wraps onto more lines").to.be.greaterThan(30);
        });
      cy.getByData("leaderboard-card").should("contain.text", "Wolfeschlegelsteinhausenbergerdorff-Featherstonehaugh");
    });

    it("still shows a chart tooltip on hover, inside the chart, after the chart narrows", () => {
      stubDashboard(LONG_LABEL_RESPONSES);
      visitAs(leadership.username);
      cy.get('[data-dashboard-status="loading"]').should("not.exist");
      setLayout(1400, "expanded");
      setLayout(1024, "collapsed");

      // Recharts renders its own markup, so these are its class names rather than data-qa hooks.
      cy.getByData("event-activity-chart").find(".recharts-bar-rectangle").last().trigger("mousemove");
      cy.getByData("event-activity-chart")
        .find(".recharts-tooltip-wrapper")
        .should("contain.text", `${LONG_EVENT} (5)`)
        .and(($tooltip) => {
          const tooltip = $tooltip[0].getBoundingClientRect();
          const chart = $tooltip[0].closest('[data-qa="event-activity-chart"]')!.getBoundingClientRect();
          expect(tooltip.left).to.be.at.least(chart.left - 1);
          expect(tooltip.right).to.be.at.most(chart.right + 1);
        });
    });

    it("keeps the grid intact while every card is loading", () => {
      (Object.keys(DASHBOARD_ENDPOINTS) as EndpointName[]).forEach((name) => {
        cy.intercept("GET", DASHBOARD_ENDPOINTS[name], (request) => {
          request.on("response", (response) => {
            response.setDelay(60_000);
          });
        });
      });
      visitAs(leadership.username);
      cy.get('[data-dashboard-status="loading"]').should("have.length", 7);

      checkEveryLayout(leadership.order);
    });

    it("keeps the grid intact when every card is empty", () => {
      stubDashboard(EMPTY_RESPONSES);
      visitAs(leadership.username);
      cy.get('[data-dashboard-status="loading"]').should("not.exist");
      cy.getByData("event-activity-card").should("have.attr", "data-dashboard-status", "empty");

      checkEveryLayout(leadership.order);
    });

    it("keeps the grid intact when every endpoint fails", () => {
      (Object.keys(DASHBOARD_ENDPOINTS) as EndpointName[]).forEach((name) => {
        cy.intercept("GET", DASHBOARD_ENDPOINTS[name], { statusCode: 500, body: { error: "unavailable" } });
      });
      visitAs(leadership.username);
      cy.get('[data-dashboard-status="error"]', { timeout: 15_000 }).should("have.length", 7);

      checkEveryLayout(leadership.order);
    });

    it("keeps the data tables inside their cards on a phone", () => {
      // Load at phone width, as a phone would: shrinking a desktop page instead keeps
      // the expanded sidenav, which opens as an overlay over the cards.
      cy.viewport(375, 900);
      stubDashboard(LONG_LABEL_RESPONSES);
      visitAs(leadership.username);
      cy.get('[data-dashboard-status="loading"]').should("not.exist");

      // The mobile header is fixed, so leave room above each summary before clicking it.
      cy.getByData("event-activity-card")
        .contains("summary", "View event entry data")
        .scrollIntoView({ offset: { top: -200, left: 0 } })
        .click();
      cy.getByData("signup-timeline-card")
        .contains("summary", "View daily signup data")
        .scrollIntoView({ offset: { top: -200, left: 0 } })
        .click();
      cy.getByData("signup-timeline-card").find("details").should("have.attr", "open");

      assertNoOverflow();
      // Six columns cannot fit a phone, so the table scrolls in a keyboard-reachable region.
      cy.getByData("signup-timeline-table")
        .should("have.attr", "tabindex", "0")
        .and("have.attr", "aria-labelledby")
        .then(() => {
          cy.getByData("signup-timeline-table").should(($region) => {
            expect($region[0].scrollWidth).to.be.greaterThan($region[0].clientWidth);
          });
        });
    });
  });
});
