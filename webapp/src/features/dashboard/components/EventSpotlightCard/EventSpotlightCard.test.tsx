/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import "@testing-library/jest-dom";
import { useSpotlight } from "../../hooks/useDashboardQueries";
import { EventSpotlightCard } from "./EventSpotlightCard";
import { SAMPLE_TERM } from "../../fixtures/sampleTerm";

jest.mock("../../hooks/useDashboardQueries", () => ({ useSpotlight: jest.fn() }));
const mocked = useSpotlight as jest.Mock;

const renderCard = () =>
  render(
    <MemoryRouter>
      <EventSpotlightCard semesterId="s1" />
    </MemoryRouter>,
  );

describe("EventSpotlightCard", () => {
  afterEach(() => jest.clearAllMocks());

  it("leads with the event's name, not a count", () => {
    mocked.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.spotlight, refetch: jest.fn() });
    renderCard();

    expect(screen.getByRole("heading", { name: "Thursday Night NLH" })).toBeInTheDocument();
    expect(screen.getByText("47")).toBeInTheDocument();
    expect(screen.getByText("12")).toBeInTheDocument();
  });

  it("links to the event it is featuring", () => {
    mocked.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.spotlight, refetch: jest.fn() });
    renderCard();

    expect(screen.getByRole("link", { name: "Open event" })).toHaveAttribute("href", "/admin/events/412");
  });

  it("counts down to an event that has not started instead of claiming it is live", () => {
    mocked.mockReturnValue({
      isLoading: false,
      isError: false,
      refetch: jest.fn(),
      data: { ...SAMPLE_TERM.spotlight, startDate: "2099-01-08T23:00:00Z" },
    });
    renderCard();

    expect(screen.queryByText(/Live/)).not.toBeInTheDocument();
    expect(screen.getByText(/Scheduled/)).toBeInTheDocument();
  });

  it("says nothing is scheduled when there is no event", () => {
    mocked.mockReturnValue({ isLoading: false, isError: false, data: null, refetch: jest.fn() });
    renderCard();

    expect(screen.getByText("No events scheduled this term.")).toBeInTheDocument();
  });
});
