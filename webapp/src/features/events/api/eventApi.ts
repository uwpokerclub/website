import { apiClient } from "@/lib/apiClient";
import { Event } from "@/types";

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

export async function createEvent(semesterId: string, eventData: CreateEventRequest): Promise<Event> {
  return apiClient<Event>(`v2/semesters/${semesterId}/events`, {
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

export async function fetchEvent(semesterId: string, eventId: number): Promise<Event> {
  return apiClient<Event>(`v2/semesters/${semesterId}/events/${eventId}`);
}

export async function updateEvent(semesterId: string, eventId: number, eventData: UpdateEventRequest): Promise<Event> {
  return apiClient<Event>(`v2/semesters/${semesterId}/events/${eventId}`, {
    method: "PATCH",
    body: eventData,
  });
}

export async function fetchEvents(
  semesterId: string,
  params: { limit: number; offset: number; search?: string },
): Promise<{ data: Event[]; total: number }> {
  let query = `?limit=${params.limit}&offset=${params.offset}`;
  if (params.search) {
    query += `&search=${encodeURIComponent(params.search)}`;
  }
  const response = await apiClient<{ data: Event[]; total: number }>(`v2/semesters/${semesterId}/events${query}`);
  return { data: response.data ?? [], total: response.total ?? 0 };
}

export async function endEvent(semesterId: string, eventId: number): Promise<void> {
  return apiClient<void>(`v2/semesters/${semesterId}/events/${eventId}/end`, { method: "POST" });
}

export async function restartEvent(semesterId: string, eventId: number): Promise<void> {
  return apiClient<void>(`v2/semesters/${semesterId}/events/${eventId}/restart`, { method: "POST" });
}

export async function rebuyEvent(semesterId: string, eventId: number): Promise<void> {
  return apiClient<void>(`v2/semesters/${semesterId}/events/${eventId}/rebuy`, { method: "POST" });
}
