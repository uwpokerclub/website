import { apiClient } from "@/lib/apiClient";
import { ClockState } from "@/types";

export async function fetchClock(semesterId: string, eventId: number): Promise<ClockState> {
  return apiClient<ClockState>(`v2/semesters/${semesterId}/events/${eventId}/clock`);
}

export async function pauseClock(semesterId: string, eventId: number): Promise<ClockState> {
  return apiClient<ClockState>(`v2/semesters/${semesterId}/events/${eventId}/clock/pause`, { method: "POST" });
}

export async function resumeClock(semesterId: string, eventId: number): Promise<ClockState> {
  return apiClient<ClockState>(`v2/semesters/${semesterId}/events/${eventId}/clock/resume`, { method: "POST" });
}

export async function adjustClock(semesterId: string, eventId: number, deltaSeconds: number): Promise<ClockState> {
  return apiClient<ClockState>(`v2/semesters/${semesterId}/events/${eventId}/clock/adjust`, {
    method: "POST",
    body: { deltaSeconds },
  });
}

export async function setClockLevel(semesterId: string, eventId: number, index: number): Promise<ClockState> {
  return apiClient<ClockState>(`v2/semesters/${semesterId}/events/${eventId}/clock/level`, {
    method: "POST",
    body: { index },
  });
}
