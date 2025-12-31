import { z } from "zod";

export const createSemesterSchema = z
  .object({
    name: z.string().min(1, "Semester name is required"),
    startDate: z.string().min(1, "Start date is required"),
    endDate: z.string().min(1, "End date is required"),
    startingBudget: z.number().min(0, "Starting budget must be >= 0"),
    membershipFee: z.number().min(0, "Membership fee must be >= 0"),
    membershipDiscountFee: z.number().min(0, "Discounted membership fee must be >= 0"),
    rebuyFee: z.number().min(0, "Rebuy fee must be >= 0"),
    meta: z.string().optional(),
  })
  .refine((data) => new Date(data.endDate) > new Date(data.startDate), {
    message: "End date must be after start date",
    path: ["endDate"],
  });

export type CreateSemesterFormData = z.infer<typeof createSemesterSchema>;
