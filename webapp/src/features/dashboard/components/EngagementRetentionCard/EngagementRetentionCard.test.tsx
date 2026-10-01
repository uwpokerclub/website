/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";
import { useEngagementDashboard } from "../../hooks/useDashboardQueries";
import { EngagementRetentionCard } from "./EngagementRetentionCard";
import type { EngagementDashboardResponse, EngagementStats } from "../../api/dashboardApi";

jest.mock("../../hooks/useDashboardQueries", () => ({
  useEngagementDashboard: jest.fn(),
}));

const mockedUseEngagementDashboard = useEngagementDashboard as jest.Mock;

function stats(overrides: Partial<EngagementStats> = {}): EngagementStats {
  return {
    players: 142,
    medianEventsAttended: 2.5,
    playedOnceCount: 48,
    playedOnceShare: 0.338,
    tenPlusCount: 16,
    ...overrides,
  };
}

function withComparison(current: EngagementStats, comparisonStats: EngagementStats): EngagementDashboardResponse {
  return {
    current,
    comparison: { semester: { id: "fall-2025", name: "Fall 2025" }, stats: comparisonStats },
  };
}

function chipFor(label: string): HTMLElement {
  const el = screen.getByText(label).closest('[data-qa="delta-chip"]');
  if (!el) throw new Error(`No delta chip found for label "${label}"`);
  return el as HTMLElement;
}

describe("EngagementRetentionCard", () => {
  afterEach(() => {
    jest.clearAllMocks();
  });

  it("renders engagement figures, their population qualifier, and appropriately-sentimented deltas", () => {
    mockedUseEngagementDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: withComparison(
        stats(),
        stats({ players: 100, medianEventsAttended: 2, playedOnceShare: 0.4, tenPlusCount: 10 }),
      ),
      refetch: jest.fn(),
    });

    render(<EngagementRetentionCard semesterId="s1" />);

    expect(screen.getByText("Members who entered at least one event this term.")).toBeInTheDocument();
    expect(screen.getByText("142")).toBeInTheDocument();
    expect(screen.getByText("2.5")).toBeInTheDocument();
    expect(screen.getByText("33.8%")).toBeInTheDocument();
    expect(screen.getByText("16")).toBeInTheDocument();
    expect(chipFor("Up 42 (42%) from Fall 2025")).toHaveAttribute("data-tone", "positive");
    expect(chipFor("Up 0.5 (25%) from Fall 2025")).toHaveAttribute("data-tone", "positive");
    expect(chipFor("Down 6.2 percentage points from Fall 2025")).toHaveAttribute("data-tone", "positive");
    expect(chipFor("Up 6 (60%) from Fall 2025")).toHaveAttribute("data-tone", "positive");
  });

  it("formats valid 0% and 100% shares", () => {
    mockedUseEngagementDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: withComparison(stats({ playedOnceShare: 0 }), stats({ playedOnceShare: 1 })),
      refetch: jest.fn(),
    });

    const { rerender } = render(<EngagementRetentionCard semesterId="s1" />);
    expect(screen.getByText("0%")).toBeInTheDocument();

    mockedUseEngagementDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: withComparison(stats({ playedOnceShare: 1 }), stats({ playedOnceShare: 0 })),
      refetch: jest.fn(),
    });
    rerender(<EngagementRetentionCard semesterId="s1" />);
    expect(screen.getByText("100%")).toBeInTheDocument();
  });

  it("shows the loading state when query data is absent", () => {
    mockedUseEngagementDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: undefined,
      refetch: jest.fn(),
    });

    render(<EngagementRetentionCard semesterId="s1" />);

    expect(screen.getByRole("status", { name: "Loading Engagement & Retention" })).toBeInTheDocument();
  });

  it("isolates a request error in this card and retries its own query", async () => {
    const user = userEvent.setup();
    const refetch = jest.fn();
    mockedUseEngagementDashboard.mockReturnValue({ isLoading: false, isError: true, data: undefined, refetch });

    render(<EngagementRetentionCard semesterId="s1" />);

    expect(screen.getByRole("alert")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Retry" }));
    expect(refetch).toHaveBeenCalledTimes(1);
  });

  it("uses the start-of-term state instead of zero figures when there are no players", () => {
    mockedUseEngagementDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: withComparison(
        stats({ players: 0, medianEventsAttended: 0, playedOnceCount: 0, playedOnceShare: 0, tenPlusCount: 0 }),
        stats(),
      ),
      refetch: jest.fn(),
    });

    render(<EngagementRetentionCard semesterId="s1" />);

    expect(screen.getByText("No event entries recorded yet this term.")).toBeInTheDocument();
    expect(screen.queryByText("DISTINCT PLAYERS")).not.toBeInTheDocument();
  });

  it("renders no chips and a plain note when there is no comparable term", () => {
    mockedUseEngagementDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: { current: stats(), comparison: null },
      refetch: jest.fn(),
    });

    const { container } = render(<EngagementRetentionCard semesterId="s1" />);

    expect(container.querySelectorAll('[data-qa="delta-chip"]')).toHaveLength(0);
    expect(screen.getByText("No comparable term to compare against yet.")).toBeInTheDocument();
  });
});
