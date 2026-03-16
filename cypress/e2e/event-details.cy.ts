import {
  SEMESTER,
  EVENT,
  ENDED_EVENT,
  MEMBERS,
  ACTIVE_EVENT_PARTICIPANTS,
  ENDED_EVENT_PARTICIPANTS,
} from "../seed";
import { getUserForMember } from "../support/helpers";

// Helper to visit event details page
const visitEventDetails = (eventId: string) => {
  cy.visit(`/admin/events/${eventId}`);
};

describe("EventDetails", () => {
  before(() => {
    cy.resetDatabase();
  });

  context("stubbed tests", () => {
    context("active event", () => {
      beforeEach(() => {
        cy.login();
        cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
        cy.intercept("GET", /\/api\/v2\/semesters\/.*\/events\/1$/, { fixture: "event-details.json" }).as("getEvent");
        cy.intercept("GET", /\/api\/v2\/semesters\/.*\/events\/1\/entries/, { fixture: "event-entries.json" }).as("getEntries");

        visitEventDetails(EVENT.id);
        // Wait for page to load by verifying event title is visible
        cy.contains("h1", EVENT.name).should("be.visible");
      });

      context("event info and stats", () => {
        it("displays event header with title, format, date, and active status", () => {
          // Verify event title
          cy.contains("h1", EVENT.name).should("be.visible");

          // Verify format
          cy.contains("Format").should("be.visible");
          cy.contains(EVENT.format).should("be.visible");

          // Verify date is displayed
          cy.contains("Date").should("be.visible");

          // Verify active status badge
          cy.contains("Active").should("be.visible");
        });

        it("displays stats bar with Total Entries, Players, Rebuys, Points Multiplier", () => {
          const expectedPlayers = ACTIVE_EVENT_PARTICIPANTS.length;
          const expectedRebuys = EVENT.rebuys;
          const expectedTotal = expectedPlayers + expectedRebuys;

          // Total Entries
          cy.getByData("stat-total-entries").should("contain", String(expectedTotal));

          // Players
          cy.getByData("stat-players").should("contain", String(expectedPlayers));

          // Rebuys
          cy.getByData("stat-rebuys").should("contain", String(expectedRebuys));

          // Points Multiplier
          cy.getByData("stat-points-multiplier").should("contain", `${EVENT.pointsMultiplier}x`);
        });
      });

      context("entries tab", () => {
        it("displays participants list with correct columns", () => {
          // Verify table headers (using exist due to CSS overflow clipping)
          cy.contains("th", "#").should("exist");
          cy.contains("th", "First Name").should("exist");
          cy.contains("th", "Last Name").should("exist");
          cy.contains("th", "Student Number").should("exist");
          cy.contains("th", "Signed Out At").should("exist");
          cy.contains("th", "Place").should("exist");
          cy.contains("th", "Actions").should("exist");

          // Verify participant count display
          const expectedPlayers = ACTIVE_EVENT_PARTICIPANTS.length;
          cy.contains(`${expectedPlayers} Entries`).should("be.visible");
          cy.contains(`${expectedPlayers} Players`).should("be.visible");

          // Verify a participant is visible (use exist due to CSS clipping)
          const firstParticipant = ACTIVE_EVENT_PARTICIPANTS[0];
          const user = getUserForMember(firstParticipant.membershipId);
          if (user) {
            cy.contains(user.firstName).should("exist");
            cy.contains(user.lastName).should("exist");
          }
        });

      });

      context("structure tab", () => {
        it("displays blind levels table", () => {
          // Click Structure subtab
          cy.contains("button", "Structure").click();

          // Verify structure table headers
          cy.contains("th", "Level").should("be.visible");
          cy.contains("th", "Small Blind").should("be.visible");
          cy.contains("th", "Big Blind").should("be.visible");
          cy.contains("th", "Ante").should("be.visible");
          cy.contains("th", "Time").should("be.visible");

          // Verify some blind level data (first level: 10/20/0/15)
          cy.contains("td", "10").should("exist");
          cy.contains("td", "20").should("exist");
        });
      });

      context("end event flow", () => {
        it("opens and cancels end event modal", () => {
          // Click End Event button
          cy.getByData("end-event-btn").click();

          // Verify modal appears
          cy.getByData("end-event-modal").should("exist");
          cy.contains("Are you sure you want to end the event?").should("exist");

          // Click cancel/go back
          cy.getByData("end-event-cancel-btn").click();

          // Verify modal closed
          cy.getByData("end-event-modal").should("not.exist");

          // Verify event still active (use exist due to CSS overflow clipping)
          cy.contains("Active").should("exist");
        });
      });
    });

    context("ended event", () => {
      beforeEach(() => {
        cy.login();
        cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
        cy.intercept("GET", /\/api\/v2\/semesters\/.*\/events\/2$/, { fixture: "ended-event-details.json" }).as("getEvent");
        cy.intercept("GET", /\/api\/v2\/semesters\/.*\/events\/2\/entries/, { fixture: "ended-event-entries.json" }).as("getEntries");

        visitEventDetails(ENDED_EVENT.id);
        // Wait for page to load by verifying event title is visible
        cy.contains("h1", ENDED_EVENT.name).should("be.visible");
      });

      it("displays ended state with correct UI", () => {
        // Status badge
        cy.contains("Ended").should("be.visible");

        // Placements are displayed
        const firstPlaceParticipant = ENDED_EVENT_PARTICIPANTS.find(
          (p) => p.placement === 1
        );
        if (firstPlaceParticipant) {
          const user = getUserForMember(firstPlaceParticipant.membershipId);
          if (user) {
            // Verify first place participant is visible (use exist due to CSS clipping)
            cy.contains(user.firstName).should("exist");
          }
        }

        // Verify Place column shows placements (use exist due to CSS clipping)
        cy.getByData("entries-table").within(() => {
          cy.contains("td", "1").should("exist");
          cy.contains("td", "2").should("exist");
        });

        // Action buttons hidden
        cy.getByData("register-members-btn").should("not.exist");
        cy.getByData("rebuy-btn").should("not.exist");
        cy.getByData("end-event-btn").should("not.exist");

        // Entry action buttons hidden
        cy.getByData("sign-in-btn").should("not.exist");
        cy.getByData("sign-out-btn").should("not.exist");
        cy.getByData("remove-btn").should("not.exist");
      });
    });

    context("error states", () => {
      beforeEach(() => {
        cy.login();
        cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
      });

      it("displays error when event fetch fails", () => {
        cy.intercept("GET", /\/api\/v2\/semesters\/.*\/events\/\d+$/, {
          statusCode: 500,
          body: { message: "Internal server error" },
        }).as("getEventError");

        visitEventDetails(EVENT.id);

        // Verify error state is displayed
        cy.contains("Internal server error").should("be.visible");
        cy.contains("Retry").should("exist");
      });

      it("handles rebuy API failure gracefully", () => {
        cy.intercept("GET", /\/api\/v2\/semesters\/.*\/events\/1$/, { fixture: "event-details.json" }).as("getEvent");
        cy.intercept("GET", /\/api\/v2\/semesters\/.*\/events\/1\/entries/, { fixture: "event-entries.json" }).as("getEntries");
        cy.intercept("POST", /\/api\/v2\/semesters\/.*\/events\/\d+\/rebuy/, {
          statusCode: 500,
          body: { message: "Failed to record rebuy" },
        }).as("rebuyError");

        visitEventDetails(EVENT.id);
        // Wait for page to load
        cy.contains("h1", EVENT.name).should("be.visible");

        // Click rebuy button
        cy.getByData("rebuy-btn").click();
        cy.wait("@rebuyError");

        // Verify error toast (use exist and longer timeout)
        cy.contains("Failed to record rebuy", { timeout: 5000 }).should("exist");

        // Verify rebuy count unchanged
        cy.getByData("stat-rebuys").should("contain", String(EVENT.rebuys));
      });

      it("handles end event API failure gracefully", () => {
        cy.intercept("GET", /\/api\/v2\/semesters\/.*\/events\/1$/, { fixture: "event-details.json" }).as("getEvent");
        cy.intercept("GET", /\/api\/v2\/semesters\/.*\/events\/1\/entries/, { fixture: "event-entries.json" }).as("getEntries");

        visitEventDetails(EVENT.id);
        // Wait for page to load
        cy.contains("h1", EVENT.name).should("be.visible");

        // Override end event intercept with error
        cy.intercept("POST", /\/api\/v2\/semesters\/.*\/events\/\d+\/end/, {
          statusCode: 500,
          body: { message: "Failed to end event" },
        }).as("endEventError");

        // Click End Event button
        cy.getByData("end-event-btn").click();
        cy.getByData("end-event-modal").should("exist");

        // Confirm end event
        cy.getByData("end-event-confirm-btn").click();
        cy.wait("@endEventError");

        // Verify error message in modal
        cy.getByData("end-event-error-alert").should("be.visible");
        cy.contains("Failed to end event").should("be.visible");

        // Modal should still be open
        cy.getByData("end-event-modal").should("exist");
      });
    });
  });

  context("contract tests", () => {
    before(() => {
      cy.resetDatabase();
    });

    beforeEach(() => {
      cy.login();

      // API intercepts for verifying calls
      cy.intercept("POST", /\/api\/v2\/semesters\/.*\/events\/\d+\/rebuy/).as("rebuy");
      cy.intercept("POST", /\/api\/v2\/semesters\/.*\/events\/\d+\/end/).as("endEvent");
      cy.intercept("POST", /\/api\/v2\/semesters\/.*\/events\/\d+\/restart/).as("restartEvent");
      cy.intercept("POST", /\/api\/v2\/semesters\/.*\/events\/\d+\/entries\/.*\/sign-out/).as("signOut");
      cy.intercept("POST", /\/api\/v2\/semesters\/.*\/events\/\d+\/entries\/.*\/sign-in/).as("signIn");
      cy.intercept("DELETE", /\/api\/v2\/semesters\/.*\/events\/\d+\/entries\/.*/).as("removeEntry");
    });

    it("should search entries and show no results for unmatched search", () => {
      visitEventDetails(EVENT.id);
      cy.contains("h1", EVENT.name).should("be.visible");

      const nonExistentName = "ZZZNonExistent";
      cy.getByData("input-search").type(nonExistentName);
      cy.contains("No results found").should("be.visible");
      cy.contains(`matching "${nonExistentName}"`).should("be.visible");
    });

    it("should load active event page from real API", () => {
      visitEventDetails(EVENT.id);
      cy.contains("h1", EVENT.name).should("be.visible");
      cy.contains("Active").should("be.visible");
      cy.getByData("stat-players").should("contain", String(ACTIVE_EVENT_PARTICIPANTS.length));
    });

    it("should sign out and sign in a participant", () => {
      visitEventDetails(EVENT.id);
      cy.contains("h1", EVENT.name).should("be.visible");

      // Sign out
      cy.getByData("sign-out-btn").first().click();
      cy.wait("@signOut").then((interception) => {
        expect(interception.response?.statusCode).to.eq(200);
      });
      cy.getByData("sign-in-btn").should("exist");

      // Sign back in
      cy.getByData("sign-in-btn").first().click();
      cy.wait("@signIn").then((interception) => {
        expect(interception.response?.statusCode).to.eq(200);
      });
      cy.getByData("sign-out-btn").should("exist");
    });

    it("should end event and restart it", () => {
      visitEventDetails(EVENT.id);
      cy.contains("h1", EVENT.name).should("be.visible");

      // End event
      cy.getByData("end-event-btn").click();
      cy.getByData("end-event-modal").should("exist");
      cy.getByData("end-event-confirm-btn").click();

      cy.wait("@endEvent").then((interception) => {
        expect(interception.response?.statusCode).to.eq(204);
      });

      cy.getByData("end-event-modal").should("not.exist");
      cy.contains("Ended").should("exist");
      cy.getByData("end-event-btn").should("not.exist");
      cy.getByData("restart-event-btn").should("exist");

      // Restart
      cy.getByData("restart-event-btn").click();
      cy.wait("@restartEvent").then((interception) => {
        expect(interception.response?.statusCode).to.eq(204);
      });

      cy.contains("Active").should("exist");
      cy.getByData("restart-event-btn").should("not.exist");
      cy.getByData("end-event-btn").should("exist");
    });
  });
});
