import { useFormContext } from "react-hook-form";
import { FormField, Input, Textarea } from "@uwpokerclub/components";
import type { SemesterSetupFormData } from "./schema";
import styles from "./SemesterSetupWizard.module.css";

export function FeesAndBudgetStep() {
  const {
    register,
    formState: { errors },
  } = useFormContext<SemesterSetupFormData>();

  return (
    <section data-qa="semester-wizard-step-fees-budget">
      <p className={styles.stepDescription}>Set the fees and budget for this semester.</p>
      <div className={styles.formGrid}>
        <FormField label="Starting Budget ($)" htmlFor="startingBudget" required error={errors.startingBudget?.message}>
          {(props) => (
            <Input
              {...props}
              {...register("startingBudget", { valueAsNumber: true })}
              type="number"
              min={0}
              step={0.01}
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
              error={!!errors.rebuyFee}
              fullWidth
              data-qa="input-semester-rebuyFee"
            />
          )}
        </FormField>
        <FormField
          label="Free Trial Events"
          htmlFor="freeTrialLimit"
          required
          error={errors.freeTrialLimit?.message}
          hint="Events an unpaid member may attend before their free trial runs out. 0 disables the free trial."
        >
          {(props) => (
            <Input
              {...props}
              {...register("freeTrialLimit", { valueAsNumber: true })}
              type="number"
              min={0}
              max={255}
              step={1}
              error={!!errors.freeTrialLimit}
              fullWidth
              data-qa="input-semester-freeTrialLimit"
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
    </section>
  );
}
