describe("Semester setup prompt", () => {
  const pastSemester = {
    id: "past-semester",
    name: "Winter 2026",
    meta: "",
    startDate: "2026-01-01T00:00:00Z",
    endDate: "2026-04-30T00:00:00Z",
    startingBudget: 0,
    currentBudget: 0,
    membershipFee: 10,
    membershipDiscountFee: 5,
    rebuyFee: 2,
    freeTrialLimit: 0,
  };

  const visitDashboard = (semesters: unknown[], statusCode = 200, role = "president") => {
    cy.intercept("GET", "/api/v2/session", {
      username: "test_president",
      role,
      permissions: { semester: { create: true, get: true, list: true, edit: true } },
    });
    cy.intercept("GET", "/api/v2/semesters", {
      statusCode,
      body: statusCode === 200 ? { data: semesters, total: semesters.length } : { message: "Unavailable" },
    }).as("getSemesters");
    cy.intercept("GET", "/api/v2/officer-transitions/completed", {
      id: "completed-transition",
      status: "completed",
      presidentUsername: "test_president",
    }).as("getCompletedTransition");
    cy.visit("/admin/dashboard");
    cy.wait("@getSemesters");
    cy.wait("@getCompletedTransition");
  };

  beforeEach(() => {
    cy.resetDatabase();
    cy.clearAllSessionStorage();
  });

  it("appears only when the president has no future semester", () => {
    visitDashboard([]);
    cy.getByData("semester-setup-prompt").should("be.visible");

    visitDashboard([{ ...pastSemester, endDate: "2099-12-31T00:00:00Z" }]);
    cy.getByData("semester-setup-prompt").should("not.exist");
  });

  it("ignores malformed legacy ranges and remains usable when the read fails", () => {
    visitDashboard([{ ...pastSemester, startDate: "2027-01-01T00:00:00Z", endDate: "2026-04-30T00:00:00Z" }]);
    cy.getByData("semester-setup-prompt").should("be.visible");

    visitDashboard([], 500);
    cy.getByData("semester-setup-prompt").should("not.exist");
    cy.getByData("sidenav").should("be.visible");
  });

  it("dismisses locally without blocking the dashboard", () => {
    visitDashboard([]);
    cy.getByData("semester-setup-prompt-dismiss").click();
    cy.getByData("semester-setup-prompt").should("not.exist");
    cy.getByData("sidenav").should("be.visible");
  });

  it("never renders for a non-president", () => {
    visitDashboard([], 200, "executive");
    cy.getByData("semester-setup-prompt").should("not.exist");
  });

  it("opens the existing wizard and clears after it creates a semester", () => {
    visitDashboard([]);
    cy.intercept("POST", "/api/v2/semesters", (request) => {
      request.reply({ statusCode: 201, body: { ...pastSemester, id: "created-semester", name: "Fall 2027", endDate: "2027-12-31T00:00:00Z" } });
    }).as("createSemester");

    cy.getByData("semester-setup-prompt-open").click();
    cy.getByData("semester-setup-wizard").should("be.visible");
    cy.getByData("semester-term").select("fall");
    cy.getByData("input-semester-startDate").type("2027-09-01");
    cy.getByData("input-semester-endDate").type("2027-12-31");
    cy.getByData("semester-wizard-next-btn").click();
    cy.getByData("semester-wizard-next-btn").click();
    cy.getByData("create-semester-submit-btn").click();
    cy.wait("@createSemester");
    cy.getByData("semester-setup-prompt").should("not.exist");
  });

  it("has no axe violations for the prompt and wizard", () => {
    visitDashboard([]);
    cy.injectAxe();
    cy.checkA11y('[data-qa="semester-setup-prompt"]');
    cy.getByData("semester-setup-prompt-open").click();
    cy.checkA11y('[data-qa="semester-setup-wizard"]');
  });
});
