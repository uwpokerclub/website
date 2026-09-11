/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";
import { DashboardCard } from "./DashboardCard";

describe("DashboardCard", () => {
  it("renders the title and an optional action slot", () => {
    render(
      <DashboardCard title="Event Spotlight" status="ready" action={<button>View all</button>}>
        {() => <p>body</p>}
      </DashboardCard>,
    );

    expect(screen.getByRole("heading", { name: "Event Spotlight" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "View all" })).toBeInTheDocument();
  });

  it("shows a skeleton and does not call the render function while loading", () => {
    const children = jest.fn(() => <p>body</p>);
    render(
      <DashboardCard title="Event Spotlight" status="loading">
        {children}
      </DashboardCard>,
    );

    const skeleton = screen.getByRole("status", { name: "Loading Event Spotlight" });
    expect(skeleton).toHaveAttribute("aria-busy", "true");
    expect(children).not.toHaveBeenCalled();
    expect(screen.queryByText("body")).not.toBeInTheDocument();
  });

  it("shows an alert with a retry button and does not call the render function on error", async () => {
    const user = userEvent.setup();
    const children = jest.fn(() => <p>body</p>);
    const onRetry = jest.fn();
    render(
      <DashboardCard title="Event Spotlight" status="error" errorMessage="Could not load spotlight" onRetry={onRetry}>
        {children}
      </DashboardCard>,
    );

    const alert = screen.getByRole("alert");
    expect(alert).toHaveTextContent("Could not load spotlight");

    const retryButton = screen.getByRole("button", { name: "Retry" });
    expect(retryButton.tagName).toBe("BUTTON");

    await user.click(retryButton);
    expect(onRetry).toHaveBeenCalledTimes(1);
    expect(children).not.toHaveBeenCalled();
  });

  it("shows the empty message and does not call the render function when empty", () => {
    const children = jest.fn(() => <p>body</p>);
    render(
      <DashboardCard title="Event Spotlight" status="empty" emptyMessage="No events scheduled">
        {children}
      </DashboardCard>,
    );

    expect(screen.getByText("No events scheduled")).toBeInTheDocument();
    expect(children).not.toHaveBeenCalled();
  });

  it("calls the render function and shows its result when ready", () => {
    const children = jest.fn(() => <p>Spotlight body</p>);
    render(
      <DashboardCard title="Event Spotlight" status="ready">
        {children}
      </DashboardCard>,
    );

    expect(children).toHaveBeenCalledTimes(1);
    expect(screen.getByText("Spotlight body")).toBeInTheDocument();
  });
});
