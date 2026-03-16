import { USERS, USERS_WITHOUT_MEMBERSHIPS } from "../seed";

// Helper to get the member search input
const getMemberSearchInput = () => {
  return cy.getByData("member-search-combobox");
};

describe("Member Registration", () => {
  before(() => {
    cy.resetDatabase();
  });

  context("stubbed tests", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
      cy.intercept("GET", /\/api\/v2\/semesters\/.*\/memberships/, { fixture: "memberships.json" }).as("getMemberships");
      cy.visit("/admin/members");
      cy.getByData("members-table").should("exist");

      // Set up API intercepts for search/create
      cy.intercept("GET", "/api/v2/members?name=*").as("searchByName");
      cy.intercept("GET", "/api/v2/members?email=*").as("searchByEmail");
      cy.intercept("GET", "/api/v2/members?id=*").as("searchById");
      cy.intercept("POST", "/api/v2/members").as("createMember");
      cy.intercept("POST", "/api/v2/semesters/*/memberships").as(
        "createMembership"
      );
    });

    context("Search Mode", () => {
      context("modal behavior", () => {
        it("should open in search mode with combobox by default", () => {
          cy.getByData("register-member-btn").click();
          cy.getByData("register-member-modal").should("exist");
          cy.getByData("member-search-combobox").should("exist");
          cy.getByData("toggle-new-member-btn").should("exist");
          getMemberSearchInput().should("exist");
        });
      });

      context("membership configuration", () => {
        it("should show discounted checkbox only when paid is checked", () => {
          cy.getByData("register-member-btn").click();
          cy.getByData("checkbox-discounted").should("not.exist");
          cy.getByData("checkbox-paid").click();
          cy.getByData("checkbox-discounted").should("be.visible");
        });

        it("should hide discounted when paid is unchecked", () => {
          cy.getByData("register-member-btn").click();
          cy.getByData("checkbox-paid").click();
          cy.getByData("checkbox-discounted").should("be.visible");
          cy.getByData("checkbox-discounted").click();
          cy.getByData("checkbox-paid").click();
          cy.getByData("checkbox-discounted").should("not.exist");
        });
      });

      context("submission", () => {
        it("should validate member must be selected", () => {
          cy.getByData("register-member-btn").click();
          cy.getByData("register-submit-btn").click();
          // Form should show validation error - member selection required
          cy.contains("Please select a member").should("be.visible");
        });
      });
    });

    context("Create Mode", () => {
      context("mode switching", () => {
        it("should toggle between search and create modes", () => {
          cy.getByData("register-member-btn").click();

          // Switch to create mode
          cy.getByData("toggle-new-member-btn").click();
          cy.getByData("input-studentId").should("be.visible");
          cy.getByData("input-firstName").should("be.visible");
          cy.getByData("input-lastName").should("be.visible");
          cy.getByData("input-email").should("be.visible");
          cy.getByData("select-faculty").should("be.visible");

          // Switch back to search mode
          cy.getByData("toggle-search-btn").click();
          cy.getByData("member-search-combobox").should("exist");
          cy.getByData("input-studentId").should("not.exist");
        });
      });

      context("form validation", () => {
        it("should validate required fields", () => {
          cy.getByData("register-member-btn").click();
          cy.getByData("toggle-new-member-btn").click();
          cy.getByData("register-submit-btn").click();

          // Check for validation errors on required fields (use exist instead of visible due to modal scrolling)
          cy.contains("Student ID is required").should("exist");
          cy.contains("First name is required").should("exist");
          cy.contains("Last name is required").should("exist");
          cy.contains("Email is required").should("exist");
          cy.contains("Please select a faculty").should("exist");
        });

        it("should validate student ID is numeric only", () => {
          cy.getByData("register-member-btn").click();
          cy.getByData("toggle-new-member-btn").click();
          cy.getByData("input-studentId").type("abc123");
          cy.getByData("register-submit-btn").click();
          cy.contains("Student ID must contain only numbers").should("be.visible");
        });

        it("should validate email format", () => {
          cy.getByData("register-member-btn").click();
          cy.getByData("toggle-new-member-btn").click();
          cy.getByData("input-studentId").type("12345678");
          cy.getByData("input-firstName").type("Test");
          cy.getByData("input-lastName").type("User");
          cy.getByData("input-email").type("invalid-email");
          cy.getByData("select-faculty").select("Science");
          cy.getByData("register-submit-btn").click();
          cy.contains("Invalid email").should("be.visible");
        });
      });
    });

    context("General", () => {
      it("should close modal when cancel is clicked", () => {
        cy.getByData("register-member-btn").click();
        cy.getByData("register-member-modal").should("exist");
        cy.getByData("register-cancel-btn").click();
        cy.getByData("register-member-modal").should("not.exist");
      });
    });
  });

  context("contract tests", () => {
    before(() => {
      cy.resetDatabase();
    });

    beforeEach(() => {
      cy.login();
      cy.visit("/admin/members");
      cy.getByData("members-table").should("exist");

      cy.intercept("GET", "/api/v2/members?name=*").as("searchByName");
      cy.intercept("GET", "/api/v2/members?email=*").as("searchByEmail");
      cy.intercept("GET", "/api/v2/members?id=*").as("searchById");
      cy.intercept("POST", "/api/v2/members").as("createMember");
      cy.intercept("POST", "/api/v2/semesters/*/memberships").as(
        "createMembership"
      );
    });

    it("should search by name and select member", () => {
      const targetUser = USERS[0]; // Heinrik Drust

      cy.getByData("register-member-btn").click();
      getMemberSearchInput().type(targetUser.firstName);
      cy.wait("@searchByName");
      cy.getByData(`combobox-option-${targetUser.id}`).should("exist");
    });

    it("should search by email", () => {
      const targetUser = USERS[0]; // hdrust0@merriam-webster.com

      cy.getByData("register-member-btn").click();
      getMemberSearchInput().type(targetUser.email);
      cy.wait("@searchByEmail");
      cy.getByData(`combobox-option-${targetUser.id}`).should("exist");
    });

    it("should search by student ID", () => {
      const targetUser = USERS[0]; // 62958169

      cy.getByData("register-member-btn").click();
      getMemberSearchInput().type(targetUser.id);
      cy.wait("@searchById");
      cy.getByData(`combobox-option-${targetUser.id}`).should("exist");
    });

    it("should show no results message", () => {
      cy.getByData("register-member-btn").click();
      getMemberSearchInput().type("nonexistentuser12345");
      cy.wait("@searchByName");
      cy.contains("No members found").should("exist");
    });

    it("should register existing member successfully", () => {
      const targetUser = USERS_WITHOUT_MEMBERSHIPS[0]; // Unregistered TestUser

      cy.getByData("register-member-btn").click();
      getMemberSearchInput().type(targetUser.firstName);
      cy.wait("@searchByName");
      cy.getByData(`combobox-option-${targetUser.id}`).click();
      cy.getByData("checkbox-paid").click();
      cy.getByData("register-submit-btn").click();

      cy.wait("@createMembership").its("response.statusCode").should("eq", 201);
      cy.contains("registered successfully").should("exist");
    });

    it("should create new member successfully", () => {
      cy.fixture("users.json").then((users) => {
        const newUser = users[0]; // Port Heikkinen from fixture

        cy.getByData("register-member-btn").click();
        cy.getByData("toggle-new-member-btn").click();

        cy.getByData("input-studentId").type(String(newUser.id));
        cy.getByData("input-firstName").type(newUser.firstName);
        cy.getByData("input-lastName").type(newUser.lastName);
        cy.getByData("input-email").type(newUser.email);
        cy.getByData("select-faculty").select(newUser.faculty);
        cy.getByData("input-questId").type(newUser.questId);
        cy.getByData("checkbox-paid").click();

        cy.getByData("register-submit-btn").click();

        cy.wait("@createMember").its("response.statusCode").should("eq", 201);
        cy.wait("@createMembership").its("response.statusCode").should("eq", 201);
        // Verify success via toast message
        cy.contains("registered successfully").should("exist");
      });
    });
  });
});
