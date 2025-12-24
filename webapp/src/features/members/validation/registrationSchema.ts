import { z } from "zod";
import { FACULTIES } from "../../../data/constants";

// Faculty enum from constants with custom error message
const facultySchema = z.enum(FACULTIES as [string, ...string[]], {
  errorMap: () => ({ message: "Please select a faculty" }),
});

export type Faculty = z.infer<typeof facultySchema>;

// Schema for member search results from API
export const memberSearchResultSchema = z.object({
  id: z.string(),
  firstName: z.string(),
  lastName: z.string(),
  email: z.string(),
  faculty: facultySchema.optional(),
});

export type MemberSearchResult = z.infer<typeof memberSearchResultSchema>;

// Schema for creating a new member
export const createMemberSchema = z.object({
  id: z.string().min(1, "Student ID is required").regex(/^\d+$/, "Student ID must contain only numbers"),
  questId: z.string().optional(),
  firstName: z.string().min(1, "First name is required"),
  lastName: z.string().min(1, "Last name is required"),
  email: z.string().min(1, "Email is required").email("Invalid email format"),
  faculty: facultySchema.refine((val) => val !== "", { message: "Faculty is required" }),
});

export type CreateMemberFormData = z.infer<typeof createMemberSchema>;

// Schema for membership configuration
export const membershipSchema = z
  .object({
    paid: z.boolean(),
    discounted: z.boolean(),
  })
  .refine((data) => !(data.discounted && !data.paid), {
    message: "A membership cannot be discounted if it is not paid",
    path: ["discounted"],
  });

export type MembershipFormData = z.infer<typeof membershipSchema>;

// Combined schema for search mode (selecting existing member + membership config)
export const searchModeSchema = z.object({
  selectedMemberId: z.string().min(1, "Please select a member"),
  membership: membershipSchema,
});

export type SearchModeFormData = z.infer<typeof searchModeSchema>;

// Combined schema for create mode (new member + membership config)
export const createModeSchema = z.object({
  newMember: createMemberSchema,
  membership: membershipSchema,
});

export type CreateModeFormData = z.infer<typeof createModeSchema>;
