import { useState, useCallback } from "react";
import { useForm, FormProvider } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Modal, Button, useToast, Input, Select } from "@uwpokerclub/components";
import { createLoginSchema, type CreateLoginFormData, LOGIN_ROLES } from "../../schemas/loginSchemas";
import { createLogin } from "../../api/loginsApi";
import styles from "./CreateLoginModal.module.css";

export interface CreateLoginModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export function CreateLoginModal({ isOpen, onClose, onSuccess }: CreateLoginModalProps) {
  const { showToast } = useToast();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const form = useForm<CreateLoginFormData>({
    resolver: zodResolver(createLoginSchema),
    defaultValues: {
      username: "",
      password: "",
      role: "" as CreateLoginFormData["role"],
    },
  });

  // Reset form when modal closes
  const handleClose = useCallback(() => {
    form.reset();
    setSubmitError(null);
    onClose();
  }, [form, onClose]);

  // Handle form submission
  const handleSubmit = async (data: CreateLoginFormData) => {
    setIsSubmitting(true);
    setSubmitError(null);

    const result = await createLogin({
      username: data.username,
      password: data.password,
      role: data.role,
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
      message: `Login "${data.username}" created successfully!`,
      variant: "success",
      duration: 3000,
    });

    onSuccess();
    handleClose();
    setIsSubmitting(false);
  };

  // Format role for display
  const formatRole = (role: string) => {
    return role
      .split("_")
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
      .join(" ");
  };

  // Footer with actions
  const footer = (
    <div className={styles.footer}>
      <Button variant="tertiary" onClick={handleClose} disabled={isSubmitting} data-qa="create-login-cancel-btn">
        Cancel
      </Button>
      <Button type="submit" form="create-login-form" disabled={isSubmitting} data-qa="create-login-submit-btn">
        {isSubmitting ? "Creating..." : "Create Login"}
      </Button>
    </div>
  );

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Create Login"
      size="md"
      footer={footer}
      data-qa="create-login-modal"
    >
      <div className={styles.content}>
        {/* Error display */}
        {submitError && (
          <div className={styles.errorAlert} data-qa="create-login-error-alert">
            {submitError}
          </div>
        )}

        <FormProvider {...form}>
          <form id="create-login-form" onSubmit={form.handleSubmit(handleSubmit)} noValidate className={styles.form}>
            {/* Username Field */}
            <div className={styles.field}>
              <label htmlFor="username" className={styles.label}>
                Username / QuestID <span className={styles.required}>*</span>
              </label>
              <Input
                id="username"
                data-qa="input-username"
                type="text"
                placeholder="e.g., jsmith or j9smith"
                {...form.register("username")}
                error={!!form.formState.errors.username}
                errorMessage={form.formState.errors.username?.message}
                fullWidth
              />
              <p className={styles.fieldHint}>Lowercase letters, numbers, and underscores only</p>
            </div>

            {/* Password Field */}
            <div className={styles.field}>
              <label htmlFor="password" className={styles.label}>
                Password <span className={styles.required}>*</span>
              </label>
              <Input
                id="password"
                data-qa="input-password"
                type="password"
                placeholder="Minimum 8 characters"
                {...form.register("password")}
                error={!!form.formState.errors.password}
                errorMessage={form.formState.errors.password?.message}
                fullWidth
              />
            </div>

            {/* Role Field */}
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
                placeholder="Select a role..."
                options={LOGIN_ROLES.map((role) => ({
                  value: role,
                  label: formatRole(role),
                }))}
                fullWidth
              />
            </div>
          </form>
        </FormProvider>
      </div>
    </Modal>
  );
}
