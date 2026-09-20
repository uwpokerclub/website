import { useAuth } from "@/hooks";
import { OfficerTransitionSection } from "@/features/officerTransitions";
import { Navigate } from "react-router-dom";

export function Executive() {
  const { hasPermission } = useAuth();

  if (!hasPermission("get", "officer-transition")) {
    return <Navigate to="/admin/dashboard" replace />;
  }

  return <OfficerTransitionSection />;
}
