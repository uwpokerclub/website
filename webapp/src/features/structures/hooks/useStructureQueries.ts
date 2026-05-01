import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  fetchStructures,
  fetchStructure,
  createStructure,
  updateStructure,
  deleteStructure,
  UpdateStructureRequest,
} from "../api/structuresApi";
import { Blind } from "@/types";

export const structureKeys = {
  all: ["structures"] as const,
  lists: () => [...structureKeys.all, "list"] as const,
  details: () => [...structureKeys.all, "detail"] as const,
  detail: (id: number) => [...structureKeys.details(), id] as const,
};

export function useStructures() {
  return useQuery({
    queryKey: structureKeys.lists(),
    queryFn: fetchStructures,
  });
}

export function useStructure(id: number | undefined) {
  return useQuery({
    queryKey: id ? structureKeys.detail(id) : structureKeys.details(),
    queryFn: () => fetchStructure(id!),
    enabled: !!id,
  });
}

export function useCreateStructure() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ name, blinds }: { name: string; blinds: Blind[] }) => createStructure(name, blinds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: structureKeys.lists() });
    },
  });
}

export function useUpdateStructure() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, updates }: { id: number; updates: UpdateStructureRequest }) => updateStructure(id, updates),
    onSuccess: (_data, { id }) => {
      queryClient.invalidateQueries({ queryKey: structureKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: structureKeys.lists() });
    },
  });
}

export function useDeleteStructure() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => deleteStructure(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: structureKeys.lists() });
    },
  });
}
