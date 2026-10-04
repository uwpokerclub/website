import { z } from "zod";

export const terms = ["fall", "winter", "spring"] as const;
export type Term = (typeof terms)[number];

export const termAndDatesSchema = z
  .object({
    term: z
      .enum(terms)
      .or(z.literal(""))
      .refine((term) => term !== "", "Select a term"),
    startDate: z.string().min(1, "Start date is required"),
    endDate: z.string().min(1, "End date is required"),
  })
  .refine((data) => !data.startDate || !data.endDate || new Date(data.endDate) > new Date(data.startDate), {
    message: "End date must be after start date",
    path: ["endDate"],
  });

export const feesAndBudgetSchema = z.object({
  startingBudget: z.number().min(0, "Starting budget must be >= 0"),
  membershipFee: z.number().min(0, "Membership fee must be >= 0"),
  membershipDiscountFee: z.number().min(0, "Discounted membership fee must be >= 0"),
  rebuyFee: z.number().min(0, "Rebuy fee must be >= 0"),
  freeTrialLimit: z
    .number()
    .int("Free trial limit must be a whole number")
    .min(0, "Free trial limit must be >= 0")
    .max(255, "Free trial limit must be <= 255"),
  meta: z.string().optional(),
});

export const semesterSetupSchema = termAndDatesSchema.extend(feesAndBudgetSchema.shape);

export type SemesterSetupFormData = z.infer<typeof semesterSetupSchema>;

export function deriveSemesterName(term: Term | "", startDate: string) {
  if (!term || !startDate) return "";

  const year = new Date(`${startDate}T00:00:00`).getFullYear();
  return `${term.charAt(0).toUpperCase()}${term.slice(1)} ${year}`;
}
