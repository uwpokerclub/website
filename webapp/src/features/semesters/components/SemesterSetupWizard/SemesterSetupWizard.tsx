import { useCallback, useState } from "react";
import { FormProvider, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button, Modal, useToast } from "@uwpokerclub/components";
import type { Semester } from "../../../../types";
import { useCreateSemester } from "../../hooks/useSemesterQueries";
import { FeesAndBudgetStep } from "./FeesAndBudgetStep";
import { ReviewStep } from "./ReviewStep";
import { deriveSemesterName, semesterSetupSchema, type SemesterSetupFormData } from "./schema";
import { TermAndDatesStep } from "./TermAndDatesStep";
import styles from "./SemesterSetupWizard.module.css";

export interface SemesterSetupWizardProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (semester: Semester) => void;
}

const steps = ["Term & Dates", "Fees & Budget", "Review"];

export function SemesterSetupWizard({ isOpen, onClose, onSuccess }: SemesterSetupWizardProps) {
  const { showToast } = useToast();
  const createSemester = useCreateSemester();
  const [step, setStep] = useState(0);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const form = useForm<SemesterSetupFormData>({
    resolver: zodResolver(semesterSetupSchema),
    defaultValues: {
      term: "",
      startDate: "",
      endDate: "",
      startingBudget: 0,
      membershipFee: 10,
      membershipDiscountFee: 5,
      rebuyFee: 2,
      freeTrialLimit: 0,
      meta: "",
    },
  });
  const values = form.watch();
  const name = deriveSemesterName(values.term, values.startDate);

  const handleClose = useCallback(() => {
    form.reset();
    setStep(0);
    setSubmitError(null);
    onClose();
  }, [form, onClose]);

  const handleNext = async () => {
    const fields: (keyof SemesterSetupFormData)[] =
      step === 0
        ? ["term", "startDate", "endDate"]
        : ["startingBudget", "membershipFee", "membershipDiscountFee", "rebuyFee", "freeTrialLimit"];
    if (await form.trigger(fields)) setStep((currentStep) => currentStep + 1);
  };

  const handleSubmit = (data: SemesterSetupFormData) => {
    setSubmitError(null);
    const semesterName = deriveSemesterName(data.term, data.startDate);
    createSemester.mutate(
      {
        name: semesterName,
        meta: data.meta || "",
        startDate: new Date(data.startDate),
        endDate: new Date(data.endDate),
        startingBudget: data.startingBudget,
        membershipFee: data.membershipFee,
        membershipDiscountFee: data.membershipDiscountFee,
        rebuyFee: data.rebuyFee,
        freeTrialLimit: data.freeTrialLimit,
      },
      {
        onSuccess: (semester) => {
          showToast({
            message: `Semester "${semesterName}" created successfully!`,
            variant: "success",
            duration: 3000,
          });
          onSuccess(semester);
          handleClose();
        },
        onError: (error) =>
          setSubmitError(error instanceof Error ? error.message : "An unexpected error occurred. Please try again."),
      },
    );
  };

  const footer = (
    <div className={styles.footer}>
      <Button
        type="button"
        variant="tertiary"
        onClick={handleClose}
        disabled={createSemester.isPending}
        data-qa="semester-wizard-cancel-btn"
      >
        Cancel
      </Button>
      {step > 0 && (
        <Button
          type="button"
          variant="tertiary"
          onClick={() => setStep((currentStep) => currentStep - 1)}
          disabled={createSemester.isPending}
          data-qa="semester-wizard-back-btn"
        >
          Back
        </Button>
      )}
      {step < steps.length - 1 ? (
        <Button type="button" onClick={handleNext} data-qa="semester-wizard-next-btn">
          Continue
        </Button>
      ) : (
        <Button
          type="button"
          onClick={form.handleSubmit(handleSubmit)}
          disabled={createSemester.isPending}
          data-qa="create-semester-submit-btn"
        >
          {createSemester.isPending ? "Creating..." : "Create Semester"}
        </Button>
      )}
    </div>
  );

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Set Up Semester"
      size="lg"
      footer={footer}
      data-qa="semester-setup-wizard"
    >
      <div className={styles.content}>
        <ol className={styles.steps} aria-label="Semester setup progress">
          {steps.map((label, index) => {
            const state = index < step ? "completed" : index === step ? "current" : "upcoming";

            return (
              <li
                key={label}
                className={`${styles.step} ${styles[state]}`}
                aria-current={state === "current" ? "step" : undefined}
              >
                <span className={styles.stepNumber}>{index + 1}</span>
                <span>{label}</span>
              </li>
            );
          })}
        </ol>
        {submitError && (
          <div className={styles.errorAlert} data-qa="create-semester-error-alert">
            {submitError}
          </div>
        )}
        <FormProvider {...form}>
          <form id="semester-setup-form" onSubmit={(event) => event.preventDefault()} noValidate>
            {step === 0 && <TermAndDatesStep />}
            {step === 1 && <FeesAndBudgetStep />}
            {step === 2 && <ReviewStep data={values} name={name} />}
          </form>
        </FormProvider>
      </div>
    </Modal>
  );
}
