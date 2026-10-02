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
  freeTrialLimit: 4,
  comparison: null,
};

describe("TrialConversionCard", () => {
  afterEach(() => jest.clearAllMocks());

  it("leads with the number who spent the trial and never paid", () => {
    mocked.mockReturnValue({ isPending: false, isError: false, data: conversion, refetch: jest.fn() });
    render(<TrialConversionCard semesterId="s1" />);

    expect(screen.getByTestId("trial-conversion-figure")).toHaveTextContent("78");
    expect(screen.getByText(/spent all 4 free entries and never bought a membership/)).toBeInTheDocument();
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

  it("is explicit that it shows a snapshot and not a conversion rate", () => {
    mocked.mockReturnValue({ isPending: false, isError: false, data: conversion, refetch: jest.fn() });
    render(<TrialConversionCard semesterId="s1" />);

    expect(screen.getByText(/not a conversion rate/)).toBeInTheDocument();
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
  it("shows the loading state rather than fabricated figures while the request is in flight", () => {
    mocked.mockReturnValue({ isPending: true, isError: false, data: undefined, refetch: jest.fn() });
    render(<TrialConversionCard semesterId="s1" />);

    expect(screen.getByRole("status")).toBeInTheDocument();
    expect(screen.queryByText("Sample data")).not.toBeInTheDocument();
    expect(screen.queryByTestId("trial-conversion-figure")).not.toBeInTheDocument();
  });
});
