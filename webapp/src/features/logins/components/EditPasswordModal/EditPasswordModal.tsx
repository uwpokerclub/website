import { useState, useCallback, useEffect } from "react";
import { useForm, FormProvider } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Modal, Button, useToast, Input } from "@uwpokerclub/components";
import { editPasswordSchema, type EditPasswordFormData } from "../../schemas/loginSchemas";
import { changePassword } from "../../api/loginsApi";
import { LoginResponse } from "../../types";
import styles from "./EditPasswordModal.module.css";

export interface EditPasswordModalProps {
  isOpen: boolean;
  login: LoginResponse | null;
  onClose: () => void;
  onSuccess: () => void;
}

export function EditPasswordModal({ isOpen, login, onClose, onSuccess }: EditPasswordModalProps) {
  const { showToast } = useToast();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const form = useForm<EditPasswordFormData>({
    resolver: zodResolver(editPasswordSchema),
    defaultValues: {
      newPassword: "",
      confirmPassword: "",
    },
  });

  // Reset form when modal opens with new login
  useEffect(() => {
    if (isOpen) {
      form.reset({
        newPassword: "",
        confirmPassword: "",
      });
      setSubmitError(null);
    }
  }, [isOpen, form]);

  // Reset form when modal closes
  const handleClose = useCallback(() => {
    form.reset();
    setSubmitError(null);
    onClose();
  }, [form, onClose]);

  // Handle form submission
  const handleSubmit = async (data: EditPasswordFormData) => {
    if (!login) return;

    setIsSubmitting(true);
    setSubmitError(null);

    const result = await changePassword(login.username, {
      newPassword: data.newPassword,
    });

    if (!result.success) {
      setSubmitError(result.error);
      showToast({
        message: result.error,
        variant: "error",
        duration: 5000,
      });
      setIsSubmitting(false);
      return;
    }

    // Success!
    showToast({
      message: `Password for "${login.username}" updated successfully!`,
      variant: "success",
      duration: 3000,
    });

    onSuccess();
    handleClose();
    setIsSubmitting(false);
  };

  // Footer with actions
  const footer = (
    <div className={styles.footer}>
      <Button variant="tertiary" onClick={handleClose} disabled={isSubmitting} data-qa="edit-password-cancel-btn">
        Cancel
      </Button>
      <Button type="submit" form="edit-password-form" disabled={isSubmitting} data-qa="edit-password-submit-btn">
        {isSubmitting ? "Updating..." : "Update Password"}
      </Button>
    </div>
  );

  if (!login) {
    return null;
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title={`Edit Password for ${login.username}`}
      size="md"
      footer={footer}
      data-qa="edit-password-modal"
    >
      <div className={styles.content}>
        {/* Error display */}
        {submitError && (
          <div className={styles.errorAlert} data-qa="edit-password-error-alert">
            {submitError}
          </div>
        )}

        <FormProvider {...form}>
          <form id="edit-password-form" onSubmit={form.handleSubmit(handleSubmit)} noValidate className={styles.form}>
            {/* New Password Field */}
            <div className={styles.field}>
              <label htmlFor="newPassword" className={styles.label}>
                New Password <span className={styles.required}>*</span>
              </label>
              <Input
                id="newPassword"
                data-qa="input-new-password"
                type="password"
                placeholder="Minimum 8 characters"
                {...form.register("newPassword")}
                error={!!form.formState.errors.newPassword}
                errorMessage={form.formState.errors.newPassword?.message}
                fullWidth
              />
            </div>

            {/* Confirm Password Field */}
            <div className={styles.field}>
              <label htmlFor="confirmPassword" className={styles.label}>
                Confirm Password <span className={styles.required}>*</span>
              </label>
              <Input
                id="confirmPassword"
                data-qa="input-confirm-password"
                type="password"
                placeholder="Re-enter new password"
                {...form.register("confirmPassword")}
                error={!!form.formState.errors.confirmPassword}
                errorMessage={form.formState.errors.confirmPassword?.message}
                fullWidth
              />
            </div>
          </form>
        </FormProvider>
      </div>
    </Modal>
  );
}
