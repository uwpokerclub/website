/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";
import { ExecutiveTeamRoster } from "./ExecutiveTeamRoster";

jest.mock("@/hooks", () => ({
  useAuth: jest.fn(),
  useCurrentSemester: jest.fn(),
}));
jest.mock("@/features/semesters/hooks/useSemesterQueries", () => ({ useSemesters: jest.fn() }));
jest.mock("../hooks/useExecutiveTeamQueries", () => ({ useExecutiveTeam: jest.fn() }));
jest.mock("../api/executiveTeamApi", () => ({ createExecutiveTeamActivationLink: jest.fn() }));
jest.mock("@/components", () => ({
  QueryErrorState: ({
    title,
    message,
    onRetry,
    "data-qa": dataQa,
    retryDataQa,
  }: {
    title: string;
    message: string;
    onRetry: () => void;
    "data-qa"?: string;
    retryDataQa?: string;
  }) => (
    <div data-qa={dataQa}>
      <h3>{title}</h3>
      <p>{message}</p>
      <button data-qa={retryDataQa} onClick={onRetry}>
        Retry
      </button>
    </div>
  ),
}));
jest.mock("@uwpokerclub/components", () => ({
  useToast: () => ({ showToast: jest.fn() }),
}));

import { useAuth, useCurrentSemester } from "@/hooks";
import { useSemesters } from "@/features/semesters/hooks/useSemesterQueries";
import { createExecutiveTeamActivationLink } from "../api/executiveTeamApi";
import { useExecutiveTeam } from "../hooks/useExecutiveTeamQueries";
import type { ExecutiveTeamMember } from "../types";

const semester = {
  id: "semester-1",
  name: "Fall 2026",
  startDate: "2026-09-01T00:00:00Z",
  endDate: "2099-12-31T00:00:00Z",
  meta: "",
  startingBudget: 0,
  currentBudget: 0,
  membershipFee: 0,
  membershipDiscountFee: 0,
  rebuyFee: 0,
  freeTrialLimit: 0,
};

function member(overrides: Partial<ExecutiveTeamMember> = {}): ExecutiveTeamMember {
  return {
    memberId: 1,
    firstName: "Ada",
    lastName: "Lovelace",
    questId: "adalovelace",
    username: "adalovelace",
    role: "executive",
    accountStatus: "active",
    activationOutstanding: false,
    hasLogin: true,
    hasExecutiveMembership: true,
    ...overrides,
  };
}

function queryResult(overrides: Partial<ReturnType<typeof useExecutiveTeam>> = {}) {
  return {
    data: { members: [] },
    isPending: false,
    isError: false,
    error: null,
    refetch: jest.fn(),
    ...overrides,
  } as ReturnType<typeof useExecutiveTeam>;
}

function setup(
  overrides: {
    semesterValue?: typeof semester | null;
    canView?: boolean;
    query?: Partial<ReturnType<typeof useExecutiveTeam>>;
    semesterLoading?: boolean;
    semestersError?: boolean;
  } = {},
) {
  (useAuth as jest.Mock).mockReturnValue({ hasPermission: () => overrides.canView ?? true });
  (useCurrentSemester as jest.Mock).mockReturnValue({
    currentSemester: overrides.semesterValue === undefined ? semester : overrides.semesterValue,
    loading: overrides.semesterLoading ?? false,
  });
  (useSemesters as jest.Mock).mockReturnValue({
    isLoading: false,
    isError: overrides.semestersError ?? false,
    refetch: jest.fn(),
  });
  (useExecutiveTeam as jest.Mock).mockReturnValue(queryResult(overrides.query));
}

