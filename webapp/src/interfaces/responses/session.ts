/**
 * @fileoverview This file contains the interfaces for the API responses on the /api/v2/session endpoint.
 */

import { Role } from "@/types/roles";

/**
 * @interface Permissions
 * @description This interface defines all the available permissions for all resources.
 */
interface Permissions {
  create: boolean;
  get: boolean;
  list: boolean;
  edit: boolean;
  delete: boolean;
  end: boolean;
  restart: boolean;
  rebuy: boolean;
  signin: boolean;
  signout: boolean;
  export: boolean;
}

/**
 * @interface PermissionList
 * @description This interface defines the permissions for each resource. To add a new resource, add it to the interface
 * and add any new permissions to the Permissions interface. Then use the Pick/Omit utility types to select which
 * permissions are available for the resource.
 */
export interface PermissionList {
  [key: string]: {
    [key: string]:
      | boolean
      | {
          [key: string]: boolean;
        };
  };
  user: Pick<Permissions, "create" | "get" | "list" | "edit" | "delete">;
  event: Pick<Permissions, "create" | "get" | "list" | "edit" | "end" | "restart" | "rebuy"> & {
    participant: Pick<Permissions, "create" | "get" | "list" | "signin" | "signout" | "delete">;
  };
  login: Pick<Permissions, "create">;
  membership: Pick<Permissions, "create" | "get" | "list" | "edit">;
  semester: Pick<Permissions, "create" | "get" | "list" | "edit"> & {
    rankings: Pick<Permissions, "get" | "list" | "export">;
    transaction: Pick<Permissions, "create" | "get" | "list" | "edit" | "delete">;
  };
  structure: Pick<Permissions, "create" | "get" | "list" | "edit">;
}

export type Resources = keyof PermissionList;

export type Actions = keyof Permissions;

export type SubResources = "participant" | "rankings" | "transaction";

/**
 * @interface UserSession
 * @description This interface defines the response for the GET /api/v2/session endpoint.
 */
export interface UserSession {
  username: string;
  role: Role;
  permissions: PermissionList;
}
