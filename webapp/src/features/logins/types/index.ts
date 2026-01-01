import { Role } from "@/types/roles";

/**
 * Linked member information for display
 */
export interface LinkedMember {
  id: number;
  firstName: string;
  lastName: string;
}

/**
 * Login response from API (matches backend LoginWithMember)
 */
export interface LoginResponse {
  username: string;
  role: Role;
  linkedMember: LinkedMember | null;
}

/**
 * Request payload for creating a new login
 */
export interface CreateLoginRequest {
  username: string;
  password: string;
  role: Role;
}

/**
 * Request payload for changing password
 */
export interface ChangePasswordRequest {
  newPassword: string;
}
