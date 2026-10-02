/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import "@testing-library/jest-dom";
import { useEventActivity } from "../../hooks/useDashboardQueries";
import { EventActivityCard } from "./EventActivityCard";
import { SAMPLE_TERM } from "../../fixtures/sampleTerm";

jest.mock("../../hooks/useDashboardQueries", () => ({ useEventActivity: jest.fn() }));
const renderCard = () =>
  render(
    <MemoryRouter>
      <EventActivityCard semesterId="s1" />
    </MemoryRouter>,
  );

const mocked = useEventActivity as jest.Mock;

describe("EventActivityCard", () => {
  afterEach(() => jest.clearAllMocks());

  it("leads with events run against events scheduled", () => {
    mocked.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.eventActivity, refetch: jest.fn() });
    renderCard();

    expect(screen.getByText("11")).toBeInTheDocument();
    expect(screen.getByText(/of 14 events run this term/)).toBeInTheDocument();
  });

  it("reports the remaining events, the entry total and the average field", () => {
    mocked.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.eventActivity, refetch: jest.fn() });
    renderCard();

    expect(screen.getByText(/3 still scheduled/)).toBeInTheDocument();
    expect(screen.getByText(/486 entries/)).toBeInTheDocument();
    expect(screen.getByText(/average field 44.2/)).toBeInTheDocument();
  });

  it("shows a start-of-term state rather than a meter at zero", () => {
    mocked.mockReturnValue({
      isLoading: false,
      isError: false,
      refetch: jest.fn(),
      data: {
        current: { eventsRun: 0, eventsScheduled: 0, totalEntries: 0, averageFieldSize: 0, series: [] },
        comparison: null,
      },
    });
    renderCard();

    expect(screen.getByText("No events scheduled yet this term.")).toBeInTheDocument();
  });

  it("states what counts as a run event", () => {
    mocked.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.eventActivity, refetch: jest.fn() });
    renderCard();

    expect(screen.getByText(/counts as run once it has ended/)).toBeInTheDocument();
  });
});
