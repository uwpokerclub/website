/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import "@testing-library/jest-dom";
import { useSignups } from "../../hooks/useDashboardQueries";
import { SignupTimelineCard } from "./SignupTimelineCard";
import { SAMPLE_TERM } from "../../fixtures/sampleTerm";

jest.mock("../../hooks/useDashboardQueries", () => ({ useSignups: jest.fn() }));
const renderCard = () =>
  render(
    <MemoryRouter>
      <SignupTimelineCard semesterId="s1" />
    </MemoryRouter>,
  );

const mocked = useSignups as jest.Mock;

describe("SignupTimelineCard", () => {
  afterEach(() => jest.clearAllMocks());

  it("leads with the term total and the busiest day", () => {
    mocked.mockReturnValue({ isPending: false, isError: false, data: SAMPLE_TERM.signups, refetch: jest.fn() });
    renderCard();

    expect(screen.getByText("967")).toBeInTheDocument();
    expect(screen.getByText(/81 on the busiest day/)).toBeInTheDocument();
  });

  it("explains that older memberships are dated from first participation", () => {
    mocked.mockReturnValue({ isPending: false, isError: false, data: SAMPLE_TERM.signups, refetch: jest.fn() });
    renderCard();

    expect(screen.getByText(/dated from the member's first event/)).toBeInTheDocument();
    // The backfill means historical terms are no longer permanently empty.
    expect(screen.queryByText(/stay empty/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/collecting data since/i)).not.toBeInTheDocument();
  });

  it("treats an empty series as a fact about the term, not a wait", () => {
    mocked.mockReturnValue({
      isPending: false,
      isError: false,
      refetch: jest.fn(),
      data: { series: [], eventDates: [], dataStartsAt: "2026-09-01", total: 0 },
    });
    renderCard();

    expect(screen.getByText(/No signups recorded for this term/)).toBeInTheDocument();
    expect(screen.getByText(/Dated signups begin September 2026/)).toBeInTheDocument();
  });

  it("marks itself as sample data while the endpoint returns nothing", () => {
    mocked.mockReturnValue({ isPending: false, isError: true, data: undefined, refetch: jest.fn() });
    renderCard();

    expect(screen.getByText("Sample data")).toBeInTheDocument();
  });
  it("shows the loading state rather than a fabricated chart while the request is in flight", () => {
    mocked.mockReturnValue({ isPending: true, isError: false, data: undefined, refetch: jest.fn() });
    renderCard();

    expect(screen.getByRole("status")).toBeInTheDocument();
    expect(screen.queryByText("967")).not.toBeInTheDocument();
  });
});
