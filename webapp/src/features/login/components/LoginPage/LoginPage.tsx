import { useState } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { useAuth } from "@/hooks";
import { LoginForm } from "../LoginForm";
import uwpscLogo from "@/assets/uwpsc_logo.svg";
import styles from "./LoginPage.module.css";

/**
 * LoginPage - Main login page with branding and card layout
 *
 * Features:
 * - UWPSC logo branding
 * - Centered card layout
 * - Error banner for API errors
 * - Redirect after successful login
 */
export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { login, error } = useAuth();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const from = (location.state as { from: { pathname: string } })?.from?.pathname || "/admin";

  const handleLogin = async (username: string, password: string) => {
    setIsSubmitting(true);

    try {
      await login(username, password, () => {
        navigate(from, { replace: true });
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className={styles.container}>
      <div className={styles.logoContainer}>
        <img src={uwpscLogo} alt="UWPSC Logo" className={styles.logo} />
      </div>

      <div className={styles.card}>
        <div className={styles.header}>
          <h1 className={styles.title}>Welcome Back</h1>
          <p className={styles.subtitle}>Sign in to access the executive dashboard</p>
        </div>

        {error && (
          <div data-qa="login-error-banner" role="alert" className={styles.errorAlert}>
            {error}
          </div>
        )}

        <LoginForm onSubmit={handleLogin} isSubmitting={isSubmitting} />
      </div>

      <p className={styles.footer}>University of Waterloo Poker Studies Club</p>
    </div>
  );
}
