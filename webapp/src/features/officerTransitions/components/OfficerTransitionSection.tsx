import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Button } from "@uwpokerclub/components";
import { useAuth } from "@/hooks";
import { StageOfficerTransitionModal } from "./StageOfficerTransitionModal";
import { TransitionReceipt } from "./TransitionReceipt";
import { PendingTransitionBanner } from "./PendingTransitionBanner";
import { officerTransitionKeys, useCurrentOfficerTransition } from "../hooks/useOfficerTransitionQueries";
import type { StageOfficerTransitionResponse } from "../types";
import styles from "./OfficerTransition.module.css";

export function OfficerTransitionSection() {
  const { hasPermission } = useAuth();
  const queryClient = useQueryClient();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [receipt, setReceipt] = useState<StageOfficerTransitionResponse | null>(null);
  const currentTransition = useCurrentOfficerTransition(true);
  const canStage = hasPermission("create", "officer-transition");

  return (
    <div className={styles.page}>
      <header className={styles.pageHeader}>
        <p className={styles.eyebrow}>Executive operations</p>
        <h1 id="officer-transition-heading">Officer Transition</h1>
        <p className={styles.pageSubtitle}>Prepare the next elected team without interrupting the current one.</p>
      </header>

      <div className={styles.pageContent}>
        {currentTransition.data?.status === "pending" ? (
          <PendingTransitionBanner transition={currentTransition.data} />
        ) : canStage ? (
          <section
            className={styles.transitionCard}
            aria-labelledby="officer-transition-heading"
            data-qa="officer-transition-section"
          >
            <div className={styles.cardCopy}>
              <h2>Finish Semester</h2>
              <p>Confirm the four elected officers and securely hand each person their activation link.</p>
            </div>
            <div className={styles.cardAction}>
              <Button onClick={() => setIsModalOpen(true)} data-qa="officer-transition-start">
                Start officer transition
              </Button>
              <p>Access stays unchanged until the incoming president activates their account.</p>
            </div>
          </section>
        ) : null}
      </div>

      <StageOfficerTransitionModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onStaged={(staged) => {
          setReceipt(staged);
          void queryClient.invalidateQueries({ queryKey: officerTransitionKeys.current });
        }}
      />
      <TransitionReceipt receipt={receipt} onClose={() => setReceipt(null)} />
    </div>
  );
}
