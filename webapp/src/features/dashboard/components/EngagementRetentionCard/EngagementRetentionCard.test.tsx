/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";
import { useEngagementDashboard, useMembershipsDashboard } from "../../hooks/useDashboardQueries";
import { EngagementRetentionCard } from "./EngagementRetentionCard";
import { SAMPLE_TERM } from "../../fixtures/sampleTerm";

jest.mock("../../hooks/useDashboardQueries", () => ({
  useEngagementDashboard: jest.fn(),
  useMembershipsDashboard: jest.fn(),
}));

const mockedEngagement = useEngagementDashboard as jest.Mock;
const mockedMemberships = useMembershipsDashboard as jest.Mock;

function ready() {
  mockedEngagement.mockReturnValue({
    isLoading: false,
    isError: false,
    data: SAMPLE_TERM.engagement,
    refetch: jest.fn(),
  });
  mockedMemberships.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.memberships });
}

describe("EngagementRetentionCard", () => {
  afterEach(() => jest.clearAllMocks());

  it("leads with distinct players against total memberships", () => {
    ready();
    render(<EngagementRetentionCard semesterId="s1" />);

    expect(screen.getByText("824")).toBeInTheDocument();
    expect(screen.getByText(/of 967 members played at least one event/)).toBeInTheDocument();
  });

  it("splits the population into three attendance buckets that sum to the total", () => {
    ready();
    render(<EngagementRetentionCard semesterId="s1" />);

    const buckets = screen.getAllByRole("listitem").map((item) => item.textContent);
    expect(buckets).toEqual(["313 played once", "470 played 2–9", "41 regulars, 10+"]);
  });

  it("places both rates against their historical band rather than showing a bare delta", () => {
    ready();
    render(<EngagementRetentionCard semesterId="s1" />);

    // Median 2 and 38% played-once are both normal despite being down year-over-year.
    expect(screen.getAllByText("normal")).toHaveLength(2);
  });

  it("states the population it counts", () => {
    ready();
    render(<EngagementRetentionCard semesterId="s1" />);

    expect(screen.getByText(/Members who entered at least one event this term/)).toBeInTheDocument();
  });

  it("renders a start-of-term state rather than zeroes when nobody has played", () => {
    mockedEngagement.mockReturnValue({
      isLoading: false,
      isError: false,
      refetch: jest.fn(),
      data: {
        current: {
          players: 0,
          medianEventsAttended: 0,
          playedOnceCount: 0,
          playedOnceShare: 0,
          tenPlusCount: 0,
        },
        comparison: null,
      },
    });
    mockedMemberships.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.memberships });
    render(<EngagementRetentionCard semesterId="s1" />);

    expect(screen.getByText("No event entries recorded yet this term.")).toBeInTheDocument();
  });

  it("renders its own error state", () => {
    mockedEngagement.mockReturnValue({ isLoading: false, isError: true, data: undefined, refetch: jest.fn() });
    mockedMemberships.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.memberships });
    render(<EngagementRetentionCard semesterId="s1" />);

    expect(screen.getByRole("button", { name: "Retry" })).toBeInTheDocument();
  });

  it("omits the member total when the memberships query has not resolved", () => {
    mockedEngagement.mockReturnValue({
      isLoading: false,
      isError: false,
      data: SAMPLE_TERM.engagement,
      refetch: jest.fn(),
    });
    mockedMemberships.mockReturnValue({ isLoading: true, isError: false, data: undefined });
    render(<EngagementRetentionCard semesterId="s1" />);

    expect(screen.getByText("members played at least one event")).toBeInTheDocument();
  });
});
