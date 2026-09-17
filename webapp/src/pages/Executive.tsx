import { useAuth } from "@/hooks";
import { OfficerTransitionSection } from "@/features/officerTransitions";

export function Executive() {
  const { hasPermission } = useAuth();
  return hasPermission("get", "officer-transition") ? <OfficerTransitionSection /> : null;
}
