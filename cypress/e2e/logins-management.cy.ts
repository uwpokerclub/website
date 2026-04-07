import { LOGINS } from "../seed";

describe("Logins Management", () => {
  before(() => {
    cy.resetDatabase();
  });

  context("navigation", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
      cy.intercept("GET", "/api/v2/logins*", { fixture: "logins.json" }).as("getLogins");
    });

    it("should navigate to logins page from sidenav", () => {
      cy.visit("/admin/dashboard");
      cy.getByData("sidenav").should("exist");
      cy.getByData("sidenav-webmaster-section").should("exist");
      cy.getByData("nav-link-manage-logins").click();
      cy.location("pathname").should("eq", "/admin/logins");
    });

    it("should display page header", () => {
      cy.visit("/admin/logins");
      cy.contains("h1", "Manage Logins").should("be.visible");
      cy.contains("Create and manage login accounts").should("be.visible");
    });
  });

  context("logins table", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
      cy.intercept("GET", "/api/v2/logins*", { fixture: "logins.json" }).as("getLogins");
      cy.visit("/admin/logins");
      cy.getByData("logins-table").should("exist");
    });

    it("should display table with correct headers, logins, and role badges", () => {
      // Column headers
      cy.getByData("sort-username-header").should("be.visible");
      cy.getByData("sort-role-header").should("be.visible");
      cy.getByData("sort-linkedMember-header").should("be.visible");

      // All seeded logins
      cy.get("[data-qa^='login-row-']").should("have.length", LOGINS.length);

      // Role badges
      LOGINS.forEach((login) => {
        cy.getByData(`login-role-${login.username}`).should("exist");
      });
    });

    it("should display login with linked member correctly", () => {
      const loginWithMember = LOGINS.find((l) => l.linkedMember)!;

      cy.getByData(`login-row-${loginWithMember.username}`).within(() => {
        cy.getByData(`login-username-${loginWithMember.username}`).should(
          "contain",
          loginWithMember.username
        );
        cy.getByData(`login-role-${loginWithMember.username}`).should("exist");
        cy.getByData(`login-linkedMember-${loginWithMember.username}`).should(
          "contain",
          loginWithMember.linkedMember!.firstName
        );
      });
    });

    it("should display login without linked member correctly", () => {
      const loginWithoutMember = LOGINS.find((l) => !l.linkedMember)!;

      cy.getByData(`login-linkedMember-${loginWithoutMember.username}`).should(
        "contain",
        "No linked member"
      );
    });
  });

  context("sorting", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
      cy.intercept("GET", "/api/v2/logins*", { fixture: "logins.json" }).as("getLogins");
      cy.visit("/admin/logins");
      cy.getByData("logins-table").should("exist");
    });

    it("should have clickable sort headers", () => {
      cy.getByData("sort-username-header").should("exist");
      cy.getByData("sort-role-header").should("exist");
      cy.getByData("sort-linkedMember-header").should("exist");

      // Verify clicking headers doesn't cause errors
      cy.getByData("sort-username-header").click();
      cy.getByData("sort-role-header").click();
      cy.getByData("sort-linkedMember-header").click();

      // Table should still be visible
      cy.getByData("logins-table").should("exist");
    });
  });

  context("create login modal", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
      cy.intercept("GET", "/api/v2/logins*", { fixture: "logins.json" }).as("getLogins");
      cy.visit("/admin/logins");
      cy.getByData("logins-table").should("exist");
    });

    it("should open modal when create button is clicked", () => {
      cy.getByData("create-login-btn").click();
      cy.getByData("create-login-modal").should("exist");
    });

    it("should close modal when cancel is clicked", () => {
      cy.getByData("create-login-btn").click();
      cy.getByData("create-login-modal").should("exist");

      cy.getByData("create-login-cancel-btn").click();
      cy.getByData("create-login-modal").should("not.exist");
    });

    it("should show validation errors for empty form", () => {
      cy.getByData("create-login-btn").click();
      cy.getByData("create-login-modal").should("exist");

      // Try to submit empty form
      cy.getByData("create-login-submit-btn").click();

      // Should show validation errors (form should still be open)
      cy.getByData("create-login-modal").should("exist");
    });

    it("should show validation error for short password", () => {
      cy.getByData("create-login-btn").click();
      cy.getByData("create-login-modal").should("exist");

      cy.getByData("input-username").type("newuser");
      cy.getByData("input-password").type("short");
      cy.getByData("select-role").select("executive");

      cy.getByData("create-login-submit-btn").click();

      // Should show validation error (modal still open)
      cy.getByData("create-login-modal").should("exist");
    });

    it("should show error for duplicate username", () => {
      cy.getByData("create-login-btn").click();
      cy.getByData("create-login-modal").should("exist");

      // Try to create with existing username
      cy.getByData("input-username").type("e2e_user");
      cy.getByData("input-password").type("password123");
      cy.getByData("select-role").select("executive");

      cy.intercept("POST", "/api/v2/logins", {
        statusCode: 409,
        body: { message: "Username already exists" },
      }).as("createLoginDuplicate");

      cy.getByData("create-login-submit-btn").click();

      cy.wait("@createLoginDuplicate");

      // Should show error (modal stays open)
      cy.getByData("create-login-error-alert").should("exist");
    });
  });

  context("edit password modal", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
      cy.intercept("GET", "/api/v2/logins*", { fixture: "logins.json" }).as("getLogins");
      cy.visit("/admin/logins");
      cy.getByData("logins-table").should("exist");
    });

    it("should open, display username, and close", () => {
      cy.getByData("edit-password-btn-test_executive").click();
      cy.getByData("edit-password-modal").should("exist");
      cy.getByData("edit-password-modal").should("contain", "test_executive");

      cy.getByData("edit-password-cancel-btn").click();
      cy.getByData("edit-password-modal").should("not.exist");
    });

    it("should show validation error when passwords do not match", () => {
      cy.getByData("edit-password-btn-test_executive").click();
      cy.getByData("edit-password-modal").should("exist");

      cy.getByData("input-new-password").type("newpassword123");
      cy.getByData("input-confirm-password").type("differentpassword");

      cy.getByData("edit-password-submit-btn").click();

      // Should show validation error (modal stays open)
      cy.getByData("edit-password-modal").should("exist");
    });
  });

  context("delete login modal", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
      cy.intercept("GET", "/api/v2/logins*", { fixture: "logins.json" }).as("getLogins");
      cy.visit("/admin/logins");
      cy.getByData("logins-table").should("exist");
    });

    it("should open, display username, and close", () => {
      cy.getByData("delete-login-btn-test_executive").click();
      cy.getByData("delete-login-modal").should("exist");
      cy.getByData("delete-login-modal").should("contain", "test_executive");

      cy.getByData("delete-login-cancel-btn").click();
      cy.getByData("delete-login-modal").should("not.exist");
    });

    it("should show linked member warning", () => {
      const loginWithMember = LOGINS.find((l) => l.linkedMember)!;

      cy.getByData(`delete-login-btn-${loginWithMember.username}`).click();
      cy.getByData("delete-login-modal").should("exist");

      // Should show linked member info
      cy.getByData("delete-login-modal").should(
        "contain",
        loginWithMember.linkedMember!.firstName
      );
    });
  });

  context("empty state", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
    });

    it("should display empty state when no logins exist", () => {
      // Mock the API to return empty array
      cy.intercept("GET", "/api/v2/logins*", {
        statusCode: 200,
        body: { data: [], total: 0 },
      }).as("getEmptyLogins");

      cy.visit("/admin/logins");
      cy.wait("@getEmptyLogins");

      cy.getByData("logins-empty").should("be.visible");
    });
  });

  context("loading state", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
    });

    it("should display loading spinner while fetching data", () => {
      // Delay the API response to see loading state
      cy.intercept("GET", "/api/v2/logins*", {
        statusCode: 200,
        body: { data: [], total: 0 },
        delay: 1000,
      }).as("getLoginsDelayed");

      cy.visit("/admin/logins");
      cy.getByData("logins-loading").should("be.visible");
    });
  });

  context("error state", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
    });

    it("should display error state on API failure", () => {
      cy.intercept("GET", "/api/v2/logins*", {
        statusCode: 500,
        body: { message: "Internal server error" },
      }).as("getLoginsError");

      cy.visit("/admin/logins");
      cy.wait("@getLoginsError");

      cy.getByData("logins-error").should("be.visible");
      cy.getByData("retry-btn").should("exist");
    });
  });

  context("contract tests", () => {
    before(() => {
      cy.resetDatabase();
    });

    beforeEach(() => {
      cy.login();
    });

    it("should load logins list from real API", () => {
      cy.visit("/admin/logins");
      cy.getByData("logins-table").should("exist");
      cy.get("[data-qa^='login-row-']").should("have.length", LOGINS.length);
    });

    it("should search logins and filter results", () => {
      cy.intercept("GET", "/api/v2/logins*").as("getLogins");
      cy.visit("/admin/logins");
      cy.wait("@getLogins");
      cy.getByData("logins-table").should("exist");

      cy.getByData("input-logins-search").type("hdrust0");
      cy.wait("@getLogins").its("request.url").should("include", "search=hdrust0");
      cy.getByData("logins-results-info").should("contain", "hdrust0");
      cy.get("[data-qa^='login-row-']").should("have.length", 1);
      cy.getByData("login-row-hdrust0").should("exist");
    });

    it("should create and delete a login", () => {
      cy.visit("/admin/logins");
      cy.getByData("logins-table").should("exist");

      // Create
      const newUsername = "testuser_cypress";
      cy.getByData("create-login-btn").click();
      cy.getByData("create-login-modal").should("exist");
      cy.getByData("input-username").type(newUsername);
      cy.getByData("input-password").type("password123");
      cy.getByData("select-role").select("executive");
      cy.getByData("create-login-submit-btn").click();
      cy.getByData("create-login-modal").should("not.exist");
      cy.getByData(`login-row-${newUsername}`).should("exist");

      // Delete
      cy.getByData(`delete-login-btn-${newUsername}`).click();
      cy.getByData("delete-login-modal").should("exist");
      cy.getByData("delete-login-confirm-btn").click();
      cy.getByData("delete-login-modal").should("not.exist");
      cy.getByData(`login-row-${newUsername}`).should("not.exist");
    });

    it("should update password successfully", () => {
      cy.visit("/admin/logins");
      cy.getByData("logins-table").should("exist");

      cy.getByData("edit-password-btn-test_executive").click();
      cy.getByData("edit-password-modal").should("exist");
      cy.getByData("input-new-password").type("newpassword123");
      cy.getByData("input-confirm-password").type("newpassword123");
      cy.getByData("edit-password-submit-btn").click();
      cy.getByData("edit-password-modal").should("not.exist");
    });
  });
});
