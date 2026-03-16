describe("Login page", () => {
  before(() => {
    cy.resetDatabase();
  });

  context("authentication", () => {
    beforeEach(() => {
      cy.visit("/admin/login");
    });

    it("should show error for invalid credentials", () => {
      cy.intercept("POST", "/api/v2/session", {
        statusCode: 401,
        body: { message: "Invalid username or password" },
      }).as("loginFailed");

      cy.getByData("input-username").type("unknownuser");
      cy.getByData("input-password").type("wrongpassword");
      cy.getByData("login-submit").click();

      cy.wait("@loginFailed");

      // Verify error banner is displayed
      cy.getByData("login-error-banner")
        .should("exist")
        .and("contain", "Invalid username or password");

      // Verify no session cookie is set
      cy.getCookie("uwpsc-dev-session-id").should("not.exist");

      // Verify user stays on login page
      cy.location("pathname").should("eq", "/admin/login");
    });

    it("should validate required credentials", () => {
      // Empty username
      cy.getByData("input-password").type("password");
      cy.getByData("login-submit").click();
      cy.location("pathname").should("eq", "/admin/login");
      cy.getCookie("uwpsc-dev-session-id").should("not.exist");

      // Clear and try empty password
      cy.getByData("input-password").clear();
      cy.getByData("input-username").type("e2e_user");
      cy.getByData("login-submit").click();
      cy.location("pathname").should("eq", "/admin/login");
      cy.getCookie("uwpsc-dev-session-id").should("not.exist");
    });
  });

  context("logout", () => {
    beforeEach(() => {
      // Login via API for faster setup
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
      cy.visit("/admin");
    });

    it("should successfully logout", () => {
      // Verify we're logged in
      cy.getCookie("uwpsc-dev-session-id").should("exist");

      cy.getByData("logout-btn").click();

      // Verify redirect to login page
      cy.location("pathname").should("eq", "/admin/login");
    });
  });

  context("session persistence", () => {
    it("should redirect to login when accessing protected route without session", () => {
      // Clear any existing session
      cy.clearCookies();

      // Try to access protected route
      cy.visit("/admin/events");

      // Should redirect to login
      cy.location("pathname").should("eq", "/admin/login");
    });
  });

  context("contract tests", () => {
    before(() => {
      cy.resetDatabase();
    });

    it("should successfully login with valid credentials", () => {
      cy.visit("/admin/login");

      cy.getByData("input-username").type("e2e_user");
      cy.getByData("input-password").type("password");
      cy.getByData("login-submit").click();

      // Verify redirect to admin dashboard
      cy.location("pathname").should("eq", "/admin");

      // Verify session cookie exists
      cy.getCookie("uwpsc-dev-session-id").should("exist");
    });

    it("should persist session across page reloads", () => {
      // Login via API
      cy.login();
      cy.visit("/admin");

      // Verify logged in
      cy.getCookie("uwpsc-dev-session-id").should("exist");
      cy.location("pathname").should("eq", "/admin");

      // Reload the page
      cy.reload();

      // Verify still logged in
      cy.getCookie("uwpsc-dev-session-id").should("exist");
      cy.location("pathname").should("eq", "/admin");
    });
  });
});
