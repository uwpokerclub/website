import { SAMPLE_TERM } from "./sampleTerm";

describe("SAMPLE_TERM", () => {
  it("reads as one coherent term across every card", () => {
    const { memberships, engagement, conversion } = SAMPLE_TERM;

    expect(engagement.current.players).toBeLessThanOrEqual(memberships.current.total);
    expect(conversion.current.players).toBe(engagement.current.players);

    const buckets =
      conversion.current.paid +
      conversion.current.trialSpent +
      conversion.current.trialOpen +
      conversion.current.executive;
    expect(buckets).toBe(conversion.current.players);
  });

  it("keeps the engagement buckets within the player total", () => {
    const { players, playedOnceCount, tenPlusCount } = SAMPLE_TERM.engagement.current;

    expect(playedOnceCount + tenPlusCount).toBeLessThanOrEqual(players);
  });

  it("sums the membership buckets to the total", () => {
    const { total, paid, unpaid, discounted, executive, new: isNew, returning } = SAMPLE_TERM.memberships.current;

    expect(paid + unpaid + discounted + executive).toBe(total);
    expect(isNew + returning).toBe(total);
  });

  it("agrees with itself on the signup total", () => {
    const seriesTotal = SAMPLE_TERM.signups.series.reduce((sum, point) => sum + point.admin + point.discord, 0);

    expect(seriesTotal).toBeLessThanOrEqual(SAMPLE_TERM.signups.total);
  });
});
