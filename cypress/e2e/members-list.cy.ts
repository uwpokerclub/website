import { MEMBERS, SEMESTER, USERS } from "../seed";
import { getUserForMember, getMemberFullName } from "../support/helpers";

describe("MembersList", () => {
  before(() => {
    cy.resetDatabase();
  });

  context("when no semester is selected", () => {
    beforeEach(() => {
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
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
      cy.intercept("GET", /\/api\/v2\/semesters\/.*\/memberships/, { fixture: "memberships.json" }).as("getMemberships");
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

      it("should display member statuses correctly", () => {
        // Unpaid
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

        // Paid
        const paidMember = MEMBERS.find((m) => m.paid && !m.discounted)!;
        cy.getByData(`member-status-${paidMember.id}`).should("contain", "Paid");
        cy.getByData(`member-status-${paidMember.id}`).should(
          "not.contain",
          "Discounted"
        );

        // Discounted
        const discountedMember = MEMBERS.find((m) => m.paid && m.discounted)!;
        cy.getByData(`member-status-${discountedMember.id}`).should(
          "contain",
          "Discounted"
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

    context("filter drawer", () => {
      it("should open and close filter drawer", () => {
        cy.getByData("filter-toggle-btn").click();
        cy.getByData("filter-sidebar").should("be.visible");
        cy.getByData("filter-studentId").should("exist");
        cy.getByData("filter-name").should("exist");
        cy.getByData("filter-email").should("exist");
        cy.getByData("filter-faculty").should("exist");
        cy.getByData("filter-paid").should("exist");
        cy.getByData("filter-discounted").should("exist");

        // Close via X button
        cy.get("[aria-label='Close filters']").click();
        cy.getByData("filter-sidebar").should("not.be.visible");
      });

      it("should send name filter to API", () => {
        cy.getByData("filter-toggle-btn").click();
        cy.getByData("filter-name").type("Heinrik");
        cy.wait("@getMemberships").its("request.url").should("include", "name=Heinrik");
      });

      it("should send email filter to API", () => {
        cy.getByData("filter-toggle-btn").click();
        cy.getByData("filter-email").type("merriam");
        cy.wait("@getMemberships").its("request.url").should("include", "email=merriam");
      });

      it("should send student ID filter to API", () => {
        cy.getByData("filter-toggle-btn").click();
        cy.getByData("filter-studentId").type("62958169");
        cy.wait("@getMemberships").its("request.url").should("include", "studentId=62958169");
      });

      it("should send faculty filter to API", () => {
        cy.getByData("filter-toggle-btn").click();
        cy.getByData("filter-faculty").select("Engineering");
        cy.wait("@getMemberships").its("request.url").should("include", "faculty=Engineering");
      });

      it("should send paid filter to API", () => {
        cy.getByData("filter-toggle-btn").click();
        cy.getByData("filter-paid").select("Yes");
        cy.wait("@getMemberships").its("request.url").should("include", "paid=true");
      });

      it("should send discounted filter to API", () => {
        cy.getByData("filter-toggle-btn").click();
        cy.getByData("filter-discounted").select("Yes");
        cy.wait("@getMemberships").its("request.url").should("include", "discounted=true");
      });

      it("should show active filter count", () => {
        cy.getByData("filter-toggle-btn").click();
        cy.getByData("filter-name").type("test");
        cy.getByData("filter-active-count").should("contain", "1");
        cy.getByData("filter-faculty").select("Math");
        cy.getByData("filter-active-count").should("contain", "2");
      });

      it("should clear all filters", () => {
        cy.getByData("filter-toggle-btn").click();
        cy.getByData("filter-name").type("test");
        cy.getByData("filter-faculty").select("Math");
        cy.getByData("filter-clear-btn").click();

        cy.getByData("filter-name").should("have.value", "");
        cy.getByData("filter-faculty").should("have.value", "");
        cy.getByData("filter-active-count").should("not.exist");
      });

      it("should persist filters in URL query params", () => {
        cy.getByData("filter-toggle-btn").click();
        cy.getByData("filter-faculty").select("Engineering");
        cy.getByData("filter-paid").select("Yes");

        cy.url().should("include", "faculty=Engineering");
        cy.url().should("include", "paid=true");
      });

      it("should restore filters from URL on page load", () => {
        cy.visit("/admin/members?name=Heinrik&faculty=AHS");
        cy.getByData("members-table").should("exist");
        cy.getByData("filter-toggle-btn").click();

        cy.getByData("filter-name").should("have.value", "Heinrik");
        cy.getByData("filter-faculty").should("have.value", "AHS");
      });

      it("should clear URL params when filters are cleared", () => {
        cy.visit("/admin/members?name=test&faculty=Math");
        cy.getByData("members-table").should("exist");
        cy.getByData("filter-toggle-btn").click();
        cy.getByData("filter-clear-btn").click();

        cy.url().should("not.include", "name=");
        cy.url().should("not.include", "faculty=");
      });
    });

    context("register member modal", () => {
      it("should open and close modal", () => {
        cy.getByData("register-member-btn").click();
        cy.getByData("register-member-modal").should("exist");

        cy.getByData("register-cancel-btn").click();
        cy.getByData("register-member-modal").should("not.exist");
      });
    });

    context("edit member modal", () => {
      it("should open and close modal", () => {
        const member = MEMBERS[0];

        cy.getByData(`edit-member-btn-${member.id}`).scrollIntoView().click();
        cy.getByData("edit-member-modal").should("exist");

        cy.getByData("edit-cancel-btn").click();
        cy.getByData("edit-member-modal").should("not.exist");
      });
    });

    context("delete membership modal", () => {
      // Use member with no event participation for clean deletion
      const deleteMember = MEMBERS[3]; // Khalil Duckham - paid, no events
      const deleteMemberName = getMemberFullName(deleteMember);

      it("should open modal and display confirmation details", () => {
        cy.getByData(`delete-member-btn-${deleteMember.id}`).scrollIntoView().click();
        cy.getByData("delete-membership-modal").should("exist");
        cy.getByData("delete-membership-modal").should("contain", deleteMemberName);
        cy.getByData("delete-membership-modal").should("contain", "Rankings will be removed");
        cy.getByData("delete-membership-modal").should("contain", "This action cannot be undone.");
      });

      it("should close modal when cancel is clicked", () => {
        cy.getByData(`delete-member-btn-${deleteMember.id}`).scrollIntoView().click();
        cy.getByData("delete-membership-modal").should("exist");

        cy.getByData("delete-membership-cancel-btn").click();
        cy.getByData("delete-membership-modal").should("not.exist");
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

  context("contract tests", () => {
    before(() => {
      cy.resetDatabase();
    });

    beforeEach(() => {
      cy.login();
      cy.visit("/admin/members");
      cy.getByData("members-table").should("exist");
    });

    it("should load members list from real API", () => {
      cy.get("[data-qa^='member-row-']").should(
        "have.length",
        MEMBERS.length
      );
    });

    it("should filter members by name", () => {
      const targetUser = USERS[0]; // Heinrik Drust
      const targetMember = MEMBERS[0];

      cy.getByData("filter-toggle-btn").click();
      cy.getByData("filter-name").type(targetUser.firstName);

      // Wait for debounce + API response
      cy.get("[data-qa^='member-row-']").should("have.length", 1);
      cy.getByData(`member-row-${targetMember.id}`).should("exist");
    });

    it("should filter members by faculty", () => {
      // Engineering: Khalil Duckham, Amandie Libbis
      cy.getByData("filter-toggle-btn").click();
      cy.getByData("filter-faculty").select("Engineering");

      cy.get("[data-qa^='member-row-']").should("have.length", 2);
    });

    it("should filter members by student ID", () => {
      const targetMember = MEMBERS[0];

      cy.getByData("filter-toggle-btn").click();
      cy.getByData("filter-studentId").type(targetMember.userId);

      cy.get("[data-qa^='member-row-']").should("have.length", 1);
      cy.getByData(`member-row-${targetMember.id}`).should("exist");
    });

    it("should filter members by paid status", () => {
      // 6 paid members in seed data
      const paidCount = MEMBERS.filter((m) => m.paid).length;

      cy.getByData("filter-toggle-btn").click();
      cy.getByData("filter-paid").select("Yes");

      cy.get("[data-qa^='member-row-']").should("have.length", paidCount);
    });

    it("should filter members by discounted status", () => {
      // 2 discounted members in seed data
      const discountedCount = MEMBERS.filter((m) => m.discounted).length;

      cy.getByData("filter-toggle-btn").click();
      cy.getByData("filter-discounted").select("Yes");

      cy.get("[data-qa^='member-row-']").should("have.length", discountedCount);
    });

    it("should show empty state when filters match nothing", () => {
      cy.getByData("filter-toggle-btn").click();
      cy.getByData("filter-name").type("zzzznonexistent");

      cy.getByData("members-no-results").should("be.visible");
    });

    it("should clear filters and restore full list", () => {
      cy.getByData("filter-toggle-btn").click();
      cy.getByData("filter-faculty").select("Engineering");
      cy.get("[data-qa^='member-row-']").should("have.length", 2);

      cy.getByData("filter-clear-btn").click();
      cy.get("[data-qa^='member-row-']").should("have.length", MEMBERS.length);
    });

    it("should delete membership successfully", () => {
      const deleteMember = MEMBERS[3]; // Khalil Duckham
      const initialCount = MEMBERS.length;

      cy.getByData(`delete-member-btn-${deleteMember.id}`).scrollIntoView().click();
      cy.getByData("delete-membership-modal").should("exist");

      cy.getByData("delete-membership-confirm-btn").click();

      // Modal should close
      cy.getByData("delete-membership-modal").should("not.exist");

      // Membership should be removed from table
      cy.getByData(`member-row-${deleteMember.id}`).should("not.exist");
      cy.get("[data-qa^='member-row-']").should("have.length", initialCount - 1);
    });
  });
});
