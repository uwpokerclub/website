import { useFormContext } from "react-hook-form";
import { FormField, Input, Select } from "@uwpokerclub/components";
import type { SemesterSetupFormData } from "./schema";
import styles from "./SemesterSetupWizard.module.css";

const termOptions = [
  { value: "fall", label: "Fall" },
  { value: "winter", label: "Winter" },
  { value: "spring", label: "Spring" },
];

export function TermAndDatesStep() {
  const {
    register,
    formState: { errors },
  } = useFormContext<SemesterSetupFormData>();

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
    </section>
  );
}
