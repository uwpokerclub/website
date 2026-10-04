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
    ComposedChart: ({
      children,
      data,
      margin,
    }: {
      children: ReactNode;
      data: { comparisonTotal?: number }[];
      margin: { left: number };
    }) => (
      <div
        data-chart-values={data.map((point) => point.comparisonTotal ?? "null").join(",")}
        data-chart-margin-left={margin.left}
      >
        {children}
      </div>
    ),
    Line: ({ name, connectNulls }: { name: string; connectNulls: boolean }) => (
      <span data-comparison-line={name} data-connect-nulls={connectNulls} />
    ),
    CartesianGrid: () => null,
    Legend: () => null,
    ReferenceLine: ({ x }: { x: string }) => <span data-event-date={x} />,
    ResponsiveContainer: Container,
    Tooltip: () => null,
    XAxis: () => null,
    YAxis: ({
      width,
      allowDecimals,
      tickFormatter,
    }: {
      width: number;
      allowDecimals: boolean;
      tickFormatter: (value: number) => string;
    }) => (
      <span data-y-axis-width={width} data-allow-decimals={allowDecimals} data-formatted-tick={tickFormatter(141)} />
    ),
  };
});

jest.mock("../../hooks/useDashboardQueries", () => ({ useSignups: jest.fn() }));
const response: SignupsResponse = {
  total: 10,
  dataStartsAt: "2026-09-01",
  eventDates: ["2026-09-12", "2026-09-19"],
  series: [
    { date: "2026-09-01", admin: 0, discord: 0, unknown: 0 },
    { date: "2026-09-02", admin: 1, discord: 0, unknown: 0 },
    { date: "2026-09-12", admin: 2, discord: 3, unknown: 5 },
  ],
  comparison: null,
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

  it("aligns the single prior daily-total line by elapsed calendar day and leaves uncovered days null", () => {
    const { container } = renderCardWithData({
      ...response,
      comparison: {
        semester: { id: "fall-2025", name: "Fall 2025" },
        dailyTotals: [
          { elapsedDay: 0, total: 3 },
          { elapsedDay: 1, total: 0 },
        ],
      },
    });

    expect(container.querySelector("[data-comparison-line='Fall 2025 daily total']")).toHaveAttribute(
      "data-connect-nulls",
      "false",
    );
    expect(container.querySelector("[data-chart-values]")).toHaveAttribute("data-chart-values", "3,0,null");
    expect(container.querySelectorAll("[data-event-date]")).toHaveLength(response.eventDates.length);
  });

  it("reserves a left gutter wide enough for integer signup count ticks", () => {
    const { container } = renderCardWithData(response);

    expect(container.querySelector("[data-chart-margin-left]")).toHaveAttribute("data-chart-margin-left", "0");
    expect(container.querySelector("[data-y-axis-width]")).toHaveAttribute("data-y-axis-width", "48");
    expect(container.querySelector("[data-y-axis-width]")).toHaveAttribute("data-allow-decimals", "false");
    expect(container.querySelector("[data-y-axis-width]")).toHaveAttribute("data-formatted-tick", "141");
  });

  it("shows the current series only when the comparison is unavailable or empty", () => {
    renderCardWithData({
      ...response,
      comparison: { semester: { id: "fall-2025", name: "Fall 2025" }, dailyTotals: [] },
    });
    expect(document.querySelector("[data-comparison-line]")).not.toBeInTheDocument();

    renderCardWithData(response);
    expect(document.querySelector("[data-comparison-line]")).not.toBeInTheDocument();
  });

  it("treats an empty series as a fact about the term, not a wait", () => {
    renderCardWithData({ series: [], eventDates: [], dataStartsAt: "2026-09-01", total: 0, comparison: null });

    expect(screen.getByText(/No signups recorded for this term/)).toBeInTheDocument();
    expect(screen.getByText(/Dated signups begin September 2026/)).toBeInTheDocument();
  });

  it("keeps historical comparison visible when the current term has only zero-count days", () => {
    const zeroResponse: SignupsResponse = {
      total: 0,
      dataStartsAt: "2026-09-01",
      eventDates: [],
      series: [
        { date: "2026-09-01", admin: 0, discord: 0, unknown: 0 },
        { date: "2026-09-02", admin: 0, discord: 0, unknown: 0 },
      ],
      comparison: {
        semester: { id: "fall-2025", name: "Fall 2025" },
        dailyTotals: [
          { elapsedDay: 0, total: 4 },
          { elapsedDay: 1, total: 0 },
        ],
      },
    };
    renderCardWithData(zeroResponse);

    expect(screen.getByRole("status")).toHaveTextContent("No signups recorded for this term");
    expect(document.querySelector("[data-comparison-line='Fall 2025 daily total']")).toBeInTheDocument();
    expect(document.querySelector("[data-chart-values]")).toHaveAttribute("data-chart-values", "4,0");
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
