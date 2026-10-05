/** @jest-environment jsdom */
import { fireEvent, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import "@testing-library/jest-dom";
import { ErrorBoundary } from "./ErrorBoundary";

function BrokenChunk(): ReactNode {
  throw new Error("Chunk failed to load");
}

describe("ErrorBoundary recovery action", () => {
  it("waits for an explicit reload click after a child fails", () => {
    const reload = jest.fn();
    const consoleError = jest.spyOn(console, "error").mockImplementation(() => {});

    try {
      render(
        <ErrorBoundary recoveryAction={{ label: "Reload app", onClick: reload }}>
          <BrokenChunk />
        </ErrorBoundary>,
      );

      expect(screen.getByRole("button", { name: "Reload app" })).toBeInTheDocument();
      expect(reload).not.toHaveBeenCalled();
      fireEvent.click(screen.getByRole("button", { name: "Reload app" }));
      expect(reload).toHaveBeenCalledTimes(1);
    } finally {
      consoleError.mockRestore();
    }
  });
});
