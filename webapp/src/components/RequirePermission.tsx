import { useAuth } from "@/hooks/useAuth";
import { Actions, Resources, SubResources } from "@/interfaces/responses";
import { ReactNode } from "react";

interface RequirePermissionProps {
  resource: Resources;
  subResource?: SubResources;
  action: Actions;
  children: ReactNode;
}

export default function RequirePermission({ resource, subResource, action, children }: RequirePermissionProps) {
  const { hasPermission } = useAuth();

  if (!hasPermission(action, resource, subResource)) {
    return <>Unauthorized Component</>;
  }

  return <>{children}</>;
}
