/** @jest-environment jsdom */
import { render } from "@testing-library/react";
import "@testing-library/jest-dom";
import { ClockActions } from "./ClockActions";

jest.mock("../../../components", () => ({
  Icon: ({ iconType }: { iconType: string }) => <span data-qa={`icon-${iconType}`} />,
}));

const mockHasPermission = jest.fn();
jest.mock("@/hooks/useAuth", () => ({ useAuth: () => ({ hasPermission: mockHasPermission }) }));

function baseProps(overrides: Partial<React.ComponentProps<typeof ClockActions>> = {}) {
  return {
    isPaused: false,
    onStart: jest.fn(),
    onPause: jest.fn(),
    onStepBack: jest.fn(),
    onStepForward: jest.fn(),
    onSubtractTime: jest.fn(),
    onAddTime: jest.fn(),
    ...overrides,
  };
}

describe("ClockActions", () => {
  beforeEach(() => {
    mockHasPermission.mockReset();
  });

  it("is absent for a role below tournament director", () => {
    mockHasPermission.mockReturnValue(false);
    const { container } = render(<ClockActions {...baseProps()} />);

    expect(container).toBeEmptyDOMElement();
  });

  it("checks the clock control permission, not a generic RequirePermission gate", () => {
    mockHasPermission.mockReturnValue(true);
    render(<ClockActions {...baseProps()} />);

    expect(mockHasPermission).toHaveBeenCalledWith("control", "event", "clock");
  });

  it("is present for a tournament director", () => {
    mockHasPermission.mockReturnValue(true);
    render(<ClockActions {...baseProps()} />);

    expect(document.querySelector('[data-qa="toggle-timer-btn"]')).toBeInTheDocument();
  });
});
