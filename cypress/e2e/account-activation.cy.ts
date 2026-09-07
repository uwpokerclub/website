describe("Account activation", () => {
  const token = "activation-token";
  let verificationResponses: Array<"success" | "invalid" | "unavailable">;
  let sessionIdentity: string | null;
  let completeSucceeds: boolean;

  beforeEach(() => {
    verificationResponses = ["success", "success"];
    sessionIdentity = null;
    completeSucceeds = false;

    cy.intercept("GET", "/api/v2/session", (request) => {
      if (!sessionIdentity) {
        request.reply({ statusCode: 401, body: { message: "Unauthenticated" } });
        return;
      }
      request.reply({ statusCode: 200, body: { username: sessionIdentity, role: "president", permissions: {} } });
    });
    cy.intercept("GET", "/api/v2/semesters", { body: { data: [] } });
    cy.intercept("POST", "/api/v2/activations/verify", (request) => {
      expect(request.body).to.deep.equal({ token });
      expect(request.url).not.to.contain(token);
      const response = verificationResponses.shift() ?? "success";
      if (response === "unavailable") {
        request.reply({ statusCode: 503, body: { message: "Service unavailable" } });
        return;
      }
      if (response === "invalid") {
        request.reply({ statusCode: 401, body: { message: "Invalid or expired activation token" } });
        return;
      }
      request.reply({ firstName: "Ada", lastName: "Lovelace" });
    }).as("verifyActivation");
    cy.intercept("POST", "/api/v2/activations/complete", (request) => {
      if (completeSucceeds) {
        sessionIdentity = "ada";
        request.reply({ statusCode: 201 });
        return;
      }
      request.reply({ statusCode: 500, body: { message: "Service unavailable" } });
    }).as("completeActivation");
  });

  it("reads the token from the fragment, clears it, and activates the account", () => {
    completeSucceeds = true;
    cy.visit(`/activate#token=${token}`);

    cy.wait("@verifyActivation");
    cy.location("hash").should("eq", "");
    cy.getByData("activation-heading").should("contain", "Set your password");
    cy.contains("You're activating an account for Ada Lovelace.").should("be.visible");
    cy.getByData("activation-password").type("correct horse");
    cy.getByData("activation-confirm-password").type("correct horse");
    cy.getByData("activation-submit").click();

    cy.wait("@completeActivation");
    cy.location("pathname").should("eq", "/admin/dashboard");
  });

  it("shows actionable invalid-token copy", () => {
    verificationResponses = ["invalid", "invalid"];
    cy.visit(`/activate#token=${token}`);

    cy.wait("@verifyActivation");
    cy.getByData("activation-error").should("contain", "This activation link is invalid");
    cy.getByData("activation-error").should("contain", "Ask whoever sent you this link for a new one");
  });

  it("lets the user retry verification after a transient failure", () => {
    verificationResponses = ["unavailable", "success"];
    cy.visit(`/activate#token=${token}`);

    cy.wait("@verifyActivation");
    cy.getByData("activation-error").should("contain", "Check your connection and try again");
    cy.getByData("activation-retry").click();
    cy.wait("@verifyActivation");
    cy.getByData("activation-heading").should("contain", "Set your password");
  });

  it("replaces a cached session with the activated identity", () => {
    sessionIdentity = "outgoing";
    completeSucceeds = true;
    cy.visit(`/activate#token=${token}`);
    cy.wait("@verifyActivation");
    cy.getByData("activation-password").type("correct horse");
    cy.getByData("activation-confirm-password").type("correct horse");
    cy.getByData("activation-submit").click();

    cy.wait("@completeActivation");
    cy.getByData("user-name").should("have.text", "ada");
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
