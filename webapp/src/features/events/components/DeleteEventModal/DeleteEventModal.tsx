import { useCallback, useState } from "react";
import { Modal, Button, Input } from "@uwpokerclub/components";
import { useDeleteEvent } from "../../hooks/useEventQueries";

import styles from "./DeleteEventModal.module.css";

type DeleteEventModalProps = {
  show: boolean;
  semesterId: string;
  event: { id: number; name: string };
  onClose: () => void;
  onDeleted: () => void;
  onError: (message: string) => void;
};

export function DeleteEventModal({ show, semesterId, event, onClose, onDeleted, onError }: DeleteEventModalProps) {
  const [confirmText, setConfirmText] = useState("");
  const deleteEventMutation = useDeleteEvent();
  const isSubmitting = deleteEventMutation.isPending;
  const isConfirmed = confirmText === event.name;

  const handleClose = useCallback(() => {
    setConfirmText("");
    onClose();
  }, [onClose]);

  const handleSubmit = useCallback(async () => {
    try {
      await deleteEventMutation.mutateAsync({ semesterId, eventId: event.id });
      setConfirmText("");
      onDeleted();
    } catch (err) {
      setConfirmText("");
      onClose();
      onError(err instanceof Error ? err.message : "Failed to delete event");
    }
  }, [semesterId, event.id, onClose, onDeleted, onError, deleteEventMutation]);

  const footer = (
    <div className={styles.footer}>
      <Button variant="tertiary" onClick={handleClose} disabled={isSubmitting} data-qa="delete-event-cancel-btn">
        Cancel
      </Button>
      <Button
        variant="destructive"
        onClick={handleSubmit}
        disabled={isSubmitting || !isConfirmed}
        data-qa="delete-event-confirm-btn"
      >
        {isSubmitting ? "Deleting..." : "Delete Event"}
      </Button>
    </div>
  );

  return (
    <Modal
      isOpen={show}
      onClose={handleClose}
      title="Delete Event"
      size="md"
      footer={footer}
      data-qa="delete-event-modal"
    >
      <p>
        Are you sure you want to delete <strong>&quot;{event.name}&quot;</strong>? This will permanently remove the
        event and all of its entries.
      </p>
      <p>
        <strong>This action cannot be undone.</strong>
      </p>
      <div className={styles.confirmInput}>
        <label className={styles.confirmLabel} htmlFor="delete-confirm-name">
          Type <strong>{event.name}</strong> to confirm
        </label>
        <Input
          id="delete-confirm-name"
          type="text"
          value={confirmText}
          onChange={(e) => setConfirmText(e.target.value)}
          fullWidth
          data-qa="input-delete-confirm-name"
        />
      </div>
    </Modal>
  );
}
