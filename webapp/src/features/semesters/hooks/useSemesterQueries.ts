import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { fetchSemesters, createSemester } from "../api/semesterApi";

export const semesterKeys = {
  all: ["semesters"] as const,
  list: () => [...semesterKeys.all, "list"] as const,
};

export function useSemesters() {
  return useQuery({
    queryKey: semesterKeys.list(),
    queryFn: fetchSemesters,
  });
}

export function useCreateSemester() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createSemester,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: semesterKeys.list() });
    },
  });
}
