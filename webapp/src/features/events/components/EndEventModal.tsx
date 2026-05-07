import { Modal, Button } from "@uwpokerclub/components";
import { useCallback, useState } from "react";
import { useEndEvent } from "../hooks/useEventQueries";

import styles from "./EndEventModal.module.css";

type EndEventModalProps = {
  show: boolean;
  semesterId: string;
  eventId: number;
  onClose: () => void;
  onSuccess: () => void;
};

export function EndEventModal({ show, semesterId, eventId, onClose, onSuccess }: EndEventModalProps) {
  const [error, setError] = useState("");
  const endEventMutation = useEndEvent();
  const isSubmitting = endEventMutation.isPending;

  const handleSubmit = useCallback(async () => {
    setError("");

    try {
      await endEventMutation.mutateAsync({ semesterId, eventId });
      onSuccess();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to end event");
    }
  }, [semesterId, eventId, onClose, onSuccess, endEventMutation]);

  const handleClose = useCallback(() => {
    setError("");
    onClose();
  }, [onClose]);

  const footer = (
    <div className={styles.footer}>
      <Button variant="tertiary" onClick={handleClose} disabled={isSubmitting} data-qa="end-event-cancel-btn">
        Go back
      </Button>
      <Button variant="destructive" onClick={handleSubmit} disabled={isSubmitting} data-qa="end-event-confirm-btn">
        {isSubmitting ? "Ending..." : "End event"}
      </Button>
    </div>
  );

  return (
    <Modal
      isOpen={show}
      onClose={handleClose}
      title="Are you sure you want to end the event?"
      size="md"
      footer={footer}
      data-qa="end-event-modal"
    >
      {error && (
        <div className={styles.errorAlert} role="alert" data-qa="end-event-error-alert">
          {error}
        </div>
      )}
      <p>
        Ending the event will calculate the amount of points each player will get and update the rankings. This action
        should only be taken once the tournament has ended and the final standings have been confirmed.
      </p>
      <p>
        <strong>
          This action cannot be undone. If you are sure you want to end the event, click the &quot;End event&quot;
          button.
        </strong>
      </p>
    </Modal>
  );
}
