import { useFormContext } from "react-hook-form";
import { FormField, Input, Textarea } from "@uwpokerclub/components";
import type { CreateSemesterFormData } from "./createSemesterSchema";
import styles from "./SemesterDetailsForm.module.css";

/**
 * SemesterDetailsForm - Form fields for semester details
 *
 * Fields: Name, Start Date, End Date, Starting Budget, Membership Fee,
 * Discounted Membership Fee, Rebuy Fee, Additional Details
 * Uses react-hook-form context. Must be wrapped in a FormProvider.
 */
export function SemesterDetailsForm() {
  const {
    register,
    formState: { errors },
  } = useFormContext<CreateSemesterFormData>();

  return (
    <div className={styles.form}>
      <div className={styles.fullWidth}>
        <FormField label="Semester Name" htmlFor="name" required error={errors.name?.message}>
          {(props) => (
            <Input
              {...props}
              {...register("name")}
              type="text"
              placeholder="e.g., Fall 2025"
              error={!!errors.name}
              fullWidth
              data-qa="input-semester-name"
            />
          )}
        </FormField>
      </div>

      <FormField label="Start Date" htmlFor="startDate" required error={errors.startDate?.message}>
        {(props) => (
          <Input
            {...props}
            {...register("startDate")}
            type="date"
            error={!!errors.startDate}
            fullWidth
            data-qa="input-semester-startDate"
          />
        )}
      </FormField>

      <FormField label="End Date" htmlFor="endDate" required error={errors.endDate?.message}>
        {(props) => (
          <Input
            {...props}
            {...register("endDate")}
            type="date"
            error={!!errors.endDate}
            fullWidth
            data-qa="input-semester-endDate"
          />
        )}
      </FormField>

      <FormField label="Starting Budget ($)" htmlFor="startingBudget" required error={errors.startingBudget?.message}>
        {(props) => (
          <Input
            {...props}
            {...register("startingBudget", { valueAsNumber: true })}
            type="number"
            min={0}
            step={0.01}
            placeholder="0.00"
            error={!!errors.startingBudget}
            fullWidth
            data-qa="input-semester-startingBudget"
          />
        )}
      </FormField>

      <FormField label="Membership Fee ($)" htmlFor="membershipFee" required error={errors.membershipFee?.message}>
        {(props) => (
          <Input
            {...props}
            {...register("membershipFee", { valueAsNumber: true })}
            type="number"
            min={0}
            placeholder="10"
            error={!!errors.membershipFee}
            fullWidth
            data-qa="input-semester-membershipFee"
          />
        )}
      </FormField>

      <FormField
        label="Discounted Membership Fee ($)"
        htmlFor="membershipDiscountFee"
        required
        error={errors.membershipDiscountFee?.message}
      >
        {(props) => (
          <Input
            {...props}
            {...register("membershipDiscountFee", { valueAsNumber: true })}
            type="number"
            min={0}
            placeholder="5"
            error={!!errors.membershipDiscountFee}
            fullWidth
            data-qa="input-semester-membershipDiscountFee"
          />
        )}
      </FormField>

      <FormField label="Rebuy Fee ($)" htmlFor="rebuyFee" required error={errors.rebuyFee?.message}>
        {(props) => (
          <Input
            {...props}
            {...register("rebuyFee", { valueAsNumber: true })}
            type="number"
            min={0}
            placeholder="2"
            error={!!errors.rebuyFee}
            fullWidth
            data-qa="input-semester-rebuyFee"
          />
        )}
      </FormField>

      <div className={styles.fullWidth}>
        <FormField label="Additional Details" htmlFor="meta">
          {(props) => (
            <Textarea
              {...props}
              {...register("meta")}
              placeholder="Optional notes about the semester..."
              rows={3}
              fullWidth
              data-qa="input-semester-meta"
            />
          )}
        </FormField>
      </div>
    </div>
  );
}
