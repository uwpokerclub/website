/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import "@testing-library/jest-dom";
import { CardActionLink } from "./CardActionLink";

describe("CardActionLink", () => {
  it("renders a router link when given a to prop", () => {
    render(
      <MemoryRouter>
        <CardActionLink to="/admin/rankings">View all</CardActionLink>
      </MemoryRouter>,
    );

    const link = screen.getByRole("link", { name: "View all" });
    expect(link).toHaveAttribute("href", "/admin/rankings");
  });

  it("renders a plain anchor when given an href prop", () => {
    render(<CardActionLink href="https://example.com/events">View all</CardActionLink>);

    const link = screen.getByRole("link", { name: "View all" });
    expect(link.tagName).toBe("A");
    expect(link).toHaveAttribute("href", "https://example.com/events");
  });
});
