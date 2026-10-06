/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";
import { DashboardGrid, DashboardGridItem } from "./DashboardGrid";

const items: DashboardGridItem[] = [
  { id: "engagement", lead: true, spans: { medium: 12, wide: 8 }, content: <p>Lead card</p> },
  { id: "termAtAGlance", lead: false, spans: { medium: 6, wide: 4 }, content: <p>Second card</p> },
  { id: "signupTimeline", lead: false, spans: { medium: 12, wide: 12 }, content: <p>Third card</p> },
];

describe("DashboardGrid", () => {
  it("marks the lead cell and renders the rest as plain cells", () => {
    render(<DashboardGrid data-qa="dashboard-grid" items={items} />);

    expect(screen.getByTestId("dashboard-lead")).toHaveTextContent("Lead card");
    expect(screen.getByTestId("dashboard-lead")).toHaveAttribute("data-qa", "dashboard-lead");
    expect(screen.getByText("Second card").parentElement).toHaveAttribute("data-qa", "dashboard-cell");
  });

  it("reads cells in the order given — the order assistive tech follows", () => {
    const { container } = render(<DashboardGrid data-qa="dashboard-grid" items={items} />);

    const text = Array.from(container.querySelectorAll("p")).map((node) => node.textContent);
    expect(text).toEqual(["Lead card", "Second card", "Third card"]);
  });

  it("hands each cell's spans to the stylesheet", () => {
    render(<DashboardGrid data-qa="dashboard-grid" items={items} />);

    const cell = screen.getByText("Second card").parentElement as HTMLElement;
    expect(cell).toHaveAttribute("data-card", "termAtAGlance");
    expect(cell.style.getPropertyValue("--dashboard-span-medium")).toBe("6");
    expect(cell.style.getPropertyValue("--dashboard-span-wide")).toBe("4");
  });

  it("still renders a single tagged container", () => {
    const { container } = render(<DashboardGrid data-qa="dashboard-grid" items={[]} />);

    expect(container.querySelector('[data-qa="dashboard-grid"]')).toBeInTheDocument();
  });
});
