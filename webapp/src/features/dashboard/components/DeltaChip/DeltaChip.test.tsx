/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";
import { DeltaChip } from "./DeltaChip";

function chipFor(label: string): HTMLElement {
  const el = screen.getByText(label).closest('[data-qa="delta-chip"]');
  if (!el) throw new Error(`No delta chip found for label "${label}"`);
  return el as HTMLElement;
}

describe("DeltaChip", () => {
  it("renders an increase with an upward arrow, plus sign, and percent", () => {
    render(<DeltaChip current={142} comparison={100} comparisonLabel="Fall 2025" sentiment="positive-is-good" />);

    const chip = chipFor("Up 42 (42%) from Fall 2025");
    expect(chip).toHaveTextContent("▲ +42 (+42%) vs Fall 2025");
  });

  it("renders a decrease with a downward arrow and the real minus sign", () => {
    render(<DeltaChip current={88} comparison={100} comparisonLabel="Fall 2025" sentiment="positive-is-good" />);

    const chip = chipFor("Down 12 (12%) from Fall 2025");
    expect(chip).toHaveTextContent("▼ −12 (−12%) vs Fall 2025");
  });

  it("colors an increase as positive when sentiment is positive-is-good", () => {
    render(<DeltaChip current={142} comparison={100} comparisonLabel="Fall 2025" sentiment="positive-is-good" />);

    expect(chipFor("Up 42 (42%) from Fall 2025")).toHaveAttribute("data-tone", "positive");
  });

  it("colors an increase as negative when sentiment is negative-is-good", () => {
    render(<DeltaChip current={142} comparison={100} comparisonLabel="Fall 2025" sentiment="negative-is-good" />);

    expect(chipFor("Up 42 (42%) from Fall 2025")).toHaveAttribute("data-tone", "negative");
  });

  it("colors a decrease as positive when sentiment is negative-is-good", () => {
    render(<DeltaChip current={88} comparison={100} comparisonLabel="Fall 2025" sentiment="negative-is-good" />);

    expect(chipFor("Down 12 (12%) from Fall 2025")).toHaveAttribute("data-tone", "positive");
  });

  it("always renders neutral color regardless of direction when sentiment is neutral", () => {
    const { rerender } = render(
      <DeltaChip current={142} comparison={100} comparisonLabel="Fall 2025" sentiment="neutral" />,
    );
    expect(chipFor("Up 42 (42%) from Fall 2025")).toHaveAttribute("data-tone", "neutral");

    rerender(<DeltaChip current={88} comparison={100} comparisonLabel="Fall 2025" sentiment="neutral" />);
    expect(chipFor("Down 12 (12%) from Fall 2025")).toHaveAttribute("data-tone", "neutral");
  });

  it("shows the absolute change only, without a percent, when the comparison value is 0", () => {
    render(<DeltaChip current={12} comparison={0} comparisonLabel="Fall 2025" sentiment="positive-is-good" />);

    const chip = chipFor("Up 12 from Fall 2025");
    expect(chip).toHaveTextContent("▲ +12 vs Fall 2025");
    expect(chip).not.toHaveTextContent("%");
  });

  it("shows a neutral 'No change' status when current equals comparison, never a coloured 0%", () => {
    render(<DeltaChip current={100} comparison={100} comparisonLabel="Fall 2025" sentiment="positive-is-good" />);

    const chip = chipFor("No change from Fall 2025");
    expect(chip).toHaveAttribute("data-tone", "neutral");
    expect(chip).toHaveTextContent("No change vs Fall 2025");
    expect(chip).not.toHaveTextContent("0%");
  });

  it("shows the comparison term inline as 'vs <term>' when not compact", () => {
    render(<DeltaChip current={142} comparison={100} comparisonLabel="Fall 2025" sentiment="positive-is-good" />);

    expect(chipFor("Up 42 (42%) from Fall 2025")).toHaveTextContent("vs Fall 2025");
  });

  it("hides the visible 'vs <term>' text but keeps the full accessible label when compact", () => {
    render(
      <DeltaChip current={142} comparison={100} comparisonLabel="Fall 2025" sentiment="positive-is-good" compact />,
    );

    const chip = chipFor("Up 42 (42%) from Fall 2025");
    expect(chip).not.toHaveTextContent("vs Fall 2025");
    expect(chip).toHaveTextContent("▲ +42 (+42%)");
  });

  it("shows '<1%' rather than a misleading '0%' when a real change rounds down to zero percent", () => {
    render(<DeltaChip current={1003} comparison={1000} comparisonLabel="Fall 2025" sentiment="positive-is-good" />);

    const chip = chipFor("Up 3 (less than 1%) from Fall 2025");
    expect(chip).toHaveTextContent("▲ +3 (+<1%) vs Fall 2025");
    expect(chip).not.toHaveTextContent("0%");
  });

  it("does not expose role=status, so a card with many chips does not become a wall of live-region announcements", () => {
    render(<DeltaChip current={142} comparison={100} comparisonLabel="Fall 2025" sentiment="positive-is-good" />);

    expect(screen.queryByRole("status")).not.toBeInTheDocument();
  });

  it("hides the visible text from assistive tech and exposes only the full sentence", () => {
    render(<DeltaChip current={142} comparison={100} comparisonLabel="Fall 2025" sentiment="positive-is-good" />);

    const visible = screen.getByText("▲ +42 (+42%) vs Fall 2025");
    expect(visible).toHaveAttribute("aria-hidden", "true");
  });
});
