import { Actions, Resources, SubResources, UserSession } from "@/interfaces/responses";
import { createContext } from "react";
interface IAuthContext {
  user: UserSession | null;
  loading: boolean;
  error: string;
  login: (username: string, password: string, cb: () => void) => Promise<void>;
  logout: (cb: () => void) => Promise<void>;
  hasPermission: (action: Actions, resource: Resources, subResource?: SubResources) => boolean;
}

export const AuthContext = createContext<IAuthContext | null>(null);
