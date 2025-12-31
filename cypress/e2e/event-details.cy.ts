import {
  SEMESTER,
  EVENT,
  ENDED_EVENT,
  MEMBERS,
  USERS,
  ACTIVE_EVENT_PARTICIPANTS,
  ENDED_EVENT_PARTICIPANTS,
} from "../seed";

// Helper to get user by membership
const getUserForMember = (membershipId: string) => {
  const member = MEMBERS.find((m) => m.id === membershipId);
  return USERS.find((u) => u.id === member?.userId);
};

// Helper to visit event details page
const visitEventDetails = (eventId: string) => {
  cy.visit(`/admin/events/${eventId}`);
};

describe("EventDetails", () => {
  beforeEach(() => {
    cy.exec("npm run db:reset && npm run db:seed", { timeout: 30000 });
    cy.login("e2e_user", "password");

    // API intercepts
    cy.intercept("GET", /\/api\/v2\/semesters\/.*\/events\/\d+$/).as("getEvent");
    cy.intercept("GET", /\/api\/v2\/semesters\/.*\/events\/\d+\/entries/).as(
      "getEntries"
    );
    cy.intercept("POST", /\/api\/v2\/semesters\/.*\/events\/\d+\/rebuy/).as(
      "rebuy"
    );
    cy.intercept("POST", /\/api\/v2\/semesters\/.*\/events\/\d+\/end/).as(
      "endEvent"
    );
    cy.intercept("POST", /\/api\/v2\/semesters\/.*\/events\/\d+\/restart/).as(
      "restartEvent"
    );
    cy.intercept(
      "POST",
      /\/api\/v2\/semesters\/.*\/events\/\d+\/entries\/.*\/sign-out/
    ).as("signOut");
    cy.intercept(
      "POST",
      /\/api\/v2\/semesters\/.*\/events\/\d+\/entries\/.*\/sign-in/
    ).as("signIn");
    cy.intercept(
      "DELETE",
      /\/api\/v2\/semesters\/.*\/events\/\d+\/entries\/.*/
    ).as("removeEntry");
  });

  context("active event", () => {
    beforeEach(() => {
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
        cy.contains("Total Entries")
          .parent()
          .should("contain", String(expectedTotal));

        // Players
        cy.contains("Players").parent().should("contain", String(expectedPlayers));

        // Rebuys
        cy.contains("Rebuys").parent().should("contain", String(expectedRebuys));

        // Points Multiplier
        cy.contains("Points Multiplier")
          .parent()
          .should("contain", `${EVENT.pointsMultiplier}x`);
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

      it("searches participants by name", () => {
        const firstParticipant = ACTIVE_EVENT_PARTICIPANTS[0];
        const user = getUserForMember(firstParticipant.membershipId);

        if (user) {
          // Type search query
          cy.getByData("input-search").type(user.firstName);

          // Wait for debounce and verify results
          cy.contains(`matching "${user.firstName}"`).should("be.visible");
          cy.contains(user.firstName).should("be.visible");

          // Verify showing filtered count
          cy.contains("Showing 1 of 1 entries").should("be.visible");
        }
      });

      it("shows no results state for unmatched search", () => {
        const nonExistentName = "ZZZNonExistent";

        cy.getByData("input-search").type(nonExistentName);

        // Wait for debounce and verify no results
        cy.contains("No results found").should("be.visible");
        cy.contains(`matching "${nonExistentName}"`).should("be.visible");
      });

      context("participant actions", () => {
        it("signs out a participant", () => {
          // Find the first sign-out button and click it
          cy.getByData("sign-out-btn").first().click();

          cy.wait("@signOut").then((interception) => {
            expect(interception.response?.statusCode).to.eq(200);
          });

          // After sign out, sign-in button should appear
          cy.getByData("sign-in-btn").should("exist");
        });

        it("signs in a signed-out participant", () => {
          // First sign out a participant
          cy.getByData("sign-out-btn").first().click();
          cy.wait("@signOut");

          // Wait for UI to update
          cy.getByData("sign-in-btn").should("exist");

          // Sign back in
          cy.getByData("sign-in-btn").first().click();

          cy.wait("@signIn").then((interception) => {
            expect(interception.response?.statusCode).to.eq(200);
          });

          // Sign-out button should reappear
          cy.getByData("sign-out-btn").should("exist");
        });

        it("removes a participant", () => {
          // Get initial count
          const initialCount = ACTIVE_EVENT_PARTICIPANTS.length;

          // Click remove button
          cy.getByData("remove-btn").first().click();

          cy.wait("@removeEntry").then((interception) => {
            expect([200, 204]).to.include(interception.response?.statusCode);
          });

          // Verify entry count decremented
          cy.contains(`${initialCount - 1} Entries`).should("be.visible");
        });

        it("updates entry count after removal", () => {
          const initialPlayers = ACTIVE_EVENT_PARTICIPANTS.length;

          // Remove a participant
          cy.getByData("remove-btn").first().click();
          cy.wait("@removeEntry");

          // Verify both stats updated
          cy.contains(`${initialPlayers - 1} Players`).should("be.visible");
          cy.contains("Players")
            .parent()
            .should("contain", String(initialPlayers - 1));
        });
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

        // Verify some blind level data (first level: 10/20/20/5)
        cy.contains("td", "10").should("exist");
        cy.contains("td", "20").should("exist");
      });
    });

    context("rebuy functionality", () => {
      it("increments rebuy count and shows success toast", () => {
        const initialRebuys = EVENT.rebuys;

        // Click rebuy button
        cy.getByData("rebuy-btn").click();

        cy.wait("@rebuy").then((interception) => {
          expect(interception.response?.statusCode).to.eq(204);
        });

        // Verify rebuy count incremented
        cy.contains("Rebuys")
          .parent()
          .should("contain", String(initialRebuys + 1));

        // Verify toast message (use exist and longer timeout)
        cy.contains("Rebuy recorded", { timeout: 5000 }).should("exist");
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

      it("confirms end event and updates UI state", () => {
        // Click End Event button
        cy.getByData("end-event-btn").click();
        cy.getByData("end-event-modal").should("exist");

        // Confirm end event
        cy.getByData("end-event-confirm-btn").click();

        cy.wait("@endEvent").then((interception) => {
          expect(interception.response?.statusCode).to.eq(204);
        });

        // Verify modal closed
        cy.getByData("end-event-modal").should("not.exist");

        // Verify status changed to Ended (use exist due to CSS overflow clipping)
        cy.contains("Ended").should("exist");

        // Verify End Event button hidden, Restart Event visible
        cy.getByData("end-event-btn").should("not.exist");
        cy.getByData("restart-event-btn").should("exist");

        // Verify other action buttons hidden
        cy.getByData("rebuy-btn").should("not.exist");
        cy.getByData("register-members-btn").should("not.exist");

        // Verify entry action buttons are hidden
        cy.getByData("sign-out-btn").should("not.exist");
        cy.getByData("remove-btn").should("not.exist");
      });
    });
  });

  context("ended event", () => {
    beforeEach(() => {
      visitEventDetails(ENDED_EVENT.id);
      // Wait for page to load by verifying event title is visible
      cy.contains("h1", ENDED_EVENT.name).should("be.visible");
    });

    it("displays ended status badge", () => {
      cy.contains("Ended").should("be.visible");
    });

    it("shows placements in entries table", () => {
      // Verify placements are displayed
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
      cy.get("table").within(() => {
        cy.contains("td", "1").should("exist");
        cy.contains("td", "2").should("exist");
      });
    });

    it("hides action buttons (Register, Rebuy, End)", () => {
      cy.getByData("register-members-btn").should("not.exist");
      cy.getByData("rebuy-btn").should("not.exist");
      cy.getByData("end-event-btn").should("not.exist");
    });

    it("hides entry action buttons (sign in/out, remove)", () => {
      cy.getByData("sign-in-btn").should("not.exist");
      cy.getByData("sign-out-btn").should("not.exist");
      cy.getByData("remove-btn").should("not.exist");
    });

    context("restart event flow", () => {
      it("restarts event and restores active state", () => {
        // Click Restart Event
        cy.getByData("restart-event-btn").click();

        cy.wait("@restartEvent").then((interception) => {
          expect(interception.response?.statusCode).to.eq(204);
        });

        // Verify status changed to Active (use exist due to CSS clipping)
        cy.contains("Active").should("exist");

        // Verify Restart button hidden, other buttons visible
        cy.getByData("restart-event-btn").should("not.exist");
        cy.getByData("end-event-btn").should("exist");
        cy.getByData("rebuy-btn").should("exist");
        cy.getByData("register-members-btn").should("exist");

        // Verify success toast (use exist and longer timeout)
        cy.contains("Event restarted", { timeout: 5000 }).should("exist");
      });
    });
  });

  context("error states", () => {
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
      // Override rebuy intercept with error BEFORE visiting
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
      cy.contains("Rebuys").parent().should("contain", String(EVENT.rebuys));
    });

    it("handles end event API failure gracefully", () => {
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
