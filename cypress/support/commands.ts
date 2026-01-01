/// <reference types="cypress" />

// eslint-disable-next-line @typescript-eslint/no-namespace
declare namespace Cypress {
  interface Chainable {
    /**
     * Get element by data-qa attribute
     * @param dataTestAttribute - The value of the data-qa attribute
     * @example cy.getByData("login-submit").click()
     */
    getByData(dataTestAttribute: string): Chainable<JQuery<HTMLElement>>;

    /**
     * Login to the application via API
     * @param username - Username for authentication (defaults to TEST_USERNAME env var)
     * @param password - Password for authentication (defaults to TEST_PASSWORD env var)
     * @example cy.login("e2e_user", "password")
     */
    login(username?: string, password?: string): Chainable<void>;

    /**
     * Reset and seed the test database
     * @example cy.resetDatabase()
     */
    resetDatabase(): Chainable<Cypress.Exec>;
  }
}

/**
 * Get element by data-qa attribute
 */
Cypress.Commands.add("getByData", (selector: string) => {
  return cy.get(`[data-qa="${selector}"]`);
});

/**
 * Login to the application via API
 * Uses environment variables if credentials not provided
 */
Cypress.Commands.add("login", (username?: string, password?: string) => {
  const user = username ?? Cypress.env("TEST_USERNAME") ?? "e2e_user";
  const pass = password ?? Cypress.env("TEST_PASSWORD") ?? "password";

  cy.request("POST", "/api/v2/session", {
    username: user,
    password: pass,
  });
  cy.getCookie("uwpsc-dev-session-id").should("exist");
});

/**
 * Reset and seed the test database
 */
Cypress.Commands.add("resetDatabase", () => {
  return cy.exec("npm run db:reset && npm run db:seed", { timeout: 30000 });
});
