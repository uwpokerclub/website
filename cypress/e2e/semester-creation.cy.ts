describe("CreateSemesterModal", () => {
  beforeEach(() => {
    cy.resetDatabase();
    cy.login();
    cy.visit("/admin");
    cy.getByData("sidenav").should("exist");
  });

  context("permission-gated visibility", () => {
    it("should display create semester option in dropdown for users with permission", () => {
      // Default e2e_user has WEBMASTER role with semester.create permission
      // Open dropdown first
      cy.getByData("semester-dropdown").click();
      cy.getByData("create-semester-btn").should("exist");
    });

    it("should display create semester button on semesters list page", () => {
      cy.visit("/admin/semesters");
      cy.getByData("create-semester-btn").should("exist");
    });
  });

  context("modal interactions from sidenav", () => {
    beforeEach(() => {
      // Open dropdown then click create
      cy.getByData("semester-dropdown").click();
      cy.getByData("create-semester-btn").click();
      cy.getByData("create-semester-modal").should("exist");
    });

    it("should open and close modal via cancel button", () => {
      cy.getByData("create-semester-modal").should("exist");
      cy.getByData("create-semester-cancel-btn").click();
      cy.getByData("create-semester-modal").should("not.exist");
    });

    it("should reset form on close and reopen", () => {
      // Fill in some values
      cy.getByData("input-semester-name").type("Test Semester");
      cy.getByData("input-semester-startingBudget").clear().type("500");

      // Close modal
      cy.getByData("create-semester-cancel-btn").click();
      cy.getByData("create-semester-modal").should("not.exist");

      // Reopen modal (need to open dropdown first since it closes with modal)
      cy.getByData("semester-dropdown").click();
      cy.getByData("create-semester-btn").click();
      cy.getByData("create-semester-modal").should("exist");

      // Verify form is reset
      cy.getByData("input-semester-name").should("have.value", "");
    });
  });

  context("form validation", () => {
    beforeEach(() => {
      // Open dropdown then click create
      cy.getByData("semester-dropdown").click();
      cy.getByData("create-semester-btn").click();
      cy.getByData("create-semester-modal").should("exist");
    });

    it("should show validation error for empty semester name", () => {
      // Fill required fields except name
      cy.getByData("input-semester-startDate").type("2025-09-01");
      cy.getByData("input-semester-endDate").type("2025-12-31");

      // Submit
      cy.getByData("create-semester-submit-btn").scrollIntoView().click();

      // Verify validation error
      cy.contains("Semester name is required").should("exist");
      cy.getByData("create-semester-modal").should("exist");
    });

    it("should show validation error for missing start date", () => {
      cy.getByData("input-semester-name").type("Test Semester");
      cy.getByData("input-semester-endDate").type("2025-12-31");

      // Submit
      cy.getByData("create-semester-submit-btn").scrollIntoView().click();

      // Verify validation error
      cy.contains("Start date is required").should("exist");
      cy.getByData("create-semester-modal").should("exist");
    });

    it("should show validation error for missing end date", () => {
      cy.getByData("input-semester-name").type("Test Semester");
      cy.getByData("input-semester-startDate").type("2025-09-01");

      // Submit
      cy.getByData("create-semester-submit-btn").scrollIntoView().click();

      // Verify validation error
      cy.contains("End date is required").should("exist");
      cy.getByData("create-semester-modal").should("exist");
    });

    it("should show validation error when end date is before start date", () => {
      cy.getByData("input-semester-name").type("Test Semester");
      cy.getByData("input-semester-startDate").type("2025-12-31");
      cy.getByData("input-semester-endDate").type("2025-09-01");

      // Submit
      cy.getByData("create-semester-submit-btn").scrollIntoView().click();

      // Verify validation error
      cy.contains("End date must be after start date").should("exist");
      cy.getByData("create-semester-modal").should("exist");
    });

    it("should fill all semester details fields", () => {
      cy.getByData("input-semester-name").type("Winter 2026");
      cy.getByData("input-semester-startDate").type("2026-01-01");
      cy.getByData("input-semester-endDate").type("2026-04-30");
      cy.getByData("input-semester-startingBudget").clear().type("1000");
      cy.getByData("input-semester-membershipFee").clear().type("15");
      cy.getByData("input-semester-membershipDiscountFee").clear().type("10");
      cy.getByData("input-semester-rebuyFee").clear().type("5");
      cy.getByData("input-semester-meta").type("Winter semester notes");

      // Verify values
      cy.getByData("input-semester-name").should("have.value", "Winter 2026");
      cy.getByData("input-semester-startDate").should(
        "have.value",
        "2026-01-01",
      );
      cy.getByData("input-semester-endDate").should("have.value", "2026-04-30");
      cy.getByData("input-semester-startingBudget").should(
        "have.value",
        "1000",
      );
      cy.getByData("input-semester-membershipFee").should("have.value", "15");
      cy.getByData("input-semester-membershipDiscountFee").should(
        "have.value",
        "10",
      );
      cy.getByData("input-semester-rebuyFee").should("have.value", "5");
      cy.getByData("input-semester-meta").should(
        "have.value",
        "Winter semester notes",
      );
    });
  });

  context("successful submission from sidenav", () => {
    it("should create semester and auto-select it in dropdown", () => {
      cy.intercept("POST", "/api/v2/semesters").as("createSemester");

      // Open dropdown then click create
      cy.getByData("semester-dropdown").click();
      cy.getByData("create-semester-btn").click();
      cy.getByData("create-semester-modal").should("exist");

      // Fill all required fields
      cy.getByData("input-semester-name").type("Winter 2026");
      cy.getByData("input-semester-startDate").type("2026-01-01");
      cy.getByData("input-semester-endDate").type("2026-04-30");
      cy.getByData("input-semester-startingBudget").clear().type("1000");
      cy.getByData("input-semester-membershipFee").clear().type("15");
      cy.getByData("input-semester-membershipDiscountFee").clear().type("10");
      cy.getByData("input-semester-rebuyFee").clear().type("5");
      cy.getByData("input-semester-meta").type("Winter semester notes");

      // Submit
      cy.getByData("create-semester-submit-btn").scrollIntoView().click();

      // Verify API call
      cy.wait("@createSemester").then((interception) => {
        expect(interception.request.body).to.deep.include({
          name: "Winter 2026",
          membershipFee: 15,
          membershipDiscountFee: 10,
          rebuyFee: 5,
        });
        expect(interception.response?.statusCode).to.eq(201);
      });

      // Verify modal closes
      cy.getByData("create-semester-modal").should("not.exist");

      // Verify new semester is selected in dropdown
      cy.getByData("semester-dropdown").should("contain", "Winter 2026");
    });
  });

  context("successful submission from semesters list", () => {
    it("should create semester and add it to the list", () => {
      cy.intercept("POST", "/api/v2/semesters").as("createSemester");

      cy.visit("/admin/semesters");
      cy.getByData("create-semester-btn").click();
      cy.getByData("create-semester-modal").should("exist");

      // Fill all required fields
      cy.getByData("input-semester-name").type("Spring 2027");
      cy.getByData("input-semester-startDate").type("2027-05-01");
      cy.getByData("input-semester-endDate").type("2027-08-31");
      cy.getByData("input-semester-startingBudget").clear().type("500");
      cy.getByData("input-semester-membershipFee").clear().type("12");
      cy.getByData("input-semester-membershipDiscountFee").clear().type("8");
      cy.getByData("input-semester-rebuyFee").clear().type("3");

      // Submit
      cy.getByData("create-semester-submit-btn").scrollIntoView().click();

      // Verify API call
      cy.wait("@createSemester").then((interception) => {
        expect(interception.request.body).to.deep.include({
          name: "Spring 2027",
          membershipFee: 12,
          membershipDiscountFee: 8,
          rebuyFee: 3,
        });
        expect(interception.response?.statusCode).to.eq(201);
      });

      // Verify modal closes
      cy.getByData("create-semester-modal").should("not.exist");

      // Verify new semester appears in the list
      cy.contains("Spring 2027").should("exist");
    });
  });

  context("error handling", () => {
    it("should display error when API fails", () => {
      cy.intercept("POST", "/api/v2/semesters", {
        statusCode: 500,
        body: { message: "Internal server error" },
      }).as("createSemesterError");

      // Open dropdown then click create
      cy.getByData("semester-dropdown").click();
      cy.getByData("create-semester-btn").click();
      cy.getByData("create-semester-modal").should("exist");

      // Fill all required fields
      cy.getByData("input-semester-name").type("Test Semester");
      cy.getByData("input-semester-startDate").type("2025-09-01");
      cy.getByData("input-semester-endDate").type("2025-12-31");

      // Submit
      cy.getByData("create-semester-submit-btn").scrollIntoView().click();

      // Wait for the failed request
      cy.wait("@createSemesterError");

      // Verify error is displayed and modal stays open
      cy.getByData("create-semester-error-alert").should("exist");
      cy.getByData("create-semester-modal").should("exist");
    });

    it("should display error when validation fails on server", () => {
      cy.intercept("POST", "/api/v2/semesters", {
        statusCode: 400,
        body: { message: "Invalid semester data: name already exists" },
      }).as("createSemesterValidationError");

      // Open dropdown then click create
      cy.getByData("semester-dropdown").click();
      cy.getByData("create-semester-btn").click();
      cy.getByData("create-semester-modal").should("exist");

      // Fill all required fields
      cy.getByData("input-semester-name").type("Duplicate Semester");
      cy.getByData("input-semester-startDate").type("2025-09-01");
      cy.getByData("input-semester-endDate").type("2025-12-31");

      // Submit
      cy.getByData("create-semester-submit-btn").scrollIntoView().click();

      // Wait for the failed request
      cy.wait("@createSemesterValidationError");

      // Verify error is displayed and modal stays open
      cy.getByData("create-semester-error-alert").should(
        "contain",
        "Invalid semester data",
      );
      cy.getByData("create-semester-modal").should("exist");
    });
  });
});
