import { sendAPIRequest } from "../../../lib/sendAPIRequest";
import { Semester } from "../../../types";
import { APIErrorResponse } from "../../../types/error";

/**
 * Result type for API operations that may fail
 */
export type ApiResult<T> = { success: true; data: T } | { success: false; error: string };

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

/**
 * Create a new semester
 * @param data - Semester data
 * @returns Created semester or error
 */
export async function createSemester(data: CreateSemesterRequest): Promise<ApiResult<Semester>> {
  const { status, data: responseData } = await sendAPIRequest<Semester | APIErrorResponse>(
    "v2/semesters",
    "POST",
    data as unknown as Record<string, unknown>,
  );

  if (status === 201) {
    return { success: true, data: responseData as Semester };
  }

  if (status === 400) {
    const errorResponse = responseData as APIErrorResponse | undefined;
    return {
      success: false,
      error: errorResponse?.message ?? "Invalid semester data",
    };
  }

  const errorResponse = responseData as APIErrorResponse | undefined;
  return {
    success: false,
    error: errorResponse?.message ?? "Failed to create semester",
  };
}
