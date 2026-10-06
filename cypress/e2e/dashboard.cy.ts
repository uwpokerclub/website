const SWITCHED_SEMESTER_ID = "1f536973-5fb8-4c52-95fe-14cc479af31c";

const CARD_QAS = [
  "event-spotlight-card",
  "quick-actions-card",
  "term-at-a-glance-card",
  "signup-timeline-card",
  "event-activity-card",
  "engagement-retention-card",
  "leaderboard-card",
  "trial-conversion-card",
] as const;

const REAL_DASHBOARD_ALIASES = [
  "spotlight",
  "memberships",
  "engagement",
  "events",
  "signups",
  "conversion",
  "rankings",
] as const;

type RoleLayout = {
  username: string;
  /** Every card in the role's preset order, lead first — the grid's DOM and visual order. */
  order: readonly (typeof CARD_QAS)[number][];
};

const ROLE_LAYOUTS: Record<string, RoleLayout> = {
  Ops: {
    username: "dashboard_ops",
    order: [
      "event-spotlight-card",
      "quick-actions-card",
      "event-activity-card",
      "leaderboard-card",
      "term-at-a-glance-card",
      "engagement-retention-card",
      "signup-timeline-card",
      "trial-conversion-card",
    ],
  },
  Records: {
    username: "dashboard_records",
    order: [
      "trial-conversion-card",
      "term-at-a-glance-card",
      "signup-timeline-card",
      "event-spotlight-card",
      "event-activity-card",
      "engagement-retention-card",
      "leaderboard-card",
      "quick-actions-card",
    ],
  },
  Leadership: {
    username: "dashboard_leadership",
    order: [
      "engagement-retention-card",
      "event-activity-card",
      "term-at-a-glance-card",
      "trial-conversion-card",
      "event-spotlight-card",
      "signup-timeline-card",
      "leaderboard-card",
      "quick-actions-card",
    ],
  },
};

function interceptRealDashboardRequests() {
  cy.intercept("GET", /\/api\/v2\/semesters\/[^/]+\/dashboard\/spotlight$/).as("spotlight");
  cy.intercept("GET", /\/api\/v2\/semesters\/[^/]+\/dashboard\/memberships$/).as("memberships");
  cy.intercept("GET", /\/api\/v2\/semesters\/[^/]+\/dashboard\/engagement$/).as("engagement");
  cy.intercept("GET", /\/api\/v2\/semesters\/[^/]+\/dashboard\/events$/).as("events");
  cy.intercept("GET", /\/api\/v2\/semesters\/[^/]+\/dashboard\/signups$/).as("signups");
  cy.intercept("GET", /\/api\/v2\/semesters\/[^/]+\/dashboard\/conversion$/).as("conversion");
  cy.intercept("GET", /\/api\/v2\/semesters\/[^/]+\/rankings\?limit=5$/).as("rankings");
}

function waitForRealDashboardRequests() {
  return cy.wait(REAL_DASHBOARD_ALIASES.map((alias) => `@${alias}`)).then((interceptions) => {
    interceptions.forEach((interception) => expect(interception.response?.statusCode).to.equal(200));
  });
}

function waitForSiblingDashboardRequests(excluded: (typeof REAL_DASHBOARD_ALIASES)[number]) {
  return cy.wait(REAL_DASHBOARD_ALIASES.filter((alias) => alias !== excluded).map((alias) => `@${alias}`));
}

function visitDashboard(username = "e2e_user") {
  cy.resetDatabase();
  cy.login(username, "password");
  interceptRealDashboardRequests();
  cy.visit("/admin/dashboard");
  return waitForRealDashboardRequests();
}

function visitDashboardWithEventActivity(response: object) {
  cy.resetDatabase();
  cy.login("e2e_user", "password");
  interceptRealDashboardRequests();
  cy.intercept("GET", /\/api\/v2\/semesters\/[^/]+\/dashboard\/events$/, {
    statusCode: 200,
    body: response,
  }).as("eventActivityState");
  cy.visit("/admin/dashboard");
  cy.wait("@eventActivityState").its("response.statusCode").should("equal", 200);
  return waitForSiblingDashboardRequests("events");
}

