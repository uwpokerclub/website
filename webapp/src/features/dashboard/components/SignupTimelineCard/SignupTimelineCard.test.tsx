/** @jest-environment jsdom */
import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import type { ReactNode } from "react";
import "@testing-library/jest-dom";
import { useSignups } from "../../hooks/useDashboardQueries";
import { SignupTimelineCard } from "./SignupTimelineCard";
import type { SignupsResponse } from "../../api/dashboardApi";

jest.mock("recharts", () => {
  const Container = ({ children }: { children: ReactNode }) => <div>{children}</div>;

  return {
    Area: ({ name }: { name: string }) => <span data-series={name}>{name}</span>,
    AreaChart: Container,
    CartesianGrid: () => null,
    Legend: () => null,
    ReferenceLine: ({ x }: { x: string }) => <span data-event-date={x} />,
    ResponsiveContainer: Container,
    Tooltip: () => null,
    XAxis: () => null,
    YAxis: () => null,
  };
});

jest.mock("../../hooks/useDashboardQueries", () => ({ useSignups: jest.fn() }));
const response: SignupsResponse = {
  total: 10,
  dataStartsAt: "2026-09-01",
  eventDates: ["2026-09-12", "2026-09-19"],
  series: [
    { date: "2026-09-01", admin: 0, discord: 0, unknown: 0 },
    { date: "2026-09-12", admin: 2, discord: 3, unknown: 5 },
  ],
};

const renderCard = () =>
  render(
    <MemoryRouter>
      <SignupTimelineCard semesterId="s1" />
    </MemoryRouter>,
  );

const mocked = useSignups as jest.Mock;

describe("SignupTimelineCard", () => {
  afterEach(() => jest.clearAllMocks());

  it("renders real admin, Discord, and unknown data and includes all buckets in the busiest day", () => {
    const { container } = renderCardWithData(response);

    expect(screen.getByText("10")).toBeInTheDocument();
    expect(screen.getByText(/10 on the busiest day/)).toBeInTheDocument();
    expect(screen.getByText("Admin")).toBeInTheDocument();
    expect(screen.getByText("Discord")).toBeInTheDocument();
    expect(screen.getByText("Unknown")).toBeInTheDocument();
    expect(container.querySelectorAll("[data-event-date]")).toHaveLength(response.eventDates.length);
  });

  it("explains that older memberships are dated from first participation", () => {
    renderCardWithData(response);

    expect(screen.getByText(/dated from the member's first event/)).toBeInTheDocument();
    // The backfill means historical terms are no longer permanently empty.
    expect(screen.queryByText(/stay empty/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/collecting data since/i)).not.toBeInTheDocument();
  });

  it("treats an empty series as a fact about the term, not a wait", () => {
    renderCardWithData({ series: [], eventDates: [], dataStartsAt: "2026-09-01", total: 0 });

    expect(screen.getByText(/No signups recorded for this term/)).toBeInTheDocument();
    expect(screen.getByText(/Dated signups begin September 2026/)).toBeInTheDocument();
  });

  it("keeps a failed request card-local and retries without showing sample data", () => {
    const refetch = jest.fn();
    mocked.mockReturnValue({ isPending: false, isError: true, data: undefined, refetch });
    renderCard();

    expect(screen.getByRole("alert")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Retry" })).toBeInTheDocument();
    expect(screen.queryByText("Sample data")).not.toBeInTheDocument();
    expect(screen.queryByText("967")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    expect(refetch).toHaveBeenCalledTimes(1);
  });
  it("shows the loading state rather than a fabricated chart while the request is in flight", () => {
    mocked.mockReturnValue({ isPending: true, isError: false, data: undefined, refetch: jest.fn() });
    renderCard();

    expect(screen.getByRole("status")).toBeInTheDocument();
    expect(screen.queryByText("967")).not.toBeInTheDocument();
  });
});

function renderCardWithData(data: SignupsResponse) {
  mocked.mockReturnValue({ isPending: false, isError: false, data, refetch: jest.fn() });
  return renderCard();
}