describe("ExecutiveTeamRoster", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    setup();
  });

  it("renders account states, missing login, and protected-role guidance clearly", () => {
    (useExecutiveTeam as jest.Mock).mockReturnValue(
      queryResult({
        data: {
          members: [
            member({ memberId: 1, accountStatus: "active" }),
            member({
              memberId: 2,
              firstName: "Grace",
              lastName: "Hopper",
              role: "tournament_director",
              accountStatus: "pending_activation",
              activationOutstanding: true,
            }),
            member({
              memberId: 3,
              firstName: "Katherine",
              lastName: "Johnson",
              role: "president",
              accountStatus: "pending_activation",
              activationOutstanding: true,
            }),
            member({
              memberId: 4,
              firstName: "No",
              lastName: "Login",
              questId: "nologin",
              username: null,
              role: null,
              accountStatus: null,
              hasLogin: false,
            }),
            member({ memberId: 5, firstName: "Disabled", lastName: "Account", accountStatus: "disabled" }),
          ],
        },
      }),
    );

    const { container } = render(<ExecutiveTeamRoster />);

    expect(screen.getByText("Active")).toBeInTheDocument();
    expect(screen.getAllByText("Pending activation")).toHaveLength(2);
    expect(screen.getByText("Disabled")).toBeInTheDocument();
    expect(screen.getByText("No account")).toBeInTheDocument();
    expect(screen.getByText("Not assigned (no account)")).toBeInTheDocument();
    expect(screen.getByText("Managed through officer transition.")).toBeInTheDocument();
    expect(screen.getByText("No Login")).toBeInTheDocument();
    expect(screen.getByText("nologin")).toBeInTheDocument();
    expect(screen.getByText("No account").closest("tr")).toHaveTextContent("No Login");
    expect(screen.getByRole("button", { name: /generate replacement link/i })).toBeInTheDocument();
    expect(container.querySelector('[data-qa="executive-team-row-3"]')).not.toHaveTextContent(
      "Generate replacement link",
    );
  });

  it.each([
    ["past", { startDate: "2020-01-01", endDate: "2020-12-31" }],
    ["unknown", { startDate: "2026-09-01", endDate: "invalid" }],
    ["reversed", { startDate: "2026-12-01", endDate: "2026-09-01" }],
  ])("does not offer activation actions for %s semester dates", (_label, dates) => {
    (useCurrentSemester as jest.Mock).mockReturnValue({
      currentSemester: { ...semester, ...dates },
      loading: false,
    });
    (useExecutiveTeam as jest.Mock).mockReturnValue(
      queryResult({
        data: { members: [member({ accountStatus: "pending_activation", activationOutstanding: true })] },
      }),
    );

    render(<ExecutiveTeamRoster />);

    expect(screen.getByText(/read-only/i, { selector: '[data-qa="executive-team-read-only"]' })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /generate replacement link/i })).not.toBeInTheDocument();
    expect(createExecutiveTeamActivationLink).not.toHaveBeenCalled();
  });

  it("generates a replacement link on request and keeps it available if clipboard copy fails", async () => {
    const user = userEvent.setup();
    (useExecutiveTeam as jest.Mock).mockReturnValue(
      queryResult({
        data: { members: [member({ accountStatus: "pending_activation", activationOutstanding: true })] },
      }),
    );
    (createExecutiveTeamActivationLink as jest.Mock).mockResolvedValue({ activationToken: "one-time-token" });
    Object.defineProperty(navigator, "clipboard", {
      configurable: true,
      value: { writeText: jest.fn().mockRejectedValue(new Error("clipboard denied")) },
    });

    render(<ExecutiveTeamRoster />);
    await user.click(screen.getByRole("button", { name: "Generate replacement link" }));

    const link = await screen.findByLabelText("Replacement activation link for Ada Lovelace");
    expect(link).toHaveValue("http://localhost/activate#token=one-time-token");
    await user.click(screen.getByRole("button", { name: "Copy link" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Select and copy the link manually.");
    expect(link).toHaveValue("http://localhost/activate#token=one-time-token");
    expect(createExecutiveTeamActivationLink).toHaveBeenCalledTimes(1);
  });

  it("shows generation errors without a link and retries roster loading", async () => {
    const user = userEvent.setup();
    const refetch = jest.fn();
    (useExecutiveTeam as jest.Mock).mockReturnValue(
      queryResult({ isError: true, error: new Error("service unavailable"), refetch }),
    );

    const { rerender } = render(<ExecutiveTeamRoster />);
    expect(screen.getByText("service unavailable")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Retry" }));
    expect(refetch).toHaveBeenCalledTimes(1);

    (useExecutiveTeam as jest.Mock).mockReturnValue(
      queryResult({ data: { members: [member({ accountStatus: "pending_activation" })] } }),
    );
    (createExecutiveTeamActivationLink as jest.Mock).mockRejectedValue(new Error("replacement unavailable"));
    rerender(<ExecutiveTeamRoster />);
    await user.click(screen.getByRole("button", { name: "Generate replacement link" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("replacement unavailable");
    expect(screen.queryByLabelText(/replacement activation link/i)).not.toBeInTheDocument();
  });

  it("shows loading, no-semester, and permission states without roster data", () => {
    (useExecutiveTeam as jest.Mock).mockReturnValue(queryResult({ isPending: true }));
    const { rerender } = render(<ExecutiveTeamRoster />);
    expect(screen.getByRole("status")).toHaveTextContent("Loading executive roster");

    (useCurrentSemester as jest.Mock).mockReturnValue({ currentSemester: null, loading: false });
    rerender(<ExecutiveTeamRoster />);
    expect(screen.getByText(/please select a semester/i)).toBeInTheDocument();

    (useAuth as jest.Mock).mockReturnValue({ hasPermission: () => false });
    rerender(<ExecutiveTeamRoster />);
    expect(screen.getByText(/do not have permission/i)).toBeInTheDocument();
  });

  it("keeps the roster query disabled without executive-team GET permission", () => {
    setup({ canView: false });

    render(<ExecutiveTeamRoster />);

    expect(useExecutiveTeam).toHaveBeenCalledWith(semester.id, false);
    expect(screen.getByText(/do not have permission/i)).toBeInTheDocument();
  });
});
