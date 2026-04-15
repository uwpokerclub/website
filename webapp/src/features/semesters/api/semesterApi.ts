import { apiClient } from "../../../lib/apiClient";
import { Semester } from "../../../types";

export async function fetchSemesters(): Promise<Semester[]> {
  const response = await apiClient<{ data: Semester[] }>("v2/semesters");
  return response.data ?? [];
}

/**
 * Request type for creating a semester
 */
export interface CreateSemesterRequest {
  name: string;
  meta?: string;
  startDate: Date;
  endDate: Date;
  startingBudget: number;
  membershipFee: number;
  membershipDiscountFee: number;
  rebuyFee: number;
}

export async function createSemester(data: CreateSemesterRequest): Promise<Semester> {
  return apiClient<Semester>("v2/semesters", {
    method: "POST",
    body: data,
  });
}
