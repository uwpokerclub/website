import { SEMESTER } from "../seed";
import { Transaction } from "../types";

describe("Semesters page", () => {
  beforeEach(() => {
    cy.exec("npm run db:reset && npm run db:seed");
    cy.login("e2e_user", "password");
    cy.visit("/admin/semesters");
  });

  it("should create a new semester", () => {
    cy.getByData("create-semester-btn").click();
    cy.location("pathname").should("eq", "/admin/semesters/new");

    cy.getByData("name").type("Winter 2025");
    cy.getByData("start-date").type("2025-01-01");
    cy.getByData("end-date").type("2025-04-30");
    cy.getByData("starting-budget").type("1000");
    cy.getByData("membership-fee").type("10");
    cy.getByData("discounted-membership-fee").type("7");
    cy.getByData("rebuy-fee").type("2");
    cy.getByData("additional-details").type("Integration test run");
    cy.getByData("form-submit").click();

    // Check that the page is redirected back to the semesters list
    cy.location("pathname").should("eq", "/admin/semesters/");

    // Check that the new semester is visible in the page
    cy.getByData("semester-name").should("exist").contains("Winter 2025");
  });

  context("view a semester", () => {
    it("should show the semester info page", () => {
      cy.getByData(`view-semester-${SEMESTER.id}`).click();
      cy.location("pathname").should("eq", `/admin/semesters/${SEMESTER.id}`);
    });
  });

  context("transactions", () => {
    beforeEach(() => {
      cy.visit(`/admin/semesters/${SEMESTER.id}`);
    });

    it("should create a new transaction", () => {
      cy.intercept(
        "POST",
        `/api/semesters/${SEMESTER.id}/transactions`
      ).as("createTransaction");

      /**
       * Create a new transaction that removes money
       */

      // Open the new transaction modal
      cy.getByData("new-transaction-btn").click();
      cy.getByData("modal").should("be.visible");
      cy.getByData("modal-title").should("contain", "New transaction");

      // Fill out form fields and submit
      cy.getByData("input-description").type("New cards");
      cy.getByData("input-amount").type("-35.00");
      cy.getByData("modal-submit-btn").click();

      // The modal should be closed at this point
      cy.getByData("modal").should("not.be.visible");

      // Reload the page to get new data
      cy.reload();

      // Check that the budget was updated
      cy.getByData("current-budget-card").should("contain", "$65.00");
      // Check that the transaction appears in the list
      cy.wait<Cypress.RequestBody, Transaction>("@createTransaction").then(
        ({ response }) => {
          cy.getByData(`transaction-${response.body.id}`).within(() => {
            cy.getByData(`${response.body.id}-amount`).should(
              "contain",
              "-$35.00"
            );
          });
        }
      );

      /**
       * Create a new transaction that adds money
       */
      // Open the new transaction modal
      cy.getByData("new-transaction-btn").click();

      // Fill out form fields and submit
      cy.getByData("input-description").type("New cards");
      cy.getByData("input-amount").type("10.00");
      cy.getByData("modal-submit-btn").click();

      // The modal should be closed at this point
      cy.getByData("modal").should("not.be.visible");

      // Reload the page to get new data
      cy.reload();

      // Check that the budget was updated
      cy.getByData("current-budget-card").should("contain", "$75.00");
      // Check that the transaction appears in the list
      cy.wait<Cypress.RequestBody, Transaction>("@createTransaction").then(
        ({ response }) => {
          cy.getByData(`transaction-${response.body.id}`).within(() => {
            cy.getByData(`${response.body.id}-amount`).should(
              "contain",
              "$10.00"
            );
          });
        }
      );
    });

    it("should delete an existing transaction", () => {
        cy.request(
          "POST",
          `/api/semesters/${SEMESTER.id}/transactions`,
          {
            description: "New cards",
            amount: -35.0,
          }
        ).as("transaction");

      // Find row that has the transaction and click the delete button for it
      cy.get<Cypress.Response<Transaction>>("@transaction").then((response) => {
        cy.getByData(`transaction-${response.body.id}`).within(() => {
          cy.getByData(`${response.body.id}-delete-btn`).click();

          cy.getByData(`transaction-${response.body.id}`).should("not.exist");
        });
      });
    });
  });
});
