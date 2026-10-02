/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import "@testing-library/jest-dom";
import { useAuth } from "@/hooks";
import { QuickActionsCard } from "./QuickActionsCard";

jest.mock("@/hooks", () => ({ useAuth: jest.fn() }));
const mocked = useAuth as jest.Mock;

const renderCard = () =>
  render(
    <MemoryRouter>
      <QuickActionsCard />
    </MemoryRouter>,
  );

describe("QuickActionsCard", () => {
  afterEach(() => jest.clearAllMocks());

  it("offers every destination the session is permitted", () => {
    mocked.mockReturnValue({ hasPermission: () => true });
    renderCard();

    expect(screen.getByRole("link", { name: "Events" })).toHaveAttribute("href", "/admin/events");
    expect(screen.getByRole("link", { name: "Members" })).toHaveAttribute("href", "/admin/members");
    expect(screen.getByRole("link", { name: "Rankings" })).toHaveAttribute("href", "/admin/rankings");
  });

  it("hides destinations the session cannot reach", () => {
    mocked.mockReturnValue({
      hasPermission: (_action: string, resource: string) => resource === "semester",
    });
    renderCard();

    expect(screen.queryByRole("link", { name: "Events" })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Rankings" })).toBeInTheDocument();
  });

  it("says so when the session can reach none of them", () => {
    mocked.mockReturnValue({ hasPermission: () => false });
    renderCard();

    expect(screen.getByText("No actions available for your role.")).toBeInTheDocument();
  });
});
