/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";
import { RangeTrack } from "./RangeTrack";

const base = {
  label: "Median events played",
  display: "2",
  min: 0,
  max: 5,
  bandLow: 2,
  bandHigh: 3,
};

describe("RangeTrack", () => {
  it("calls a value inside the historical band normal", () => {
    render(<RangeTrack {...base} value={2} />);

    expect(screen.getByText("normal")).toBeInTheDocument();
  });

  it("calls a value under the band low, and over it high", () => {
    const { rerender } = render(<RangeTrack {...base} value={1} />);
    expect(screen.getByText("below usual")).toBeInTheDocument();

    rerender(<RangeTrack {...base} value={4} display="4" />);
    expect(screen.getByText("above usual")).toBeInTheDocument();
  });

  it("spells the band out for assistive technology rather than relying on position", () => {
    render(<RangeTrack {...base} value={2} comparison={{ value: 3, label: "Fall 2025" }} />);

    expect(
      screen.getByText("2, normal for this club — the usual range is 2 to 3. Fall 2025 was 3."),
    ).toBeInTheDocument();
  });
});
