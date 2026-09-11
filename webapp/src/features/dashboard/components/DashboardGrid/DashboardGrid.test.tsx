/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";
import { DashboardGrid } from "./DashboardGrid";

describe("DashboardGrid", () => {
  it("renders all of its children", () => {
    render(
      <DashboardGrid>
        <div>Card one</div>
        <div>Card two</div>
        <div>Card three</div>
      </DashboardGrid>,
    );

    expect(screen.getByText("Card one")).toBeInTheDocument();
    expect(screen.getByText("Card two")).toBeInTheDocument();
    expect(screen.getByText("Card three")).toBeInTheDocument();
  });

  it("wraps its children in a single container element", () => {
    const { container } = render(
      <DashboardGrid data-qa="dashboard-grid">
        <div>Card one</div>
        <div>Card two</div>
      </DashboardGrid>,
    );

    const grid = container.querySelector('[data-qa="dashboard-grid"]');
    expect(grid).toBeInTheDocument();
    expect(grid?.children).toHaveLength(2);
  });
});
