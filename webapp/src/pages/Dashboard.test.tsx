/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import "@testing-library/jest-dom";
import { ROLES } from "@/types/roles";
import { Semester } from "@/types/semester";
import { DASHBOARD_CARDS } from "@/features/dashboard/dashboardCards";

jest.mock("@/hooks/useAuth", () => ({ useAuth: jest.fn() }));
jest.mock("@/hooks/useCurrentSemester", () => ({ useCurrentSemester: jest.fn() }));
// dashboardApi.ts imports the real apiClient, which reads `import.meta.env` —
// not valid under Jest's CommonJS transform (see useClockQueries.test.ts).
// This suite only cares about card titles/order, not card data.
const pending = () => ({ isLoading: true, isError: false, data: undefined, refetch: jest.fn() });
jest.mock("@/features/dashboard/hooks/useDashboardQueries", () => ({
  useMembershipsDashboard: jest.fn(() => pending()),
  useEngagementDashboard: jest.fn(() => pending()),
  useSpotlight: jest.fn(() => pending()),
  useEventActivity: jest.fn(() => pending()),
  useSignups: jest.fn(() => pending()),
  useTrialConversion: jest.fn(() => pending()),
  useTopRankings: jest.fn(() => pending()),
}));

import { useAuth } from "@/hooks/useAuth";
import { useCurrentSemester } from "@/hooks/useCurrentSemester";
import { Dashboard } from "./Dashboard";

const mockedUseAuth = useAuth as jest.Mock;
const mockedUseCurrentSemester = useCurrentSemester as jest.Mock;

const semester: Semester = {
  id: "sem-1",
  name: "Fall 2026",
  meta: "",
  startDate: "2026-09-01",
  endDate: "2026-12-01",
  startingBudget: 0,
  currentBudget: 0,
  membershipFee: 0,
  membershipDiscountFee: 0,
  rebuyFee: 0,
  freeTrialLimit: 0,
};

function renderDashboard() {
  return render(
    <MemoryRouter>
      <Dashboard />
    </MemoryRouter>,
  );
}

function cardTitleOrder() {
  return screen.getAllByRole("heading", { level: 3 }).map((heading) => heading.textContent);
}

function titlesIn(testId: string) {
  return Array.from(screen.getByTestId(testId).querySelectorAll("h3")).map((heading) => heading.textContent);
}

afterEach(() => {
  delete DASHBOARD_CARDS.spotlight;
});

describe("Dashboard", () => {
  it("no longer renders the ComingSoon placeholder", () => {
    mockedUseAuth.mockReturnValue({ user: { role: ROLES.EXECUTIVE } });
    mockedUseCurrentSemester.mockReturnValue({ currentSemester: semester });

    renderDashboard();

    expect(screen.queryByText(/under construction/i)).not.toBeInTheDocument();
  });

  it("leads with the role's first preset card", () => {
    mockedUseAuth.mockReturnValue({ user: { role: ROLES.PRESIDENT } });
    mockedUseCurrentSemester.mockReturnValue({ currentSemester: semester });

    renderDashboard();

    expect(screen.getByTestId("dashboard-lead")).toHaveTextContent("Engagement & retention");
  });

  it("routes chart cards to the wide lane and stat cards to the rail, preset order intact", () => {
    mockedUseAuth.mockReturnValue({ user: { role: ROLES.PRESIDENT } });
    mockedUseCurrentSemester.mockReturnValue({ currentSemester: semester });

    renderDashboard();

    expect(titlesIn("dashboard-wide")).toEqual(["Event activity", "Signup timeline"]);
    expect(titlesIn("dashboard-rail")).toEqual([
      "Memberships",
      "Trial conversion",
      "Event spotlight",
      "Leaderboard",
      "Quick actions",
    ]);
  });

  it("gives a different role a different lead card", () => {
    mockedUseAuth.mockReturnValue({ user: { role: ROLES.TREASURER } });
    mockedUseCurrentSemester.mockReturnValue({ currentSemester: semester });

    renderDashboard();

    expect(screen.getByTestId("dashboard-lead")).toHaveTextContent("Trial conversion");
  });

  it("renders all eight cards exactly once", () => {
    mockedUseAuth.mockReturnValue({ user: { role: ROLES.EXECUTIVE } });
    mockedUseCurrentSemester.mockReturnValue({ currentSemester: semester });

    renderDashboard();

    const titles = cardTitleOrder();
    expect(titles).toHaveLength(8);
    expect(new Set(titles).size).toBe(8);
  });

  it("falls back to the Ops preset when there is no user yet", () => {
    mockedUseAuth.mockReturnValue({ user: null });
    mockedUseCurrentSemester.mockReturnValue({ currentSemester: semester });

    renderDashboard();

    expect(screen.getByTestId("dashboard-lead")).toHaveTextContent("Event spotlight");
  });

  it("shows a no-semester state instead of the grid when there is no current semester", () => {
    mockedUseAuth.mockReturnValue({ user: { role: ROLES.EXECUTIVE } });
    mockedUseCurrentSemester.mockReturnValue({ currentSemester: null });

    const { container } = renderDashboard();

    expect(screen.getByText("Please select a semester to view the dashboard.")).toBeInTheDocument();
    expect(container.querySelector('[data-qa="dashboard-grid"]')).not.toBeInTheDocument();
  });

  it("shows the title but no subtitle when there is no current semester", () => {
    mockedUseAuth.mockReturnValue({ user: { role: ROLES.EXECUTIVE } });
    mockedUseCurrentSemester.mockReturnValue({ currentSemester: null });

    renderDashboard();

    expect(screen.getByRole("heading", { level: 1, name: "Dashboard" })).toBeInTheDocument();
    expect(screen.queryByText(semester.name)).not.toBeInTheDocument();
  });

  it("shows the current semester's name as the header subtitle", () => {
    mockedUseAuth.mockReturnValue({ user: { role: ROLES.EXECUTIVE } });
    mockedUseCurrentSemester.mockReturnValue({ currentSemester: semester });

    renderDashboard();

    expect(screen.getByRole("heading", { level: 1, name: "Dashboard" })).toBeInTheDocument();
    expect(screen.getByText("Fall 2026")).toBeInTheDocument();
  });

  it("renders a registered card component instead of the placeholder and passes semesterId", () => {
    const TestSpotlightCard = jest.fn(({ semesterId }: { semesterId: string }) => (
      <div data-qa="test-spotlight-card">{semesterId}</div>
    ));
    DASHBOARD_CARDS.spotlight = TestSpotlightCard;

    mockedUseAuth.mockReturnValue({ user: { role: ROLES.EXECUTIVE } });
    mockedUseCurrentSemester.mockReturnValue({ currentSemester: semester });

    renderDashboard();

    expect(TestSpotlightCard).toHaveBeenCalledWith({ semesterId: "sem-1" }, undefined);
    expect(screen.getByText("sem-1")).toBeInTheDocument();
  });
});
