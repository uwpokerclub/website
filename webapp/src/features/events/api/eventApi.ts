import { apiClient } from "../../../lib/apiClient";
import { Structure, StructureWithBlinds, Blind } from "../../../types";

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

export async function fetchStructures(): Promise<Structure[]> {
  const response = await apiClient<{ data: Structure[]; total: number }>("v2/structures");
  return response.data ?? [];
}

export async function fetchStructure(id: number): Promise<StructureWithBlinds> {
  return apiClient<StructureWithBlinds>(`v2/structures/${id}`);
}

export async function createStructure(name: string, blinds: Blind[]): Promise<StructureWithBlinds> {
  return apiClient<StructureWithBlinds>("v2/structures", {
    method: "POST",
    body: { name, blinds },
  });
}

/**
 * Request type for updating a structure
 */
export interface UpdateStructureRequest {
  name?: string;
  blinds?: Blind[];
}

export async function updateStructure(id: number, updates: UpdateStructureRequest): Promise<StructureWithBlinds> {
  return apiClient<StructureWithBlinds>(`v2/structures/${id}`, {
    method: "PATCH",
    body: updates,
  });
}

export async function deleteStructure(id: number): Promise<void> {
  return apiClient<void>(`v2/structures/${id}`, { method: "DELETE" });
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

export async function createEvent(semesterId: string, eventData: CreateEventRequest): Promise<EventResponse> {
  return apiClient<EventResponse>(`v2/semesters/${semesterId}/events`, {
    method: "POST",
    body: { ...eventData, semesterId },
  });
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

export async function fetchEvent(semesterId: string, eventId: number): Promise<EventResponse> {
  return apiClient<EventResponse>(`v2/semesters/${semesterId}/events/${eventId}`);
}

export async function updateEvent(
  semesterId: string,
  eventId: number,
  eventData: UpdateEventRequest,
): Promise<EventResponse> {
  return apiClient<EventResponse>(`v2/semesters/${semesterId}/events/${eventId}`, {
    method: "PATCH",
    body: eventData,
  });
}
