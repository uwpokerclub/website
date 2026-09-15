import { useMutation } from "@tanstack/react-query";
import { stageOfficerTransition } from "../api/officerTransitionsApi";

export function useStageOfficerTransition() {
  return useMutation({ mutationFn: stageOfficerTransition });
}
