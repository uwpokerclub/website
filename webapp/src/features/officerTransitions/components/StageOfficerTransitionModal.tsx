import { useRef, useState } from "react";
import { Button, FormField, Input, Modal } from "@uwpokerclub/components";
import { ApiError } from "@/lib/apiClient";
import { normalizeQuestId, resolveQuestId } from "../api/officerTransitionsApi";
import { useStageOfficerTransition } from "../hooks/useOfficerTransitionQueries";
import { OFFICER_ROLES, type OfficerRole, type ResolvedQuestId, type StageOfficerTransitionResponse } from "../types";
import styles from "./OfficerTransition.module.css";

const labels: Record<OfficerRole, string> = {
  president: "President",
  vice_president: "Vice President",
  secretary: "Secretary",
  treasurer: "Treasurer",
};

type FieldState = {
  value: string;
  resolved?: ResolvedQuestId;
  error?: string;
  resolving: boolean;
};
type Fields = Record<OfficerRole, FieldState>;

const emptyFields = (): Fields =>
  Object.fromEntries(OFFICER_ROLES.map((role) => [role, { value: "", resolving: false }])) as Fields;

export interface StageOfficerTransitionModalProps {
  isOpen: boolean;
  onClose: () => void;
  onStaged: (receipt: StageOfficerTransitionResponse) => void;
}

export function StageOfficerTransitionModal({ isOpen, onClose, onStaged }: StageOfficerTransitionModalProps) {
  const [fields, setFields] = useState<Fields>(emptyFields);
  const [step, setStep] = useState<"form" | "confirm">("form");
  const [submitError, setSubmitError] = useState("");
  const requests = useRef<Record<OfficerRole, number>>({
    president: 0,
    vice_president: 0,
    secretary: 0,
    treasurer: 0,
  });
  const controllers = useRef<Partial<Record<OfficerRole, AbortController>>>({});
  const stageMutation = useStageOfficerTransition();

  const invalidate = (role: OfficerRole, value: string) => {
    requests.current[role] += 1;
    controllers.current[role]?.abort();
    setFields((current) => ({
      ...current,
      [role]: { value, resolving: false },
    }));
    setSubmitError("");
  };

  const resolve = async (role: OfficerRole) => {
    const value = normalizeQuestId(fields[role].value);
    const request = ++requests.current[role];
    controllers.current[role]?.abort();
    if (!value) {
      setFields((current) => ({
        ...current,
        [role]: {
          value: current[role].value,
          resolving: false,
          error: `${labels[role]} Quest ID is required.`,
        },
      }));
      return;
    }
    const duplicate = OFFICER_ROLES.some((other) => other !== role && normalizeQuestId(fields[other].value) === value);
    if (duplicate) {
      setFields((current) => ({
        ...current,
        [role]: {
          value: current[role].value,
          resolving: false,
          error: "Quest IDs must be distinct.",
        },
      }));
      return;
    }
    const controller = new AbortController();
    controllers.current[role] = controller;
    setFields((current) => ({
      ...current,
      [role]: { value: current[role].value, resolving: true },
    }));
    try {
      const resolved = await resolveQuestId(value, controller.signal);
      if (requests.current[role] !== request) return;
      setFields((current) => ({
        ...current,
        [role]: { value: current[role].value, resolving: false, resolved },
      }));
    } catch (error) {
      if (controller.signal.aborted || requests.current[role] !== request) return;
      const message =
        error instanceof ApiError ? error.message : `${labels[role]} Quest ID must resolve to exactly one member.`;
      setFields((current) => ({
        ...current,
        [role]: {
          value: current[role].value,
          resolving: false,
          error: message,
        },
      }));
    }
  };

  const ready = OFFICER_ROLES.every((role) => Boolean(fields[role].resolved) && !fields[role].resolving);
  const submit = async () => {
    setSubmitError("");
    try {
      const response = await stageMutation.mutateAsync({
        presidentQuestId: fields.president.value,
        vicePresidentQuestId: fields.vice_president.value,
        secretaryQuestId: fields.secretary.value,
        treasurerQuestId: fields.treasurer.value,
      });
      if (!response.activationTokens.president) throw new Error("The president activation link was not returned.");
      onStaged(response);
      onClose();
    } catch (error) {
      setStep("form");
      setSubmitError(error instanceof Error ? error.message : "Unable to stage the officer transition.");
    }
  };

  const footer =
    step === "form" ? (
      <>
        <Button variant="tertiary" onClick={onClose} disabled={stageMutation.isPending}>
          Cancel
        </Button>
        <Button onClick={() => setStep("confirm")} disabled={!ready} data-qa="officer-transition-review">
          Review transition
        </Button>
      </>
    ) : (
      <>
        <Button variant="tertiary" onClick={() => setStep("form")} disabled={stageMutation.isPending}>
          Back
        </Button>
        <Button onClick={() => void submit()} loading={stageMutation.isPending} data-qa="officer-transition-submit">
          Stage transition
        </Button>
      </>
    );

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Finish Semester"
      size="md"
      footer={<div className={styles.footer}>{footer}</div>}
      data-qa="officer-transition-modal"
    >
      {step === "form" ? (
        <div className={styles.form}>
          {OFFICER_ROLES.map((role) => {
            const field = fields[role];
            return (
              <FormField
                key={role}
                label={`${labels[role]} Quest ID`}
                htmlFor={`transition-${role}`}
                required
                error={field.error}
              >
                {(props) => (
                  <>
                    <Input
                      {...props}
                      id={`transition-${role}`}
                      data-qa={`officer-transition-${role}`}
                      value={field.value}
                      onChange={(event) => invalidate(role, event.target.value)}
                      onBlur={() => void resolve(role)}
                      error={Boolean(field.error)}
                      fullWidth
                    />
                    {field.resolving && <p>Resolving…</p>}
                    {field.resolved && (
                      <p data-qa={`officer-transition-${role}-name`}>
                        {field.resolved.nominee.firstName} {field.resolved.nominee.lastName}
                      </p>
                    )}
                  </>
                )}
              </FormField>
            );
          })}
          {submitError && (
            <p role="alert" className={styles.error} data-qa="officer-transition-error">
              {submitError}
            </p>
          )}
        </div>
      ) : (
        <div className={styles.confirmation}>
          <p>Confirm the next officer team:</p>
          <ul>
            {OFFICER_ROLES.map((role) => (
              <li key={role}>
                {labels[role]}: {fields[role].resolved?.nominee.firstName} {fields[role].resolved?.nominee.lastName}
              </li>
            ))}
          </ul>
          <p>
            <strong>No access changes yet.</strong> The previous team keeps working until the incoming president
            activates their account.
          </p>
          {submitError && (
            <p role="alert" className={styles.error}>
              {submitError}
            </p>
          )}
        </div>
      )}
    </Modal>
  );
}
