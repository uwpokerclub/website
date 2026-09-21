describe("SemesterSetupWizard fee prefill", () => {
  beforeEach(() => {
    cy.resetDatabase();
    cy.login();
    cy.intercept("GET", "/api/v2/semesters", {
      delay: 500,
      body: {
        data: [
          {
            id: "previous-semester",
            name: "Winter 2027",
            startDate: "2027-01-01T00:00:00Z",
            endDate: "2027-04-30T00:00:00Z",
            startingBudget: 250,
            currentBudget: 250,
            membershipFee: 15,
            membershipDiscountFee: 10,
            rebuyFee: 0,
            freeTrialLimit: 4,
            meta: "",
          },
        ],
        total: 1,
      },
    }).as("getSemesters");
    cy.visit("/admin/dashboard");
  });

  it("prefills fees when opened before the semester query resolves", () => {
    cy.getByData("semester-dropdown").click();
    cy.getByData("create-semester-btn").click();
    cy.getByData("semester-setup-wizard").should("exist");
    cy.wait("@getSemesters");

    cy.getByData("semester-term").select("spring");
    cy.getByData("input-semester-startDate").type("2027-05-01");
    cy.getByData("input-semester-endDate").type("2027-08-31");
    cy.getByData("semester-wizard-next-btn").click();

    cy.getByData("input-semester-startingBudget").should("have.value", "250");
    cy.getByData("input-semester-membershipFee").should("have.value", "15");
    cy.getByData("input-semester-membershipDiscountFee").should("have.value", "10");
    cy.getByData("input-semester-rebuyFee").should("have.value", "0");
    cy.getByData("input-semester-freeTrialLimit").should("have.value", "4");
  });
});
