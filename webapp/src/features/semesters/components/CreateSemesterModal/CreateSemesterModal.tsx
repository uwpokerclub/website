import { useState, useCallback } from "react";
import { useForm, FormProvider } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Modal, Button, useToast } from "@uwpokerclub/components";
import { SemesterDetailsForm } from "./SemesterDetailsForm";
import { createSemesterSchema, type CreateSemesterFormData } from "./createSemesterSchema";
import { useCreateSemester } from "../../hooks/useSemesterQueries";
import type { Semester } from "../../../../types";
import styles from "./CreateSemesterModal.module.css";

export interface CreateSemesterModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (semester: Semester) => void;
}

/**
 * CreateSemesterModal - Modal for creating new semesters
 *
 * Creates a new semester and calls onSuccess with the created semester.
 * The parent component should handle auto-selecting the new semester.
 */
export function CreateSemesterModal({ isOpen, onClose, onSuccess }: CreateSemesterModalProps) {
  const { showToast } = useToast();
  const createSemester = useCreateSemester();

  const [submitError, setSubmitError] = useState<string | null>(null);

  const form = useForm<CreateSemesterFormData>({
    resolver: zodResolver(createSemesterSchema),
    defaultValues: {
      name: "",
      startDate: "",
      endDate: "",
      startingBudget: 0,
      membershipFee: 10,
      membershipDiscountFee: 5,
      rebuyFee: 2,
      meta: "",
    },
  });

  // Reset form when modal closes
  const handleClose = useCallback(() => {
    form.reset();
    setSubmitError(null);
    onClose();
  }, [form, onClose]);

  // Handle form submission
  const handleSubmit = (data: CreateSemesterFormData) => {
    setSubmitError(null);

    createSemester.mutate(
      {
        name: data.name,
        meta: data.meta || "",
        startDate: new Date(data.startDate),
        endDate: new Date(data.endDate),
        startingBudget: data.startingBudget,
        membershipFee: data.membershipFee,
        membershipDiscountFee: data.membershipDiscountFee,
        rebuyFee: data.rebuyFee,
      },
      {
        onSuccess: (semester) => {
          showToast({
            message: `Semester "${data.name}" created successfully!`,
            variant: "success",
            duration: 3000,
          });
          onSuccess(semester);
          handleClose();
        },
        onError: (err) => {
          setSubmitError(err instanceof Error ? err.message : "An unexpected error occurred. Please try again.");
        },
      },
    );
  };

  // Footer with actions
  const footer = (
    <div className={styles.footer}>
      <Button
        variant="tertiary"
        onClick={handleClose}
        disabled={createSemester.isPending}
        data-qa="create-semester-cancel-btn"
      >
        Cancel
      </Button>
      <Button
        type="submit"
        form="create-semester-form"
        disabled={createSemester.isPending}
        data-qa="create-semester-submit-btn"
      >
        {createSemester.isPending ? "Creating..." : "Create Semester"}
      </Button>
    </div>
  );

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Create Semester"
      size="lg"
      footer={footer}
      data-qa="create-semester-modal"
    >
      <div className={styles.content}>
        {/* Error display */}
        {submitError && (
          <div className={styles.errorAlert} data-qa="create-semester-error-alert">
            {submitError}
          </div>
        )}

        <FormProvider {...form}>
          <form id="create-semester-form" onSubmit={form.handleSubmit(handleSubmit)} noValidate>
            <SemesterDetailsForm />
          </form>
        </FormProvider>
      </div>
    </Modal>
  );
}
