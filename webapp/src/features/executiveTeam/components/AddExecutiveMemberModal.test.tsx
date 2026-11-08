/** @jest-environment jsdom */
/* eslint-disable @typescript-eslint/no-var-requires */
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom";
import type React from "react";
import { AddExecutiveMemberModal } from "./AddExecutiveMemberModal";

jest.mock("@/lib/apiClient", () => {
  class MockApiError extends Error {
    constructor(
      public status: number,
      public type: string,
      message: string,
    ) {
      super(message);
    }
  }
  return { ApiError: MockApiError, apiClient: jest.fn() };
});

jest.mock("@uwpokerclub/components", () => {
  const React = require("react");
  const MockInput = React.forwardRef(
    (
      props: React.InputHTMLAttributes<HTMLInputElement> & Record<string, unknown>,
      ref: React.Ref<HTMLInputElement>,
    ) => {
      const { error, fullWidth, ...inputProps } = props;
      void error;
      void fullWidth;
      return React.createElement("input", { ...inputProps, ref });
    },
  );
  MockInput.displayName = "MockInput";
  return {
    Button: ({ children, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) =>
      React.createElement("button", { type: "button", ...props }, children),
    FormField: ({
      label,
      htmlFor,
      children,
    }: {
      label: string;
      htmlFor?: string;
      children: (props: { id?: string }) => React.ReactNode;
    }) => React.createElement("label", { htmlFor }, label, children({ id: htmlFor })),
    Input: MockInput,
    Select: ({
      options,
      ...props
    }: { options: Array<{ value: string; label: string }> } & React.SelectHTMLAttributes<HTMLSelectElement>) =>
      React.createElement(
        "select",
        props,
        options.map((option) =>
          React.createElement("option", { key: option.value, value: option.value }, option.label),
        ),
      ),
    Modal: ({
      isOpen,
      title,
      children,
      footer,
    }: {
      isOpen: boolean;
      title: string;
      children: React.ReactNode;
      footer: React.ReactNode;
    }) =>
      isOpen ? React.createElement("section", null, React.createElement("h2", null, title), children, footer) : null,
    useToast: () => ({ showToast: jest.fn() }),
  };
});

jest.mock("@/features/members/components/RegisterMemberModal/MemberSearch", () => ({
  MemberSearch: ({ onSelect }: { onSelect: (member: unknown) => void }) => {
    const React = require("react");
    return React.createElement(
      "div",
      null,
      React.createElement(
        "button",
        {
          type: "button",
          "data-qa": "select-test-member",
          onClick: () =>
            onSelect({
              id: "41",
              firstName: "Ada",
              lastName: "Lovelace",
              email: "ada@example.com",
              faculty: "Math",
              questId: "adalovelace",
            }),
        },
        "Select Ada",
      ),
      React.createElement(
        "button",
        {
          type: "button",
          "data-qa": "select-member-missing-quest-id",
          onClick: () =>
            onSelect({
              id: "42",
              firstName: "Grace",
              lastName: "Hopper",
              email: "grace@example.com",
              faculty: "Math",
              questId: "",
            }),
        },
        "Select Grace",
      ),
    );
  },
}));

jest.mock("@/features/members/components/RegisterMemberModal/NewMemberForm", () => ({
  NewMemberForm: () => {
    const React = require("react");
    const { register } = require("react-hook-form").useFormContext();
    return React.createElement(
      "div",
      null,
      React.createElement("input", { ...register("newMember.id"), "data-qa": "new-id" }),
      React.createElement("input", { ...register("newMember.questId"), "data-qa": "new-quest-id" }),
      React.createElement("input", { ...register("newMember.firstName"), "data-qa": "new-first-name" }),
      React.createElement("input", { ...register("newMember.lastName"), "data-qa": "new-last-name" }),
      React.createElement("input", { ...register("newMember.email"), "data-qa": "new-email" }),
      React.createElement("input", { ...register("newMember.faculty"), "data-qa": "new-faculty" }),
    );
  },
}));

jest.mock("@/features/members/api/memberRegistrationApi", () => ({
  createMember: jest.fn(),
  fetchMemberships: jest.fn(),
  updateMember: jest.fn(),
}));
jest.mock("../api/executiveTeamApi", () => ({
  addExecutiveTeamMember: jest.fn(),
  createExecutiveTeamActivationLink: jest.fn(),
}));

import { createMember, fetchMemberships, updateMember } from "@/features/members/api/memberRegistrationApi";
import { addExecutiveTeamMember } from "../api/executiveTeamApi";
import { ApiError } from "@/lib/apiClient";

function getQa(qa: string): HTMLElement {
  const element = document.querySelector<HTMLElement>(`[data-qa="${qa}"]`);
  if (!element) throw new Error(`Missing data-qa element: ${qa}`);
  return element;
}

function waitForQa(qa: string): Promise<HTMLElement> {
  return waitFor(() => getQa(qa));
}

const semester = {
  id: "semester-1",
  name: "Fall 2026",
  startDate: "2026-09-01",
  endDate: "2026-12-31",
  meta: "",
  startingBudget: 100,
  currentBudget: 100,
  membershipFee: 20,
  membershipDiscountFee: 8,
  rebuyFee: 5,
  freeTrialLimit: 0,
};

const pendingResponse = {
  memberId: 41,
  username: "adalovelace",
  role: "executive" as const,
  accountStatus: "pending_activation" as const,
  activationToken: "one-time-token",
};

function setup() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const user = userEvent.setup();
  const view = render(
    <QueryClientProvider client={client}>
      <AddExecutiveMemberModal isOpen semester={semester} onClose={jest.fn()} />
    </QueryClientProvider>,
  );
  return { ...view, client, user };
}

