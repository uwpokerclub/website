import { EVENT } from "../seed";
import {
  TRIAL_EXHAUSTED_REGISTERED_MEMBER,
  TRIAL_EXHAUSTED_UNREGISTERED_MEMBER,
  TRIAL_AVAILABLE_UNPAID_MEMBER,
  TRIAL_EXHAUSTED_PAID_MEMBER,
} from "../support/helpers";

const SHADE = "rgb(254, 226, 226)";

describe("EventRegistrationModal", () => {
  before(() => {
    cy.resetDatabase();
  });

  beforeEach(() => {
    cy.login();
    cy.intercept("GET", "/api/v2/semesters", { fixture: "semesters.json" }).as("getSemesters");
    cy.intercept("GET", /\/api\/v2\/semesters\/.*\/events\/1$/, { fixture: "event-details.json" }).as("getEvent");
    cy.intercept("GET", /\/api\/v2\/semesters\/.*\/events\/1\/entries/, { fixture: "event-entries.json" }).as(
      "getEntries",
    );
    cy.intercept("GET", /\/api\/v2\/semesters\/.*\/memberships/, { fixture: "memberships.json" }).as("getMemberships");

    cy.visit(`/admin/events/${EVENT.id}`);
    cy.getByData("register-members-btn").click();
    cy.getByData("panel-available").should("be.visible");
  });

  context("free trial status", () => {
    it("shades an unregistered member whose free trial is used up", () => {
      cy.getByData("panel-available").within(() => {
        cy.getByData(`member-row-${TRIAL_EXHAUSTED_UNREGISTERED_MEMBER.id}`)
          .should("have.css", "background-color", SHADE)
          .and("have.attr", "title", "Free trial used up");
      });
    });

    it("shades a registered member whose free trial is used up", () => {
      cy.getByData("panel-registered").within(() => {
        cy.getByData(`member-row-${TRIAL_EXHAUSTED_REGISTERED_MEMBER.id}`)
          .should("have.css", "background-color", SHADE)
          .and("have.attr", "title", "Free trial used up");
      });
    });

    it("does not shade an unpaid member whose free trial is still available", () => {
      cy.getByData("panel-registered").within(() => {
        cy.getByData(`member-row-${TRIAL_AVAILABLE_UNPAID_MEMBER.id}`)
          .should("not.have.css", "background-color", SHADE)
          .and("not.have.attr", "title");
      });
    });

    it("does not shade a paid member carrying a stale exhausted flag", () => {
      cy.getByData("panel-available").within(() => {
        cy.getByData(`member-row-${TRIAL_EXHAUSTED_PAID_MEMBER.id}`)
          .should("not.have.css", "background-color", SHADE)
          .and("not.have.attr", "title");
      });
    });

    it("exposes the trial status as text for screen readers", () => {
      cy.getByData("panel-available").within(() => {
        cy.getByData(`member-row-${TRIAL_EXHAUSTED_UNREGISTERED_MEMBER.id}`).should(
          "contain",
          "Free trial used up",
        );
      });
    });
  });
});
