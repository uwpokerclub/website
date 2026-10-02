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
  lead: (typeof CARD_QAS)[number];
  wide: readonly (typeof CARD_QAS)[number][];
  rail: readonly (typeof CARD_QAS)[number][];
};

const ROLE_LAYOUTS: Record<string, RoleLayout> = {
  Ops: {
    username: "test_executive",
    lead: "event-spotlight-card",
    wide: ["event-activity-card", "engagement-retention-card", "signup-timeline-card"],
    rail: ["quick-actions-card", "leaderboard-card", "term-at-a-glance-card", "trial-conversion-card"],
  },
  Records: {
    username: "dashboard_records",
    lead: "trial-conversion-card",
    wide: ["signup-timeline-card", "event-activity-card", "engagement-retention-card"],
    rail: ["term-at-a-glance-card", "event-spotlight-card", "leaderboard-card", "quick-actions-card"],
  },
  Leadership: {
    username: "dashboard_leadership",
    lead: "engagement-retention-card",
    wide: ["event-activity-card", "signup-timeline-card"],
    rail: [
      "term-at-a-glance-card",
      "trial-conversion-card",
      "event-spotlight-card",
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

function visitDashboard(username = "e2e_user") {
  cy.resetDatabase();
  cy.login(username, "password");
  interceptRealDashboardRequests();
  cy.visit("/admin/dashboard");
  return waitForRealDashboardRequests();
}

function expectCardOrder(containerQa: string, cards: readonly string[]) {
  cy.getByData(containerQa)
    .children()
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
        cy.getByData("trial-conversion-card").contains("Sample data").should("not.exist");
        cy.getByData("trial-conversion-card").find('[data-qa="dashboard-card-error"]').should("not.exist");
        cy.getByData("dashboard-lead").children().should("have.attr", "data-qa", layout.lead);
        expectCardOrder("dashboard-wide-stack", layout.wide);
        expectCardOrder("dashboard-rail-lane", layout.rail);
      });
    });
  });

  it("renders real conversion data without the sample fallback", () => {
    visitDashboard();

    cy.getByData("trial-conversion-card").scrollIntoView();
    cy.getByData("trial-conversion-card").contains("Sample data").should("not.exist");
    cy.getByData("trial-conversion-card").find('[data-qa="dashboard-card-error"]').should("not.exist");
    cy.get('[data-testid="trial-conversion-figure"]').should("have.text", "0 players");
    cy.getByData("trial-conversion-card")
      .contains("used all 4 free entries and are still unpaid")
      .should("be.visible");
    cy.getByData("trial-conversion-card").contains("No comparable term to compare against yet.").should("be.visible");
    cy.getByData("trial-conversion-rate").contains("No tracked trial players yet").should("be.visible");
    cy.getByData("trial-conversion-card")
      .contains("Only includes trials recorded since tracking began. Earlier trial and payment history is unavailable.")
      .should("be.visible");
    cy.getByData("trial-conversion-card").contains(/untracked|observed-ever|purchase proxy/i).should("not.exist");
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

    cy.getByData("event-activity-card")
      .find('[data-qa="dashboard-card-error"]', { timeout: 10_000 })
      .should("be.visible");
    CARD_QAS.filter((cardQa) => cardQa !== "event-activity-card").forEach((cardQa) => {
      cy.getByData(cardQa).should("exist").find('[data-qa="dashboard-card-error"]').should("not.exist");
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

    cy.getByData("trial-conversion-card")
      .scrollIntoView()
      .find('[data-qa="dashboard-card-error"]', { timeout: 10_000 })
      .should("be.visible");
    cy.getByData("trial-conversion-card").contains("Sample data").should("not.exist");
    CARD_QAS.filter((cardQa) => cardQa !== "trial-conversion-card").forEach((cardQa) => {
      cy.getByData(cardQa).should("exist").find('[data-qa="dashboard-card-error"]').should("not.exist");
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

  context("responsive lanes", () => {
    beforeEach(() => visitDashboard());

    it("uses two lanes when the dashboard container is wide", () => {
      cy.viewport(1440, 900);
      cy.getByData("dashboard-wide-lane").then(($wide) => {
        cy.getByData("dashboard-rail-lane").then(($rail) => {
          const wide = $wide[0].getBoundingClientRect();
          const rail = $rail[0].getBoundingClientRect();

          expect(rail.left).to.be.greaterThan(wide.left);
          expect(rail.top).to.equal(wide.top);
        });
      });
    });

    it("collapses to one lane when the dashboard container is narrow", () => {
      cy.viewport(768, 900);
      cy.getByData("dashboard-wide-lane").then(($wide) => {
        cy.getByData("dashboard-rail-lane").then(($rail) => {
          const wide = $wide[0].getBoundingClientRect();
          const rail = $rail[0].getBoundingClientRect();

          expect(rail.top).to.be.greaterThan(wide.bottom);
          expect(rail.left).to.equal(wide.left);
        });
      });
    });
  });
});
