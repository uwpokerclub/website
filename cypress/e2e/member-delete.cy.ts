import { MEMBERS } from "../seed";
import { getUserForMember, PAID_MEMBER } from "../support/helpers";

// Helper to open edit modal for a specific member
const openEditModal = (memberId: string) => {
  cy.getByData(`edit-member-btn-${memberId}`)
    .scrollIntoView()
    .click({ force: true });
  cy.getByData("edit-member-modal").should("exist");
  cy.getByData("input-firstName").should("be.visible");
};

describe("DeleteMemberModal", () => {
  beforeEach(() => {
    cy.resetDatabase();
    cy.login();
    cy.visit("/admin/members");
    cy.getByData("members-table").should("exist");
  });

  context("danger zone visibility", () => {
    it("should display danger zone in edit modal", () => {
      const member = PAID_MEMBER;

      openEditModal(member.id);

      cy.getByData("danger-zone").scrollIntoView().should("exist");
      cy.getByData("danger-zone").should("contain", "Danger Zone");
      cy.getByData("danger-zone").should(
        "contain",
        "Permanently delete this member"
      );
      cy.getByData("delete-member-btn").should("exist");
    });
  });

  context("modal behavior", () => {
    // Use PAID_MEMBER (Khalil Duckham) - no event participation
    const deleteMember = PAID_MEMBER;
    const deleteUser = getUserForMember(deleteMember);
    const deleteMemberName = `${deleteUser.firstName} ${deleteUser.lastName}`;

    it("should open delete modal when delete button is clicked", () => {
      openEditModal(deleteMember.id);

      cy.getByData("delete-member-btn").scrollIntoView().click();

      cy.getByData("delete-member-modal").should("exist");
    });

    it("should display member name in confirmation message", () => {
      openEditModal(deleteMember.id);
      cy.getByData("delete-member-btn").scrollIntoView().click();

      cy.getByData("delete-member-modal").should("contain", deleteMemberName);
    });

    it("should display cascade warning", () => {
      openEditModal(deleteMember.id);
      cy.getByData("delete-member-btn").scrollIntoView().click();

      cy.getByData("delete-member-modal").should(
        "contain",
        "all their memberships across all semesters"
      );
      cy.getByData("delete-member-modal").should(
        "contain",
        "This action cannot be undone."
      );
    });

    it("should close delete modal when cancel is clicked", () => {
      openEditModal(deleteMember.id);
      cy.getByData("delete-member-btn").scrollIntoView().click();
      cy.getByData("delete-member-modal").should("exist");

      cy.getByData("delete-member-cancel-btn").click();

      cy.getByData("delete-member-modal").should("not.exist");
      // Edit modal should still be open
      cy.getByData("edit-member-modal").should("exist");
    });

    it("should delete member successfully", () => {
      const initialCount = MEMBERS.length;

      cy.intercept("DELETE", `/api/v2/members/${deleteMember.userId}`).as(
        "deleteMember"
      );

      openEditModal(deleteMember.id);
      cy.getByData("delete-member-btn").scrollIntoView().click();
      cy.getByData("delete-member-modal").should("exist");

      cy.getByData("delete-member-confirm-btn").click();

      // Verify API call
      cy.wait("@deleteMember").then((interception) => {
        expect(interception.response?.statusCode).to.eq(204);
      });

      // Both modals should close
      cy.getByData("delete-member-modal").should("not.exist");
      cy.getByData("edit-member-modal").should("not.exist");

      // Verify success toast
      cy.contains("deleted successfully").should("be.visible");

      // Member should be removed from table
      cy.getByData(`member-row-${deleteMember.id}`).should("not.exist");
      cy.get("[data-qa^='member-row-']").should(
        "have.length",
        initialCount - 1
      );
    });
  });

  context("error handling", () => {
    const deleteMember = PAID_MEMBER;

    it("should show error when API fails", () => {
      cy.intercept("DELETE", `/api/v2/members/${deleteMember.userId}`, {
        statusCode: 500,
        body: { message: "Internal server error" },
      }).as("deleteMemberError");

      openEditModal(deleteMember.id);
      cy.getByData("delete-member-btn").scrollIntoView().click();
      cy.getByData("delete-member-modal").should("exist");

      cy.getByData("delete-member-confirm-btn").click();

      cy.wait("@deleteMemberError");

      // Error should be displayed
      cy.getByData("delete-member-error-alert").should("exist");

      // Modal should stay open
      cy.getByData("delete-member-modal").should("exist");
    });

    it("should show error when member not found", () => {
      cy.intercept("DELETE", `/api/v2/members/${deleteMember.userId}`, {
        statusCode: 404,
        body: { message: "Member not found" },
      }).as("deleteMemberNotFound");

      openEditModal(deleteMember.id);
      cy.getByData("delete-member-btn").scrollIntoView().click();

      cy.getByData("delete-member-confirm-btn").click();

      cy.wait("@deleteMemberNotFound");

      cy.getByData("delete-member-error-alert").should("exist");
      cy.getByData("delete-member-modal").should("exist");
    });
  });
});
