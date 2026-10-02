/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import "@testing-library/jest-dom";
import { useTopRankings } from "../../hooks/useDashboardQueries";
import { LeaderboardCard } from "./LeaderboardCard";
import { SAMPLE_TERM } from "../../fixtures/sampleTerm";

jest.mock("../../hooks/useDashboardQueries", () => ({ useTopRankings: jest.fn() }));
const mocked = useTopRankings as jest.Mock;

const renderCard = () =>
  render(
    <MemoryRouter>
      <LeaderboardCard semesterId="s1" />
    </MemoryRouter>,
  );

describe("LeaderboardCard", () => {
  afterEach(() => jest.clearAllMocks());

  it("ranks players by the position the API computed, not the array index", () => {
    mocked.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.rankings, refetch: jest.fn() });
    renderCard();

    expect(screen.getByText("Robin Chen")).toBeInTheDocument();
    expect(screen.getByText("412")).toBeInTheDocument();
    expect(screen.getAllByRole("listitem")).toHaveLength(5);
  });

  it("honours a tie rather than renumbering it", () => {
    mocked.mockReturnValue({
      isLoading: false,
      isError: false,
      refetch: jest.fn(),
      data: [
        { id: 1, firstName: "Robin", lastName: "Chen", points: 412, position: 1 },
        { id: 2, firstName: "Asha", lastName: "Patel", points: 412, position: 1 },
        { id: 3, firstName: "Michael", lastName: "Osei", points: 300, position: 3 },
      ],
    });
    renderCard();

    const ranks = screen.getAllByRole("listitem").map((item) => item.textContent?.slice(0, 1));
    expect(ranks).toEqual(["1", "1", "3"]);
  });

  it("says so when nobody has scored yet", () => {
    mocked.mockReturnValue({ isLoading: false, isError: false, data: [], refetch: jest.fn() });
    renderCard();

    expect(screen.getByText("No points awarded yet this term.")).toBeInTheDocument();
  });
});
