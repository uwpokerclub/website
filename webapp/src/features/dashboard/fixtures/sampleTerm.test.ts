import { SAMPLE_TERM } from "./sampleTerm";

describe("SAMPLE_TERM", () => {
  it("reads as one coherent term across every card", () => {
    const { memberships, engagement } = SAMPLE_TERM;

    expect(engagement.current.players).toBeLessThanOrEqual(memberships.current.total);
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
});
