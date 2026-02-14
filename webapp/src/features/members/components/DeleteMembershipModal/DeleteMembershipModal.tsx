import { Modal, Button, useToast } from "@uwpokerclub/components";
import { useCallback, useState } from "react";
import { deleteMembership } from "../../api/memberRegistrationApi";
import { Membership } from "@/types";
import styles from "./DeleteMembershipModal.module.css";

export interface DeleteMembershipModalProps {
  isOpen: boolean;
  membership: Membership | null;
  semesterId: string;
  onClose: () => void;
  onSuccess: () => void;
}

export function DeleteMembershipModal({
  isOpen,
  membership,
  semesterId,
  onClose,
  onSuccess,
}: DeleteMembershipModalProps) {
  const { showToast } = useToast();
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = useCallback(async () => {
    if (!membership) return;

    setIsSubmitting(true);
    setError("");

    const result = await deleteMembership(semesterId, membership.id);

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
      message: `Membership for ${membership.user.firstName} ${membership.user.lastName} deleted successfully`,
      variant: "success",
      duration: 3000,
    });

    onSuccess();
    onClose();
    setIsSubmitting(false);
  }, [membership, semesterId, onClose, onSuccess, showToast]);

  const handleClose = useCallback(() => {
    setError("");
    setIsSubmitting(false);
    onClose();
  }, [onClose]);

  const footer = (
    <div className={styles.footer}>
      <Button variant="tertiary" onClick={handleClose} disabled={isSubmitting} data-qa="delete-membership-cancel-btn">
        Cancel
      </Button>
      <Button
        variant="destructive"
        onClick={handleSubmit}
        disabled={isSubmitting}
        data-qa="delete-membership-confirm-btn"
      >
        {isSubmitting ? "Deleting..." : "Delete Membership"}
      </Button>
    </div>
  );

  if (!membership) {
    return null;
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Delete Membership"
      size="md"
      footer={footer}
      data-qa="delete-membership-modal"
    >
      {error && (
        <div className={styles.errorAlert} role="alert" data-qa="delete-membership-error-alert">
          {error}
        </div>
      )}
      <p>
        Are you sure you want to delete the membership for{" "}
        <strong>
          {membership.user.firstName} {membership.user.lastName}
        </strong>
        ?
      </p>
      <p>This will permanently delete the membership. Rankings will be removed and event entries will be unlinked.</p>
      <p>
        <strong>This action cannot be undone.</strong>
      </p>
    </Modal>
  );
}
