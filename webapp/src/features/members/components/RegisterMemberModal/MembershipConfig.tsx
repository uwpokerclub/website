import { useFormContext, useWatch } from "react-hook-form";
import { Checkbox } from "@uwpokerclub/components";
import styles from "./MembershipConfig.module.css";

// Base type for forms with membership configuration
interface FormWithMembership {
  membership: {
    paid: boolean;
    discounted: boolean;
    executive: boolean;
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

  // Watch the executive value to conditionally show paid/discounted
  const isExecutive = useWatch<FormWithMembership>({ name: "membership.executive" });

  // Get membership errors
  const membershipErrors = errors.membership as
    | {
        paid?: { message?: string };
        discounted?: { message?: string };
        executive?: { message?: string };
      }
    | undefined;

  // Handle paid checkbox change - if unchecking paid, also uncheck discounted
  const handlePaidChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const checked = e.target.checked;
    if (!checked) {
      setValue("membership.discounted", false);
    }
  };

  const handleExecutiveChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const checked = e.target.checked;
    if (checked) {
      setValue("membership.paid", false);
      setValue("membership.discounted", false);
    }
  };

  return (
    <div className={styles.container}>
      <h3 className={styles.heading}>Membership Status</h3>
      <p className={styles.description}>This section is only for executive members.</p>

      <div className={styles.checkboxGroup}>
        <Checkbox
          {...register("membership.executive", {
            onChange: handleExecutiveChange,
          })}
          data-qa="checkbox-executive"
          label="Executive Member"
        />

        {!isExecutive && (
          <Checkbox
            {...register("membership.paid", { onChange: handlePaidChange })}
            data-qa="checkbox-paid"
            label="Paid"
          />
        )}

        {!isExecutive && isPaid && (
          <Checkbox {...register("membership.discounted")} data-qa="checkbox-discounted" label="Discounted" />
        )}
      </div>

      {membershipErrors?.executive?.message && (
        <p className={styles.error} data-qa="membership-error-executive">
          {membershipErrors.executive.message}
        </p>
      )}

      {membershipErrors?.discounted?.message && (
        <p className={styles.error} data-qa="membership-error">
          {membershipErrors.discounted.message}
        </p>
      )}
    </div>
  );
}
