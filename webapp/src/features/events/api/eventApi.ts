import { sendAPIRequest } from "../../../lib/sendAPIRequest";
import { Structure, StructureWithBlinds, Blind } from "../../../types";
import { APIErrorResponse } from "../../../types/error";

/**
 * Event type for API responses
 */
export interface EventResponse {
  id: number;
  name: string;
  format: string;
  notes: string;
  semesterId: string;
  startDate: string;
  state: number;
  rebuys: number;
  pointsMultiplier: number;
  structureId: number;
  structure?: Structure;
}

/**
 * Result type for API operations that may fail
 */
export type ApiResult<T> = { success: true; data: T } | { success: false; error: string };

/**
 * Fetch all existing structures
 * @returns Array of structures or error
 */
export async function fetchStructures(): Promise<ApiResult<Structure[]>> {
  const { status, data } = await sendAPIRequest<{ data: Structure[]; total: number } | APIErrorResponse>(
    "v2/structures",
  );

  if (status >= 200 && status < 300) {
    const listResponse = data as { data: Structure[]; total: number };
    return { success: true, data: listResponse?.data ?? [] };
  }

  const errorResponse = data as APIErrorResponse | undefined;
  return {
    success: false,
    error: errorResponse?.message ?? "Failed to fetch structures",
  };
}

/**
 * Fetch a single structure by ID
 * @param id - Structure ID
 * @returns Structure with blinds or error
 */
export async function fetchStructure(id: number): Promise<ApiResult<StructureWithBlinds>> {
  const { status, data } = await sendAPIRequest<StructureWithBlinds | APIErrorResponse>(`v2/structures/${id}`);

  if (status >= 200 && status < 300) {
    return { success: true, data: data as StructureWithBlinds };
  }

  if (status === 404) {
    return { success: false, error: "Structure not found" };
  }

  const errorResponse = data as APIErrorResponse | undefined;
  return {
    success: false,
    error: errorResponse?.message ?? "Failed to fetch structure",
  };
}

/**
 * Create a new structure with blind levels
 * @param name - Structure name
 * @param blinds - Array of blind levels
 * @returns Created structure or error
 */
export async function createStructure(name: string, blinds: Blind[]): Promise<ApiResult<StructureWithBlinds>> {
  const { status, data } = await sendAPIRequest<StructureWithBlinds | APIErrorResponse>("v2/structures", "POST", {
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
 * Request type for updating a structure
 */
export interface UpdateStructureRequest {
  name?: string;
  blinds?: Blind[];
}

/**
 * Update an existing structure (partial update)
 * @param id - Structure ID
 * @param updates - Partial structure data to update
 * @returns Updated structure or error
 */
export async function updateStructure(
  id: number,
  updates: UpdateStructureRequest,
): Promise<ApiResult<StructureWithBlinds>> {
  const { status, data } = await sendAPIRequest<StructureWithBlinds | APIErrorResponse>(
    `v2/structures/${id}`,
    "PATCH",
    updates as unknown as Record<string, unknown>,
  );

  if (status >= 200 && status < 300) {
    return { success: true, data: data as StructureWithBlinds };
  }

  if (status === 404) {
    return { success: false, error: "Structure not found" };
  }

  const errorResponse = data as APIErrorResponse | undefined;
  return {
    success: false,
    error: errorResponse?.message ?? "Failed to update structure",
  };
}

/**
 * Delete a structure
 * @param id - Structure ID
 * @returns Success or error
 */
export async function deleteStructure(id: number): Promise<ApiResult<void>> {
  const { status, data } = await sendAPIRequest<void | APIErrorResponse>(`v2/structures/${id}`, "DELETE");

  if (status === 204) {
    return { success: true, data: undefined };
  }

  if (status === 404) {
    return { success: false, error: "Structure not found" };
  }

  const errorResponse = data as APIErrorResponse | undefined;
  return {
    success: false,
    error: errorResponse?.message ?? "Failed to delete structure",
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
export async function createEvent(
  semesterId: string,
  eventData: CreateEventRequest,
): Promise<ApiResult<EventResponse>> {
  const payload = {
    ...eventData,
    semesterId,
  };

  const { status, data } = await sendAPIRequest<EventResponse | APIErrorResponse>(
    `v2/semesters/${semesterId}/events`,
    "POST",
    payload as unknown as Record<string, unknown>,
  );

  if (status === 201) {
    return { success: true, data: data as EventResponse };
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

/**
 * Request type for updating an event
 */
export interface UpdateEventRequest {
  name: string;
  format: string;
  notes?: string;
  startDate: string;
  pointsMultiplier: number;
}

/**
 * Fetch a single event by ID
 * @param semesterId - The semester ID
 * @param eventId - The event ID
 * @returns Event or error
 */
export async function fetchEvent(semesterId: string, eventId: number): Promise<ApiResult<EventResponse>> {
  const { status, data } = await sendAPIRequest<EventResponse | APIErrorResponse>(
    `v2/semesters/${semesterId}/events/${eventId}`,
  );

  if (status >= 200 && status < 300) {
    return { success: true, data: data as EventResponse };
  }

  const errorResponse = data as APIErrorResponse | undefined;
  return {
    success: false,
    error: errorResponse?.message ?? "Failed to fetch event",
  };
}

/**
 * Update an existing event
 * @param semesterId - The semester ID
 * @param eventId - The event ID to update
 * @param eventData - Updated event data
 * @returns Updated event or error
 */
export async function updateEvent(
  semesterId: string,
  eventId: number,
  eventData: UpdateEventRequest,
): Promise<ApiResult<EventResponse>> {
  const { status, data } = await sendAPIRequest<EventResponse | APIErrorResponse>(
    `v2/semesters/${semesterId}/events/${eventId}`,
    "PATCH",
    eventData as unknown as Record<string, unknown>,
  );

  if (status >= 200 && status < 300) {
    return { success: true, data: data as EventResponse };
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
    error: errorResponse?.message ?? "Failed to update event",
  };
}
