import { SEMESTER } from "../seed";

const firstSemester = {
  id: SEMESTER.id,
  name: "Spring 2026",
  meta: "",
  startDate: "2026-01-01T00:00:00Z",
  endDate: "2099-12-31T00:00:00Z",
  startingBudget: 100,
  currentBudget: 100,
  membershipFee: 10,
  membershipDiscountFee: 7,
  rebuyFee: 2,
  freeTrialLimit: 4,
};
const secondSemester = {
  ...firstSemester,
  id: "df4c1c55-738e-4caf-9f54-81e439290e15",
  name: "Spring 2025",
};

const rosterMembers = [
  {
    memberId: 701,
    firstName: "Active",
    lastName: "Member",
    questId: "activemember",
    username: "activemember",
    role: "executive",
    accountStatus: "active",
    activationOutstanding: false,
    hasLogin: true,
    hasExecutiveMembership: true,
  },
  {
    memberId: 702,
    firstName: "Pending",
    lastName: "Member",
    questId: "pendingmember",
    username: "pendingmember",
    role: "tournament_director",
    accountStatus: "pending_activation",
    activationOutstanding: true,
    hasLogin: true,
    hasExecutiveMembership: true,
  },
  {
    memberId: 703,
    firstName: "Disabled",
    lastName: "Member",
    questId: "disabledmember",
    username: "disabledmember",
    role: "executive",
    accountStatus: "disabled",
    activationOutstanding: false,
    hasLogin: true,
    hasExecutiveMembership: true,
  },
  {
    memberId: 704,
    firstName: "No",
    lastName: "Login",
    questId: "nologin",
    username: null,
    role: null,
    accountStatus: null,
    activationOutstanding: false,
    hasLogin: false,
    hasExecutiveMembership: true,
  },
  {
    memberId: 705,
    firstName: "Protected",
    lastName: "Officer",
    questId: "protectedofficer",
    username: "protectedofficer",
    role: "president",
    accountStatus: "pending_activation",
    activationOutstanding: true,
    hasLogin: true,
    hasExecutiveMembership: true,
  },
];

