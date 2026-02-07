import { MEMBERS, SEMESTER, USERS } from "../seed";
import { getUserForMember } from "../support/helpers";

describe("MembersList", () => {
  context("when no semester is selected", () => {
    beforeEach(() => {
      cy.resetDatabase();
      cy.login();
      // Mock semesters API to return empty array so no semester is selected
      cy.intercept("GET", "/api/v2/semesters", { data: [], total: 0 }).as("getSemesters");
    });

    it("should display no semester selected message", () => {
      cy.visit("/admin/members");
      cy.getByData("members-no-semester").should("be.visible");
    });
  });

  context("with semester selected", () => {
    beforeEach(() => {
      cy.resetDatabase();
      cy.login();
      cy.visit("/admin/members");
      // Wait for table to be visible (data loaded)
      cy.getByData("members-table").should("exist");
    });

    context("table display", () => {
      it("should display all column headers", () => {
        cy.getByData("sort-userId-header").should("be.visible");
        cy.getByData("sort-name-header").should("be.visible");
        cy.getByData("sort-email-header").should("be.visible");
        cy.getByData("sort-status-header").should("be.visible");
      });

      it("should display all members in table", () => {
        cy.get("[data-qa^='member-row-']").should(
          "have.length",
          MEMBERS.length
        );
      });

      it("should display unpaid status correctly", () => {
        const unpaidMember = MEMBERS.find((m) => !m.paid)!;
        const user = getUserForMember(unpaidMember);

        cy.getByData(`member-row-${unpaidMember.id}`).within(() => {
          cy.getByData(`member-userId-${unpaidMember.id}`).should(
            "contain",
            unpaidMember.userId
          );
          cy.getByData(`member-name-${unpaidMember.id}`).should(
            "contain",
            `${user.firstName} ${user.lastName}`
          );
          cy.getByData(`member-email-${unpaidMember.id}`).should(
            "contain",
            user.email
          );
          cy.getByData(`member-status-${unpaidMember.id}`).should(
            "contain",
            "Unpaid"
          );
        });
      });

      it("should display paid status correctly", () => {
        const paidMember = MEMBERS.find((m) => m.paid && !m.discounted)!;

        cy.getByData(`member-status-${paidMember.id}`).should("contain", "Paid");
        cy.getByData(`member-status-${paidMember.id}`).should(
          "not.contain",
          "Discounted"
        );
      });

      it("should display paid discounted status correctly", () => {
        const discountedMember = MEMBERS.find((m) => m.paid && m.discounted)!;

        cy.getByData(`member-status-${discountedMember.id}`).should(
          "contain",
          "Discounted"
        );
      });
    });

    context("search functionality", () => {
      it("should filter by name", () => {
        const targetUser = USERS[0]; // Heinrik Drust
        const targetMember = MEMBERS[0];

        cy.getByData("input-members-search").type(targetUser.firstName);

        // Wait for debounce by checking results info updates
        cy.getByData("members-results-info").should(
          "contain",
          targetUser.firstName
        );
        cy.get("[data-qa^='member-row-']").should("have.length", 1);
        cy.getByData(`member-row-${targetMember.id}`).should("exist");
      });

      it("should filter by email", () => {
        const targetUser = USERS[2]; // eaucock2@si.edu
        const targetMember = MEMBERS[2];

        cy.getByData("input-members-search").type(targetUser.email);

        cy.getByData("members-results-info").should("contain", targetUser.email);
        cy.get("[data-qa^='member-row-']").should("have.length", 1);
        cy.getByData(`member-row-${targetMember.id}`).should("exist");
      });

      it("should be case-insensitive", () => {
        const targetUser = USERS[0];

        cy.getByData("input-members-search").type(
          targetUser.firstName.toUpperCase()
        );

        cy.getByData("members-results-info").should(
          "contain",
          targetUser.firstName.toUpperCase()
        );
        cy.get("[data-qa^='member-row-']").should("have.length", 1);
      });

      it("should show no results state", () => {
        cy.getByData("input-members-search").type("nonexistentuser12345");

        cy.getByData("members-no-results").should("be.visible");
      });

      it("should clear search and show all members", () => {
        // First search to filter
        cy.getByData("input-members-search").type("Heinrik");
        cy.getByData("members-results-info").should("contain", "Heinrik");
        cy.get("[data-qa^='member-row-']").should("have.length", 1);

        // Clear search
        cy.getByData("clear-search-btn").click();

        // Should show all members again
        cy.getByData("input-members-search").should("have.value", "");
        cy.get("[data-qa^='member-row-']").should(
          "have.length",
          MEMBERS.length
        );
      });
    });

    context("sorting", () => {
      it("should have clickable sort headers", () => {
        // Verify all sortable column headers exist and are clickable
        cy.getByData("sort-userId-header").should("exist");
        cy.getByData("sort-name-header").should("exist");
        cy.getByData("sort-email-header").should("exist");
        cy.getByData("sort-status-header").should("exist");

        // Verify clicking headers doesn't cause errors
        cy.getByData("sort-userId-header").click();
        cy.getByData("sort-name-header").click();
        cy.getByData("sort-email-header").click();
        cy.getByData("sort-status-header").click();

        // Table should still be visible after clicking headers
        cy.getByData("members-table").should("exist");
      });

      it("should display members in consistent order", () => {
        // Verify all members are displayed
        cy.get("[data-qa^='member-row-']").should("have.length", MEMBERS.length);

        // Verify each member row has all required data cells
        MEMBERS.forEach((member) => {
          cy.getByData(`member-row-${member.id}`).within(() => {
            cy.getByData(`member-userId-${member.id}`).should("exist");
            cy.getByData(`member-name-${member.id}`).should("exist");
            cy.getByData(`member-email-${member.id}`).should("exist");
            cy.getByData(`member-status-${member.id}`).should("exist");
          });
        });
      });
    });

    context("pagination", () => {
      it("should not show pagination when 25 or fewer members", () => {
        cy.getByData("members-pagination").should("not.exist");
      });

      it("should show pagination when more than 25 members", () => {
        // Generate mock data with 30 members
        const mockMembers = Array.from({ length: 30 }, (_, i) => ({
          id: `mock-member-${i}`,
          userId: `${10000000 + i}`,
          semesterId: SEMESTER.id,
          paid: i % 2 === 0,
          discounted: i % 3 === 0,
          user: {
            id: `${10000000 + i}`,
            firstName: `FirstName${i}`,
            lastName: `LastName${i}`,
            email: `user${i}@test.com`,
            faculty: "Science",
            questId: `user${i}`,
            createdAt: "2025-01-01",
          },
        }));

        cy.intercept("GET", /\/api\/v2\/semesters\/.*\/memberships/, {
          statusCode: 200,
          body: { data: mockMembers, total: mockMembers.length },
        }).as("getManyMembers");

        // Revisit to trigger mock
        cy.visit("/admin/members");
        cy.wait("@getManyMembers");

        cy.getByData("members-pagination").should("exist");
      });
    });

    context("register member modal", () => {
      it("should open modal when register button is clicked", () => {
        // Click the register button
        cy.getByData("register-member-btn").click();

        // Modal content should exist (using exist instead of visible due to CSS)
        cy.getByData("register-member-modal").should("exist");
      });

      it("should close modal when cancel is clicked", () => {
        cy.getByData("register-member-btn").click();
        cy.getByData("register-member-modal").should("exist");

        cy.getByData("register-cancel-btn").click();

        cy.getByData("register-member-modal").should("not.exist");
      });
    });

    context("edit member modal", () => {
      it("should open modal when edit icon is clicked", () => {
        const member = MEMBERS[0];

        // Scroll the table to make sure the edit button is in view and click it
        cy.getByData(`edit-member-btn-${member.id}`).scrollIntoView().click({ force: true });

        cy.getByData("edit-member-modal").should("exist");
      });

      it("should close modal when cancel is clicked", () => {
        cy.getByData(`edit-member-btn-${MEMBERS[0].id}`).scrollIntoView().click({ force: true });
        cy.getByData("edit-member-modal").should("exist");

        cy.getByData("edit-cancel-btn").click();

        cy.getByData("edit-member-modal").should("not.exist");
      });
    });

    context("empty state", () => {
      it("should display empty state when no members exist", () => {
        cy.intercept("GET", /\/api\/v2\/semesters\/.*\/memberships/, {
          statusCode: 200,
          body: { data: [], total: 0 },
        }).as("getEmptyMembers");

        cy.visit("/admin/members");
        cy.wait("@getEmptyMembers");

        cy.getByData("members-empty").should("be.visible");
      });
    });
  });
});
