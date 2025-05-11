import { useAuth } from "@/hooks/useAuth";
import { ReactNode } from "react";
import { Navigate, useLocation } from "react-router-dom";

interface RequireAuthProps {
  children: ReactNode;
}

export default function RequireAuth({ children }: RequireAuthProps) {
  const { user, loading } = useAuth();
  const location = useLocation();

  if (loading) {
    return <>Loading Component</>;
  }

  // User not authenticated
  if (!user) {
    return <Navigate to={"/admin/login"} state={{ from: location }} replace />;
  }

  return <>{children}</>;
}
