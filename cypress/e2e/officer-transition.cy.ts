const roles = [
  "president",
  "vice_president",
  "secretary",
  "treasurer",
] as const;
const questIds = ["hdrust0", "dhousegoe1", "eaucock2", "kduckham3"];
const nomineeNames = [
  "Heinrik Drust",
  "Doretta Housegoe",
  "Elita Aucock",
  "Khalil Duckham",
];
const seededPassword = "password";

describe("Officer transition", () => {
  beforeEach(() => {
    Cypress.session.clearAllSavedSessions();
    cy.resetDatabase();
  });

  const stageDialog = () => cy.getByData("officer-transition-modal");
  const receiptDialog = () => cy.getByData("officer-transition-receipt");

  const openTransition = (clipboardUnavailable = false) => {
    cy.login("test_president", seededPassword);
    cy.intercept("GET", "/api/v2/members?*").as("resolveQuestId");
    cy.intercept("POST", "/api/v2/officer-transitions").as("stageTransition");
    cy.visit("/admin/executive", {
      onBeforeLoad(window) {
        if (!clipboardUnavailable) return;
        Object.defineProperty(window.navigator, "clipboard", {
          configurable: true,
          value: {
            writeText: () => Promise.reject(new Error("denied")),
          },
        });
      },
    });
    cy.getByData("officer-transition-start").click();
  };

  const resolveAll = () => {
    roles.forEach((role, index) => {
      cy.getByData(`officer-transition-${role}`).type(questIds[index]).blur();
      cy.wait("@resolveQuestId");
      stageDialog()
        .find(`[data-qa="officer-transition-${role}-name"]`)
        .should("contain", nomineeNames[index]);
    });
  };

  it("stages a transition while the outgoing president remains authorized", () => {
    openTransition();
    resolveAll();
    cy.getByData("officer-transition-review").click();
    cy.contains("No access changes yet.").should("be.visible");
    cy.getByData("officer-transition-submit").click();
    cy.wait("@stageTransition").its("response.statusCode").should("eq", 201);
    receiptDialog().should("have.attr", "aria-modal", "true");
    cy.location("origin").then((origin) => {
      receiptDialog()
        .find('[data-qa="officer-transition-link-president"]')
        .should("have.attr", "readonly");
      receiptDialog()
        .find('[data-qa="officer-transition-link-president"]')
        .invoke("val")
        .should((value) => {
          const link = new URL(String(value));
          expect(link.origin).to.equal(origin);
          expect(link.pathname).to.equal("/activate");
          expect(link.hash).to.match(/^#token=[A-Za-z0-9_-]+$/);
        });
    });
    cy.request("/api/v2/session").its("body.role").should("eq", "president");
  });

  it("keeps the pending team visible after president activation so remaining officers can activate", () => {
    openTransition();
    resolveAll();
    cy.getByData("officer-transition-review").click();
    cy.getByData("officer-transition-submit").click();
    cy.wait("@stageTransition").its("response.statusCode").should("eq", 201);
    cy.on("window:confirm", () => true);
    cy.getByData("officer-transition-receipt-close").click();

    cy.getByData("pending-transition-banner").should("be.visible");
    cy.getByData("pending-transition-nominees").should("contain", "Heinrik Drust");
    cy.getByData("pending-transition-president").should("contain", "Activation required");
    cy.getByData("pending-transition-banner")
      .should("contain", "Current access transfers when Heinrik Drust activates.")
      .and("contain", "until every incoming officer has activated their account.");
    cy.getByData("pending-transition-reissue-president").click();
    cy.getByData("pending-transition-link-president")
      .invoke("val")
      .should("be.a", "string")
      .then((link) => {
        cy.visit(String(link));
      });

    cy.getByData("activation-heading").should("contain", "Set your password");
    cy.getByData("activation-password").type("replacement password");
    cy.getByData("activation-confirm-password").type("replacement password");
    cy.getByData("activation-submit").click();
    cy.location("pathname").should("eq", "/admin/dashboard");

    cy.visit("/admin/executive");
    cy.getByData("pending-transition-banner").should("be.visible");
    cy.getByData("pending-transition-president").should("contain", "Account already active");
    cy.getByData("pending-transition-cancel").should("not.exist");
  });

  it("removes the pending banner after a confirmed cancellation", () => {
    openTransition();
    resolveAll();
    cy.getByData("officer-transition-review").click();
    cy.getByData("officer-transition-submit").click();
    cy.wait("@stageTransition").its("response.statusCode").should("eq", 201);
    cy.on("window:confirm", () => true);
    cy.getByData("officer-transition-receipt-close").click();
    cy.getByData("pending-transition-banner").should("be.visible");

    cy.intercept("POST", /\/api\/v2\/officer-transitions\/[^/]+\/cancel/).as("cancelTransition");
    cy.getByData("pending-transition-cancel").click();
    cy.wait("@cancelTransition").its("response.statusCode").should("eq", 204);
    cy.getByData("pending-transition-banner").should("not.exist");
    cy.getByData("officer-transition-section").should("exist");
  });

  it("keeps an ambiguity error inline", () => {
    openTransition();
    cy.intercept("GET", "/api/v2/members?questId=hdrust0&limit=2", {
      body: {
        data: [{ questId: "hdrust0" }, { questId: "hdrust0" }],
        total: 2,
      },
    }).as("ambiguousQuestId");
    cy.getByData("officer-transition-president").type("hdrust0").blur();
    cy.wait("@ambiguousQuestId");
    cy.contains(
      "President Quest ID must resolve to exactly one member.",
    ).should("be.visible");
  });

  it("rejects duplicate Quest IDs before staging", () => {
    openTransition();
    cy.getByData("officer-transition-president").type("hdrust0").blur();
    cy.wait("@resolveQuestId");
    cy.getByData("officer-transition-vice_president").type("HDRUST0").blur();
    cy.contains("Quest IDs must be distinct.").should("be.visible");
    cy.getByData("officer-transition-review").should("be.disabled");

    cy.getByData("officer-transition-president").clear().type("alibbis4").blur();
    cy.wait("@resolveQuestId");
    cy.contains("Quest IDs must be distinct.").should("not.exist");
  });

  it("keeps resolved form values after a pending-transition conflict", () => {
    openTransition();
    resolveAll();
    cy.intercept("POST", "/api/v2/officer-transitions", {
      statusCode: 409,
      body: { message: "an officer transition is already pending" },
    }).as("pendingConflict");
    cy.getByData("officer-transition-review").click();
    cy.getByData("officer-transition-submit").click();
    cy.wait("@pendingConflict");
    cy.getByData("officer-transition-error").should(
      "contain",
      "already pending",
    );
    cy.getByData("officer-transition-president").should(
      "have.value",
      "hdrust0",
    );
    stageDialog()
      .find('[data-qa="officer-transition-president-name"]')
      .should("contain", "Heinrik Drust");
  });

  it("offers manual copy guidance when the clipboard is unavailable", () => {
    openTransition(true);
    resolveAll();
    cy.intercept("POST", "/api/v2/officer-transitions", {
      fixture: "officer-transition.json",
    }).as("stagedTransition");
    cy.getByData("officer-transition-review").click();
    cy.getByData("officer-transition-submit").click();
    cy.wait("@stagedTransition");
    receiptDialog()
      .find('[data-qa="officer-transition-copy-president"]')
      .click();
    cy.contains(
      '[role="alert"]',
      "Copy failed. Select and copy the link manually.",
    ).should("exist");
  });

  it("redirects a tournament director away from executive management", () => {
    cy.login("e2e_user", seededPassword);
    cy.request("PATCH", "/api/v2/logins/test_executive", {
      role: "tournament_director",
    });
    cy.clearCookies();
    cy.login("test_executive", seededPassword);
    cy.visit("/admin/executive");
    cy.location("pathname").should("eq", "/admin/dashboard");
    cy.getByData("sidenav-officers-section").should("not.exist");
  });
});
