/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";
import { useMembershipsDashboard } from "../../hooks/useDashboardQueries";
import { TermAtAGlanceCard } from "./TermAtAGlanceCard";
import { SAMPLE_TERM } from "../../fixtures/sampleTerm";

jest.mock("../../hooks/useDashboardQueries", () => ({
  useMembershipsDashboard: jest.fn(),
}));

const mockedUseMembershipsDashboard = useMembershipsDashboard as jest.Mock;

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

  it("states there is nothing to compare against rather than showing a zero delta", () => {
    mockedUseMembershipsDashboard.mockReturnValue({
      isLoading: false,
      isError: false,
      data: { ...SAMPLE_TERM.memberships, comparison: null },
      refetch: jest.fn(),
    });
    render(<TermAtAGlanceCard semesterId="s1" />);

    expect(screen.getByText("No comparable term to compare against yet.")).toBeInTheDocument();
    expect(screen.queryByTestId("delta-chip")).not.toBeInTheDocument();
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
    const user = userEvent.setup();
    const refetch = jest.fn();
    mockedUseMembershipsDashboard.mockReturnValue({ isLoading: false, isError: true, data: undefined, refetch });
    render(<TermAtAGlanceCard semesterId="s1" />);

    await user.click(screen.getByRole("button", { name: "Retry" }));
    expect(refetch).toHaveBeenCalledTimes(1);
  });
});
