import { LOGINS } from "../seed";

describe("Logins Management", () => {
  beforeEach(() => {
    cy.resetDatabase();
    cy.login();
  });

  context("navigation", () => {
    it("should navigate to logins page from sidenav", () => {
      cy.visit("/admin");
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
      cy.visit("/admin/logins");
      cy.getByData("logins-table").should("exist");
    });

    it("should display all column headers", () => {
      cy.getByData("sort-username-header").should("be.visible");
      cy.getByData("sort-role-header").should("be.visible");
      cy.getByData("sort-linkedMember-header").should("be.visible");
    });

    it("should display all seeded logins", () => {
      cy.get("[data-qa^='login-row-']").should("have.length", LOGINS.length);
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

    it("should display role badges", () => {
      LOGINS.forEach((login) => {
        cy.getByData(`login-role-${login.username}`).should("exist");
      });
    });
  });

  context("search functionality", () => {
    beforeEach(() => {
      cy.visit("/admin/logins");
      cy.getByData("logins-table").should("exist");
    });

    it("should filter by username", () => {
      cy.getByData("input-logins-search").type("hdrust0");

      cy.getByData("logins-results-info").should("contain", "hdrust0");
      cy.get("[data-qa^='login-row-']").should("have.length", 1);
      cy.getByData("login-row-hdrust0").should("exist");
    });

    it("should filter by role", () => {
      cy.getByData("input-logins-search").type("executive");

      cy.getByData("logins-results-info").should("contain", "executive");
      // Should show test_executive and hdrust0 (both executive role)
      cy.get("[data-qa^='login-row-']").should("have.length", 2);
    });

    it("should be case-insensitive", () => {
      cy.getByData("input-logins-search").type("PRESIDENT");

      cy.getByData("logins-results-info").should("contain", "PRESIDENT");
      cy.get("[data-qa^='login-row-']").should("have.length", 1);
    });

    it("should show no results state", () => {
      cy.getByData("input-logins-search").type("nonexistentuser12345");

      cy.getByData("logins-no-results").should("be.visible");
    });

    it("should clear search and show all logins", () => {
      // First search to filter
      cy.getByData("input-logins-search").type("hdrust0");
      cy.getByData("logins-results-info").should("contain", "hdrust0");
      cy.get("[data-qa^='login-row-']").should("have.length", 1);

      // Clear search
      cy.getByData("clear-search-btn").click();

      // Should show all logins again
      cy.getByData("input-logins-search").should("have.value", "");
      cy.get("[data-qa^='login-row-']").should("have.length", LOGINS.length);
    });
  });

  context("sorting", () => {
    beforeEach(() => {
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

    it("should create a new login successfully", () => {
      const newUsername = `testuser_${Date.now()}`;

      cy.getByData("create-login-btn").click();
      cy.getByData("create-login-modal").should("exist");

      cy.getByData("input-username").type(newUsername);
      cy.getByData("input-password").type("password123");
      cy.getByData("select-role").select("executive");

      cy.getByData("create-login-submit-btn").click();

      // Modal should close
      cy.getByData("create-login-modal").should("not.exist");

      // New login should appear in the table
      cy.getByData(`login-row-${newUsername}`).should("exist");
    });

    it("should show error for duplicate username", () => {
      cy.getByData("create-login-btn").click();
      cy.getByData("create-login-modal").should("exist");

      // Try to create with existing username
      cy.getByData("input-username").type("e2e_user");
      cy.getByData("input-password").type("password123");
      cy.getByData("select-role").select("executive");

      cy.getByData("create-login-submit-btn").click();

      // Should show error (modal stays open)
      cy.getByData("create-login-error-alert").should("exist");
    });
  });

  context("edit password modal", () => {
    beforeEach(() => {
      cy.visit("/admin/logins");
      cy.getByData("logins-table").should("exist");
    });

    it("should open modal when edit button is clicked", () => {
      cy.getByData("edit-password-btn-test_executive").click();
      cy.getByData("edit-password-modal").should("exist");
    });

    it("should display username in modal title", () => {
      cy.getByData("edit-password-btn-test_executive").click();
      cy.getByData("edit-password-modal").should("contain", "test_executive");
    });

    it("should close modal when cancel is clicked", () => {
      cy.getByData("edit-password-btn-test_executive").click();
      cy.getByData("edit-password-modal").should("exist");

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

    it("should update password successfully", () => {
      cy.getByData("edit-password-btn-test_executive").click();
      cy.getByData("edit-password-modal").should("exist");

      cy.getByData("input-new-password").type("newpassword123");
      cy.getByData("input-confirm-password").type("newpassword123");

      cy.getByData("edit-password-submit-btn").click();

      // Modal should close
      cy.getByData("edit-password-modal").should("not.exist");
    });
  });

  context("delete login modal", () => {
    beforeEach(() => {
      cy.visit("/admin/logins");
      cy.getByData("logins-table").should("exist");
    });

    it("should open modal when delete button is clicked", () => {
      cy.getByData("delete-login-btn-test_executive").click();
      cy.getByData("delete-login-modal").should("exist");
    });

    it("should display username in confirmation message", () => {
      cy.getByData("delete-login-btn-test_executive").click();
      cy.getByData("delete-login-modal").should("contain", "test_executive");
    });

    it("should close modal when cancel is clicked", () => {
      cy.getByData("delete-login-btn-test_executive").click();
      cy.getByData("delete-login-modal").should("exist");

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

    it("should delete login successfully", () => {
      const initialCount = LOGINS.length;

      cy.getByData("delete-login-btn-test_executive").click();
      cy.getByData("delete-login-modal").should("exist");

      cy.getByData("delete-login-confirm-btn").click();

      // Modal should close
      cy.getByData("delete-login-modal").should("not.exist");

      // Login should be removed from table
      cy.getByData("login-row-test_executive").should("not.exist");
      cy.get("[data-qa^='login-row-']").should("have.length", initialCount - 1);
    });
  });

  context("empty state", () => {
    it("should display empty state when no logins exist", () => {
      // Mock the API to return empty array
      cy.intercept("GET", "/api/v2/logins", {
        statusCode: 200,
        body: [],
      }).as("getEmptyLogins");

      cy.visit("/admin/logins");
      cy.wait("@getEmptyLogins");

      cy.getByData("logins-empty").should("be.visible");
    });
  });

  context("loading state", () => {
    it("should display loading spinner while fetching data", () => {
      // Delay the API response to see loading state
      cy.intercept("GET", "/api/v2/logins", {
        statusCode: 200,
        body: [],
        delay: 1000,
      }).as("getLoginsDelayed");

      cy.visit("/admin/logins");
      cy.getByData("logins-loading").should("be.visible");
    });
  });

  context("error state", () => {
    it("should display error state on API failure", () => {
      cy.intercept("GET", "/api/v2/logins", {
        statusCode: 500,
        body: { message: "Internal server error" },
      }).as("getLoginsError");

      cy.visit("/admin/logins");
      cy.wait("@getLoginsError");

      cy.getByData("logins-error").should("be.visible");
      cy.getByData("retry-btn").should("exist");
    });
  });
});
