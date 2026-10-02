/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";
import { StatBar } from "./StatBar";

const segments = [
  { key: "paid", label: "paid", value: 712, color: "var(--dash-cat-1)" },
  { key: "unpaid", label: "unpaid", value: 189, color: "var(--dash-cat-2)" },
];

describe("StatBar", () => {
  it("labels every segment with its value, so no segment depends on colour alone", () => {
    render(<StatBar segments={segments} ariaLabel="Membership buckets" />);

    const items = screen.getAllByRole("listitem").map((item) => item.textContent);
    expect(items).toEqual(["712 paid", "189 unpaid"]);
  });

  it("exposes the breakdown to assistive technology as a list", () => {
    render(<StatBar segments={segments} ariaLabel="Membership buckets" />);

    expect(screen.getByRole("list", { name: "Membership buckets" })).toBeInTheDocument();
    expect(screen.getAllByRole("listitem")).toHaveLength(2);
  });

  it("drops zero-value segments from the bar but keeps them in the legend", () => {
    render(
      <StatBar
        segments={[...segments, { key: "exec", label: "exec", value: 0, color: "var(--dash-cat-3)" }]}
        ariaLabel="Membership buckets"
      />,
    );

    expect(screen.getAllByRole("listitem")).toHaveLength(3);
    expect(screen.getByTestId("statbar-track").children).toHaveLength(2);
  });
});
