import PointsService from "./PointsService";

describe("PointsService", () => {
  it("should calculate correct points for 64 players", () => {
    const ps = new PointsService(64);

    const player1Points = ps.calculatePoints(1);
    const player2Points = ps.calculatePoints(50);
    const player3Points = ps.calculatePoints(35);

    expect(player1Points).toBe(40);
    expect(player2Points).toBe(1);
    expect(player3Points).toBe(2);
  });

  it("should calculate correct points for 25 players", () => {
    const ps = new PointsService(25);

    const player1Points = ps.calculatePoints(4);
    const player2Points = ps.calculatePoints(6);
    const player3Points = ps.calculatePoints(11);

    expect(player1Points).toBe(10);
    expect(player2Points).toBe(8);
    expect(player3Points).toBe(4);
  });

  it("should calculate correct points for 43 players", () => {
    const ps = new PointsService(43);

    const player1Points = ps.calculatePoints(5);
    const player2Points = ps.calculatePoints(40);
    const player3Points = ps.calculatePoints(30);

    expect(player1Points).toBe(15);
    expect(player2Points).toBe(1);
    expect(player3Points).toBe(2);
  });
});
