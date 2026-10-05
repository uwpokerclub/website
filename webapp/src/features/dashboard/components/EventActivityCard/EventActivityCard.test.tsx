/** @jest-environment jsdom */
import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import type { ReactNode } from "react";
import "@testing-library/jest-dom";
import type { EventActivityResponse } from "../../api/dashboardApi";
import { useEventActivity } from "../../hooks/useDashboardQueries";
import { EventActivityCard } from "./EventActivityCard";
import { SAMPLE_TERM } from "../../fixtures/sampleTerm";

jest.mock("recharts", () => {
  const Container = ({ children }: { children: ReactNode }) => <div>{children}</div>;

  return {
    BarChart: ({ children, margin }: { children: ReactNode; margin: { left: number } }) => (
      <div data-chart-margin-left={margin.left}>{children}</div>
    ),
    Bar: () => null,
    CartesianGrid: () => null,
    ReferenceLine: ({ y, ifOverflow, strokeDasharray }: { y: number; ifOverflow: string; strokeDasharray: string }) => (
      <span data-reference-y={y} data-if-overflow={ifOverflow} data-line-style={strokeDasharray} />
    ),
    ResponsiveContainer: Container,
    Tooltip: () => null,
    XAxis: () => null,
    YAxis: ({
      width,
      allowDecimals,
      tickFormatter,
      domain,
    }: {
      width: number;
      allowDecimals: boolean;
      tickFormatter: (value: number) => string;
      domain: [number, number];
    }) => (
      <span
        data-y-axis-width={width}
        data-allow-decimals={allowDecimals}
        data-formatted-tick={tickFormatter(154)}
        data-y-axis-domain={domain.join(",")}
      />
    ),
  };
});

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
    fireEvent.click(screen.getByText("View event entry data"));

    // eventsRun and eventsScheduled are disjoint, so the term total is their sum.
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

  it("labels and extends the prior average reference line on the count axis", () => {
    const data = {
      ...SAMPLE_TERM.eventActivity,
      current: {
        ...SAMPLE_TERM.eventActivity.current,
        averageFieldSize: 42.6,
        series: SAMPLE_TERM.eventActivity.current.series.map((point) => ({
          ...point,
          entries: Math.min(point.entries, 30),
        })),
      },
      comparison: { semester: { id: "fall-2025", name: "Fall 2025" }, averageFieldSize: 154.2 },
    };
    mocked.mockReturnValue({ isLoading: false, isError: false, data, refetch: jest.fn() });
    const { container } = renderCard();

    expect(screen.getByText("Current average: 42.6 players per completed event")).toBeInTheDocument();
    expect(
      screen.getByText("Fall 2025: 154.2 players per completed event at the same elapsed-term span"),
    ).toBeInTheDocument();
    expect(container.querySelectorAll("[data-reference-y]")).toHaveLength(2);
    expect(container.querySelector("[data-reference-y='42.6']")).toHaveAttribute("data-if-overflow", "extendDomain");
    expect(container.querySelector("[data-reference-y='154.2']")).toHaveAttribute("data-if-overflow", "extendDomain");
    expect(container.querySelector("[data-reference-y='42.6']")).toHaveAttribute("data-line-style", "4 4");
    expect(container.querySelector("[data-reference-y='154.2']")).toHaveAttribute("data-line-style", "2 3");
    expect(container.querySelector("[data-y-axis-domain]")).toHaveAttribute("data-y-axis-domain", "0,155");
    expect(data.comparison.averageFieldSize).toBeGreaterThan(
      Math.max(...data.current.series.map((point) => point.entries)),
    );
  });

  it("distinguishes no completed events from completed zero-entry events", () => {
    const noCompletedEvents: EventActivityResponse = {
      current: { eventsRun: 0, eventsScheduled: 1, totalEntries: 0, averageFieldSize: 0, series: [] },
      comparison: { semester: { id: "fall-2025", name: "Fall 2025" }, averageFieldSize: null },
    };
    mocked.mockReturnValue({ isLoading: false, isError: false, data: noCompletedEvents, refetch: jest.fn() });
    const { container, rerender } = renderCard();

    expect(screen.getByText("No completed events yet.")).toBeInTheDocument();
    expect(screen.getByText("Fall 2025: no completed events by this point.")).toBeInTheDocument();
    expect(screen.queryByText(/Current average/)).not.toBeInTheDocument();
    expect(container.querySelectorAll("[data-reference-y]")).toHaveLength(0);

    const zeroEntryEvent: EventActivityResponse = {
      current: {
        eventsRun: 1,
        eventsScheduled: 0,
        totalEntries: 0,
        averageFieldSize: 0,
        series: [{ id: 1, name: "Week 1", startDate: "2026-09-12T23:00:00Z", entries: 0 }],
      },
      comparison: null,
    };
    mocked.mockReturnValue({ isLoading: false, isError: false, data: zeroEntryEvent, refetch: jest.fn() });
    rerender(
      <MemoryRouter>
        <EventActivityCard semesterId="s1" />
      </MemoryRouter>,
    );

    expect(screen.getByText("Current average: 0.0 players per completed event")).toBeInTheDocument();
    expect(screen.queryByText("No completed events yet.")).not.toBeInTheDocument();
    expect(container.querySelectorAll("[data-reference-y]")).toHaveLength(1);
  });

  it("reserves a left gutter wide enough for integer count ticks in the existing chart", () => {
    const { container } = renderCardWithSample();

    expect(container.querySelector("[data-chart-margin-left]")).toHaveAttribute("data-chart-margin-left", "0");
    expect(container.querySelector("[data-y-axis-width]")).toHaveAttribute("data-y-axis-width", "48");
    expect(container.querySelector("[data-y-axis-width]")).toHaveAttribute("data-allow-decimals", "false");
    expect(container.querySelector("[data-y-axis-width]")).toHaveAttribute("data-formatted-tick", "154");
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

  it("provides event entries in expandable table details", () => {
    mocked.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.eventActivity, refetch: jest.fn() });
    renderCard();
    expect(screen.getByRole("table", { name: /Event entries for/ })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "Entries" })).toBeInTheDocument();
    expect(screen.getByText("View event entry data")).toBeInTheDocument();
  });

  it("never reports more events run than the term holds", () => {
    mocked.mockReturnValue({
      isPending: false,
      isError: false,
      refetch: jest.fn(),
      data: {
        current: { eventsRun: 3, eventsScheduled: 2, totalEntries: 90, averageFieldSize: 30, series: [] },
        comparison: null,
      },
    });
    renderCard();

    // Previously rendered "3 of 2 events run this term" with 0 still scheduled.
    expect(screen.getByText(/of 5 events run this term/)).toBeInTheDocument();
    expect(screen.getByText(/2 still scheduled/)).toBeInTheDocument();
    expect(screen.queryByText(/players per completed event at the same elapsed-term span/)).not.toBeInTheDocument();
    expect(document.querySelectorAll("[data-reference-y]")).toHaveLength(1);
  });

  it("shows zero current and prior averages as real values", () => {
    mocked.mockReturnValue({
      isLoading: false,
      isError: false,
      refetch: jest.fn(),
      data: {
        current: {
          eventsRun: 1,
          eventsScheduled: 0,
          totalEntries: 0,
          averageFieldSize: 0,
          series: [{ id: 1, name: "Empty event", startDate: "2026-09-01", entries: 0 }],
        },
        comparison: { semester: { id: "fall-2025", name: "Fall 2025" }, averageFieldSize: 0 },
      },
    });
    renderCard();

    expect(screen.getByText("Current average: 0.0 players per completed event")).toBeInTheDocument();
    expect(
      screen.getByText("Fall 2025: 0.0 players per completed event at the same elapsed-term span"),
    ).toBeInTheDocument();
    expect(document.querySelectorAll("[data-reference-y='0']")).toHaveLength(2);
  });
});

function renderCardWithSample() {
  mocked.mockReturnValue({ isLoading: false, isError: false, data: SAMPLE_TERM.eventActivity, refetch: jest.fn() });
  return renderCard();
}
