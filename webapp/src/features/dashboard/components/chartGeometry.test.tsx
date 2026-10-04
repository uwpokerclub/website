/** @jest-environment jsdom */
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import type { ReactElement } from "react";
import "@testing-library/jest-dom";
import { SAMPLE_TERM } from "../fixtures/sampleTerm";
import type { EventActivityResponse, SignupsResponse } from "../api/dashboardApi";
import { useEventActivity, useSignups } from "../hooks/useDashboardQueries";
import { EventActivityCard } from "./EventActivityCard/EventActivityCard";
import { SignupTimelineCard } from "./SignupTimelineCard/SignupTimelineCard";

jest.mock("recharts", () => {
  const React = jest.requireActual<typeof import("react")>("react");
  const recharts = jest.requireActual("recharts");
  return {
    ...recharts,
    ResponsiveContainer: ({ children }: { children: ReactElement }) =>
      React.cloneElement(children as ReactElement<{ width?: number; height?: number }>, { width: 640, height: 180 }),
  };
});
jest.mock("../hooks/useDashboardQueries", () => ({ useEventActivity: jest.fn(), useSignups: jest.fn() }));

const signupsHook = useSignups as jest.Mock;
const eventsHook = useEventActivity as jest.Mock;

describe("dashboard chart SVG geometry", () => {
  afterEach(() => jest.clearAllMocks());

  it("renders multi-digit signup tick labels inside a full-width left axis", async () => {
    const data: SignupsResponse = {
      total: 380,
      dataStartsAt: "2026-09-01",
      eventDates: [],
      series: [
        { date: "2026-09-01", admin: 141, discord: 0, unknown: 0 },
        { date: "2026-09-02", admin: 119, discord: 0, unknown: 0 },
        { date: "2026-09-03", admin: 120, discord: 0, unknown: 0 },
      ],
      comparison: null,
    };
    signupsHook.mockReturnValue({ data, isPending: false, isError: false, refetch: jest.fn() });
    const { container } = render(
      <MemoryRouter>
        <SignupTimelineCard semesterId="current" />
      </MemoryRouter>,
    );

    const labels = Array.from(container.querySelectorAll(".recharts-yAxis .recharts-cartesian-axis-tick-value"));
    const numericTicks = labels.map((label) => Number(label.textContent?.replace(/,/g, "")));
    const chartGrid = container.querySelector(".recharts-cartesian-grid-horizontal line");

    await waitFor(() => expect(container.querySelectorAll("path.recharts-area-area").length).toBeGreaterThan(0));
    expect(numericTicks.some((tick) => tick >= 141)).toBe(true);
    expect(labels.every((label) => Number(label.getAttribute("width")) === 48)).toBe(true);
    expect(
      labels.every((label) => Number(label.getAttribute("x")) > (label.textContent?.replace(/,/g, "").length ?? 0) * 7),
    ).toBe(true);
    expect(Number(chartGrid?.getAttribute("x1"))).toBeGreaterThanOrEqual(48);
  });

  it("keeps event tick labels inside the axis and the larger prior average within the plot domain", async () => {
    const data: EventActivityResponse = {
      ...SAMPLE_TERM.eventActivity,
      current: {
        ...SAMPLE_TERM.eventActivity.current,
        averageFieldSize: 42.6,
        series: SAMPLE_TERM.eventActivity.current.series.map((point) => ({ ...point, entries: 30 })),
      },
      comparison: { semester: { id: "winter-2024", name: "Winter 2024" }, averageFieldSize: 154.2 },
    };
    eventsHook.mockReturnValue({ data, isLoading: false, isError: false, refetch: jest.fn() });
    const { container } = render(
      <MemoryRouter>
        <EventActivityCard semesterId="current" />
      </MemoryRouter>,
    );

    expect(screen.getByText(/Winter 2024: 154.2 players per completed event/)).toBeInTheDocument();
    const labels = Array.from(container.querySelectorAll(".recharts-yAxis .recharts-cartesian-axis-tick-value"));
    const numericTicks = labels.map((label) => Number(label.textContent?.replace(/,/g, "")));
    const gridLines = Array.from(container.querySelectorAll(".recharts-cartesian-grid-horizontal line"));
    const gridYs = gridLines.map((line) => Number(line.getAttribute("y1")));
    const gridTop = Math.min(...gridYs);
    const gridBottom = Math.max(...gridYs);
    const referenceLines = Array.from(container.querySelectorAll(".recharts-reference-line-line"));
    const priorLine = referenceLines.find((line) => line.getAttribute("stroke-dasharray") === "2 3");
    const currentLine = referenceLines.find((line) => line.getAttribute("stroke-dasharray") === "4 4");
    const priorY = Number(priorLine?.getAttribute("y1"));
    const currentY = Number(currentLine?.getAttribute("y1"));

    await waitFor(() =>
      expect(container.querySelectorAll(".recharts-bar-rectangle .recharts-rectangle").length).toBeGreaterThan(0),
    );
    expect(numericTicks.some((tick) => tick >= 154.2)).toBe(true);
    expect(labels.every((label) => Number(label.getAttribute("width")) === 48)).toBe(true);
    expect(
      labels.every((label) => Number(label.getAttribute("x")) > (label.textContent?.replace(/,/g, "").length ?? 0) * 7),
    ).toBe(true);
    expect(priorLine).toBeInTheDocument();
    expect(currentLine).toBeInTheDocument();
    expect(priorLine?.getAttribute("y1")).toEqual(priorLine?.getAttribute("y2"));
    expect(priorY).toBeLessThan(currentY);
    expect(priorY).toBeGreaterThanOrEqual(gridTop);
    expect(priorY).toBeLessThanOrEqual(gridBottom);
  });
});
