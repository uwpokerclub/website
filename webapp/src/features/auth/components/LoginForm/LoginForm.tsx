import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { FormField, Input, Button } from "@uwpokerclub/components";
import { loginSchema, type LoginFormData } from "../../validation/loginSchema";
import styles from "./LoginForm.module.css";

type LoginFormProps = {
  onSubmit: (username: string, password: string) => Promise<void>;
  isSubmitting: boolean;
};

/**
 * LoginForm - Form component for user authentication
 *
 * Uses react-hook-form with Zod validation.
 * Displays inline validation errors via FormField.
 */
export function LoginForm({ onSubmit, isSubmitting }: LoginFormProps) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      username: "",
      password: "",
    },
  });

  const handleFormSubmit = async (data: LoginFormData) => {
    await onSubmit(data.username, data.password);
  };

  return (
    <form onSubmit={handleSubmit(handleFormSubmit)} className={styles.form} noValidate>
      <FormField label="Username" htmlFor="username" required error={errors.username?.message}>
        {(props) => (
          <Input
            {...props}
            {...register("username")}
            name="username"
            type="text"
            placeholder="Enter your username"
            error={!!errors.username}
            fullWidth
            autoComplete="username"
            data-qa="input-username"
          />
        )}
      </FormField>

      <FormField label="Password" htmlFor="password" required error={errors.password?.message}>
        {(props) => (
          <Input
            {...props}
            {...register("password")}
            name="password"
            type="password"
            placeholder="Enter your password"
            error={!!errors.password}
            fullWidth
            autoComplete="current-password"
            data-qa="input-password"
          />
        )}
      </FormField>

      <div className={styles.submitContainer}>
        <Button
          type="submit"
          variant="primary"
          size="large"
          loading={isSubmitting}
          disabled={isSubmitting}
          fullWidth
          data-qa="login-submit"
        >
          {isSubmitting ? "Signing in..." : "Sign In"}
        </Button>
      </div>
    </form>
  );
}
