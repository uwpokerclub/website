/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";
import { DeltaChip } from "./DeltaChip";

describe("DeltaChip", () => {
  it("renders an increase with an upward arrow, plus sign, and percent", () => {
    render(<DeltaChip current={142} comparison={100} comparisonLabel="Fall 2025" sentiment="positive-is-good" />);

    const chip = screen.getByRole("status", { name: "Up 42 (42%) from Fall 2025" });
    expect(chip).toHaveTextContent("▲ +42 (+42%) vs Fall 2025");
  });

  it("renders a decrease with a downward arrow and minus sign", () => {
    render(<DeltaChip current={88} comparison={100} comparisonLabel="Fall 2025" sentiment="positive-is-good" />);

    const chip = screen.getByRole("status", { name: "Down 12 (12%) from Fall 2025" });
    expect(chip).toHaveTextContent("▼ -12 (-12%) vs Fall 2025");
  });

  it("colors an increase as positive when sentiment is positive-is-good", () => {
    render(<DeltaChip current={142} comparison={100} comparisonLabel="Fall 2025" sentiment="positive-is-good" />);

    expect(screen.getByRole("status")).toHaveAttribute("data-tone", "positive");
  });

  it("colors an increase as negative when sentiment is negative-is-good", () => {
    render(<DeltaChip current={142} comparison={100} comparisonLabel="Fall 2025" sentiment="negative-is-good" />);

    expect(screen.getByRole("status")).toHaveAttribute("data-tone", "negative");
  });

  it("colors a decrease as positive when sentiment is negative-is-good", () => {
    render(<DeltaChip current={88} comparison={100} comparisonLabel="Fall 2025" sentiment="negative-is-good" />);

    expect(screen.getByRole("status")).toHaveAttribute("data-tone", "positive");
  });

  it("always renders neutral color regardless of direction when sentiment is neutral", () => {
    const { rerender } = render(
      <DeltaChip current={142} comparison={100} comparisonLabel="Fall 2025" sentiment="neutral" />,
    );
    expect(screen.getByRole("status")).toHaveAttribute("data-tone", "neutral");

    rerender(<DeltaChip current={88} comparison={100} comparisonLabel="Fall 2025" sentiment="neutral" />);
    expect(screen.getByRole("status")).toHaveAttribute("data-tone", "neutral");
  });

  it("shows the absolute change only, without a percent, when the comparison value is 0", () => {
    render(<DeltaChip current={12} comparison={0} comparisonLabel="Fall 2025" sentiment="positive-is-good" />);

    const chip = screen.getByRole("status", { name: "Up 12 from Fall 2025" });
    expect(chip).toHaveTextContent("▲ +12 vs Fall 2025");
    expect(chip).not.toHaveTextContent("%");
  });

  it("shows a neutral 'No change' status when current equals comparison, never a coloured 0%", () => {
    render(<DeltaChip current={100} comparison={100} comparisonLabel="Fall 2025" sentiment="positive-is-good" />);

    const chip = screen.getByRole("status", { name: "No change from Fall 2025" });
    expect(chip).toHaveAttribute("data-tone", "neutral");
    expect(chip).toHaveTextContent("No change vs Fall 2025");
    expect(chip).not.toHaveTextContent("0%");
  });

  it("shows the comparison term inline as 'vs <term>' when not compact", () => {
    render(<DeltaChip current={142} comparison={100} comparisonLabel="Fall 2025" sentiment="positive-is-good" />);

    expect(screen.getByRole("status")).toHaveTextContent("vs Fall 2025");
  });

  it("hides the visible 'vs <term>' text but keeps the full accessible label when compact", () => {
    render(
      <DeltaChip current={142} comparison={100} comparisonLabel="Fall 2025" sentiment="positive-is-good" compact />,
    );

    const chip = screen.getByRole("status", { name: "Up 42 (42%) from Fall 2025" });
    expect(chip).not.toHaveTextContent("vs Fall 2025");
    expect(chip).toHaveTextContent("▲ +42 (+42%)");
  });
});
