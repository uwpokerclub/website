import { useState } from "react";
import { Button } from "@uwpokerclub/components";
import { StageOfficerTransitionModal } from "./StageOfficerTransitionModal";
import { TransitionReceipt } from "./TransitionReceipt";
import type { StageOfficerTransitionResponse } from "../types";
import styles from "./OfficerTransition.module.css";

export function OfficerTransitionSection() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [receipt, setReceipt] = useState<StageOfficerTransitionResponse | null>(null);

  return (
    <section
      className={styles.section}
      aria-labelledby="officer-transition-heading"
      data-qa="officer-transition-section"
    >
      <h1 id="officer-transition-heading">Officer Transition</h1>
      <p>Stage the next elected officer team at the end of the semester.</p>
      <Button onClick={() => setIsModalOpen(true)} data-qa="officer-transition-start">
        Finish Semester
      </Button>
      <StageOfficerTransitionModal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} onStaged={setReceipt} />
      <TransitionReceipt receipt={receipt} onClose={() => setReceipt(null)} />
    </section>
  );
}
