const roles = [
  "president",
  "vice_president",
  "secretary",
  "treasurer",
] as const;
const questIds = ["hdrust0", "dhousegoe1", "eaucock2", "kduckham3"];
const seededPassword = "password";

describe("Officer transition", () => {
  beforeEach(() => {
    cy.resetDatabase();
  });

  const openTransition = () => {
    cy.login("test_president", seededPassword);
    cy.intercept("GET", "/api/v2/members?*").as("resolveQuestId");
    cy.intercept("POST", "/api/v2/officer-transitions").as("stageTransition");
    cy.visit("/admin/executive");
    cy.getByData("officer-transition-start").click();
  };

  const resolveAll = () => {
    roles.forEach((role, index) => {
      cy.getByData(`officer-transition-${role}`).type(questIds[index]).blur();
      cy.wait("@resolveQuestId");
      cy.getByData(`officer-transition-${role}-name`).should("be.visible");
    });
  };

  it("stages a transition while the outgoing president remains authorized", () => {
    openTransition();
    resolveAll();
    cy.getByData("officer-transition-review").click();
    cy.contains("No access changes yet.").should("be.visible");
    cy.getByData("officer-transition-submit").click();
    cy.wait("@stageTransition").its("response.statusCode").should("eq", 201);
    cy.getByData("officer-transition-receipt").should("be.visible");
    cy.getByData("officer-transition-link-president")
      .should("have.value")
      .and("contain", "/activate#token=");
    cy.request("/api/v2/session").its("body.role").should("eq", "president");
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
    cy.getByData("officer-transition-president-name").should("be.visible");
  });

  it("offers manual copy guidance when the clipboard is unavailable", () => {
    openTransition();
    resolveAll();
    cy.intercept("POST", "/api/v2/officer-transitions", {
      fixture: "officer-transition.json",
    }).as("stagedTransition");
    cy.getByData("officer-transition-review").click();
    cy.getByData("officer-transition-submit").click();
    cy.wait("@stagedTransition");
    cy.window().then((window) =>
      cy
        .stub(window.navigator.clipboard, "writeText")
        .rejects(new Error("denied")),
    );
    cy.getByData("officer-transition-copy-president").click();
    cy.contains("Copy failed. Select and copy the link manually.").should(
      "be.visible",
    );
  });

  it("keeps the executive page accessible but hides transition controls from a tournament director", () => {
    cy.login("e2e_user", seededPassword);
    cy.request("PATCH", "/api/v2/logins/test_executive", {
      role: "tournament_director",
    });
    cy.clearCookies();
    cy.login("test_executive", seededPassword);
    cy.visit("/admin/executive");
    cy.getByData("officer-transition-section").should("not.exist");
  });
});
