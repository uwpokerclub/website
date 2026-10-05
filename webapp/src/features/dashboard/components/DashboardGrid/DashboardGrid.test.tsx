/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";
import { DashboardGrid } from "./DashboardGrid";

describe("DashboardGrid", () => {
  it("renders the lead card, the wide lane and the rail", () => {
    render(
      <DashboardGrid
        lead={<div>Lead card</div>}
        wide={[<div key="a">Wide one</div>, <div key="b">Wide two</div>]}
        rail={[<div key="c">Rail one</div>]}
      />,
    );

    expect(screen.getByText("Lead card")).toBeInTheDocument();
    expect(screen.getByText("Wide one")).toBeInTheDocument();
    expect(screen.getByText("Rail one")).toBeInTheDocument();
  });

  it("puts each card in the slot it was given", () => {
    render(
      <DashboardGrid
        data-qa="dashboard-grid"
        lead={<div>Lead card</div>}
        wide={[<div key="a">Wide one</div>]}
        rail={[<div key="c">Rail one</div>]}
      />,
    );

    expect(screen.getByTestId("dashboard-lead")).toHaveTextContent("Lead card");
    expect(screen.getByTestId("dashboard-wide")).toHaveTextContent("Wide one");
    expect(screen.getByTestId("dashboard-rail")).toHaveTextContent("Rail one");
  });

  it("reads lead, then the wide lane, then the rail — the order assistive tech follows", () => {
    const { container } = render(
      <DashboardGrid
        data-qa="dashboard-grid"
        lead={<p>Lead card</p>}
        wide={[<p key="a">Wide one</p>]}
        rail={[<p key="c">Rail one</p>]}
      />,
    );

    const text = Array.from(container.querySelectorAll("p")).map((node) => node.textContent);
    expect(text).toEqual(["Lead card", "Wide one", "Rail one"]);
  });

  it("still renders a single tagged container", () => {
    const { container } = render(<DashboardGrid data-qa="dashboard-grid" lead={<div>Lead</div>} wide={[]} rail={[]} />);

    expect(container.querySelector('[data-qa="dashboard-grid"]')).toBeInTheDocument();
  });
});
