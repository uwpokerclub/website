import { EVENT, ENDED_EVENT, SEMESTER } from "../seed";

describe("ListEvents", () => {
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
      cy.visit("/admin/events");
      cy.getByData("events-no-semester").should("be.visible");
    });
  });

  context("error state", () => {
    beforeEach(() => {
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
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
      cy.login();
      cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
      cy.intercept("GET", /\/api\/v2\/semesters\/.*\/events/, { fixture: "events.json" }).as("getEvents");
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

    context("navigation", () => {
      it("should navigate to event details when clicking event name", () => {
        cy.getByData(`event-name-${EVENT.id}`).click();

        cy.location("pathname").should("include", `/admin/events/${EVENT.id}`);
      });
    });

    context("event actions menu", () => {
      it("should show edit and end options for active events", () => {
        cy.getByData(`actions-menu-btn-${EVENT.id}`).click();

        cy.getByData(`edit-event-btn-${EVENT.id}`).should("be.visible");
        // Use exist instead of visible due to CSS overflow clipping
        cy.getByData(`end-event-btn-${EVENT.id}`).should("exist");
      });

      it("should show restart option for ended events", () => {
        cy.getByData(`actions-menu-btn-${ENDED_EVENT.id}`).click();

        // Use exist instead of visible due to CSS overflow clipping
        cy.getByData(`restart-event-btn-${ENDED_EVENT.id}`).should("exist");
      });
    });

    context("end event confirmation", () => {
      it("should open and cancel end event confirmation modal", () => {
        cy.getByData(`actions-menu-btn-${EVENT.id}`).click();
        cy.getByData(`end-event-btn-${EVENT.id}`).click();

        // Modal should be open (using exist due to CSS)
        cy.getByData(`end-confirm-modal-${EVENT.id}`).should("exist");

        cy.getByData(`end-confirm-cancel-btn-${EVENT.id}`).click();

        cy.getByData(`end-confirm-modal-${EVENT.id}`).should("not.exist");
      });
    });

    context("create event modal", () => {
      it("should open and cancel create event modal", () => {
        cy.getByData("create-event-btn").click();

        // Modal content should exist (using exist instead of visible due to CSS)
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
          body: { data: mockEvents, total: mockEvents.length },
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
          body: { data: [], total: 0 },
        }).as("getEmptyEvents");

        cy.visit("/admin/events");
        cy.wait("@getEmptyEvents");

        cy.getByData("events-empty").should("be.visible");
      });
    });
  });

  context("contract tests", () => {
    before(() => {
      cy.resetDatabase();
    });

    beforeEach(() => {
      cy.login();
    });

    it("should filter events by search and clear search", () => {
      cy.visit("/admin/events");
      cy.getByData("events-table").should("exist");

      cy.getByData("input-events-search").type(EVENT.name);
      cy.getByData("events-results-info").should("contain", EVENT.name);
      cy.getByData(`event-name-${EVENT.id}`).should("exist");
      cy.getByData(`event-name-${ENDED_EVENT.id}`).should("not.exist");

      // Clear search
      cy.getByData("clear-search-btn").click();
      cy.getByData("input-events-search").should("have.value", "");
      cy.getByData(`event-name-${EVENT.id}`).should("exist");
      cy.getByData(`event-name-${ENDED_EVENT.id}`).should("exist");
    });

    it("should load events list from real API", () => {
      cy.visit("/admin/events");
      cy.getByData("events-table").should("exist");
      cy.getByData(`event-name-${EVENT.id}`).should("contain", EVENT.name);
      cy.getByData(`event-name-${ENDED_EVENT.id}`).should("contain", ENDED_EVENT.name);
    });

    it("should end an event via real API", () => {
      cy.intercept("POST", /\/api\/v2\/semesters\/.*\/events\/.*\/end/).as("endEvent");

      cy.visit("/admin/events");
      cy.getByData("events-table").should("exist");

      cy.getByData(`actions-menu-btn-${EVENT.id}`).click();
      cy.getByData(`end-event-btn-${EVENT.id}`).click();
      cy.getByData(`end-confirm-btn-${EVENT.id}`).click();

      cy.wait("@endEvent").then((interception) => {
        expect(interception.response?.statusCode).to.eq(204);
      });

      cy.getByData(`event-status-${EVENT.id}`).should("contain", "Ended");
    });
  });
});
