describe("SemesterSetupWizard", () => {
  const openWizard = () => {
    cy.getByData("semester-dropdown").click();
    cy.getByData("create-semester-btn").click();
    cy.getByData("semester-setup-wizard").should("exist");
  };

  const fillTermAndDates = (term: "fall" | "winter" | "spring", startDate: string, endDate: string) => {
    cy.getByData("semester-term").select(term);
    cy.getByData("input-semester-startDate").type(startDate);
    cy.getByData("input-semester-endDate").type(endDate);
  };

  const continueToFees = () => cy.getByData("semester-wizard-next-btn").click();

  const continueToReview = () => {
    continueToFees();
    cy.getByData("semester-wizard-next-btn").click();
  };

  beforeEach(() => {
    cy.resetDatabase();
    cy.login();
    cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
    cy.visit("/admin/dashboard");
    cy.getByData("sidenav").should("exist");
  });

  it("opens from the semester selector and resets when cancelled", () => {
    openWizard();
    cy.getByData("semester-term").select("winter");
    cy.getByData("semester-wizard-cancel-btn").click();
    cy.getByData("semester-setup-wizard").should("not.exist");

    openWizard();
    cy.getByData("semester-term").should("contain", "Select a term");
  });

  it("keeps the generated semester name out of the term-and-dates step", () => {
    openWizard();
    fillTermAndDates("fall", "2027-09-01", "2027-12-31");
    cy.getByData("semester-derived-name").should("not.exist");
  });

  it("does not expose a free-text name input", () => {
    openWizard();
    cy.getByData("input-semester-name").should("not.exist");
  });

  it("blocks step one until its fields are valid, including date order", () => {
    openWizard();
    continueToFees();
    cy.contains("Select a term").should("exist");

    fillTermAndDates("fall", "2027-12-31", "2027-09-01");
    continueToFees();
    cy.contains("End date must be after start date").should("exist");
    cy.getByData("semester-wizard-step-term-dates").should("exist");
  });

  it("prefills editable fees from the semester with the latest start date", () => {
    cy.intercept("POST", "/api/v2/semesters").as("createSemester");
    cy.intercept("GET", "/api/v2/semesters", {
      data: [
        {
          id: "older-semester",
          name: "Fall 2026",
          startDate: "2026-09-01T00:00:00Z",
          endDate: "2026-12-31T00:00:00Z",
          startingBudget: 100,
          currentBudget: 100,
          membershipFee: 10,
          membershipDiscountFee: 5,
          rebuyFee: 2,
          freeTrialLimit: 0,
          meta: "",
        },
        {
          id: "latest-semester",
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
      total: 2,
    }).as("getLatestSemesters");
    cy.visit("/admin/dashboard");
    cy.wait("@getLatestSemesters");

    openWizard();
    fillTermAndDates("spring", "2027-05-01", "2027-08-31");
    continueToFees();

    cy.getByData("input-semester-startingBudget").should("have.value", "250");
    cy.getByData("input-semester-membershipFee").should("have.value", "15");
    cy.getByData("input-semester-membershipDiscountFee").should("have.value", "10");
    cy.getByData("input-semester-rebuyFee").should("have.value", "0");
    cy.getByData("input-semester-freeTrialLimit").should("have.value", "4");

    cy.getByData("input-semester-membershipFee").clear().type("20");
    cy.getByData("semester-wizard-next-btn").click();
    cy.getByData("semester-review-membershipFee").should("have.text", "$20");
    cy.getByData("create-semester-submit-btn").click();

    cy.wait("@createSemester").then((interception) => {
      expect(interception.request.body).to.deep.include({
        membershipFee: 20,
        rebuyFee: 0,
      });
      expect(interception.response?.statusCode).to.eq(201);
    });
  });

  it("uses hardcoded fee defaults when no previous semester exists", () => {
    cy.intercept("GET", "/api/v2/semesters", { data: [], total: 0 }).as("getEmptySemesters");
    cy.visit("/admin/dashboard");
    cy.wait("@getEmptySemesters");

    openWizard();
    fillTermAndDates("spring", "2027-05-01", "2027-08-31");
    continueToFees();

    cy.getByData("input-semester-startingBudget").should("have.value", "0");
    cy.getByData("input-semester-membershipFee").should("have.value", "10");
    cy.getByData("input-semester-membershipDiscountFee").should("have.value", "5");
    cy.getByData("input-semester-rebuyFee").should("have.value", "2");
    cy.getByData("input-semester-freeTrialLimit").should("have.value", "0");
  });

  it("warns about a mismatched term and start month without blocking progress", () => {
    openWizard();
    fillTermAndDates("fall", "2027-06-01", "2027-08-31");

    cy.getByData("semester-term-date-warning").should("contain", "Fall usually starts in September. This starts in June.");
    continueToFees();
    cy.getByData("semester-wizard-step-fees-budget").should("exist");
  });

  it("shows the exact submission values on review", () => {
    cy.intercept("POST", "/api/v2/semesters").as("createSemester");
    openWizard();
    fillTermAndDates("winter", "2028-01-08", "2028-04-30");
    continueToFees();
    cy.getByData("input-semester-startingBudget").clear().type("1000");
    cy.getByData("input-semester-membershipFee").clear().type("15");
    cy.getByData("input-semester-membershipDiscountFee").clear().type("10");
    cy.getByData("input-semester-rebuyFee").clear().type("5");
    cy.getByData("input-semester-freeTrialLimit").clear().type("4");
    cy.getByData("input-semester-meta").type("Winter semester notes");
    cy.getByData("semester-wizard-next-btn").click();

    cy.get("@createSemester.all").should("have.length", 0);
    cy.getByData("semester-wizard-step-review").should("exist");
    cy.getByData("semester-review-name").should("have.text", "Winter 2028");
    cy.getByData("semester-review-term").should("have.text", "Winter");
    cy.getByData("semester-review-startDate").should("have.text", "2028-01-08");
    cy.getByData("semester-review-endDate").should("have.text", "2028-04-30");
    cy.getByData("semester-review-startingBudget").should("have.text", "$1000");
    cy.getByData("semester-review-membershipFee").should("have.text", "$15");
    cy.getByData("semester-review-membershipDiscountFee").should("have.text", "$10");
    cy.getByData("semester-review-rebuyFee").should("have.text", "$5");
    cy.getByData("semester-review-freeTrialLimit").should("have.text", "4");
    cy.getByData("semester-review-meta").should("have.text", "Winter semester notes");
  });

  it("creates the previewed semester and selects it", () => {
    cy.intercept("POST", "/api/v2/semesters").as("createSemester");
    cy.visit("/admin/dashboard");
    openWizard();
    fillTermAndDates("winter", "2028-01-08", "2028-04-30");
    continueToReview();
    cy.getByData("create-semester-submit-btn").click();

    cy.wait("@createSemester").then((interception) => {
      expect(interception.request.body).to.deep.include({
        name: "Winter 2028",
        startingBudget: 100,
        membershipFee: 10,
        membershipDiscountFee: 7,
        rebuyFee: 2,
        freeTrialLimit: 4,
      });
      expect(interception.response?.statusCode).to.eq(201);
    });

    cy.getByData("semester-setup-wizard").should("not.exist");
    cy.getByData("semester-dropdown").should("contain", "Winter 2028");
  });
});
