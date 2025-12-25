import { sendAPIRequest } from "../../../lib/sendAPIRequest";
import { Event, Structure, StructureWithBlinds, Blind } from "../../../types";
import { APIErrorResponse } from "../../../types/error";

/**
 * Result type for API operations that may fail
 */
export type ApiResult<T> = { success: true; data: T } | { success: false; error: string };

/**
 * Fetch all existing structures
 * @returns Array of structures or error
 */
export async function fetchStructures(): Promise<ApiResult<Structure[]>> {
  const { status, data } = await sendAPIRequest<Structure[] | APIErrorResponse>("structures");

  if (status >= 200 && status < 300) {
    return { success: true, data: (data as Structure[]) ?? [] };
  }

  const errorResponse = data as APIErrorResponse | undefined;
  return {
    success: false,
    error: errorResponse?.message ?? "Failed to fetch structures",
  };
}

/**
 * Create a new structure with blind levels
 * @param name - Structure name
 * @param blinds - Array of blind levels
 * @returns Created structure or error
 */
export async function createStructure(name: string, blinds: Blind[]): Promise<ApiResult<StructureWithBlinds>> {
  const { status, data } = await sendAPIRequest<StructureWithBlinds | APIErrorResponse>("structures", "POST", {
    name,
    blinds,
  });

  if (status === 201) {
    return { success: true, data: data as StructureWithBlinds };
  }

  const errorResponse = data as APIErrorResponse | undefined;
  return {
    success: false,
    error: errorResponse?.message ?? "Failed to create structure",
  };
}

/**
 * Request type for creating an event
 */
export interface CreateEventRequest {
  name: string;
  format: string;
  notes: string;
  startDate: Date;
  structureId: number;
  pointsMultiplier: number;
}

/**
 * Create a new event
 * @param semesterId - The semester ID to create the event in
 * @param eventData - Event data
 * @returns Created event or error
 */
export async function createEvent(semesterId: string, eventData: CreateEventRequest): Promise<ApiResult<Event>> {
  const payload = {
    ...eventData,
    semesterId,
  };

  const { status, data } = await sendAPIRequest<Event | APIErrorResponse>(
    `v2/semesters/${semesterId}/events`,
    "POST",
    payload as unknown as Record<string, unknown>,
  );

  if (status === 201) {
    return { success: true, data: data as Event };
  }

  if (status === 400) {
    const errorResponse = data as APIErrorResponse | undefined;
    return {
      success: false,
      error: errorResponse?.message ?? "Invalid event data",
    };
  }

  const errorResponse = data as APIErrorResponse | undefined;
  return {
    success: false,
    error: errorResponse?.message ?? "Failed to create event",
  };
}
