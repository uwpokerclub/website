import { z } from "zod";
import { FACULTY_VALUES } from "../constants";

export function normalizeQuestIdInput(value?: string): string {
  return (value ?? "").trim().toLowerCase();
}

export function isValidQuestId(value: string): boolean {
  return value === "" || /^[a-z0-9]+$/.test(value);
}

const optionalQuestIdSchema = z
  .string()
  .optional()
  .refine((value) => isValidQuestId(normalizeQuestIdInput(value)), "Quest ID must contain only letters and numbers");

// Faculty enum from constants with custom error message
const facultySchema = z.enum(FACULTY_VALUES, {
  error: () => "Please select a faculty",
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
  questId: optionalQuestIdSchema,
  firstName: z.string().min(1, "First name is required"),
  lastName: z.string().min(1, "Last name is required"),
  email: z.string().min(1, "Email is required").email("Invalid email format"),
  faculty: facultySchema,
});

export const executiveAddMemberSchema = createMemberSchema.extend({
  questId: z
    .string()
    .min(1, "Quest ID is required")
    .refine((value) => isValidQuestId(normalizeQuestIdInput(value)), "Quest ID must contain only letters and numbers"),
});

export type ExecutiveAddMemberFormData = z.infer<typeof executiveAddMemberSchema>;
export const executiveAddMemberFormSchema = z.object({ newMember: executiveAddMemberSchema });
export type ExecutiveAddMemberFormValues = z.infer<typeof executiveAddMemberFormSchema>;

export type CreateMemberFormData = z.infer<typeof createMemberSchema>;

// Schema for membership configuration
export const membershipSchema = z
  .object({
    paid: z.boolean(),
    discounted: z.boolean(),
    executive: z.boolean(),
  })
  .refine((data) => !(data.discounted && !data.paid), {
    message: "A membership cannot be discounted if it is not paid",
    path: ["discounted"],
  })
  .refine((data) => !(data.executive && (data.paid || data.discounted)), {
    message: "An executive membership cannot be paid or discounted",
    path: ["executive"],
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

// Schema for editing a member (no student ID - it's display only)
export const editMemberSchema = z.object({
  // Existing records can contain legacy Quest IDs. Validate changed input in the submit handler,
  // where it can be compared to the original value without blocking unrelated edits.
  questId: z.string().optional(),
  firstName: z.string().min(1, "First name is required"),
  lastName: z.string().min(1, "Last name is required"),
  email: z.string().min(1, "Email is required").email("Invalid email format"),
  faculty: facultySchema,
});

export type EditMemberFormData = z.infer<typeof editMemberSchema>;

// Combined schema for edit mode (member details + membership config)
export const editMemberMembershipSchema = z.object({
  member: editMemberSchema,
  membership: membershipSchema,
});

export type EditMemberMembershipFormData = z.infer<typeof editMemberMembershipSchema>;
