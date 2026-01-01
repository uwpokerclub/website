import { useFormContext, useWatch } from "react-hook-form";
import { Checkbox } from "@uwpokerclub/components";
import styles from "./MembershipConfig.module.css";

// Base type for forms with membership configuration
interface FormWithMembership {
  membership: {
    paid: boolean;
    discounted: boolean;
  };
}

/**
 * MembershipConfig component - Checkbox configuration for membership status
 *
 * Handles the paid/discounted logic where discounted is only shown if paid is true.
 * Uses react-hook-form context for form state management.
 * Must be wrapped in a FormProvider.
 */
export function MembershipConfig() {
  const {
    register,
    setValue,
    formState: { errors },
  } = useFormContext<FormWithMembership>();

  // Watch the paid value to conditionally show discounted
  const isPaid = useWatch<FormWithMembership>({ name: "membership.paid" });

  // Get membership errors
  const membershipErrors = errors.membership as
    | {
        paid?: { message?: string };
        discounted?: { message?: string };
      }
    | undefined;

  // Handle paid checkbox change - if unchecking paid, also uncheck discounted
  const handlePaidChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const checked = e.target.checked;
    if (!checked) {
      setValue("membership.discounted", false);
    }
  };

  return (
    <div className={styles.container}>
      <h3 className={styles.heading}>Membership Status</h3>
      <p className={styles.description}>This section is only for executive members.</p>

      <div className={styles.checkboxGroup}>
        <Checkbox
          {...register("membership.paid", { onChange: handlePaidChange })}
          data-qa="checkbox-paid"
          label="Paid"
        />

        {isPaid && <Checkbox {...register("membership.discounted")} data-qa="checkbox-discounted" label="Discounted" />}
      </div>

      {membershipErrors?.discounted?.message && (
        <p className={styles.error} data-qa="membership-error">
          {membershipErrors.discounted.message}
        </p>
      )}
    </div>
  );
}
