import { Actions, Resources, SubResources, UserSession } from "@/interfaces/responses";
import { Role } from "@/types/roles";
import { createContext } from "react";
interface IAuthContext {
  user: UserSession | null;
  loading: boolean;
  error: string;
  login: (username: string, password: string, cb: () => void) => Promise<void>;
  logout: (cb: () => void) => Promise<void>;
  hasPermission: (action: Actions, resource: Resources, subResource?: SubResources) => boolean;
  hasRoles: (roles: Role[]) => boolean;
}

export const AuthContext = createContext<IAuthContext | null>(null);
