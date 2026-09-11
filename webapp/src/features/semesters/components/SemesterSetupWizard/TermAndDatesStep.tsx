import { useFormContext } from "react-hook-form";
import { FormField, Input, Select } from "@uwpokerclub/components";
import type { SemesterSetupFormData } from "./schema";
import styles from "./SemesterSetupWizard.module.css";

const termOptions = [
  { value: "fall", label: "Fall" },
  { value: "winter", label: "Winter" },
  { value: "spring", label: "Spring" },
];

const canonicalStartMonths = {
  fall: { month: 9, label: "September" },
  winter: { month: 1, label: "January" },
  spring: { month: 5, label: "May" },
} as const;

export function TermAndDatesStep() {
  const {
    register,
    watch,
    formState: { errors },
  } = useFormContext<SemesterSetupFormData>();
  const term = watch("term");
  const startDate = watch("startDate");
  const startMonth = startDate ? new Date(`${startDate}T00:00:00`).getMonth() + 1 : undefined;
  const expectedStart = term ? canonicalStartMonths[term] : undefined;
  const mismatch = expectedStart && startMonth && startMonth !== expectedStart.month;
  const actualMonth = startDate
    ? new Intl.DateTimeFormat("en-US", {
        month: "long",
        timeZone: "UTC",
      }).format(new Date(`${startDate}T00:00:00Z`))
    : "";

  return (
    <section data-qa="semester-wizard-step-term-dates">
      <p className={styles.stepDescription}>Choose the term and its dates.</p>
      <div className={styles.termDatesGrid}>
        <FormField label="Term" htmlFor="term" required error={errors.term?.message}>
          {(props) => (
            <Select
              {...props}
              {...register("term")}
              options={termOptions}
              placeholder="Select a term"
              error={!!errors.term}
              fullWidth
              data-qa="semester-term"
            />
          )}
        </FormField>
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
      </div>
      {mismatch && (
        <p className={styles.warning} role="status" data-qa="semester-term-date-warning">
          {term.charAt(0).toUpperCase() + term.slice(1)} usually starts in {expectedStart.label}. This starts in{" "}
          {actualMonth}.
        </p>
      )}
    </section>
  );
}
