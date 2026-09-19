import { Button } from "@uwpokerclub/components";
import { useState } from "react";
import { useAuth, useCurrentSemester } from "@/hooks";
import { ROLES } from "@/types/roles";
import { useSemesters } from "../../hooks/useSemesterQueries";
import { hasFutureSemester } from "../../utils";
import { SemesterSetupWizard } from "../SemesterSetupWizard";
import { useCompletedIncomingOfficerTransition } from "@/features/officerTransitions/hooks/useOfficerTransitionQueries";
import styles from "./SemesterSetupPrompt.module.css";

export function SemesterSetupPrompt() {
  const { hasRoles } = useAuth();
  const { setCurrentSemester } = useCurrentSemester();
  const { data: semesters = [], isLoading, isError } = useSemesters();
  const [dismissed, setDismissed] = useState(false);
  const [resolved, setResolved] = useState(false);
  const [isWizardOpen, setIsWizardOpen] = useState(false);

  const isPresident = hasRoles([ROLES.PRESIDENT]);
  const completedTransition = useCompletedIncomingOfficerTransition(isPresident);
  const needsSetup = !isLoading && !isError && !hasFutureSemester(semesters, Date.now());
  const isVisible =
    isPresident &&
    !completedTransition.isLoading &&
    !completedTransition.isError &&
    completedTransition.data !== null &&
    needsSetup &&
    !dismissed &&
    !resolved;

  if (!isVisible) return null;

  return (
    <section className={styles.prompt} aria-labelledby="semester-setup-prompt-heading" data-qa="semester-setup-prompt">
      <div>
        <h1 id="semester-setup-prompt-heading" className={styles.heading}>
          Set up your semester
        </h1>
        <p className={styles.copy}>Create the upcoming semester so your club can start running events.</p>
      </div>
      <div className={styles.actions}>
        <Button type="button" onClick={() => setIsWizardOpen(true)} data-qa="semester-setup-prompt-open">
          Set up your semester
        </Button>
        <Button
          type="button"
          variant="tertiary"
          onClick={() => setDismissed(true)}
          data-qa="semester-setup-prompt-dismiss"
        >
          Dismiss
        </Button>
      </div>
      <SemesterSetupWizard
        isOpen={isWizardOpen}
        onClose={() => setIsWizardOpen(false)}
        onSuccess={(semester) => {
          setCurrentSemester(semester);
          setResolved(true);
          setIsWizardOpen(false);
        }}
      />
    </section>
  );
}
