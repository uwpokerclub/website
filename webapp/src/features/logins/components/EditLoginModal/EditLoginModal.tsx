import { useState, useCallback, useEffect } from "react";
import { useForm, FormProvider } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Modal, Button, useToast, Input, Select } from "@uwpokerclub/components";
import { editLoginSchema, type EditLoginFormData, LOGIN_ROLES } from "../../schemas/loginSchemas";
import { useUpdateLogin } from "../../hooks/useLoginQueries";
import { LoginResponse, UpdateLoginRequest } from "../../types";
import styles from "./EditLoginModal.module.css";

export interface EditLoginModalProps {
  isOpen: boolean;
  login: LoginResponse | null;
  onClose: () => void;
  onSuccess?: () => void;
}

const formatRole = (role: string) =>
  role
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");

export function EditLoginModal({ isOpen, login, onClose, onSuccess }: EditLoginModalProps) {
  const { showToast } = useToast();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const updateLoginMutation = useUpdateLogin();
  const isSubmitting = updateLoginMutation.isPending;

  const form = useForm<EditLoginFormData>({
    resolver: zodResolver(editLoginSchema),
    defaultValues: {
      role: login?.role ?? LOGIN_ROLES[0],
      newPassword: "",
      confirmPassword: "",
    },
  });

  useEffect(() => {
    if (isOpen && login) {
      form.reset({
        role: login.role,
        newPassword: "",
        confirmPassword: "",
      });
      setSubmitError(null);
    }
  }, [isOpen, login, form]);

  const handleClose = useCallback(() => {
    form.reset();
    setSubmitError(null);
    onClose();
  }, [form, onClose]);

  const handleSubmit = async (data: EditLoginFormData) => {
    if (!login) return;

    setSubmitError(null);

    const payload: UpdateLoginRequest = {};
    if (data.role !== login.role) {
      payload.role = data.role;
    }
    if (data.newPassword !== "") {
      payload.password = data.newPassword;
    }

    if (payload.role === undefined && payload.password === undefined) {
      setSubmitError("Change the role or enter a new password before saving.");
      return;
    }

    try {
      await updateLoginMutation.mutateAsync({
        username: login.username,
        data: payload,
      });

      showToast({
        message: `Login "${login.username}" updated successfully!`,
        variant: "success",
        duration: 3000,
      });

      onSuccess?.();
      handleClose();
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to update login";
      setSubmitError(message);
      showToast({
        message,
        variant: "error",
        duration: 5000,
      });
    }
  };

  const footer = (
    <div className={styles.footer}>
      <Button variant="tertiary" onClick={handleClose} disabled={isSubmitting} data-qa="edit-login-cancel-btn">
        Cancel
      </Button>
      <Button type="submit" form="edit-login-form" disabled={isSubmitting} data-qa="edit-login-submit-btn">
        {isSubmitting ? "Updating..." : "Update Login"}
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
      title={`Edit Login: ${login.username}`}
      size="md"
      footer={footer}
      data-qa="edit-login-modal"
    >
      <div className={styles.content}>
        {submitError && (
          <div className={styles.errorAlert} data-qa="edit-login-error-alert">
            {submitError}
          </div>
        )}

        <FormProvider {...form}>
          <form id="edit-login-form" onSubmit={form.handleSubmit(handleSubmit)} noValidate className={styles.form}>
            <div className={styles.field}>
              <label htmlFor="role" className={styles.label}>
                Role <span className={styles.required}>*</span>
              </label>
              <Select
                id="role"
                data-qa="select-role"
                {...form.register("role")}
                error={!!form.formState.errors.role}
                errorMessage={form.formState.errors.role?.message}
                options={LOGIN_ROLES.map((role) => ({
                  value: role,
                  label: formatRole(role),
                }))}
                fullWidth
              />
            </div>

            <hr className={styles.sectionDivider} />

            <div className={styles.field}>
              <label htmlFor="newPassword" className={styles.label}>
                New Password
              </label>
              <Input
                id="newPassword"
                data-qa="input-new-password"
                type="password"
                placeholder="Leave blank to keep current password"
                {...form.register("newPassword")}
                error={!!form.formState.errors.newPassword}
                errorMessage={form.formState.errors.newPassword?.message}
                fullWidth
              />
              <p className={styles.fieldHint}>Minimum 8 characters when changed.</p>
            </div>

            <div className={styles.field}>
              <label htmlFor="confirmPassword" className={styles.label}>
                Confirm Password
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
