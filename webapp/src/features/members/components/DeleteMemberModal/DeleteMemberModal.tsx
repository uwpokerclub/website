import { Modal, Button, useToast } from "@uwpokerclub/components";
import { useCallback, useState } from "react";
import { useDeleteMember } from "../../hooks/useMemberQueries";
import styles from "./DeleteMemberModal.module.css";

export interface DeleteMemberModalProps {
  isOpen: boolean;
  member: { id: string; firstName: string; lastName: string } | null;
  onClose: () => void;
  onSuccess?: () => void;
}

export function DeleteMemberModal({ isOpen, member, onClose, onSuccess }: DeleteMemberModalProps) {
  const { showToast } = useToast();
  const [error, setError] = useState("");
  const deleteMemberMutation = useDeleteMember();
  const isSubmitting = deleteMemberMutation.isPending;

  const handleSubmit = useCallback(async () => {
    if (!member) return;

    setError("");

    try {
      await deleteMemberMutation.mutateAsync(member.id);

      showToast({
        message: `${member.firstName} ${member.lastName} deleted successfully`,
        variant: "success",
        duration: 3000,
      });

      onSuccess?.();
      onClose();
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to delete member";
      setError(message);
    }
  }, [member, onClose, onSuccess, showToast, deleteMemberMutation]);

  const handleClose = useCallback(() => {
    setError("");
    onClose();
  }, [onClose]);

  const footer = (
    <div className={styles.footer}>
      <Button variant="tertiary" onClick={handleClose} disabled={isSubmitting} data-qa="delete-member-cancel-btn">
        Cancel
      </Button>
      <Button variant="destructive" onClick={handleSubmit} disabled={isSubmitting} data-qa="delete-member-confirm-btn">
        {isSubmitting ? "Deleting..." : "Delete Member"}
      </Button>
    </div>
  );

  if (!member) {
    return null;
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Delete Member"
      size="md"
      footer={footer}
      data-qa="delete-member-modal"
    >
      {error && (
        <div className={styles.errorAlert} role="alert" data-qa="delete-member-error-alert">
          {error}
        </div>
      )}
      <p>
        Are you sure you want to delete{" "}
        <strong>
          {member.firstName} {member.lastName}
        </strong>
        ?
      </p>
      <p>
        This will permanently delete the member and all their memberships across all semesters. Rankings will be removed
        and event entries will be unlinked.
      </p>
      <p>
        <strong>This action cannot be undone.</strong>
      </p>
    </Modal>
  );
}
