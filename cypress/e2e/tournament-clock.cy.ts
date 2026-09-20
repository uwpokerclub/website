import { SEMESTER, CLOCK_EVENT, ENDED_EVENT } from "../seed";

// Two rooms polling the same server-authoritative clock never agree to the
// second, so every assertion here is a range, not an exact digit. See #456.
const SYNC_TOLERANCE_SECONDS = 5;

type ClockState = {
  levelIndex: number;
  levelEndsAt: string;
  pausedAt: string | null;
};

const visitClockTab = (eventId: string) => {
  cy.visit(`/admin/events/${eventId}`);
  cy.getByData("clock-tab").click();
};

const parseTimerSeconds = (text: string): number => {
  const [minutes, seconds] = text.split(":").map(Number);
  return minutes * 60 + seconds;
};

const fetchClock = (eventId: string) =>
  cy
    .request(`/api/v2/semesters/${SEMESTER.id}/events/${eventId}/clock`)
    .then((response) => response.body as ClockState);

const remainingSecondsOf = (clock: ClockState): number => {
  const referenceTime = clock.pausedAt ? new Date(clock.pausedAt).getTime() : Date.now();
  return Math.round((new Date(clock.levelEndsAt).getTime() - referenceTime) / 1000);
};

describe("TournamentClock sync", () => {
  beforeEach(() => {
    cy.resetDatabase();
    cy.login();
  });

  it("persists the running clock's level and remaining time across a reload", () => {
    visitClockTab(CLOCK_EVENT.id);
    cy.getByData("level").should("contain", "Level 1");

    // Start the clock; it is created paused on first read.
    cy.getByData("toggle-timer-btn").click();

    cy.getByData("timer")
      .invoke("text")
      .then(parseTimerSeconds)
      .then((before) => {
        cy.wait(5000);
        cy.reload();
        cy.getByData("clock-tab").click();

        // Still on level 1, still running, and picking up roughly where it
        // left off rather than restarting from a full level (the old
        // localStorage bug).
        cy.getByData("level").should("contain", "Level 1");
        cy.getByData("timer").should(($timer) => {
          const after = parseTimerSeconds($timer.text());
          expect(after, "remaining time should have ticked down, not reset").to.be.lessThan(before);
          expect(before - after, "elapsed time since starting the clock").to.be.within(1, 20);
        });
      });
  });

  it("shows a second, independent client the same level and remaining time", () => {
    visitClockTab(CLOCK_EVENT.id);
    cy.getByData("toggle-timer-btn").click();
    cy.wait(3000);

    cy.getByData("timer")
      .invoke("text")
      .then(parseTimerSeconds)
      .then((uiRemainingSeconds) => {
        // A direct API read stands in for a second room's client: it shares
        // no browser state with the page above, only the server.
        fetchClock(CLOCK_EVENT.id).then((clock) => {
          expect(clock.levelIndex).to.eq(0);
          expect(clock.pausedAt, "second client should also see the clock running").to.be.null;
          expect(Math.abs(remainingSecondsOf(clock) - uiRemainingSeconds)).to.be.lessThan(SYNC_TOLERANCE_SECONDS);
        });
      });
  });

  it("propagates a pause to a second client within a poll interval", () => {
    visitClockTab(CLOCK_EVENT.id);
    cy.getByData("toggle-timer-btn").click(); // resume
    cy.wait(1000);
    cy.getByData("toggle-timer-btn").click(); // pause

    // One poll interval (2s) plus slack for the second client to observe it.
    cy.wait(2500);

    fetchClock(CLOCK_EVENT.id).then((clock) => {
      expect(clock.pausedAt, "a second client should see the pause").to.not.be.null;
    });
  });

  it("advances to the next level on expiry with no control request from either client", () => {
    cy.intercept("POST", /\/clock\/(pause|resume|adjust|level)$/).as("clockControl");

    visitClockTab(CLOCK_EVENT.id);
    cy.getByData("level").should("contain", "Level 1");

    // The only control call this test makes: starting the clock so level 1
    // (a 1 minute level, seeded specifically for this case) can expire.
    cy.getByData("toggle-timer-btn").click();
    cy.wait("@clockControl");

    // Both this room and the "second room" (the direct poll below) must
    // reach level 2 purely by deriving from level_ends_at, with nobody
    // issuing a control request when the level runs out.
    cy.getByData("level", { timeout: 75000 }).should("contain", "Level 2");

    fetchClock(CLOCK_EVENT.id).then((clock) => {
      expect(clock.levelIndex).to.eq(1);
    });

    cy.get("@clockControl.all").should("have.length", 1);
  });

  it("renders a read-only mirror for a role below tournament director", () => {
    cy.login("clock_executive");
    visitClockTab(CLOCK_EVENT.id);

    cy.getByData("level").should("be.visible");
    cy.getByData("timer").should("be.visible");
    cy.getByData("toggle-timer-btn").should("not.exist");
    cy.getByData("prev-level-btn").should("not.exist");
    cy.getByData("advance-level-btn").should("not.exist");
  });

  it("rejects control actions on an ended event", () => {
    cy.intercept("POST", /\/clock\/resume$/).as("resume");

    visitClockTab(ENDED_EVENT.id);
    cy.getByData("toggle-timer-btn").click();

    cy.wait("@resume").its("response.statusCode").should("eq", 409);

    // The failed action rolled back rather than leaving the clock running.
    cy.getByData("timer").invoke("text").then(parseTimerSeconds).should("be.greaterThan", 0);
  });
});
