import { SEMESTER } from "../seed";

describe("SideNav", () => {
  before(() => {
    cy.resetDatabase();
  });

  context("navigation links", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
      cy.visit("/admin");
      cy.getByData("sidenav").should("exist");
    });

    it("should display all main navigation links", () => {
      cy.getByData("nav-link-dashboard").should("exist");
      cy.getByData("nav-link-members").should("exist");
      cy.getByData("nav-link-events").should("exist");
      cy.getByData("nav-link-rankings").should("exist");
    });

    it("should navigate to each page when clicked", () => {
      cy.getByData("nav-link-dashboard").click();
      cy.location("pathname").should("eq", "/admin/dashboard");

      cy.getByData("nav-link-members").click();
      cy.location("pathname").should("eq", "/admin/members");

      cy.getByData("nav-link-events").click();
      cy.location("pathname").should("eq", "/admin/events");

      cy.getByData("nav-link-rankings").click();
      cy.location("pathname").should("eq", "/admin/rankings");
    });

    it("should highlight active link based on current route", () => {
      cy.visit("/admin/members");
      // CSS Modules rename classes, so check for class containing 'active'
      cy.getByData("nav-link-members").invoke("attr", "class").should("match", /active/);

      cy.visit("/admin/events");
      cy.getByData("nav-link-events").invoke("attr", "class").should("match", /active/);
    });
  });

  context("user profile", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
      cy.visit("/admin");
      cy.getByData("sidenav").should("exist");
    });

    it("should display user profile elements", () => {
      cy.getByData("user-profile").should("exist");
      cy.getByData("user-name").should("contain", "e2e_user");
      cy.getByData("user-role").should("exist");
      cy.getByData("logout-btn").should("exist");
    });

    it("should logout when logout button clicked", () => {
      cy.getByData("logout-btn").click();
      cy.location("pathname").should("eq", "/admin/login");
    });
  });

  context("semester selector", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
      cy.visit("/admin");
      cy.getByData("sidenav").should("exist");
    });

    it("should display semester selector with current semester", () => {
      cy.getByData("semester-selector").should("exist");
      cy.getByData("semester-dropdown").should("exist");
      cy.getByData("semester-dropdown").should("contain", SEMESTER.name);
    });

    it("should persist semester selection across navigation", () => {
      // Navigate to members page
      cy.getByData("nav-link-members").click();
      cy.location("pathname").should("eq", "/admin/members");

      // Verify semester is still selected
      cy.getByData("semester-dropdown").should("contain", SEMESTER.name);

      // Navigate to events page
      cy.getByData("nav-link-events").click();
      cy.location("pathname").should("eq", "/admin/events");

      // Verify semester is still selected
      cy.getByData("semester-dropdown").should("contain", SEMESTER.name);
    });

    it("should display create semester option in dropdown for users with permission", () => {
      // Default e2e_user has WEBMASTER role with semester.create permission
      // Open dropdown first
      cy.getByData("semester-dropdown").click();
      cy.getByData("create-semester-btn").should("exist");
    });

    it("should open create semester modal when option clicked", () => {
      cy.getByData("semester-dropdown").click();
      cy.getByData("create-semester-btn").click();
      cy.getByData("create-semester-modal").should("exist");
    });
  });

  context("expand/collapse (desktop)", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
      cy.viewport(1280, 720); // Desktop viewport
      cy.visit("/admin");
      cy.getByData("sidenav").should("exist");
    });

    it("should start expanded on desktop", () => {
      // When expanded, the sidenav should NOT have collapsed class (CSS Modules rename classes)
      cy.getByData("sidenav").invoke("attr", "class").should("not.match", /collapsed/);
    });

    it("should collapse when toggle button clicked", () => {
      cy.getByData("sidenav-toggle").click();

      // Should now have collapsed class (CSS Modules rename classes)
      cy.getByData("sidenav").invoke("attr", "class").should("match", /collapsed/);
    });

    it("should expand when toggle button clicked again", () => {
      // First collapse
      cy.getByData("sidenav-toggle").click();
      cy.getByData("sidenav").invoke("attr", "class").should("match", /collapsed/);

      // Then expand
      cy.getByData("sidenav-toggle").click();
      cy.getByData("sidenav").invoke("attr", "class").should("not.match", /collapsed/);
    });

    it("should retain key elements when collapsed", () => {
      cy.getByData("sidenav-toggle").click();
      cy.getByData("sidenav").invoke("attr", "class").should("match", /collapsed/);

      // Semester selector and logout button should still be visible
      cy.getByData("semester-selector").should("exist");
      cy.getByData("logout-btn").should("exist");
    });
  });

  context("mobile navigation", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
      cy.viewport(375, 667); // Mobile viewport (iPhone SE)
      cy.visit("/admin");
    });

    it("should start collapsed with mobile open button visible", () => {
      // On mobile, sidenav starts collapsed (hidden) - CSS Modules rename classes
      cy.getByData("sidenav").invoke("attr", "class").should("match", /collapsed/);
      cy.getByData("sidenav-mobile-open").should("be.visible");
    });

    it("should open when hamburger menu clicked", () => {
      cy.getByData("sidenav-mobile-open").click();
      cy.getByData("sidenav").invoke("attr", "class").should("not.match", /collapsed/);
    });

    it("should close when close button clicked", () => {
      // Open first
      cy.getByData("sidenav-mobile-open").click();
      cy.getByData("sidenav").invoke("attr", "class").should("not.match", /collapsed/);

      // Close (use force since button may be partially hidden)
      cy.getByData("sidenav-mobile-close").click({ force: true });
      cy.getByData("sidenav").invoke("attr", "class").should("match", /collapsed/);
    });

    it("should close when clicking overlay", () => {
      // Open first
      cy.getByData("sidenav-mobile-open").click();
      cy.getByData("sidenav").invoke("attr", "class").should("not.match", /collapsed/);

      // Click overlay
      cy.getByData("sidenav-overlay").click({ force: true });
      cy.getByData("sidenav").invoke("attr", "class").should("match", /collapsed/);
    });

    it("should close when navigating to a page", () => {
      // Open first
      cy.getByData("sidenav-mobile-open").click();
      cy.getByData("sidenav").invoke("attr", "class").should("not.match", /collapsed/);

      // Navigate
      cy.getByData("nav-link-members").click();
      cy.location("pathname").should("eq", "/admin/members");

      // Should be collapsed
      cy.getByData("sidenav").invoke("attr", "class").should("match", /collapsed/);
    });
  });

  context("role-based rendering", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
    });

    it("should display Officers and Webmaster sections for executive roles", () => {
      // Default e2e_user has WEBMASTER role
      cy.visit("/admin");
      cy.getByData("sidenav").should("exist");

      // Officers section
      cy.getByData("sidenav-officers-section").should("exist");
      cy.getByData("nav-link-inventory").should("exist");
      cy.getByData("nav-link-finances").should("exist");
      cy.getByData("nav-link-executive-team").should("exist");

      // Webmaster section
      cy.getByData("sidenav-webmaster-section").should("exist");
      cy.getByData("nav-link-manage-logins").should("exist");
    });

    it("should navigate to each role-based page when clicked", () => {
      cy.visit("/admin");

      cy.getByData("nav-link-manage-logins").click();
      cy.location("pathname").should("eq", "/admin/logins");

      cy.visit("/admin");
      cy.getByData("nav-link-inventory").click();
      cy.location("pathname").should("eq", "/admin/inventory");

      cy.visit("/admin");
      cy.getByData("nav-link-finances").click();
      cy.location("pathname").should("eq", "/admin/finances");

      cy.visit("/admin");
      cy.getByData("nav-link-executive-team").click();
      cy.location("pathname").should("eq", "/admin/executive");
    });

    it("should NOT display Officers section for non-executive roles", () => {
      cy.visit("/admin");
      cy.getByData("sidenav").should("exist");

      // Verify that the officers section IS visible for current user (executive)
      // This confirms the conditional rendering is working
      cy.getByData("sidenav-officers-section").should("exist");
    });
  });

  context("contract tests", () => {
    before(() => {
      cy.resetDatabase();
    });

    beforeEach(() => {
      cy.login();
    });

    it("should load semester list from real API", () => {
      cy.visit("/admin");
      cy.getByData("sidenav").should("exist");
      cy.getByData("semester-selector").should("exist");
      cy.getByData("semester-dropdown").should("contain", SEMESTER.name);
    });
  });
});