describe("Executive team roster", () => {
  beforeEach(() => {
    Cypress.session.clearAllSavedSessions();
    cy.resetDatabase();
    cy.login("test_president", "password");
  });

  it("lists an executive member added through the API and generates a replacement activation link", () => {
    const newMember = {
      id: 97000417,
      firstName: "Roster",
      lastName: "Member",
      email: "roster.member@example.com",
      faculty: "Math",
      questId: "roster417",
    };

    cy.request({ method: "POST", url: "/api/v2/members", body: newMember, log: false })
      .its("status")
      .should("eq", 201);
    cy.request({
      method: "POST",
      url: `/api/v2/semesters/${SEMESTER.id}/executive-team`,
      body: { questId: newMember.questId, role: "executive" },
      log: false,
    }).its("status").should("eq", 201);

    cy.intercept("GET", "/api/v2/semesters", {
      log: false,
      body: { data: [firstSemester], total: 1 },
    }).as("rosterSemesters");
    cy.intercept("GET", `/api/v2/semesters/${SEMESTER.id}/executive-team`).as("executiveRoster");
    cy.visit("/admin/executive");
    cy.wait("@rosterSemesters");
    cy.wait("@executiveRoster");
    cy.getByData(`executive-team-row-${newMember.id}`).should("contain", "Roster Member").and("contain", "Pending activation");
    cy.getByData(`executive-team-generate-link-${newMember.id}`).click();
    cy.getByData(`executive-team-activation-link-${newMember.id}`)
      .invoke({ log: false }, "val")
      .should((value) => expect(value).to.match(new RegExp(`^${Cypress.config().baseUrl}/activate#token=`)));
  });

  it("changes roster with the global semester selector and never keeps the previous rows", () => {
    cy.intercept("GET", "/api/v2/semesters", {
      log: false,
      body: { data: [firstSemester, secondSemester], total: 2 },
    }).as("rosterSemesters");
    cy.intercept("GET", `/api/v2/semesters/${firstSemester.id}/executive-team`, {
      log: false,
      body: { members: [rosterMembers[0]] },
    }).as("firstRoster");
    cy.intercept("GET", `/api/v2/semesters/${secondSemester.id}/executive-team`, {
      log: false,
      body: { members: [{ ...rosterMembers[0], memberId: 801, firstName: "Second", lastName: "Semester" }] },
    }).as("secondRoster");

    cy.visit("/admin/executive");
    cy.wait("@rosterSemesters");
    cy.wait("@firstRoster");
    cy.getByData("executive-team-roster").should("contain", "Active Member");

    cy.getByData("semester-dropdown").click();
    cy.getByData(`semester-option-${secondSemester.id}`).click();
    cy.wait("@secondRoster");
    cy.getByData("executive-team-row-801").should("contain", "Second Semester");
    cy.getByData("executive-team-row-701").should("not.exist");
  });

  it("shows every returned account state and keeps protected or missing-account rows action-free", () => {
    cy.intercept("GET", "/api/v2/semesters", {
      log: false,
      body: { data: [firstSemester], total: 1 },
    }).as("rosterSemesters");
    cy.intercept("GET", `/api/v2/semesters/${firstSemester.id}/executive-team`, {
      log: false,
      body: { members: rosterMembers },
    }).as("executiveRoster");

    cy.visit("/admin/executive");
    cy.wait("@rosterSemesters");
    cy.wait("@executiveRoster");
    cy.getByData("executive-team-row-701").should("contain", "Active");
    cy.getByData("executive-team-row-702").should("contain", "Pending activation");
    cy.getByData("executive-team-row-703").should("contain", "Disabled");
    cy.getByData("executive-team-row-704").should("contain", "No account").and("contain", "Not assigned (no account)");
    cy.getByData("executive-team-row-705").should("contain", "Managed through officer transition.");
    cy.getByData("executive-team-generate-link-702").should("exist");
    cy.getByData("executive-team-generate-link-705").should("not.exist");
    cy.getByData("executive-team-generate-link-704").should("not.exist");
  });

  it("keeps the generated link available for manual copy after clipboard failure", () => {
    cy.intercept("GET", "/api/v2/semesters", {
      log: false,
      body: { data: [firstSemester], total: 1 },
    }).as("rosterSemesters");
    cy.intercept("GET", `/api/v2/semesters/${firstSemester.id}/executive-team`, {
      log: false,
      body: { members: [rosterMembers[1]] },
    }).as("executiveRoster");
    cy.intercept("POST", `/api/v2/semesters/${firstSemester.id}/executive-team/pendingmember/activation-link`, {
      log: false,
      statusCode: 201,
      body: { activationToken: "test-replacement-token" },
    }).as("replaceActivationLink");

    cy.visit("/admin/executive", {
      onBeforeLoad(window) {
        Object.defineProperty(window.navigator, "clipboard", {
          configurable: true,
          value: { writeText: () => Promise.reject(new Error("clipboard denied")) },
        });
      },
    });
    cy.wait("@rosterSemesters");
    cy.wait("@executiveRoster");
    cy.getByData("executive-team-generate-link-702").click();
    cy.wait("@replaceActivationLink");
    cy.getByData("executive-team-activation-link-702").should(
      "have.value",
      `${Cypress.config().baseUrl}/activate#token=test-replacement-token`,
    );
    cy.getByData("executive-team-row-702").contains("Copy link").click();
    cy.getByData("executive-team-row-702").find('[role="alert"]').should("contain", "Select and copy the link manually");
    cy.getByData("executive-team-activation-link-702").should(
      "have.value",
      `${Cypress.config().baseUrl}/activate#token=test-replacement-token`,
    );
    cy.get("@replaceActivationLink.all").should("have.length", 1);
  });

  it("does not offer activation actions for past or malformed semesters", () => {
    const pastSemester = { ...firstSemester, startDate: "2019-01-01T00:00:00Z", endDate: "2020-04-30T00:00:00Z" };
    const malformedSemester = { ...firstSemester, endDate: "not-a-date" };
    cy.intercept("GET", "/api/v2/semesters", { log: false, body: { data: [pastSemester], total: 1 } }).as("pastSemesters");
    cy.intercept("GET", `/api/v2/semesters/${firstSemester.id}/executive-team`, {
      log: false,
      body: { members: [rosterMembers[1]] },
    }).as("executiveRoster");

    cy.visit("/admin/executive");
    cy.wait("@pastSemesters");
    cy.wait("@executiveRoster");
    cy.getByData("executive-team-roster").should("contain", "Past semester");
    cy.getByData("executive-team-generate-link-702").should("not.exist");

    cy.intercept("GET", "/api/v2/semesters", { log: false, body: { data: [malformedSemester], total: 1 } }).as("malformedSemesters");
    cy.reload();
    cy.wait("@malformedSemesters");
    cy.wait("@executiveRoster");
    cy.getByData("executive-team-roster").should("contain", "dates are unavailable or invalid");
    cy.getByData("executive-team-generate-link-702").should("not.exist");
  });
});
