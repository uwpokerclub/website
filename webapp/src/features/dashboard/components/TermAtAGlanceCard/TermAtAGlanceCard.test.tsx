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

function chipFor(label: string): HTMLElement {
  const el = screen.getByText(label).closest('[data-qa="delta-chip"]');
  if (!el) throw new Error(`No delta chip found for label "${label}"`);
  return el as HTMLElement;
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

    expect(screen.getByRole("status", { name: "Loading Memberships" })).toBeInTheDocument();
  });

  it("shows the loading state, not ready, when data is missing but the query is neither loading nor errored", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: undefined,
      refetch: jest.fn(),
    });

    render(<TermAtAGlanceCard semesterId="s1" />);

    expect(screen.getByRole("status", { name: "Loading Memberships" })).toBeInTheDocument();
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
    const chip = chipFor("Up 42 (17%) from Fall 2025");
    expect(chip).toHaveTextContent("vs Fall 2025");
  });

  it("shows a TOTAL eyebrow above the headline number, before the chip, styled like the mini-stat eyebrows", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: withComparison(stats({ total: 290 }), stats({ total: 248 })),
      refetch: jest.fn(),
    });

    render(<TermAtAGlanceCard semesterId="s1" />);

    const eyebrow = screen.getByText("TOTAL");
    const total = screen.getByText("290");
    const chip = chipFor("Up 42 (17%) from Fall 2025");

    expect(eyebrow.className).toBe(screen.getByText("Paid").className);
    expect(eyebrow.compareDocumentPosition(total)).toBe(Node.DOCUMENT_POSITION_FOLLOWING);
    expect(total.compareDocumentPosition(chip)).toBe(Node.DOCUMENT_POSITION_FOLLOWING);
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
    const paidChip = chipFor("Up 30 (17%) from Fall 2025");
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

    const unpaidChip = chipFor("Up 7 (39%) from Fall 2025");
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

    expect(chipFor("Down 7 (58%) from Fall 2025")).toHaveAttribute("data-tone", "neutral");
    expect(chipFor("Up 2 (25%) from Fall 2025")).toHaveAttribute("data-tone", "neutral");
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

    expect(screen.getByText("No change from Fall 2025")).toBeInTheDocument();
  });

  it("shows the absolute-only format for a figure whose comparison value was 0", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: withComparison(stats({ discounted: 12 }), stats({ discounted: 0 })),
      refetch: jest.fn(),
    });

    render(<TermAtAGlanceCard semesterId="s1" />);

    const chip = chipFor("Up 12 from Fall 2025");
    expect(chip).not.toHaveTextContent("%");
  });

  it("renders no chips and a plain note when there is no comparable term", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: { current: stats(), comparison: null },
      refetch: jest.fn(),
    });

    const { container } = render(<TermAtAGlanceCard semesterId="s1" />);

    expect(container.querySelectorAll('[data-qa="delta-chip"]')).toHaveLength(0);
    expect(screen.getByText("No comparable term to compare against yet.")).toBeInTheDocument();
  });
});