function expectCardOrder(cards: readonly string[]) {
  cy.getByData("dashboard-grid")
    .find("section[data-qa]")
    .then(($cards) => {
      expect($cards.toArray().map((card) => card.getAttribute("data-qa"))).to.deep.equal(cards);
    });
}

describe("Dashboard", () => {
  describe("role layouts", () => {
    Object.entries(ROLE_LAYOUTS).forEach(([roleName, layout]) => {
      it(`shows all cards in the ${roleName} layout`, () => {
        visitDashboard(layout.username);

        CARD_QAS.forEach((cardQa) => cy.getByData(cardQa).should("have.length", 1));
        CARD_QAS.forEach((cardQa) => cy.getByData(cardQa).should("have.attr", "data-dashboard-status", "ready"));
        cy.getByData("trial-conversion-card").find('[data-qa="dashboard-card-error"]').should("not.exist");
        cy.getByData("dashboard-lead").children().should("have.attr", "data-qa", layout.order[0]);
        expectCardOrder(layout.order);
      });
    });
  });

  it("renders real conversion data without the sample fallback", () => {
    visitDashboard().then((interceptions) => {
      const conversion = interceptions[REAL_DASHBOARD_ALIASES.indexOf("conversion")].response?.body;
      expect(conversion).to.have.property("conversion");
    });

    cy.getByData("trial-conversion-card").scrollIntoView();
    cy.getByData("trial-conversion-card").find('[data-qa="dashboard-card-error"]').should("not.exist");
    cy.get('[data-testid="trial-conversion-figure"]').should("have.text", "0 players");
    cy.getByData("trial-conversion-card")
      .contains("used all 4 free entries and are still unpaid")
      .should("be.visible");
    cy.getByData("trial-conversion-card").contains("Winter 2024").should("be.visible");
    cy.getByData("trial-conversion-rate").contains("No tracked trial players yet").should("be.visible");
    cy.getByData("trial-conversion-card").contains("Conversion history unavailable").should("be.visible");
    cy.getByData("trial-conversion-card")
      .contains("Full-term observed trial cohort · paid status in the current historical snapshot")
      .should("be.visible");
    cy.getByData("trial-conversion-card")
      .contains("Earlier untracked trial and payment history is unavailable.")
      .should("be.visible");
  });

  it("loads seeded comparison and signup data from the real API", () => {
    visitDashboard().then((interceptions) => {
      const signups = interceptions[REAL_DASHBOARD_ALIASES.indexOf("signups")].response?.body;
      expect(signups.series).to.have.length.greaterThan(0);
      expect(signups.total).to.be.greaterThan(0);
      expect(signups.comparison.semester.name).to.equal("Winter 2024");
      expect(signups.comparison.dailyTotals).to.have.length.greaterThan(0);

      const memberships = interceptions[REAL_DASHBOARD_ALIASES.indexOf("memberships")].response?.body;
      expect(memberships.comparison.totalAsOf).to.equal(2);
      expect(memberships.comparison.stats.total).to.equal(2);
      cy.contains("Marker: Winter 2024 had 2 memberships by this point.")
        .scrollIntoView()
        .should("be.visible");

      const events = interceptions[REAL_DASHBOARD_ALIASES.indexOf("events")].response?.body;
      expect(events.current.eventsRun).to.equal(1);
      expect(events.comparison.averageFieldSize).to.equal(null);
      cy.contains("Current average:")
        .scrollIntoView()
        .should("be.visible");
      cy.contains("Winter 2024: no completed events by this point.")
        .scrollIntoView()
        .should("be.visible");
    });

    // Scroll to the caption itself: cards above can still settle after a scroll to the
    // card's top, which would leave the caption just below the fold.
    cy.getByData("signup-timeline-card")
      .should("have.attr", "data-dashboard-status", "ready")
      .contains("memberships created")
      .scrollIntoView()
      .should("be.visible");
    cy.getByData("signup-timeline-card").contains("Winter 2024 daily total").scrollIntoView().should("be.visible");
    cy.getByData("signup-timeline-card")
      .contains("the comparison line can omit undated memberships")
      .scrollIntoView()
      .should("be.visible");
  });

  it("shows the no-completed-events state for a scheduled-only response", () => {
    visitDashboardWithEventActivity({
      current: { eventsRun: 0, eventsScheduled: 1, totalEntries: 0, averageFieldSize: 0, series: [] },
      comparison: null,
    });

    cy.getByData("event-activity-card").contains("No completed events yet.").should("be.visible");
    cy.getByData("event-activity-card").contains("Current average:").should("not.exist");
  });

  it("shows a zero average when a completed event has no entries", () => {
    visitDashboardWithEventActivity({
      current: {
        eventsRun: 1,
        eventsScheduled: 0,
        totalEntries: 0,
        averageFieldSize: 0,
        series: [{ id: 99, name: "Empty completed event", startDate: "2025-01-12T19:00:00Z", entries: 0 }],
      },
      comparison: null,
    });

    cy.getByData("event-activity-card").contains("Current average: 0.0 players per completed event").should("be.visible");
    cy.getByData("event-activity-card").contains("No completed events yet.").should("not.exist");
  });

  it("isolates an Event Activity endpoint failure to its card", () => {
    cy.resetDatabase();
    cy.login("e2e_user", "password");
    interceptRealDashboardRequests();
    cy.intercept("GET", /\/api\/v2\/semesters\/[^/]+\/dashboard\/events$/, {
      statusCode: 500,
      body: { error: "dashboard event activity failed" },
    }).as("eventActivityFailure");
    cy.visit("/admin/dashboard");
    cy.wait("@eventActivityFailure");
    waitForSiblingDashboardRequests("events");

    cy.getByData("event-activity-card")
      .find('[data-qa="dashboard-card-error"]', { timeout: 10_000 })
      .should("be.visible");
    CARD_QAS.filter((cardQa) => cardQa !== "event-activity-card").forEach((cardQa) => {
      cy.getByData(cardQa)
        .should("have.attr", "data-dashboard-status", "ready")
        .find('[data-qa="dashboard-card-error"]')
        .should("not.exist");
    });
  });

  it("isolates a Trial Conversion endpoint failure to its card without using sample data", () => {
    cy.resetDatabase();
    cy.login("e2e_user", "password");
    interceptRealDashboardRequests();
    cy.intercept("GET", /\/api\/v2\/semesters\/[^/]+\/dashboard\/conversion$/, {
      statusCode: 500,
      body: { error: "dashboard conversion failed" },
    }).as("conversionFailure");
    cy.visit("/admin/dashboard");
    cy.wait("@conversionFailure");
    waitForSiblingDashboardRequests("conversion");

    cy.getByData("trial-conversion-card")
      .scrollIntoView()
      .find('[data-qa="dashboard-card-error"]', { timeout: 10_000 })
      .should("exist");
    CARD_QAS.filter((cardQa) => cardQa !== "trial-conversion-card").forEach((cardQa) => {
      cy.getByData(cardQa)
        .should("have.attr", "data-dashboard-status", "ready")
        .find('[data-qa="dashboard-card-error"]')
        .should("not.exist");
    });
  });

  it("refetches each real card for the selected semester", () => {
    visitDashboard();

    cy.getByData("semester-dropdown").click();
    cy.getByData(`semester-option-${SWITCHED_SEMESTER_ID}`).click();

    // Memberships is shared by Term at a Glance and Engagement, so this waits for
    // its one query rather than inventing a second request for the second consumer.
    waitForRealDashboardRequests().then((interceptions) => {
      interceptions.forEach((interception) => {
        expect(interception.request.url).to.contain(`/semesters/${SWITCHED_SEMESTER_ID}/`);
      });
    });
  });
});
