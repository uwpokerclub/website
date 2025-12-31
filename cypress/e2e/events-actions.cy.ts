import { EVENT, ENDED_EVENT } from "../seed";

// Helper to open actions dropdown
const openActionsMenu = (eventId: string) => {
  cy.getByData(`actions-menu-btn-${eventId}`).scrollIntoView().click();
};

// Helper to open edit modal
const openEditModal = (eventId: string) => {
  openActionsMenu(eventId);
  cy.getByData(`edit-event-btn-${eventId}`).click();
  cy.getByData("edit-event-modal").should("exist");
  cy.getByData("input-name").should("be.visible");
};

describe("Event Actions", () => {
  beforeEach(() => {
    cy.exec("npm run db:reset && npm run db:seed", { timeout: 30000 });
    cy.login("e2e_user", "password");
    cy.visit("/admin/events");
    cy.getByData("events-table").should("exist");

    // API intercepts
    cy.intercept("PATCH", /\/api\/v2\/semesters\/.*\/events\/.*/).as(
      "editEvent"
    );
    cy.intercept("POST", /\/api\/v2\/semesters\/.*\/events\/.*\/end/).as(
      "endEvent"
    );
    cy.intercept("POST", /\/api\/v2\/semesters\/.*\/events\/.*\/restart/).as(
      "restartEvent"
    );
  });

  context("dropdown menu", () => {
    it("should open and display correct options for active event", () => {
      openActionsMenu(EVENT.id);
      // Use exist instead of visible due to CSS overflow clipping
      cy.getByData(`edit-event-btn-${EVENT.id}`).should("exist");
      cy.getByData(`end-event-btn-${EVENT.id}`).should("exist");
      cy.getByData(`restart-event-btn-${EVENT.id}`).should("not.exist");
    });

    it("should open and display correct options for ended event", () => {
      openActionsMenu(ENDED_EVENT.id);
      // Edit button is disabled (rendered as span without data-qa) for ended events
      cy.getByData(`edit-event-btn-${ENDED_EVENT.id}`).should("not.exist");
      cy.contains("Edit Event").should("exist"); // Disabled span still shows text
      cy.getByData(`restart-event-btn-${ENDED_EVENT.id}`).should("exist");
      cy.getByData(`end-event-btn-${ENDED_EVENT.id}`).should("not.exist");
    });
  });

  context("edit event", () => {
    it("should open edit modal with pre-filled data", () => {
      openEditModal(EVENT.id);
      cy.getByData("input-name").should("have.value", EVENT.name);
    });

    it("should edit event name and save changes", () => {
      const newName = "Updated Event Name";
      openEditModal(EVENT.id);
      cy.getByData("input-name").clear().type(newName);
      cy.getByData("edit-event-submit-btn").scrollIntoView().click();

      cy.wait("@editEvent").then((interception) => {
        expect(interception.request.body).to.deep.include({ name: newName });
        expect(interception.response?.statusCode).to.eq(200);
      });

      cy.getByData("edit-event-modal").should("not.exist");
      cy.getByData(`event-name-${EVENT.id}`).should("contain", newName);
    });

    it("should close modal without saving when cancel is clicked", () => {
      openEditModal(EVENT.id);
      cy.getByData("input-name").clear().type("Modified Name");
      cy.getByData("edit-event-cancel-btn").scrollIntoView().click();

      cy.getByData("edit-event-modal").should("not.exist");
      cy.getByData(`event-name-${EVENT.id}`).should("contain", EVENT.name);
    });
  });

  context("end event", () => {
    it("should show confirmation modal when clicking end", () => {
      openActionsMenu(EVENT.id);
      cy.getByData(`end-event-btn-${EVENT.id}`).click();
      cy.getByData(`end-confirm-modal-${EVENT.id}`).should("exist");
    });

    it("should cancel end event action", () => {
      openActionsMenu(EVENT.id);
      cy.getByData(`end-event-btn-${EVENT.id}`).click();
      cy.getByData(`end-confirm-modal-${EVENT.id}`).should("exist");
      cy.getByData(`end-confirm-cancel-btn-${EVENT.id}`).click();

      cy.getByData(`end-confirm-modal-${EVENT.id}`).should("not.exist");
      cy.getByData(`event-status-${EVENT.id}`).should("contain", "Active");
    });

    it("should end active event and update status to Ended", () => {
      openActionsMenu(EVENT.id);
      cy.getByData(`end-event-btn-${EVENT.id}`).click();
      cy.getByData(`end-confirm-btn-${EVENT.id}`).click();

      cy.wait("@endEvent").then((interception) => {
        expect(interception.response?.statusCode).to.eq(204);
      });

      cy.getByData(`end-confirm-modal-${EVENT.id}`).should("not.exist");
      cy.getByData(`event-status-${EVENT.id}`).should("contain", "Ended");
    });
  });

  context("restart event", () => {
    it("should restart ended event and update status to Active", () => {
      openActionsMenu(ENDED_EVENT.id);
      cy.getByData(`restart-event-btn-${ENDED_EVENT.id}`).click();

      cy.wait("@restartEvent").then((interception) => {
        expect(interception.response?.statusCode).to.eq(204);
      });

      cy.getByData(`event-status-${ENDED_EVENT.id}`).should("contain", "Active");
    });
  });

  context("error handling", () => {
    it("should handle end event API failure gracefully", () => {
      cy.intercept("POST", /\/api\/v2\/semesters\/.*\/events\/.*\/end/, {
        statusCode: 500,
        body: { message: "Internal server error" },
      }).as("endEventError");

      openActionsMenu(EVENT.id);
      cy.getByData(`end-event-btn-${EVENT.id}`).click();
      cy.getByData(`end-confirm-btn-${EVENT.id}`).click();

      cy.wait("@endEventError");
      // Status should remain Active
      cy.getByData(`event-status-${EVENT.id}`).should("contain", "Active");
    });
  });
});
