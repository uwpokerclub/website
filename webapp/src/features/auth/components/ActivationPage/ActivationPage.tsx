import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { ApiError } from "@/lib/apiClient";
import { fetchSession, sessionKeys } from "@/hooks/useSessionQuery";
import { completeActivation, type ActivationMember, verifyActivation } from "../../api/activationApi";
import styles from "./ActivationPage.module.css";

type PageState = "verifying" | "ready" | "error";

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

  useEffect(() => {
    if (!token) {
      setError(`This activation link is invalid. ${recoveryCopy}`);
      setState("error");
      return;
    }

    void verifyActivation(token)
      .then((verifiedMember) => {
        setMember(verifiedMember);
        setState("ready");
      })
      .catch((requestError: unknown) => {
        setError(activationError(requestError));
        setState("error");
      });
  }, [token]);

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
    setIsSubmitting(true);
    try {
      await completeActivation(token, password);
      await queryClient.fetchQuery({ queryKey: sessionKeys.current(), queryFn: fetchSession });
      navigate("/admin/dashboard", { replace: true });
    } catch (requestError) {
      setError(activationError(requestError));
      setState("error");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <main className={styles.container}>
      <section className={styles.card} aria-live="polite">
        {state === "verifying" && <p data-qa="activation-loading">Verifying your activation link…</p>}
        {state === "error" && (
          <div role="alert" data-qa="activation-error" className={styles.error}>
            {error}
          </div>
        )}
        {state === "ready" && member && (
          <>
            <h1 data-qa="activation-heading">
              Set a password for {member.firstName} {member.lastName}
            </h1>
            <p>Choose a password with at least 8 characters to activate your account.</p>
            <form onSubmit={submit} noValidate className={styles.form}>
              <label htmlFor="activation-password">Password</label>
              <input
                id="activation-password"
                data-qa="activation-password"
                type="password"
                autoComplete="new-password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                disabled={isSubmitting}
              />
              <label htmlFor="activation-confirm-password">Confirm password</label>
              <input
                id="activation-confirm-password"
                data-qa="activation-confirm-password"
                type="password"
                autoComplete="new-password"
                value={confirmation}
                onChange={(event) => setConfirmation(event.target.value)}
                disabled={isSubmitting}
              />
              {passwordError && (
                <p role="alert" data-qa="activation-password-error" className={styles.error}>
                  {passwordError}
                </p>
              )}
              <button type="submit" data-qa="activation-submit" disabled={isSubmitting}>
                {isSubmitting ? "Activating…" : "Activate account"}
              </button>
            </form>
          </>
        )}
      </section>
    </main>
  );
}
