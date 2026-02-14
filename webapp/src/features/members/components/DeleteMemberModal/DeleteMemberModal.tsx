import { Modal, Button, useToast } from "@uwpokerclub/components";
import { useCallback, useState } from "react";
import { deleteMember } from "../../api/memberRegistrationApi";
import styles from "./DeleteMemberModal.module.css";

export interface DeleteMemberModalProps {
  isOpen: boolean;
  member: { id: string; firstName: string; lastName: string } | null;
  onClose: () => void;
  onSuccess: () => void;
}

export function DeleteMemberModal({ isOpen, member, onClose, onSuccess }: DeleteMemberModalProps) {
  const { showToast } = useToast();
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = useCallback(async () => {
    if (!member) return;

    setIsSubmitting(true);
    setError("");

    const result = await deleteMember(member.id);

    if (!result.success) {
      setError(result.error);
      showToast({
        message: result.error,
        variant: "error",
        duration: 5000,
      });
      setIsSubmitting(false);
      return;
    }

    showToast({
      message: `${member.firstName} ${member.lastName} deleted successfully`,
      variant: "success",
      duration: 3000,
    });

    onSuccess();
    onClose();
    setIsSubmitting(false);
  }, [member, onClose, onSuccess, showToast]);

  const handleClose = useCallback(() => {
    setError("");
    setIsSubmitting(false);
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
