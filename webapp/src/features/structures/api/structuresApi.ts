import { apiClient } from "@/lib/apiClient";
import { Structure, StructureWithBlinds, Blind } from "@/types";

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