describe("AddExecutiveMemberModal", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (fetchMemberships as jest.Mock).mockResolvedValue({ data: [], total: 0 });
    (updateMember as jest.Mock).mockResolvedValue({});
    (addExecutiveTeamMember as jest.Mock).mockResolvedValue({ ...pendingResponse, activationToken: undefined });
  });

  it("blocks confirmation after membership lookup fails", async () => {
    (fetchMemberships as jest.Mock).mockRejectedValue(new Error("offline"));
    const { user } = setup();
    await user.click(getQa("select-test-member"));
    expect(await waitForQa("executive-add-membership-retry")).toBeInTheDocument();
    expect(getQa("executive-add-review")).toBeDisabled();
    expect(addExecutiveTeamMember).not.toHaveBeenCalled();
  });

  it("shows the discounted refund for a paid membership", async () => {
    (fetchMemberships as jest.Mock).mockResolvedValue({
      data: [{ id: "m-1", paid: true, discounted: true }],
      total: 1,
    });
    const { user } = setup();
    await user.click(getQa("select-test-member"));
    await waitForQa("executive-add-membership-loaded");
    await user.click(getQa("executive-add-review"));
    expect(getQa("executive-add-confirmation")).toHaveTextContent("Refund $8.00");
    expect(screen.getByText(/reduce the semester budget by the same amount/i)).toBeInTheDocument();
  });

  it("saves a newly entered missing QuestID only on confirmation", async () => {
    const { user } = setup();
    await user.click(getQa("select-member-missing-quest-id"));
    await waitForQa("executive-add-membership-loaded");
    await user.type(getQa("executive-add-quest-id"), " Grace.Hopper ");
    expect(updateMember).not.toHaveBeenCalled();
    await user.clear(getQa("executive-add-quest-id"));
    await user.type(getQa("executive-add-quest-id"), " GraceHopper ");
    await user.click(getQa("executive-add-review"));
    expect(updateMember).not.toHaveBeenCalled();
    await user.click(getQa("executive-add-confirm"));
    await waitFor(() => expect(updateMember).toHaveBeenCalledTimes(1));
    expect(updateMember).toHaveBeenCalledWith("42", expect.objectContaining({ questId: "gracehopper" }));
    expect(addExecutiveTeamMember).toHaveBeenCalledWith("semester-1", { questId: "gracehopper", role: "executive" });
  });

  it("rejects invalid newly entered QuestIDs before saving them", async () => {
    const { user } = setup();
    await user.click(getQa("select-member-missing-quest-id"));
    await waitForQa("executive-add-membership-loaded");
    await user.type(getQa("executive-add-quest-id"), "grace.hopper");
    await user.click(getQa("executive-add-review"));
    expect(screen.getByRole("alert")).toHaveTextContent("letters and numbers");
    expect(updateMember).not.toHaveBeenCalled();
    expect(addExecutiveTeamMember).not.toHaveBeenCalled();
  });

  it("creates a member once and retries only AddMember after a partial success", async () => {
    (createMember as jest.Mock).mockResolvedValue({
      id: "51",
      firstName: "Grace",
      lastName: "Hopper",
      email: "grace@example.com",
      faculty: "Math",
      questId: "gracehopper",
      createdAt: "",
    });
    (addExecutiveTeamMember as jest.Mock)
      .mockRejectedValueOnce(new Error("failed"))
      .mockResolvedValueOnce({ memberId: 51, username: "gracehopper", role: "executive", accountStatus: "active" });
    const { user } = setup();
    await user.click(getQa("executive-add-create-mode"));
    await user.type(getQa("new-id"), "51");
    await user.type(getQa("new-quest-id"), " GraceHopper ");
    await user.type(getQa("new-first-name"), "Grace");
    await user.type(getQa("new-last-name"), "Hopper");
    await user.type(getQa("new-email"), "grace@example.com");
    await user.type(getQa("new-faculty"), "Math");
    await user.click(getQa("executive-add-review"));
    await user.click(getQa("executive-add-confirm"));
    await waitForQa("executive-add-error");
    expect(getQa("executive-add-error")).toHaveTextContent(/member was saved, but the team addition failed/i);
    expect(getQa("executive-add-confirm")).toHaveTextContent("Retry team add");
    expect(createMember).toHaveBeenCalledTimes(1);
    await user.click(getQa("executive-add-confirm"));
    await waitForQa("executive-add-success");
    expect(createMember).toHaveBeenCalledTimes(1);
    expect(addExecutiveTeamMember).toHaveBeenCalledTimes(2);
    expect(addExecutiveTeamMember).toHaveBeenCalledWith("semester-1", { questId: "gracehopper", role: "executive" });
  });

  it("shows the pending activation link only when a token is returned", async () => {
    (addExecutiveTeamMember as jest.Mock).mockResolvedValue(pendingResponse);
    const { user } = setup();
    await user.click(getQa("select-test-member"));
    await waitForQa("executive-add-membership-loaded");
    await user.click(getQa("executive-add-review"));
    await user.click(getQa("executive-add-confirm"));
    const link = await waitForQa("executive-add-activation-link");
    expect(link).toHaveValue("http://localhost/activate#token=one-time-token");
    expect(screen.queryByText(/existing account password/i)).not.toBeInTheDocument();
  });

  it("does not invent an activation link when a pending response has no token", async () => {
    const { user } = setup();
    await user.click(getQa("select-test-member"));
    await waitForQa("executive-add-membership-loaded");
    await user.click(getQa("executive-add-review"));
    await user.click(getQa("executive-add-confirm"));
    expect(await waitForQa("executive-add-success")).toHaveTextContent("no activation token was returned");
    expect(document.querySelector('[data-qa="executive-add-activation-link"]')).not.toBeInTheDocument();
  });

  it("reports the preserved password when the API returns an active account", async () => {
    (addExecutiveTeamMember as jest.Mock).mockResolvedValue({
      memberId: 41,
      username: "adalovelace",
      role: "executive",
      accountStatus: "active",
    });
    const { user } = setup();
    await user.click(getQa("select-test-member"));
    await waitForQa("executive-add-membership-loaded");
    await user.click(getQa("executive-add-review"));
    await user.click(getQa("executive-add-confirm"));
    expect(await waitForQa("executive-add-success")).toHaveTextContent("existing account password is unchanged");
    expect(document.querySelector('[data-qa="executive-add-activation-link"]')).not.toBeInTheDocument();
  });

  it("uses safe guidance for a protected-role 403", async () => {
    const { user } = setup();
    (addExecutiveTeamMember as jest.Mock).mockRejectedValue(new ApiError(403, "forbidden", "raw internal detail"));
    await user.click(getQa("select-test-member"));
    await waitForQa("executive-add-membership-loaded");
    await user.click(getQa("executive-add-review"));
    await user.click(getQa("executive-add-confirm"));
    expect(await waitForQa("executive-add-error")).toHaveTextContent("protected role");
    expect(getQa("executive-add-error")).not.toHaveTextContent("raw internal detail");
  });
});
