import { Modal, Button, useToast } from "@uwpokerclub/components";
import { useCallback, useState } from "react";
import { deleteLogin } from "../../api/loginsApi";
import { LoginResponse } from "../../types";
import styles from "./DeleteLoginModal.module.css";

export interface DeleteLoginModalProps {
  isOpen: boolean;
  login: LoginResponse | null;
  onClose: () => void;
  onSuccess: () => void;
}

export function DeleteLoginModal({ isOpen, login, onClose, onSuccess }: DeleteLoginModalProps) {
  const { showToast } = useToast();
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = useCallback(async () => {
    if (!login) return;

    setIsSubmitting(true);
    setError("");

    const result = await deleteLogin(login.username);

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
      message: `Login "${login.username}" deleted successfully`,
      variant: "success",
      duration: 3000,
    });

    onSuccess();
    onClose();
    setIsSubmitting(false);
  }, [login, onClose, onSuccess, showToast]);

  const handleClose = useCallback(() => {
    setError("");
    onClose();
  }, [onClose]);

  const footer = (
    <div className={styles.footer}>
      <Button variant="tertiary" onClick={handleClose} disabled={isSubmitting} data-qa="delete-login-cancel-btn">
        Cancel
      </Button>
      <Button variant="destructive" onClick={handleSubmit} disabled={isSubmitting} data-qa="delete-login-confirm-btn">
        {isSubmitting ? "Deleting..." : "Delete Login"}
      </Button>
    </div>
  );

  if (!login) {
    return null;
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Delete Login"
      size="md"
      footer={footer}
      data-qa="delete-login-modal"
    >
      {error && (
        <div className={styles.errorAlert} role="alert" data-qa="delete-login-error-alert">
          {error}
        </div>
      )}
      <p>
        Are you sure you want to delete the login for <strong>{login.username}</strong>?
      </p>
      {login.linkedMember && (
        <p>
          This login is linked to member{" "}
          <strong>
            {login.linkedMember.firstName} {login.linkedMember.lastName}
          </strong>
          . The member record will not be affected.
        </p>
      )}
      <p>
        <strong>This action cannot be undone.</strong> The user will no longer be able to log in with this account.
      </p>
    </Modal>
  );
}
