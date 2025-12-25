import { useFormContext } from "react-hook-form";
import { FormField, Input, Select, Textarea } from "@uwpokerclub/components";
import { POKER_FORMATS, type CreateEventFormData } from "../../schemas/eventSchema";
import styles from "./EventDetailsForm.module.css";

// Transform POKER_FORMATS to Select options
const formatOptions = POKER_FORMATS.map((format) => ({
  value: format,
  label: format,
}));

/**
 * EventDetailsForm - Form fields for event details
 *
 * Fields: Name, Start Date, Format, Points Multiplier, Notes
 * Uses react-hook-form context. Must be wrapped in a FormProvider.
 */
export function EventDetailsForm() {
  const {
    register,
    formState: { errors },
  } = useFormContext<CreateEventFormData>();

  return (
    <div className={styles.form}>
      <div className={styles.fullWidth}>
        <FormField label="Event Name" htmlFor="name" required error={errors.name?.message}>
          {(props) => (
            <Input
              {...props}
              {...register("name")}
              type="text"
              placeholder="e.g., Tournament #5"
              error={!!errors.name}
              fullWidth
            />
          )}
        </FormField>
      </div>

      <FormField label="Start Date" htmlFor="startDate" required error={errors.startDate?.message}>
        {(props) => (
          <Input
            {...props}
            {...register("startDate")}
            type="datetime-local"
            error={!!errors.startDate}
            fullWidth
          />
        )}
      </FormField>

      <FormField label="Format" htmlFor="format" required error={errors.format?.message}>
        {(props) => (
          <Select
            {...props}
            {...register("format")}
            options={formatOptions}
            placeholder="Select a format"
            error={!!errors.format}
            fullWidth
          />
        )}
      </FormField>

      <FormField label="Points Multiplier" htmlFor="pointsMultiplier" required error={errors.pointsMultiplier?.message}>
        {(props) => (
          <Input
            {...props}
            {...register("pointsMultiplier", { valueAsNumber: true })}
            type="number"
            min={0}
            step={0.5}
            placeholder="1"
            error={!!errors.pointsMultiplier}
            fullWidth
          />
        )}
      </FormField>

      <div className={styles.fullWidth}>
        <FormField label="Additional Details" htmlFor="notes">
          {(props) => (
            <Textarea
              {...props}
              {...register("notes")}
              placeholder="Optional notes about the event..."
              rows={3}
              fullWidth
            />
          )}
        </FormField>
      </div>
    </div>
  );
}
