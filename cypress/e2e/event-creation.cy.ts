import { SEMESTER, STRUCTURE } from "../seed";

describe("CreateEventModal", () => {
  beforeEach(() => {
    cy.exec("npm run db:reset && npm run db:seed");
    cy.login("e2e_user", "password");
    cy.visit("/admin/events");
    cy.intercept("GET", "/api/v2/structures").as("getStructures");
    cy.getByData("create-event-btn").click();
    cy.getByData("create-event-modal").should("exist");
    cy.wait("@getStructures");
  });

  context("modal interactions", () => {
    it("should open and close modal via cancel button", () => {
      cy.getByData("create-event-modal").should("exist");
      cy.getByData("create-event-cancel-btn").click();
      cy.getByData("create-event-modal").should("not.exist");
    });

    it("should reset form on close and reopen", () => {
      // Fill in some values
      cy.getByData("input-name").type("Test Event");
      cy.getByData("select-format").select("No Limit Hold'em");

      // Close modal
      cy.getByData("create-event-cancel-btn").click();
      cy.getByData("create-event-modal").should("not.exist");

      // Reopen modal
      cy.intercept("GET", "/api/v2/structures").as("getStructures2");
      cy.getByData("create-event-btn").click();
      cy.getByData("create-event-modal").should("exist");
      cy.wait("@getStructures2");

      // Verify form is reset
      cy.getByData("input-name").should("have.value", "");
      // Select element has null value when no option is selected
      cy.getByData("select-format").invoke("val").should("satisfy", (val: string | null) => val === "" || val === null);
    });
  });

  context("event details form", () => {
    it("should fill all event details fields", () => {
      const futureDate = "2025-02-15T19:00";

      cy.getByData("input-name").type("Winter Tournament #5");
      cy.getByData("input-startDate").clear().type(futureDate);
      cy.getByData("select-format").select("No Limit Hold'em");
      cy.getByData("input-pointsMultiplier").clear().type("1.5");
      cy.getByData("input-notes").type("Test event notes");

      // Verify values
      cy.getByData("input-name").should("have.value", "Winter Tournament #5");
      cy.getByData("input-startDate").should("have.value", futureDate);
      cy.getByData("select-format").should("have.value", "No Limit Hold'em");
      cy.getByData("input-pointsMultiplier").should("have.value", "1.5");
      cy.getByData("input-notes").should("have.value", "Test event notes");
    });

    it("should show validation error for empty event name", () => {
      // Fill required fields except name
      cy.getByData("input-startDate").clear().type("2025-02-15T19:00");
      cy.getByData("select-format").select("No Limit Hold'em");
      cy.getByData("select-structureId").select(String(STRUCTURE.id));

      // Submit
      cy.getByData("create-event-submit-btn").scrollIntoView().click();

      // Verify validation error
      cy.contains("Event name is required").should("exist");
      cy.getByData("create-event-modal").should("exist");
    });

    it("should show validation error for missing start date", () => {
      // Fill required fields except start date
      cy.getByData("input-name").type("Test Event");
      cy.getByData("select-format").select("No Limit Hold'em");
      cy.getByData("select-structureId").select(String(STRUCTURE.id));

      // Submit
      cy.getByData("create-event-submit-btn").scrollIntoView().click();

      // Verify validation error
      cy.contains("Start date is required").should("exist");
      cy.getByData("create-event-modal").should("exist");
    });

    it("should show validation error for missing format", () => {
      // Fill required fields except format
      cy.getByData("input-name").type("Test Event");
      cy.getByData("input-startDate").clear().type("2025-02-15T19:00");
      cy.getByData("select-structureId").select(String(STRUCTURE.id));

      // Submit
      cy.getByData("create-event-submit-btn").scrollIntoView().click();

      // Verify validation error
      cy.contains("Please select a format").should("exist");
      cy.getByData("create-event-modal").should("exist");
    });
  });

  context("selecting existing structure", () => {
    it("should display available structures in dropdown", () => {
      cy.getByData("radio-structure-mode-select").should("be.checked");
      cy.getByData("select-structureId").should("exist");
      cy.getByData("select-structureId").find("option").should("have.length.gt", 1);
    });

    it("should select an existing structure", () => {
      cy.getByData("select-structureId").select(String(STRUCTURE.id));
      cy.getByData("select-structureId").should("have.value", String(STRUCTURE.id));
    });

    it("should show validation error when no structure selected", () => {
      // Fill all other required fields
      cy.getByData("input-name").type("Test Event");
      cy.getByData("input-startDate").clear().type("2025-02-15T19:00");
      cy.getByData("select-format").select("No Limit Hold'em");
      // Leave structure unselected (default value is 0)

      // Submit
      cy.getByData("create-event-submit-btn").scrollIntoView().click();

      // Verify validation error
      cy.contains("Please select a structure").should("exist");
      cy.getByData("create-event-modal").should("exist");
    });
  });

  context("creating new structure", () => {
    it("should switch to create structure mode", () => {
      cy.getByData("radio-structure-mode-create").click();
      cy.getByData("radio-structure-mode-create").should("be.checked");
      cy.getByData("select-structureId").should("not.exist");
      cy.getByData("input-structure-name").should("exist");
    });

    it("should show default blind level (25/50/0/15)", () => {
      cy.getByData("radio-structure-mode-create").click();

      // Verify default blind level values
      cy.getByData("blind-0-small").should("have.value", "25");
      cy.getByData("blind-0-big").should("have.value", "50");
      cy.getByData("blind-0-ante").should("have.value", "0");
      cy.getByData("blind-0-time").should("have.value", "15");
    });

    it("should fill structure name and validate required", () => {
      cy.getByData("radio-structure-mode-create").click();

      // Fill all event fields but leave structure name empty
      cy.getByData("input-name").type("Test Event");
      cy.getByData("input-startDate").clear().type("2025-02-15T19:00");
      cy.getByData("select-format").select("No Limit Hold'em");

      // Submit without structure name
      cy.getByData("create-event-submit-btn").scrollIntoView().click();

      // Verify validation error
      cy.contains("Structure name is required").should("exist");
      cy.getByData("create-event-modal").should("exist");

      // Fill structure name
      cy.getByData("input-structure-name").type("My Custom Structure");
      cy.getByData("input-structure-name").should("have.value", "My Custom Structure");
    });
  });

  context("blind levels management", () => {
    beforeEach(() => {
      cy.getByData("radio-structure-mode-create").click();
    });

    it("should add level with smart defaults (doubling)", () => {
      // Add second level (should double: 25->50, 50->100)
      cy.getByData("add-blind-btn").click();

      cy.getByData("blind-1-small").should("have.value", "50");
      cy.getByData("blind-1-big").should("have.value", "100");
      cy.getByData("blind-1-ante").should("have.value", "0");
      cy.getByData("blind-1-time").should("have.value", "15");
    });

    it("should add level with progression calculation", () => {
      // Add second level (doubling)
      cy.getByData("add-blind-btn").click();

      // Add third level (progression: diff is 25/50, so 50+25=75, 100+50=150)
      cy.getByData("add-blind-btn").click();

      cy.getByData("blind-2-small").should("have.value", "75");
      cy.getByData("blind-2-big").should("have.value", "150");
      cy.getByData("blind-2-ante").should("have.value", "0");
      cy.getByData("blind-2-time").should("have.value", "15");
    });

    it("should remove blind level when multiple exist", () => {
      // Add a second level
      cy.getByData("add-blind-btn").click();
      cy.getByData("blind-1-small").should("exist");

      // Remove the first level
      cy.getByData("remove-blind-btn-0").click();

      // The second level should now be the first
      cy.getByData("blind-0-small").should("have.value", "50");
      cy.getByData("blind-1-small").should("not.exist");
    });

    it("should not show remove button when only one level", () => {
      // Only one level by default
      cy.getByData("remove-blind-btn-0").should("not.exist");
    });

    it("should validate time must be >= 1", () => {
      cy.getByData("input-name").type("Test Event");
      cy.getByData("input-startDate").clear().type("2025-02-15T19:00");
      cy.getByData("select-format").select("No Limit Hold'em");
      cy.getByData("input-structure-name").type("Test Structure");

      // Set time to 0 (invalid - must be at least 1)
      cy.getByData("blind-0-time").clear().type("0");

      // Submit
      cy.getByData("create-event-submit-btn").scrollIntoView().click();

      // Modal should stay open due to validation failure (form doesn't submit)
      cy.getByData("create-event-modal").should("exist");

      // Verify button is not in loading state (would be "Creating..." if submitted)
      cy.getByData("create-event-submit-btn").should("contain", "Create Event");
    });
  });

  context("successful submission - existing structure", () => {
    it("should create event with existing structure and verify API payload", () => {
      cy.intercept("POST", /\/api\/v2\/semesters\/.*\/events/).as("createEvent");

      // Fill all required fields
      cy.getByData("input-name").type("Test Tournament #99");
      cy.getByData("input-startDate").clear().type("2025-02-15T19:00");
      cy.getByData("select-format").select("No Limit Hold'em");
      cy.getByData("input-pointsMultiplier").clear().type("1.5");
      cy.getByData("input-notes").type("E2E test event");
      cy.getByData("select-structureId").select(String(STRUCTURE.id));

      // Submit
      cy.getByData("create-event-submit-btn").scrollIntoView().click();

      // Verify API call
      cy.wait("@createEvent").then((interception) => {
        expect(interception.request.body).to.deep.include({
          name: "Test Tournament #99",
          format: "No Limit Hold'em",
          notes: "E2E test event",
          structureId: STRUCTURE.id,
          pointsMultiplier: 1.5,
        });
        expect(interception.response?.statusCode).to.eq(201);
      });

      // Verify modal closes
      cy.getByData("create-event-modal").should("not.exist");

      // Verify navigation to new event page
      cy.url().should("match", /\/admin\/events\/\d+$/);
    });
  });

  context("successful submission - new structure", () => {
    it("should create structure then event and navigate to new event", () => {
      cy.intercept("POST", "/api/v2/structures").as("createStructure");
      cy.intercept("POST", /\/api\/v2\/semesters\/.*\/events/).as("createEvent");

      // Fill event details
      cy.getByData("input-name").type("Custom Structure Event");
      cy.getByData("input-startDate").clear().type("2025-02-20T19:00");
      cy.getByData("select-format").select("Pot Limit Omaha");
      cy.getByData("input-pointsMultiplier").clear().type("2");

      // Switch to create structure mode and fill details
      cy.getByData("radio-structure-mode-create").click();
      cy.getByData("input-structure-name").type("E2E Test Structure");

      // Add an extra blind level
      cy.getByData("add-blind-btn").click();

      // Submit
      cy.getByData("create-event-submit-btn").scrollIntoView().click();

      // Verify structure creation API call
      cy.wait("@createStructure").then((interception) => {
        expect(interception.request.body).to.deep.include({
          name: "E2E Test Structure",
        });
        expect(interception.request.body.blinds).to.have.length(2);
        expect(interception.response?.statusCode).to.eq(201);
      });

      // Verify event creation API call
      cy.wait("@createEvent").then((interception) => {
        expect(interception.request.body).to.deep.include({
          name: "Custom Structure Event",
          format: "Pot Limit Omaha",
          pointsMultiplier: 2,
        });
        expect(interception.response?.statusCode).to.eq(201);
      });

      // Verify modal closes
      cy.getByData("create-event-modal").should("not.exist");

      // Verify navigation to new event page
      cy.url().should("match", /\/admin\/events\/\d+$/);
    });
  });

  context("error handling", () => {
    it("should display error when event creation API fails", () => {
      cy.intercept("POST", /\/api\/v2\/semesters\/.*\/events/, {
        statusCode: 500,
        body: { message: "Internal server error" },
      }).as("createEventError");

      // Fill all required fields
      cy.getByData("input-name").type("Test Event");
      cy.getByData("input-startDate").clear().type("2025-02-15T19:00");
      cy.getByData("select-format").select("No Limit Hold'em");
      cy.getByData("select-structureId").select(String(STRUCTURE.id));

      // Submit
      cy.getByData("create-event-submit-btn").scrollIntoView().click();

      // Wait for the failed request
      cy.wait("@createEventError");

      // Verify error is displayed and modal stays open
      cy.getByData("create-event-error-alert").should("exist");
      cy.getByData("create-event-modal").should("exist");
    });

    it("should display error when structure creation API fails", () => {
      cy.intercept("POST", "/api/v2/structures", {
        statusCode: 500,
        body: { message: "Structure creation failed" },
      }).as("createStructureError");

      // Fill all required fields with new structure
      cy.getByData("input-name").type("Test Event");
      cy.getByData("input-startDate").clear().type("2025-02-15T19:00");
      cy.getByData("select-format").select("No Limit Hold'em");

      // Switch to create structure mode
      cy.getByData("radio-structure-mode-create").click();
      cy.getByData("input-structure-name").type("Failed Structure");

      // Submit
      cy.getByData("create-event-submit-btn").scrollIntoView().click();

      // Wait for the failed request
      cy.wait("@createStructureError");

      // Verify error is displayed and modal stays open
      cy.getByData("create-event-error-alert").should("exist");
      cy.getByData("create-event-modal").should("exist");
    });
  });
});
