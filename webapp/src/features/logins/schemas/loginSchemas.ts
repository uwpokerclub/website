import { z } from "zod";
import { ROLES } from "@/types/roles";

/**
 * Available roles for login creation (matches backend oneof validation)
 */
export const LOGIN_ROLES = [
  ROLES.BOT,
  ROLES.EXECUTIVE,
  ROLES.TOURNAMENT_DIRECTOR,
  ROLES.SECRETARY,
  ROLES.TREASURER,
  ROLES.VICE_PRESIDENT,
  ROLES.PRESIDENT,
  ROLES.WEBMASTER,
] as const;

/**
 * Role schema with enum validation
 */
const roleSchema = z.enum(LOGIN_ROLES, {
  error: () => "Please select a role",
});

/**
 * Schema for creating a new login
 */
export const createLoginSchema = z.object({
  username: z
    .string()
    .min(1, "Username/QuestID is required")
    .regex(/^[a-z0-9_]+$/, "Username must be lowercase letters, numbers, or underscores only"),
  password: z
    .string()
    .min(8, "Password must be at least 8 characters")
    .max(128, "Password must not exceed 128 characters"),
  role: roleSchema,
});

export type CreateLoginFormData = z.infer<typeof createLoginSchema>;

/**
 * Schema for editing a login: role plus optional password change.
 * Password fields are optional — leave blank to keep the current password.
 */
export const editLoginSchema = z
  .object({
    role: roleSchema,
    newPassword: z
      .string()
      .max(128, "Password must not exceed 128 characters")
      .refine((v) => v === "" || v.length >= 8, {
        message: "Password must be at least 8 characters",
      }),
    confirmPassword: z.string(),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    message: "Passwords do not match",
    path: ["confirmPassword"],
  });

export type EditLoginFormData = z.infer<typeof editLoginSchema>;
