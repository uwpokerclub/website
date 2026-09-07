import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { Button, FormField, Input } from "@uwpokerclub/components";
import { ApiError } from "@/lib/apiClient";
import { fetchSession, sessionKeys } from "@/hooks/useSessionQuery";
import uwpscLogo from "@/assets/uwpsc_logo.svg";
import { completeActivation, type ActivationMember, verifyActivation } from "../../api/activationApi";
import loginStyles from "../LoginPage/LoginPage.module.css";
import formStyles from "./ActivationPage.module.css";

type PageState = "verifying" | "ready" | "verification-error" | "completion-error" | "signed-in-error";

const recoveryCopy = "Ask whoever sent you this link for a new one.";

function activationError(error: unknown): string {
  if (error instanceof ApiError && error.status === 401) {
    return `This activation link is invalid, expired, or has already been used. ${recoveryCopy}`;
  }
  return "We could not reach the server. Check your connection and try again.";
}

export function ActivationPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [token] = useState(() => {
    const params = new URLSearchParams(window.location.hash.slice(1));
    window.history.replaceState(null, "", `${window.location.pathname}${window.location.search}`);
    return params.get("token") ?? "";
  });
  const [state, setState] = useState<PageState>("verifying");
  const [member, setMember] = useState<ActivationMember | null>(null);
  const [password, setPassword] = useState("");
  const [confirmation, setConfirmation] = useState("");
  const [passwordError, setPasswordError] = useState("");
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const verify = useCallback(() => {
    if (!token) {
      setError(`This activation link is invalid. ${recoveryCopy}`);
      setState("verification-error");
      return;
    }

    setError("");
    setState("verifying");
    void verifyActivation(token)
      .then((verifiedMember) => {
        setMember(verifiedMember);
        setState("ready");
      })
      .catch((requestError: unknown) => {
        setError(activationError(requestError));
        setState("verification-error");
      });
  }, [token]);

  useEffect(() => {
    verify();
  }, [verify]);

  const refreshSessionAndNavigate = useCallback(async () => {
    queryClient.removeQueries({
      predicate: (query) => query.queryKey[0] !== sessionKeys.all[0],
    });
    await queryClient.fetchQuery({ queryKey: sessionKeys.current(), queryFn: fetchSession, staleTime: 0 });
    navigate("/admin/dashboard", { replace: true });
  }, [navigate, queryClient]);

  const retryDashboard = async () => {
    setError("");
    try {
      await refreshSessionAndNavigate();
    } catch {
      setError("Your password has been set, but we still could not sign you in. Check your connection and try again.");
    }
  };

  const submit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (password.length < 8) {
      setPasswordError("Password must be at least 8 characters.");
      return;
    }
    if (password !== confirmation) {
      setPasswordError("Passwords do not match.");
      return;
    }

    setPasswordError("");
    setError("");
    setIsSubmitting(true);
    try {
      await completeActivation(token, password);
      try {
        await refreshSessionAndNavigate();
      } catch {
        setError("Your password has been set, but we could not sign you in. Try again to continue to the dashboard.");
        setState("signed-in-error");
      }
    } catch (requestError) {
      setError(
        `We could not confirm that your account was activated. ${activationError(requestError)} You can try again.`,
      );
      setState("completion-error");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <main className={loginStyles.container}>
      <div className={loginStyles.logoContainer}>
        <img src={uwpscLogo} alt="UWPSC Logo" className={loginStyles.logo} />
      </div>
      <section className={loginStyles.card} aria-live="polite">
        {state === "verifying" && <p data-qa="activation-loading">Verifying your activation link…</p>}
        {(state === "verification-error" || state === "completion-error" || state === "signed-in-error") && (
          <div role="alert" data-qa="activation-error" className={loginStyles.errorAlert}>
            {error}
            {state === "verification-error" && token && (
              <div className={formStyles.retryContainer}>
                <Button type="button" variant="secondary" onClick={verify} data-qa="activation-retry">
                  Try again
                </Button>
              </div>
            )}
            {state === "signed-in-error" && (
              <div className={formStyles.retryContainer}>
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() => void retryDashboard()}
                  data-qa="activation-dashboard-retry"
                >
                  Continue to dashboard
                </Button>
              </div>
            )}
          </div>
        )}
        {(state === "ready" || state === "completion-error") && member && (
          <>
            <div className={loginStyles.header}>
              <h1 data-qa="activation-heading" className={loginStyles.title}>
                Set your password
              </h1>
              <p className={loginStyles.subtitle}>
                You&apos;re activating an account for {member.firstName} {member.lastName}.
              </p>
            </div>
            <form onSubmit={submit} noValidate className={formStyles.form}>
              <FormField label="Password" htmlFor="activation-password" required error={passwordError}>
                {(props) => (
                  <Input
                    {...props}
                    id="activation-password"
                    data-qa="activation-password"
                    type="password"
                    autoComplete="new-password"
                    value={password}
                    onChange={(event) => setPassword(event.target.value)}
                    disabled={isSubmitting}
                    error={Boolean(passwordError)}
                    fullWidth
                  />
                )}
              </FormField>
              <FormField label="Confirm password" htmlFor="activation-confirm-password" required>
                {(props) => (
                  <Input
                    {...props}
                    id="activation-confirm-password"
                    data-qa="activation-confirm-password"
                    type="password"
                    autoComplete="new-password"
                    value={confirmation}
                    onChange={(event) => setConfirmation(event.target.value)}
                    disabled={isSubmitting}
                    error={Boolean(passwordError)}
                    fullWidth
                  />
                )}
              </FormField>
              {passwordError && (
                <p role="alert" data-qa="activation-password-error" className={formStyles.error}>
                  {passwordError}
                </p>
              )}
              <div className={formStyles.submitContainer}>
                <Button
                  type="submit"
                  data-qa="activation-submit"
                  loading={isSubmitting}
                  disabled={isSubmitting}
                  fullWidth
                  size="large"
                >
                  {isSubmitting ? "Activating…" : "Activate account"}
                </Button>
              </div>
            </form>
          </>
        )}
      </section>
      <p className={loginStyles.footer}>University of Waterloo Poker Studies Club</p>
    </main>
  );
}
