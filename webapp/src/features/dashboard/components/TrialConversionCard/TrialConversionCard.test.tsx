/** @jest-environment jsdom */
import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom";
import { useTrialConversion } from "../../hooks/useDashboardQueries";
import { TrialConversionCard } from "./TrialConversionCard";
import type { ConversionResponse } from "../../api/dashboardApi";

jest.mock("../../hooks/useDashboardQueries", () => ({ useTrialConversion: jest.fn() }));
const mocked = useTrialConversion as jest.Mock;

const conversion: ConversionResponse = {
  current: { players: 824, paid: 712, trialSpent: 78, trialOpen: 14, executive: 20 },
  conversion: { numerator: 3, denominator: 4, rate: 0.75, untrackedEntrants: 5 },
  freeTrialLimit: 4,
  comparison: null,
};

describe("TrialConversionCard", () => {
  afterEach(() => jest.clearAllMocks());

  it("leads with the number who spent the trial and are currently unpaid", () => {
    mocked.mockReturnValue({ isPending: false, isError: false, data: conversion, refetch: jest.fn() });
    render(<TrialConversionCard semesterId="s1" />);

    expect(screen.getByTestId("trial-conversion-figure")).toHaveTextContent("78");
    expect(screen.getByText(/spent all 4 free entries and are currently unpaid/)).toBeInTheDocument();
  });

  it("breaks the player population into four commitment buckets", () => {
    mocked.mockReturnValue({ isPending: false, isError: false, data: conversion, refetch: jest.fn() });
    render(<TrialConversionCard semesterId="s1" />);

    const buckets = screen.getAllByRole("listitem").map((item) => item.textContent);
    expect(buckets).toEqual(["712 paid", "78 trial spent, unpaid", "14 trial still open", "20 executive, comped"]);
  });

  it("says the trial is off rather than drawing a bar of zeroes", () => {
    mocked.mockReturnValue({
      isPending: false,
      isError: false,
      refetch: jest.fn(),
      data: { ...conversion, freeTrialLimit: 0 },
    });
    render(<TrialConversionCard semesterId="s1" />);

    expect(screen.getByText("Free trial is not enabled this term.")).toBeInTheDocument();
    expect(screen.queryByRole("list")).not.toBeInTheDocument();
  });

  it("keeps the status snapshot separate from the observed tracked-cohort rate", () => {
    mocked.mockReturnValue({ isPending: false, isError: false, data: conversion, refetch: jest.fn() });
    render(<TrialConversionCard semesterId="s1" />);

    expect(screen.getByTestId("trial-conversion-rate-value")).toHaveTextContent("75%");
    expect(screen.getByText(/3 of 4 tracked trial starters/)).toBeInTheDocument();
    expect(screen.getByText(/5 untracked entrants/)).toBeInTheDocument();
    expect(screen.getByText(/not a whole-term rate/)).toBeInTheDocument();
  });

  it("shows the null-rate state even when the current status snapshot is empty", () => {
    mocked.mockReturnValue({
      isPending: false,
      isError: false,
      data: {
        ...conversion,
        current: { players: 0, paid: 0, trialSpent: 0, trialOpen: 0, executive: 0 },
        conversion: { numerator: 0, denominator: 0, rate: null, untrackedEntrants: 0 },
      },
      refetch: jest.fn(),
    });
    render(<TrialConversionCard semesterId="s1" />);

    expect(screen.getByTestId("trial-conversion-rate-value")).toHaveTextContent("Not yet measurable");
    expect(screen.getByText(/0 of 0 tracked trial starters/)).toBeInTheDocument();
  });

  it("shows an endpoint error without a sample-data fallback", () => {
    mocked.mockReturnValue({ isPending: false, isError: true, data: undefined, refetch: jest.fn() });
    render(<TrialConversionCard semesterId="s1" />);

    expect(screen.getByText("Something went wrong.")).toBeInTheDocument();
    expect(screen.queryByText("Sample data")).not.toBeInTheDocument();
    expect(screen.queryByTestId("trial-conversion-figure")).not.toBeInTheDocument();
  });

  it("does not show a sample marker for endpoint data", () => {
    mocked.mockReturnValue({ isPending: false, isError: false, data: conversion, refetch: jest.fn() });
    render(<TrialConversionCard semesterId="s1" />);

    expect(screen.queryByText("Sample data")).not.toBeInTheDocument();
  });

  it("renders an enabled comparison term with its own trial limit and breakdown", () => {
    mocked.mockReturnValue({
      isPending: false,
      isError: false,
      refetch: jest.fn(),
      data: {
        ...conversion,
        comparison: {
          semester: { id: "fall-2025", name: "Fall 2025" },
          freeTrialLimit: 2,
          stats: { players: 10, paid: 4, trialSpent: 3, trialOpen: 2, executive: 1 },
          conversion: { numerator: 2, denominator: 4, rate: 0.5, untrackedEntrants: 3 },
        },
      },
    });
    render(<TrialConversionCard semesterId="s1" />);

    expect(screen.getByText("Fall 2025")).toBeInTheDocument();
    expect(screen.getByText("Free trial limit: 2 entries.")).toBeInTheDocument();
    expect(screen.getByLabelText("Players by membership status in Fall 2025")).toHaveTextContent(
      "4 paid3 trial spent, unpaid2 trial still open1 executive, comped",
    );
    expect(screen.getByText("50%", { selector: "p" })).toBeInTheDocument();
  });

  it("shows prior disabled-trial players as unpaid rather than trial spent", () => {
    mocked.mockReturnValue({
      isPending: false,
      isError: false,
      refetch: jest.fn(),
      data: {
        ...conversion,
        freeTrialLimit: 3,
        comparison: {
          semester: { id: "fall-2025", name: "Fall 2025" },
          freeTrialLimit: 0,
          stats: { players: 10, paid: 4, trialSpent: 5, trialOpen: 0, executive: 1 },
          conversion: { numerator: 0, denominator: 0, rate: null, untrackedEntrants: 4 },
        },
      },
    });
    render(<TrialConversionCard semesterId="s1" />);

    expect(screen.getByText(/Free trial was not enabled/)).toBeInTheDocument();
    expect(screen.getByLabelText("Players by membership status in Fall 2025")).toHaveTextContent("5 unpaid, no trial");
  });

  it("notes when no comparison term is available", () => {
    mocked.mockReturnValue({ isPending: false, isError: false, data: conversion, refetch: jest.fn() });
    render(<TrialConversionCard semesterId="s1" />);

    expect(screen.getByText("No comparable term to compare against yet.")).toBeInTheDocument();
  });
  it("shows the loading state rather than fabricated figures while the request is in flight", () => {
    mocked.mockReturnValue({ isPending: true, isError: false, data: undefined, refetch: jest.fn() });
    render(<TrialConversionCard semesterId="s1" />);

    expect(screen.getByRole("status")).toBeInTheDocument();
    expect(screen.queryByText("Sample data")).not.toBeInTheDocument();
    expect(screen.queryByTestId("trial-conversion-figure")).not.toBeInTheDocument();
  });
});
