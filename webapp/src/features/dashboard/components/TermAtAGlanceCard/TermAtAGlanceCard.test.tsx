/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";
import { useMembershipsDashboard } from "../../hooks/useDashboardQueries";
import { TermAtAGlanceCard } from "./TermAtAGlanceCard";
import type { MembershipStats, MembershipsDashboardResponse } from "../../api/dashboardApi";

jest.mock("../../hooks/useDashboardQueries", () => ({
  useMembershipsDashboard: jest.fn(),
}));

const mockedUseMembershipsDashboard = useMembershipsDashboard as jest.Mock;

function stats(overrides: Partial<MembershipStats> = {}): MembershipStats {
  return {
    total: 248,
    paid: 210,
    unpaid: 18,
    discounted: 12,
    executive: 8,
    new: 90,
    returning: 158,
    ...overrides,
  };
}

function withComparison(
  current: MembershipStats,
  comparisonStats: MembershipStats,
  comparisonName = "Fall 2025",
): MembershipsDashboardResponse {
  return {
    current,
    comparison: { semester: { id: "s-2025", name: comparisonName }, stats: comparisonStats },
  };
}

describe("TermAtAGlanceCard", () => {
  afterEach(() => {
    jest.clearAllMocks();
  });

  it("shows the loading state while the query is loading", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: true,
      isError: false,
      data: undefined,
      refetch: jest.fn(),
    });

    render(<TermAtAGlanceCard semesterId="s1" />);

    expect(screen.getByRole("status", { name: "Loading Term at a Glance" })).toBeInTheDocument();
  });

  it("shows the error state and retries via refetch when the query fails", async () => {
    const user = userEvent.setup();
    const refetch = jest.fn();
    mockedUseMembershipsDashboard.mockReturnValue({ isLoading: false, isError: true, data: undefined, refetch });

    render(<TermAtAGlanceCard semesterId="s1" />);

    await user.click(screen.getByRole("button", { name: "Retry" }));
    expect(refetch).toHaveBeenCalledTimes(1);
  });

  it("shows the empty state when there are zero memberships this term", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: withComparison(
        stats({ total: 0, paid: 0, unpaid: 0, discounted: 0, executive: 0, new: 0, returning: 0 }),
        stats(),
      ),
      refetch: jest.fn(),
    });

    render(<TermAtAGlanceCard semesterId="s1" />);

    expect(screen.getByText("No memberships recorded yet this term.")).toBeInTheDocument();
  });

  it("renders the total headline with a visible 'vs <term>' delta chip when a comparison exists", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: withComparison(stats({ total: 290 }), stats({ total: 248 })),
      refetch: jest.fn(),
    });

    render(<TermAtAGlanceCard semesterId="s1" />);

    expect(screen.getByText("290")).toBeInTheDocument();
    const chip = screen.getByRole("status", { name: "Up 42 (17%) from Fall 2025" });
    expect(chip).toHaveTextContent("vs Fall 2025");
  });

  it("renders paid, unpaid, discounted, executive, new and returning as compact mini-stats without visible 'vs' text", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: withComparison(stats(), stats({ paid: 180 })),
      refetch: jest.fn(),
    });

    render(<TermAtAGlanceCard semesterId="s1" />);

    expect(screen.getByText("Paid")).toBeInTheDocument();
    expect(screen.getByText("210")).toBeInTheDocument();
    const paidChip = screen.getByRole("status", { name: /Up 30 .* from Fall 2025/ });
    expect(paidChip).not.toHaveTextContent("vs Fall 2025");
  });

  it("colors the unpaid mini-stat as negative when it increases (negative-is-good)", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: withComparison(stats({ unpaid: 25 }), stats({ unpaid: 18 })),
      refetch: jest.fn(),
    });

    render(<TermAtAGlanceCard semesterId="s1" />);

    const unpaidChip = screen.getByRole("status", { name: /Up 7 .* from Fall 2025/ });
    expect(unpaidChip).toHaveAttribute("data-tone", "negative");
  });

  it("colors discounted and executive as neutral regardless of direction", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: withComparison(stats({ discounted: 5, executive: 10 }), stats({ discounted: 12, executive: 8 })),
      refetch: jest.fn(),
    });

    render(<TermAtAGlanceCard semesterId="s1" />);

    expect(screen.getByRole("status", { name: /Down 7 .* from Fall 2025/ })).toHaveAttribute("data-tone", "neutral");
    expect(screen.getByRole("status", { name: /Up 2 .* from Fall 2025/ })).toHaveAttribute("data-tone", "neutral");
  });

  it("shows a 'No change' chip for a figure whose value is identical to the comparison", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: withComparison(
        stats({ executive: 8 }),
        stats({ total: 200, paid: 180, unpaid: 10, discounted: 5, executive: 8, new: 70, returning: 130 }),
      ),
      refetch: jest.fn(),
    });

    render(<TermAtAGlanceCard semesterId="s1" />);

    expect(screen.getByRole("status", { name: "No change from Fall 2025" })).toBeInTheDocument();
  });

  it("shows the absolute-only format for a figure whose comparison value was 0", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: withComparison(stats({ discounted: 12 }), stats({ discounted: 0 })),
      refetch: jest.fn(),
    });

    render(<TermAtAGlanceCard semesterId="s1" />);

    const chip = screen.getByRole("status", { name: "Up 12 from Fall 2025" });
    expect(chip).not.toHaveTextContent("%");
  });

  it("renders no chips and a plain note when there is no comparable term", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: { current: stats(), comparison: null },
      refetch: jest.fn(),
    });

    render(<TermAtAGlanceCard semesterId="s1" />);

    expect(screen.queryAllByRole("status", { name: /from/ })).toHaveLength(0);
    expect(screen.getByText("No comparable term to compare against yet.")).toBeInTheDocument();
  });
});
