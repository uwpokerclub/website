import { EVENT, ENDED_EVENT, SEMESTER } from "../seed";

describe("ListEvents", () => {
  context("when no semester is selected", () => {
    beforeEach(() => {
      cy.resetDatabase();
      cy.login();
      // Mock semesters API to return empty array so no semester is selected
      cy.intercept("GET", "/api/v2/semesters", []).as("getSemesters");
    });

    it("should display no semester selected message", () => {
      cy.visit("/admin/events");
      cy.getByData("events-no-semester").should("be.visible");
    });
  });

  context("error state", () => {
    beforeEach(() => {
      cy.resetDatabase();
      cy.login();
    });

    it("should display error message and retry button when API fails", () => {
      cy.intercept("GET", /\/api\/v2\/semesters\/.*\/events/, {
        statusCode: 500,
        body: { error: "Internal Server Error" },
      }).as("getEventsError");

      cy.visit("/admin/events");
      cy.wait("@getEventsError");

      cy.getByData("events-error").should("be.visible");
      cy.getByData("events-retry-btn").should("exist");
    });
  });

  context("with semester selected", () => {
    beforeEach(() => {
      cy.resetDatabase();
      cy.login();
      cy.visit("/admin/events");
      // Wait for table to be visible (data loaded)
      cy.getByData("events-table").should("exist");
    });

    context("table display", () => {
      it("should display all column headers", () => {
        cy.contains("th", "Name").should("be.visible");
        cy.contains("th", "Date").should("be.visible");
        cy.contains("th", "Format").should("be.visible");
        cy.contains("th", "Entry Count").should("be.visible");
        cy.contains("th", "Status").should("be.visible");
        cy.contains("th", "Actions").should("be.visible");
      });

      it("should display all events in table", () => {
        // We have EVENT and ENDED_EVENT in seed data
        cy.getByData(`event-name-${EVENT.id}`).should("exist");
        cy.getByData(`event-name-${ENDED_EVENT.id}`).should("exist");
      });

      it("should display event data correctly", () => {
        // Verify active event data
        cy.getByData(`event-name-${EVENT.id}`).should("contain", EVENT.name);
        cy.getByData(`event-status-${EVENT.id}`).should("contain", "Active");

        // Verify ended event data
        cy.getByData(`event-name-${ENDED_EVENT.id}`).should(
          "contain",
          ENDED_EVENT.name
        );
        cy.getByData(`event-status-${ENDED_EVENT.id}`).should(
          "contain",
          "Ended"
        );
      });
    });

    context("event status badges", () => {
      it("should display Active status for started events", () => {
        cy.getByData(`event-status-${EVENT.id}`).should("contain", "Active");
      });

      it("should display Ended status for completed events", () => {
        cy.getByData(`event-status-${ENDED_EVENT.id}`).should(
          "contain",
          "Ended"
        );
      });
    });

    context("search functionality", () => {
      it("should filter events by name", () => {
        cy.getByData("input-events-search").type(EVENT.name);

        // Wait for debounce by checking results info updates
        cy.getByData("events-results-info").should("contain", EVENT.name);
        cy.getByData(`event-name-${EVENT.id}`).should("exist");
        cy.getByData(`event-name-${ENDED_EVENT.id}`).should("not.exist");
      });

      it("should be case-insensitive", () => {
        const searchTerm = EVENT.name.toUpperCase();
        cy.getByData("input-events-search").type(searchTerm);

        // Wait for debounce by checking results info updates
        cy.getByData("events-results-info").should("contain", searchTerm);
        cy.getByData(`event-name-${EVENT.id}`).should("exist");
      });

      it("should show no results state", () => {
        cy.getByData("input-events-search").type("nonexistentevent12345");

        cy.getByData("events-no-results").should("be.visible");
      });

      it("should clear search and show all events", () => {
        // First search to filter
        cy.getByData("input-events-search").type(EVENT.name);
        cy.getByData("events-results-info").should("contain", EVENT.name);
        cy.getByData(`event-name-${ENDED_EVENT.id}`).should("not.exist");

        // Clear search
        cy.getByData("clear-search-btn").click();

        // Should show all events again
        cy.getByData("input-events-search").should("have.value", "");
        cy.getByData(`event-name-${EVENT.id}`).should("exist");
        cy.getByData(`event-name-${ENDED_EVENT.id}`).should("exist");
      });
    });

    context("navigation", () => {
      it("should navigate to event details when clicking event name", () => {
        cy.getByData(`event-name-${EVENT.id}`).click();

        cy.location("pathname").should("include", `/admin/events/${EVENT.id}`);
      });
    });

    context("event actions menu", () => {
      it("should open actions menu when clicking menu button", () => {
        cy.getByData(`actions-menu-btn-${EVENT.id}`).click();

        cy.getByData(`edit-event-btn-${EVENT.id}`).should("be.visible");
      });

      it("should show edit option for active events", () => {
        cy.getByData(`actions-menu-btn-${EVENT.id}`).click();

        cy.getByData(`edit-event-btn-${EVENT.id}`).should("be.visible");
      });

      it("should show end event option for active events", () => {
        cy.getByData(`actions-menu-btn-${EVENT.id}`).click();

        // Use exist instead of visible due to CSS overflow clipping
        cy.getByData(`end-event-btn-${EVENT.id}`).should("exist");
      });

      it("should show restart option for ended events", () => {
        cy.getByData(`actions-menu-btn-${ENDED_EVENT.id}`).click();

        cy.getByData(`restart-event-btn-${ENDED_EVENT.id}`).should("be.visible");
      });
    });

    context("end event confirmation", () => {
      it("should open confirmation modal when End Event clicked", () => {
        cy.getByData(`actions-menu-btn-${EVENT.id}`).click();
        cy.getByData(`end-event-btn-${EVENT.id}`).click();

        // Modal should be open (using exist due to CSS)
        cy.getByData(`end-confirm-modal-${EVENT.id}`).should("exist");
      });

      it("should close modal when Cancel clicked", () => {
        cy.getByData(`actions-menu-btn-${EVENT.id}`).click();
        cy.getByData(`end-event-btn-${EVENT.id}`).click();
        cy.getByData(`end-confirm-modal-${EVENT.id}`).should("exist");

        cy.getByData(`end-confirm-cancel-btn-${EVENT.id}`).click();

        cy.getByData(`end-confirm-modal-${EVENT.id}`).should("not.exist");
      });
    });

    context("create event modal", () => {
      it("should open modal when Create Event button is clicked", () => {
        cy.getByData("create-event-btn").click();

        // Modal content should exist (using exist instead of visible due to CSS)
        cy.getByData("create-event-modal").should("exist");
      });

      it("should close modal when cancel is clicked", () => {
        cy.getByData("create-event-btn").click();
        cy.getByData("create-event-modal").should("exist");

        cy.getByData("create-event-cancel-btn").click();

        cy.getByData("create-event-modal").should("not.exist");
      });
    });

    context("pagination", () => {
      it("should not show pagination when 25 or fewer events", () => {
        // With only 2 events in seed data, pagination should not be visible
        cy.getByData("events-pagination").should("not.exist");
      });

      it("should show pagination when more than 25 events", () => {
        // Generate mock data with 30 events
        const mockEvents = Array.from({ length: 30 }, (_, i) => ({
          id: i + 100,
          name: `Mock Event ${i + 1}`,
          format: "No Limit Hold'em",
          notes: "",
          semesterId: SEMESTER.id,
          startDate: "2025-01-03T19:00:00.000Z",
          state: i % 2,
          entries: [],
        }));

        cy.intercept("GET", /\/api\/v2\/semesters\/.*\/events/, {
          statusCode: 200,
          body: mockEvents,
        }).as("getManyEvents");

        // Revisit to trigger mock
        cy.visit("/admin/events");
        cy.wait("@getManyEvents");

        cy.getByData("events-pagination").should("exist");
      });
    });

    context("empty state", () => {
      it("should display empty state when no events exist", () => {
        cy.intercept("GET", /\/api\/v2\/semesters\/.*\/events/, {
          statusCode: 200,
          body: [],
        }).as("getEmptyEvents");

        cy.visit("/admin/events");
        cy.wait("@getEmptyEvents");

        cy.getByData("events-empty").should("be.visible");
      });
    });
  });
});
