describe("Login page", () => {
  beforeEach(() => {
    cy.resetDatabase();
  });

  context("authentication", () => {
    beforeEach(() => {
      cy.visit("/admin/login");
    });

    it("should successfully login with valid credentials", () => {
      cy.get("input[name=username]").type("e2e_user");
      cy.get("input[name=password]").type("password");
      cy.getByData("login-submit").click();

      // Verify redirect to admin dashboard
      cy.location("pathname").should("eq", "/admin");

      // Verify session cookie exists
      cy.getCookie("uwpsc-dev-session-id").should("exist");
    });

    it("should show error for invalid credentials", () => {
      cy.get("input[name=username]").type("unknownuser");
      cy.get("input[name=password]").type("wrongpassword");
      cy.getByData("login-submit").click();

      // Verify error banner is displayed
      cy.getByData("login-error-banner")
        .should("exist")
        .and("contain", "Invalid username or password");

      // Verify no session cookie is set
      cy.getCookie("uwpsc-dev-session-id").should("not.exist");

      // Verify user stays on login page
      cy.location("pathname").should("eq", "/admin/login");
    });

    it("should show error for empty username", () => {
      cy.get("input[name=password]").type("password");
      cy.getByData("login-submit").click();

      // Verify user stays on login page (form validation prevents submission)
      cy.location("pathname").should("eq", "/admin/login");
      cy.getCookie("uwpsc-dev-session-id").should("not.exist");
    });

    it("should show error for empty password", () => {
      cy.get("input[name=username]").type("e2e_user");
      cy.getByData("login-submit").click();

      // Verify user stays on login page (form validation prevents submission)
      cy.location("pathname").should("eq", "/admin/login");
      cy.getCookie("uwpsc-dev-session-id").should("not.exist");
    });
  });

  context("logout", () => {
    beforeEach(() => {
      // Login via API for faster setup
      cy.login();
      cy.visit("/admin");
    });

    it("should successfully logout", () => {
      // Verify we're logged in
      cy.getCookie("uwpsc-dev-session-id").should("exist");

      // Click logout button (uses aria-label as no data-qa attribute exists)
      cy.get('button[aria-label="Logout"]').click();

      // Verify redirect to login page
      cy.location("pathname").should("eq", "/admin/login");
    });
  });

  context("session persistence", () => {
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

    it("should redirect to login when accessing protected route without session", () => {
      // Clear any existing session
      cy.clearCookies();

      // Try to access protected route
      cy.visit("/admin/events");

      // Should redirect to login
      cy.location("pathname").should("eq", "/admin/login");
    });
  });
});
