describe("Account activation", () => {
  const token = "activation-token";
  let activated = false;

  beforeEach(() => {
    activated = false;
    cy.intercept("GET", "/api/v2/session", (request) => {
      if (!activated) {
        request.reply({ statusCode: 401, body: { message: "Unauthenticated" } });
        return;
      }
      request.reply({ statusCode: 200, body: { username: "ada", role: "president", permissions: {} } });
    });
    cy.intercept("POST", "/api/v2/activations/verify", (request) => {
      expect(request.body).to.deep.equal({ token });
      expect(request.url).not.to.contain(token);
      request.reply({ firstName: "Ada", lastName: "Lovelace" });
    }).as("verifyActivation");
    cy.intercept("POST", "/api/v2/activations/complete", { statusCode: 500 }).as("completeActivation");
  });

  it("reads the token from the fragment, clears it, and activates the account", () => {
    cy.intercept("POST", "/api/v2/activations/complete", (request) => {
      expect(request.body).to.deep.equal({ token, password: "correct horse" });
      expect(request.url).not.to.contain(token);
      activated = true;
      request.reply({ statusCode: 201 });
    }).as("completeActivation");

    cy.visit(`/activate#token=${token}`);

    cy.wait("@verifyActivation");
    cy.location("hash").should("eq", "");
    cy.getByData("activation-heading").should("contain", "Set a password for Ada Lovelace");
    cy.getByData("activation-password").type("correct horse");
    cy.getByData("activation-confirm-password").type("correct horse");
    cy.getByData("activation-submit").click();

    cy.wait("@completeActivation");
    cy.location("pathname").should("eq", "/admin/dashboard");
  });

  it("shows actionable invalid-token copy", () => {
    cy.intercept("POST", "/api/v2/activations/verify", {
      statusCode: 401,
      body: { message: "Invalid or expired activation token" },
    }).as("verifyInvalidActivation");

    cy.visit(`/activate#token=${token}`);

    cy.wait("@verifyInvalidActivation");
    cy.getByData("activation-error").should("contain", "This activation link is invalid");
    cy.getByData("activation-error").should("contain", "Ask whoever sent you this link for a new one");
  });

  it("rejects passwords shorter than eight characters before completing", () => {
    cy.visit(`/activate#token=${token}`);
    cy.wait("@verifyActivation");

    cy.getByData("activation-password").type("short");
    cy.getByData("activation-confirm-password").type("short");
    cy.getByData("activation-submit").click();

    cy.getByData("activation-password-error").should("contain", "at least 8 characters");
    cy.get("@completeActivation.all").should("have.length", 0);
  });
});
