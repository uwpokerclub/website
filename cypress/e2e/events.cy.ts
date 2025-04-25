import { EVENT, MEMBERS, SEMESTER, USERS } from "../seed";
import { Event, Membership, Semester, Structure, User } from "../types";

describe("Events", () => {
  beforeEach(() => {
    cy.exec("npm run db:reset && npm run db:seed");
    cy.login("e2e_user", "password");
  });

  context("Create Event", () => {
    beforeEach(() => {
      cy.visit("/admin/events");
    });

    context("No existing blind structure", () => {
      it("should create a new event and structure", () => {
        cy.intercept("POST", "/api/events").as("createEvent");

        // Click the create event button
        cy.getByData("create-event-btn").click();

        // Ensure the page is redirected to the creation page
        cy.location("pathname").should("eq", "/admin/events/new");

        // Input data into the form
        cy.getByData("input-name").type("Winter 2025 Event #2");
        cy.getByData("select-semester").select("Winter 2025");
        cy.getByData("input-date").type("2025-01-06T19:00");
        cy.getByData("select-format").select("No Limit Hold'em");
        cy.getByData("input-points-multiplier").type("1");
        cy.getByData("input-additional-details").type("E2E Test Run");

        // Create the structure
        cy.getByData("tab-new-structure").click();
        cy.getByData("input-structure-name").type("Structure B");
        cy.getByData("blind-0-small").type("10");
        cy.getByData("blind-0-big").type("20");
        cy.getByData("blind-0-ante").type("20");
        cy.getByData("blind-0-time").type("5");

        // Create 5 levels based on the inputed level
        cy.getByData("add-level-btn").click();
        cy.getByData("add-level-btn").click();
        cy.getByData("add-level-btn").click();
        cy.getByData("add-level-btn").click();

        // Submit the form
        cy.getByData("submit-btn").click();

        // Ensure the user is redirected back to the events list
        cy.location("pathname").should("eq", "/admin/events/");

        // Ensure the new event is listed
        cy.wait<Cypress.RequestBody, Event>("@createEvent").then(
          ({ response }) => {
            cy.getByData(`${response.body.id}-name`).should(
              "contain",
              "Winter 2025 Event #2"
            );
            cy.getByData(`${response.body.id}-format`).should(
              "contain",
              "No Limit Hold'em"
            );
            cy.getByData(`${response.body.id}-date`).should(
              "contain",
              "Monday, January 6, 2025 at 7:00 PM"
            );
            cy.getByData(`${response.body.id}-additional-details`).should(
              "contain",
              "E2E Test Run"
            );
          }
        );
      });
    });

    context("With existing blind structure", () => {
      it("should create a new event with an existing structure", () => {
        cy.intercept("POST", "/api/events").as("createEvent");

        // Click the create event button
        cy.getByData("create-event-btn").click();

        // Ensure the page is redirected to the creation page
        cy.location("pathname").should("eq", "/admin/events/new");

        // Input data into the form
        cy.getByData("input-name").type("Winter 2025 Event #2");
        cy.getByData("select-semester").select("Winter 2025");
        cy.getByData("input-date").type("2025-01-06T19:00");
        cy.getByData("select-format").select("No Limit Hold'em");
        cy.getByData("input-points-multiplier").type("1");
        cy.getByData("input-additional-details").type("E2E Test Run");

        // Select the structure
        cy.getByData("select-structure").select("Structure A");

        // Submit the form
        cy.getByData("submit-btn").click();

        // Ensure the user is redirected back to the events list
        cy.location("pathname").should("eq", "/admin/events/");

        // Ensure the new event is listed
        cy.wait<Cypress.RequestBody, Event>("@createEvent").then(
          ({ response }) => {
            cy.getByData(`${response.body.id}-name`).should(
              "contain",
              "Winter 2025 Event #2"
            );
            cy.getByData(`${response.body.id}-format`).should(
              "contain",
              "No Limit Hold'em"
            );
            cy.getByData(`${response.body.id}-date`).should(
              "contain",
              "Monday, January 6, 2025 at 7:00 PM"
            );
            cy.getByData(`${response.body.id}-additional-details`).should(
              "contain",
              "E2E Test Run"
            );
          }
        );
      });
    });
  });

  context("Edit Event", () => {
    it("should edit an event's details", () => {
      cy.visit("/admin/events");

      cy.getByData(`event-${EVENT.id}-card`).within(() => {
        cy.getByData("actions")
          .invoke("show")
          .within(() => {
            cy.getByData("edit-event-btn").click();
          });
      });

      cy.location("pathname").should("include", "edit");

      // Change event details
      cy.getByData("input-name").clear().type("Updated Event");
      cy.getByData("input-date").type("2025-01-14T19:00");
      cy.getByData("select-format").select("Pot Limit Omaha");
      cy.getByData("input-points-multiplier").clear().type("2");
      cy.getByData("input-additional-details")
        .clear()
        .type("Updated this event");

      cy.getByData("submit-btn").click();

      cy.location("pathname").should("eq", "/admin/events/");

      cy.getByData(`${EVENT.id}-name`).should("contain", "Updated Event");
      cy.getByData(`${EVENT.id}-format`).should("contain", "Pot Limit Omaha");
      cy.getByData(`${EVENT.id}-date`).should(
        "contain",
        "Tuesday, January 14, 2025 at 7:00 PM"
      );
      cy.getByData(`${EVENT.id}-additional-details`).should(
        "contain",
        "Updated this event"
      );
    });
  });

  context("Event Management", () => {
    beforeEach(() => {
      cy.visit(`/admin/events/${EVENT.id}`);
    });

    it("should register members for the event", () => {
      // Navigate to the event registration page
      cy.getByData("register-members-btn").click();

      // Check that page was redirected
      cy.location("pathname").should("contain", "/register");

      // Sign in each member
      cy.getByData(`member-${MEMBERS[3].id}`).within(() => {
        cy.getByData("checkbox-selected").check();
      });
      cy.getByData(`member-${MEMBERS[4].id}`).within(() => {
        cy.getByData("checkbox-selected").check();
      });

      cy.getByData("sign-in-btn").click();

      cy.getByData(`entry-${MEMBERS[3].id}`).should("be.visible");
      cy.getByData(`entry-${MEMBERS[4].id}`).should("be.visible");
    });

    context("Event Started", () => {
      beforeEach(() => {
        cy.clock();
      });

      afterEach(() => {
        cy.clock().then((clock) => clock.restore());
      });

      it("should search for and sign out a user", () => {
        cy.getByData("input-search").type(USERS[1].firstName);

        cy.getByData(`entry-${MEMBERS[0].id}`).should("not.exist");

        cy.getByData(`entry-${MEMBERS[2].id}`).should("not.exist");

        cy.getByData(`entry-${MEMBERS[1].id}`).within(() => {
          cy.getByData("sign-out-btn").click();

          cy.getByData("signed-out-at").should("not.contain", "Not Signed Out");
          cy.getByData("sign-in-btn").should("be.visible");
          cy.getByData("sign-out-btn").should("not.exist");
        });
      });

      it("should start the clock and advance a level", () => {
        // Open the tournament clock tab
        cy.getByData("clock-tab").click();

        // Check that we are on the tournament clock page
        cy.getByData("level").should("contain", "Level 1");
        cy.getByData("timer").should("contain", "5:00");

        // Check that the timer works and the level changes at the end of the timer
        cy.getByData("toggle-timer-btn").click();
        cy.tick(1000 * 60 * 5); // Tick 5 minutes;
        cy.getByData("level").should("contain", "Level 2");
        cy.getByData("timer").should("contain", "5:00");

        // Check that the timer can be paused
        cy.getByData("toggle-timer-btn").click();
        cy.tick(1000 * 60); // Advance 1 minute in time
        cy.getByData("timer").should("contain", "5:00");

        // Check that the back button works
        cy.getByData("prev-level-btn").click();
        cy.getByData("level").should("contain", "Level 1");
        cy.getByData("timer").should("contain", "5:00");

        // Check that the next button works
        cy.getByData("advance-level-btn").click();
        cy.getByData("level").should("contain", "Level 2");
        cy.getByData("timer").should("contain", "5:00");

        // Check that you can subtract a minute from the timer
        cy.getByData("sub-btn").click();
        cy.getByData("timer").should("contain", "4:00");

        // Check that you can add a minute to the timer
        cy.getByData("add-btn").click();
        cy.getByData("timer").should("contain", "5:00");
      });

      it("should sign out all entries and end the event", () => {
        cy.clock().then((clock) => clock.restore());

        cy.getByData(`entry-${MEMBERS[0].id}`).within(() => {
          cy.getByData("sign-out-btn").click();
        });
        cy.getByData(`entry-${MEMBERS[1].id}`).within(() => {
          cy.getByData("sign-out-btn").click();
        });
        cy.getByData(`entry-${MEMBERS[2].id}`).within(() => {
          cy.getByData("sign-out-btn").click();
        });

        cy.getByData("end-event-btn").click();

        // Ensure the modal was opened
        cy.getByData("modal").should("be.visible");
        cy.getByData("modal-title").should(
          "contain",
          "Are you sure you want to end the event?"
        );
        cy.getByData("modal-submit-btn").click();

        cy.getByData("event-ended-banner").should("be.visible");
      });
    });
  });
});
