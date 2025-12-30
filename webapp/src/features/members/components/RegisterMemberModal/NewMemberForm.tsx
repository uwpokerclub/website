import { useFormContext } from "react-hook-form";
import { FormField, Input, Select } from "@uwpokerclub/components";
import { FACULTIES } from "../../../../data/constants";
import type { CreateModeFormData } from "../../validation/registrationSchema";
import styles from "./NewMemberForm.module.css";

// Transform FACULTIES array to SelectOption format
const facultyOptions = FACULTIES.map((faculty) => ({
  value: faculty,
  label: faculty,
}));

/**
 * NewMemberForm component - Form fields for creating a new member
 *
 * Uses react-hook-form context for form state management.
 * Must be wrapped in a FormProvider.
 */
export function NewMemberForm() {
  const {
    register,
    formState: { errors },
  } = useFormContext<CreateModeFormData>();

  return (
    <div className={styles.form}>
      <FormField label="Student ID" htmlFor="newMember.id" required error={errors.newMember?.id?.message}>
        {(props) => (
          <Input
            {...props}
            {...register("newMember.id")}
            data-qa="input-studentId"
            type="text"
            placeholder="e.g., 20780648"
            error={!!errors.newMember?.id}
            fullWidth
          />
        )}
      </FormField>

      <FormField label="Quest ID" htmlFor="newMember.questId">
        {(props) => (
          <Input
            {...props}
            {...register("newMember.questId")}
            data-qa="input-questId"
            type="text"
            placeholder="e.g., asmahood"
            fullWidth
          />
        )}
      </FormField>

      <FormField label="First Name" htmlFor="newMember.firstName" required error={errors.newMember?.firstName?.message}>
        {(props) => (
          <Input
            {...props}
            {...register("newMember.firstName")}
            data-qa="input-firstName"
            type="text"
            placeholder="e.g., Adam"
            error={!!errors.newMember?.firstName}
            fullWidth
          />
        )}
      </FormField>

      <FormField label="Last Name" htmlFor="newMember.lastName" required error={errors.newMember?.lastName?.message}>
        {(props) => (
          <Input
            {...props}
            {...register("newMember.lastName")}
            data-qa="input-lastName"
            type="text"
            placeholder="e.g., Mahood"
            error={!!errors.newMember?.lastName}
            fullWidth
          />
        )}
      </FormField>

      <FormField label="Email" htmlFor="newMember.email" required error={errors.newMember?.email?.message}>
        {(props) => (
          <Input
            {...props}
            {...register("newMember.email")}
            data-qa="input-email"
            type="email"
            placeholder="e.g., asmahood@uwaterloo.ca"
            error={!!errors.newMember?.email}
            fullWidth
          />
        )}
      </FormField>

      <FormField label="Faculty" htmlFor="newMember.faculty" required error={errors.newMember?.faculty?.message}>
        {(props) => (
          <Select
            {...props}
            {...register("newMember.faculty")}
            data-qa="select-faculty"
            options={facultyOptions}
            placeholder="Select a faculty"
            error={!!errors.newMember?.faculty}
            fullWidth
          />
        )}
      </FormField>
    </div>
  );
}
