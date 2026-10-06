/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";
import type { TooltipProps } from "recharts";
import type { NameType, ValueType } from "recharts/types/component/DefaultTooltipContent";
import { ActiveTooltipContent } from "./ActiveTooltipContent";

const props = {
  label: "Fall 2026 Event #5",
  payload: [{ name: "Entries", value: 92, dataKey: "entries" }],
} as unknown as TooltipProps<ValueType, NameType>;

describe("ActiveTooltipContent", () => {
  it("renders nothing while the tooltip is inactive, so no hidden box is left behind", () => {
    const { container } = render(<ActiveTooltipContent {...props} active={false} />);

    expect(container).toBeEmptyDOMElement();
  });

  it("renders the default tooltip content while active", () => {
    render(<ActiveTooltipContent {...props} active />);

    expect(screen.getByText("Fall 2026 Event #5")).toBeInTheDocument();
    expect(screen.getByText("92")).toBeInTheDocument();
  });
});
