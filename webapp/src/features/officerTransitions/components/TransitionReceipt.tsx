import { useMemo } from "react";
import { Button, Modal, useToast } from "@uwpokerclub/components";
import { OFFICER_ROLES, type StageOfficerTransitionResponse } from "../types";
import styles from "./OfficerTransition.module.css";

export interface TransitionReceiptProps {
  receipt: StageOfficerTransitionResponse | null;
  onClose: () => void;
}

export function TransitionReceipt({ receipt, onClose }: TransitionReceiptProps) {
  const { showToast } = useToast();
  const links = useMemo(
    () =>
      receipt
        ? OFFICER_ROLES.flatMap((role) => {
            const token = receipt.activationTokens[role];
            const nominee = receipt.transition.nominees[role];
            return token && nominee
              ? [
                  {
                    role,
                    name: `${nominee.firstName} ${nominee.lastName}`,
                    link: `${window.location.origin}/activate#token=${encodeURIComponent(token)}`,
                  },
                ]
              : [];
          })
        : [],
    [receipt],
  );
  if (!receipt) return null;
  const dismiss = () => {
    if (
      window.confirm(
        "Close this receipt? These activation links cannot be recovered here; reissue is the only recovery.",
      )
    )
      onClose();
  };
  const copy = async (link: string) => {
    try {
      await navigator.clipboard.writeText(link);
      showToast({
        message: "Activation link copied.",
        variant: "success",
        duration: 3000,
      });
    } catch {
      showToast({
        message: "Copy failed. Select and copy the link manually.",
        variant: "error",
        duration: 5000,
      });
    }
  };
  return (
    <Modal
      isOpen
      onClose={dismiss}
      title="Officer transition staged"
      size="md"
      footer={
        <div className={styles.footer}>
          <Button onClick={dismiss} data-qa="officer-transition-receipt-close">
            Close receipt
          </Button>
        </div>
      }
      data-qa="officer-transition-receipt"
    >
      <div className={styles.receipt}>
        <p>
          <strong>No access has changed yet.</strong> The previous team keeps working until{" "}
          {receipt.transition.nominees.president?.firstName} activates their account.
        </p>
        <p>Send each link directly to that person. These are per-person secrets.</p>
        {links.map(({ role, name, link }) => (
          <div key={role} className={styles.linkRow}>
            <label htmlFor={`activation-link-${role}`}>{name}</label>
            <input
              id={`activation-link-${role}`}
              data-qa={`officer-transition-link-${role}`}
              readOnly
              value={link}
              aria-label={`${name} activation link`}
            />
            <Button variant="secondary" onClick={() => void copy(link)} data-qa={`officer-transition-copy-${role}`}>
              Copy link
            </Button>
          </div>
        ))}
        <p>Closing or reloading this page loses these links. Reissue is the only recovery.</p>
      </div>
    </Modal>
  );
}
