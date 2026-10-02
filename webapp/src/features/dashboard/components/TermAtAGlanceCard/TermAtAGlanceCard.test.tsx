/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";
import { useMembershipsDashboard } from "../../hooks/useDashboardQueries";
import { useCurrentSemester } from "@/hooks";
import { TermAtAGlanceCard } from "./TermAtAGlanceCard";
import { SAMPLE_TERM } from "../../fixtures/sampleTerm";

jest.mock("../../hooks/useDashboardQueries", () => ({
  useMembershipsDashboard: jest.fn(),
}));
jest.mock("@/hooks", () => ({ useCurrentSemester: jest.fn() }));

const mockedUseMembershipsDashboard = useMembershipsDashboard as jest.Mock;
const mockedUseCurrentSemester = useCurrentSemester as jest.Mock;

// Fall 2026: 85-day term, 19 days elapsed — the real shape of the current term.
const TERM = { startDate: "2026-09-13T00:00:00Z", endDate: "2026-12-07T00:00:00Z" };

beforeEach(() => {
  jest.useFakeTimers().setSystemTime(new Date("2026-10-02T00:00:00Z"));
  mockedUseCurrentSemester.mockReturnValue({ currentSemester: TERM });
});

afterEach(() => jest.useRealTimers());

describe("TermAtAGlanceCard", () => {
  afterEach(() => jest.clearAllMocks());

  it("shows the total with its delta and both compositions as bars", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: SAMPLE_TERM.memberships,
      refetch: jest.fn(),
    });
    render(<TermAtAGlanceCard semesterId="s1" />);

    expect(screen.getByText("967")).toBeInTheDocument();
    expect(screen.getByRole("list", { name: "Memberships by payment status" })).toBeInTheDocument();
    expect(screen.getByRole("list", { name: "Memberships by prior membership" })).toBeInTheDocument();
    expect(screen.getByText("712")).toBeInTheDocument();
    expect(screen.getByText("561")).toBeInTheDocument();
  });

  it("compares against the same point last term, not its finished total", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: SAMPLE_TERM.memberships,
      refetch: jest.fn(),
    });
    const { container } = render(<TermAtAGlanceCard semesterId="s1" />);

    // 967 now vs 847 by this point in Fall 2025 — ahead, not the "-479 vs 891 final"
    // a comparison against the completed term would have shown all term long.
    const chip = container.querySelector('[data-qa="delta-chip"]');
    expect(chip).toHaveAttribute("data-tone", "positive");
    expect(chip).toHaveTextContent("+120");
  });

  it("shows pace against the final total when the prior term has no dated memberships", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      refetch: jest.fn(),
      data: {
        ...SAMPLE_TERM.memberships,
        comparison: { ...SAMPLE_TERM.memberships.comparison!, totalAsOf: null },
      },
    });
    const { container } = render(<TermAtAGlanceCard semesterId="s1" />);

    // No delta chip: there is no honest one to show.
    expect(container.querySelector('[data-qa="delta-chip"]')).toBeNull();
    expect(screen.getByText(/of Fall 2025's final 891/)).toBeInTheDocument();
  });

  it("states there is nothing to compare against rather than showing a zero delta", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: { ...SAMPLE_TERM.memberships, comparison: null },
      refetch: jest.fn(),
    });
    const { container } = render(<TermAtAGlanceCard semesterId="s1" />);

    expect(screen.getByText("No comparable term to compare against yet.")).toBeInTheDocument();
    expect(container.querySelector('[data-qa="delta-chip"]')).toBeNull();
  });

  it("renders a start-of-term state rather than a bar of zeroes", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      refetch: jest.fn(),
      data: {
        current: { total: 0, paid: 0, unpaid: 0, discounted: 0, executive: 0, new: 0, returning: 0 },
        comparison: null,
      },
    });
    render(<TermAtAGlanceCard semesterId="s1" />);

    expect(screen.getByText("No memberships recorded yet this term.")).toBeInTheDocument();
  });

  it("renders its own error state and retries", async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const refetch = jest.fn();
    mockedUseMembershipsDashboard.mockReturnValue({ isLoading: false, isError: true, data: undefined, refetch });
    render(<TermAtAGlanceCard semesterId="s1" />);

    await user.click(screen.getByRole("button", { name: "Retry" }));
    expect(refetch).toHaveBeenCalledTimes(1);
  });

  it("tracks progress toward a full term's worth, marking where last year stood", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: SAMPLE_TERM.memberships,
      refetch: jest.fn(),
    });
    render(<TermAtAGlanceCard semesterId="s1" />);

    const track = screen.getByRole("img", { name: /progress toward/i });
    // 967 of Fall 2025's final 891 — already past a full term's worth.
    expect(track).toHaveAccessibleName(/967 of 891/);
    expect(screen.getByText(/109% of Fall 2025's final 891/)).toBeInTheDocument();
    expect(screen.getByText(/day 19 of 85/)).toBeInTheDocument();
  });

  it("marks last year's position on the track only when that figure is known", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      refetch: jest.fn(),
      data: {
        ...SAMPLE_TERM.memberships,
        comparison: { ...SAMPLE_TERM.memberships.comparison!, totalAsOf: null },
      },
    });
    const { container } = render(<TermAtAGlanceCard semesterId="s1" />);

    expect(container.querySelector('[data-qa="pace-marker"]')).toBeNull();
    expect(screen.getByText(/109% of Fall 2025's final 891/)).toBeInTheDocument();
  });

  it("shows no track when there is no comparable term", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: { ...SAMPLE_TERM.memberships, comparison: null },
      refetch: jest.fn(),
    });
    render(<TermAtAGlanceCard semesterId="s1" />);

    expect(screen.queryByRole("img", { name: /progress toward/i })).not.toBeInTheDocument();
  });
});
