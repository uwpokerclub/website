import { apiClient } from "../../../lib/apiClient";
import { Semester } from "../../../types";

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
