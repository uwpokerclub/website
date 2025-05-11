describe("Login page", () => {
  beforeEach(() => {
    cy.exec("npm run db:reset && npm run db:seed");
    cy.visit("/admin/login");
  });

  it("should successfully login", () => {
    // Input username
    cy.get("input[name=username]").type("e2e_user");

    // Input password
    cy.get("input[name=password]").type("password");

    // Click the "Login" button
    cy.getByData("login-submit").click();

    // Check to ensure that a successful login will redirect off the login page
    cy.location("pathname").should("eq", "/admin");

    cy.getCookie("uwpsc-dev-session-id").should("exist");
  });

  it("should show invalid username banner", () => {
    // Input username
    cy.get("input[name=username]").type("unknownuser");

    // Input password
    cy.get("input[name=password]").type("nopassword");

    // Click the "Login" button
    cy.getByData("login-submit").click();

    // Check that the banner is on the page
    cy.getByData("login-error-banner").should("exist").contains("Invalid username/password provided");

    cy.getCookie("uwpsc-dev-session-id").should("not.exist");
  });
});
