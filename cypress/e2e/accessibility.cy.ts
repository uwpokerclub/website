function terminalLog(violations: axe.Result[]) {
  cy.task(
    "log",
    `${violations.length} accessibility violation${violations.length === 1 ? "" : "s"} detected`,
  );
  const violationData = violations.map(
    ({ id, impact, description, nodes }) => ({
      id,
      impact,
      description,
      nodes: nodes.length,
    }),
  );
  cy.task("table", violationData);
}

describe("Accessibility", () => {
  before(() => cy.resetDatabase());

  const publicPages = [
    { name: "Join", path: "/join" },
    { name: "Sponsors", path: "/sponsors" },
    { name: "Gallery", path: "/gallery" },
  ];

  const adminPages = [
    { name: "Dashboard", path: "/admin/dashboard" },
    { name: "Events", path: "/admin/events" },
    { name: "Members", path: "/admin/members" },
    { name: "Rankings", path: "/admin/rankings" },
    { name: "Logins", path: "/admin/logins" },
  ];

  context("Public pages", () => {
    publicPages.forEach(({ name, path }) => {
      it(`${name} (${path}) should log accessibility violations`, () => {
        cy.visit(path);
        cy.injectAxe();
        cy.checkA11y(null, null, terminalLog, true);
      });
    });
  });

  context("Auth pages", () => {
    it("Login (/admin/login) should log accessibility violations", () => {
      cy.visit("/admin/login");
      cy.injectAxe();
      cy.checkA11y(null, null, terminalLog, true);
    });
  });

  context("Admin pages", () => {
    beforeEach(() => cy.login());

    adminPages.forEach(({ name, path }) => {
      it(`${name} (${path}) should log accessibility violations`, () => {
        cy.visit(path);
        cy.injectAxe();
        cy.checkA11y(null, null, terminalLog, true);
      });
    });
  });
});
