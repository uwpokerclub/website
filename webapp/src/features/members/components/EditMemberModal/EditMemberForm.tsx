import { useFormContext } from "react-hook-form";
import { FormField, Input, Select } from "@uwpokerclub/components";
import { FACULTIES } from "../../../../data/constants";
import type { EditMemberMembershipFormData } from "../../validation/registrationSchema";
import styles from "./EditMemberForm.module.css";

// Transform FACULTIES array to SelectOption format
const facultyOptions = FACULTIES.map((faculty) => ({
  value: faculty,
  label: faculty,
}));

interface EditMemberFormProps {
  studentId: string;
}

/**
 * EditMemberForm component - Form fields for editing an existing member
 *
 * Student ID is displayed as read-only.
 * Uses react-hook-form context for form state management.
 * Must be wrapped in a FormProvider.
 */
export function EditMemberForm({ studentId }: EditMemberFormProps) {
  const {
    register,
    formState: { errors },
  } = useFormContext<EditMemberMembershipFormData>();

  return (
    <div className={styles.form}>
      <FormField label="Student ID" htmlFor="studentId-display">
        {(props) => (
          <Input
            {...props}
            id="studentId-display"
            data-qa="display-studentId"
            type="text"
            value={studentId}
            disabled
            fullWidth
          />
        )}
      </FormField>

      <FormField label="Quest ID" htmlFor="member.questId">
        {(props) => (
          <Input
            {...props}
            {...register("member.questId")}
            data-qa="input-questId"
            type="text"
            placeholder="e.g., asmahood"
            fullWidth
          />
        )}
      </FormField>

      <FormField label="First Name" htmlFor="member.firstName" required error={errors.member?.firstName?.message}>
        {(props) => (
          <Input
            {...props}
            {...register("member.firstName")}
            data-qa="input-firstName"
            type="text"
            placeholder="e.g., Adam"
            error={!!errors.member?.firstName}
            fullWidth
          />
        )}
      </FormField>

      <FormField label="Last Name" htmlFor="member.lastName" required error={errors.member?.lastName?.message}>
        {(props) => (
          <Input
            {...props}
            {...register("member.lastName")}
            data-qa="input-lastName"
            type="text"
            placeholder="e.g., Mahood"
            error={!!errors.member?.lastName}
            fullWidth
          />
        )}
      </FormField>

      <FormField label="Email" htmlFor="member.email" required error={errors.member?.email?.message}>
        {(props) => (
          <Input
            {...props}
            {...register("member.email")}
            data-qa="input-email"
            type="email"
            placeholder="e.g., asmahood@uwaterloo.ca"
            error={!!errors.member?.email}
            fullWidth
          />
        )}
      </FormField>

      <FormField label="Faculty" htmlFor="member.faculty" required error={errors.member?.faculty?.message}>
        {(props) => (
          <Select
            {...props}
            {...register("member.faculty")}
            data-qa="select-faculty"
            options={facultyOptions}
            placeholder="Select a faculty"
            error={!!errors.member?.faculty}
            fullWidth
          />
        )}
      </FormField>
    </div>
  );
}
