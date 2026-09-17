import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  cancelOfficerTransition,
  fetchCurrentOfficerTransition,
  reissueOfficerTransitionLink,
  stageOfficerTransition,
} from "../api/officerTransitionsApi";
import type { OfficerRole, OfficerTransitionFormData } from "../types";

export const officerTransitionKeys = {
  current: ["officer-transition", "current"] as const,
};

export function useCurrentOfficerTransition(enabled: boolean) {
  return useQuery({
    queryKey: officerTransitionKeys.current,
    queryFn: fetchCurrentOfficerTransition,
    enabled,
    retry: false,
  });
}

export function useStageOfficerTransition() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: OfficerTransitionFormData) => stageOfficerTransition(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: officerTransitionKeys.current }),
  });
}

export function useCancelOfficerTransition() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: cancelOfficerTransition,
    onSuccess: () => queryClient.setQueryData(officerTransitionKeys.current, null),
  });
}

export function useReissueOfficerTransitionLink() {
  return useMutation({
    mutationFn: ({ id, role }: { id: string; role: OfficerRole }) => reissueOfficerTransitionLink(id, role),
  });
}
