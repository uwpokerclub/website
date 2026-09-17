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
        <h1>Executive team</h1>
        <p className={styles.pageSubtitle}>Manage executive team access, roles, and officer transitions.</p>
      </header>

      <div className={styles.pageContent}>
        <section className={styles.transitionSection} aria-labelledby="officer-transition-heading">
          <header className={styles.featureHeader}>
            <h2 id="officer-transition-heading">Officer transition</h2>
            <p>Prepare the next elected team without interrupting the current one.</p>
          </header>
          {currentTransition.data?.status === "pending" ? (
            <PendingTransitionBanner transition={currentTransition.data} />
          ) : canStage ? (
            <div className={styles.transitionCard} data-qa="officer-transition-section">
              <div className={styles.cardCopy}>
                <h3>Finish semester</h3>
                <p>Confirm the four elected officers and securely hand each person their activation link.</p>
              </div>
              <div className={styles.cardAction}>
                <Button onClick={() => setIsModalOpen(true)} data-qa="officer-transition-start">
                  Start officer transition
                </Button>
                <p>Access stays unchanged until the incoming president activates their account.</p>
              </div>
            </div>
          ) : null}
        </section>
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
