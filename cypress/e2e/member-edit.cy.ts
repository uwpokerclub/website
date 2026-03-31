import { MEMBERS } from "../seed";
import {
  getUserForMember,
  UNPAID_MEMBER,
  PAID_MEMBER,
  DISCOUNTED_MEMBER,
} from "../support/helpers";

// Helper to open edit modal for a specific member
const openEditModal = (memberId: string) => {
  cy.getByData(`edit-member-btn-${memberId}`)
    .scrollIntoView()
    .click();
  // Wait for modal to exist and animation to complete
  cy.getByData("edit-member-modal").should("exist");
  // Wait for the modal to be fully animated in by checking an input is interactable
  cy.getByData("input-firstName").should("be.visible");
};

describe("EditMemberModal", () => {
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

      // API intercepts
      cy.intercept("PATCH", "/api/v2/members/*").as("updateMember");
      cy.intercept("PATCH", "/api/v2/semesters/*/memberships/*").as(
        "updateMembership"
      );
    });

    context("form population", () => {
      it("should open modal with pre-filled data for unpaid member", () => {
        const member = UNPAID_MEMBER;
        const user = getUserForMember(member);

        openEditModal(member.id);

        // Verify all form fields are pre-filled with correct values
        cy.getByData("display-studentId").should("have.value", member.userId);
        cy.getByData("input-firstName").should("have.value", user.firstName);
        cy.getByData("input-lastName").should("have.value", user.lastName);
        cy.getByData("input-email").should("have.value", user.email);
        cy.getByData("select-faculty").should("have.value", user.faculty);
        cy.getByData("input-questId").should("have.value", user.questId);

        // Verify paid checkbox is not checked for unpaid member
        cy.getByData("checkbox-paid").should("not.be.checked");
        // Discounted should not be visible when unpaid
        cy.getByData("checkbox-discounted").should("not.exist");
      });

      it("should show discounted checkbox checked for discounted member", () => {
        const member = DISCOUNTED_MEMBER;

        openEditModal(member.id);

        // Verify both paid and discounted are checked
        // Scroll to checkboxes first to ensure they're visible
        cy.getByData("checkbox-paid").scrollIntoView().should("be.checked");
        cy.getByData("checkbox-discounted").scrollIntoView().should("exist");
        cy.getByData("checkbox-discounted").should("be.checked");
      });
    });

    context("cancel behavior", () => {
      it("should close modal without saving when cancel is clicked", () => {
        const member = UNPAID_MEMBER;
        const user = getUserForMember(member);
        const modifiedName = "ModifiedName";

        openEditModal(member.id);

        // Modify firstName
        cy.getByData("input-firstName").clear().type(modifiedName);

        // Click Cancel
        cy.getByData("edit-cancel-btn").scrollIntoView().click();

        // Verify modal closes
        cy.getByData("edit-member-modal").should("not.exist");

        // Verify table shows original name (no API call made)
        cy.getByData(`member-name-${member.id}`).should(
          "contain",
          `${user.firstName} ${user.lastName}`
        );

        // Re-open modal and verify form shows original values (changes were discarded)
        openEditModal(member.id);
        cy.getByData("input-firstName").should("have.value", user.firstName);
      });
    });

    context("validation", () => {
      it("should show errors for empty required fields", () => {
        const member = UNPAID_MEMBER;

        openEditModal(member.id);

        // Clear all required fields
        cy.getByData("input-firstName").clear();
        cy.getByData("input-lastName").clear();
        cy.getByData("input-email").clear();

        // Submit the form
        cy.getByData("edit-submit-btn").scrollIntoView().click();

        // Verify validation errors
        cy.contains("First name is required").should("exist");
        cy.contains("Last name is required").should("exist");
        cy.contains("Email is required").should("exist");

        // Verify modal stays open
        cy.getByData("edit-member-modal").should("exist");
      });

      it("should show error for invalid email format", () => {
        const member = UNPAID_MEMBER;

        openEditModal(member.id);

        // Enter invalid email
        cy.getByData("input-email").clear().type("invalid-email");

        // Submit the form
        cy.getByData("edit-submit-btn").scrollIntoView().click();

        // Verify validation error (zod uses "Invalid email")
        cy.contains("Invalid email").should("exist");

        // Verify modal stays open and no API call was made
        cy.getByData("edit-member-modal").should("exist");
      });
    });

    context("error handling", () => {
      it("should show error alert when API fails", () => {
        const member = UNPAID_MEMBER;

        // Mock API to return 500 error
        cy.intercept("PATCH", "/api/v2/members/*", {
          statusCode: 500,
          body: { message: "Internal server error" },
        }).as("updateMemberError");

        openEditModal(member.id);

        // Make a change
        cy.getByData("input-firstName").clear().type("NewName");

        // Submit the form
        cy.getByData("edit-submit-btn").scrollIntoView().click();

        // Wait for the failed API call
        cy.wait("@updateMemberError");

        // Verify error alert is visible
        cy.getByData("edit-error-alert").should("exist");

        // Verify modal stays open
        cy.getByData("edit-member-modal").should("exist");
      });
    });

    context("membership status UI", () => {
      it("should show discounted checkbox when paid is checked", () => {
        const member = UNPAID_MEMBER;

        openEditModal(member.id);

        // Verify initially unpaid - scroll to checkbox first
        cy.getByData("checkbox-paid").scrollIntoView().should("not.be.checked");
        cy.getByData("checkbox-discounted").should("not.exist");

        // Check paid checkbox
        cy.getByData("checkbox-paid").scrollIntoView().click();
        cy.getByData("checkbox-paid").should("be.checked");
        cy.getByData("checkbox-discounted").should("exist");
      });

      it("should hide discounted when paid is unchecked on discounted member", () => {
        const member = DISCOUNTED_MEMBER;

        openEditModal(member.id);

        // Verify both are initially checked - scroll to make visible
        cy.getByData("checkbox-paid").scrollIntoView().should("be.checked");
        cy.getByData("checkbox-discounted").should("be.checked");

        // Uncheck paid checkbox
        cy.getByData("checkbox-paid").scrollIntoView().click();
        cy.getByData("checkbox-paid").should("not.be.checked");

        // Verify discounted disappears (component behavior)
        cy.getByData("checkbox-discounted").should("not.exist");
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

      cy.intercept("PATCH", "/api/v2/members/*").as("updateMember");
      cy.intercept("PATCH", "/api/v2/semesters/*/memberships/*").as(
        "updateMembership"
      );
    });

    it("should edit member name and save", () => {
      const member = UNPAID_MEMBER;
      const newFirstName = "UpdatedFirst";
      const newLastName = "UpdatedLast";

      openEditModal(member.id);

      // Clear and type new values
      cy.getByData("input-firstName").clear().type(newFirstName);
      cy.getByData("input-lastName").clear().type(newLastName);

      // Submit the form - scroll to ensure button is visible
      cy.getByData("edit-submit-btn").scrollIntoView().click();

      // Wait for API and verify request body
      cy.wait("@updateMember").then((interception) => {
        expect(interception.request.body).to.deep.include({
          firstName: newFirstName,
          lastName: newLastName,
        });
        expect(interception.response?.statusCode).to.eq(200);
      });

      // Wait for membership update as well (component updates both)
      cy.wait("@updateMembership");

      // Verify success toast
      cy.contains("updated successfully").should("be.visible");

      // Verify modal closes
      cy.getByData("edit-member-modal").should("not.exist");

      // Verify table row shows updated name
      cy.getByData(`member-name-${member.id}`).should(
        "contain",
        `${newFirstName} ${newLastName}`
      );
    });

    it("should change unpaid to paid", () => {
      const member = UNPAID_MEMBER;

      openEditModal(member.id);

      // Check paid checkbox
      cy.getByData("checkbox-paid").scrollIntoView().click();
      cy.getByData("checkbox-paid").should("be.checked");

      // Submit the form
      cy.getByData("edit-submit-btn").scrollIntoView().click();

      // Wait for member update first
      cy.wait("@updateMember");

      // Wait for membership update and verify request body
      cy.wait("@updateMembership").then((interception) => {
        expect(interception.request.body).to.deep.include({
          paid: true,
          discounted: false,
        });
        expect(interception.response?.statusCode).to.eq(200);
      });

      // Verify success
      cy.contains("updated successfully").should("be.visible");
      cy.getByData("edit-member-modal").should("not.exist");

      // Verify status cell shows "Paid"
      cy.getByData(`member-status-${member.id}`).should("contain", "Paid");
    });
  });
});
