import { useState, useContext, useCallback, useEffect } from "react";
import { useForm, FormProvider } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Modal, Button, useToast } from "@uwpokerclub/components";
import { SemesterContext } from "../../../../contexts";
import { EditMemberForm } from "./EditMemberForm";
import { MembershipConfig } from "../RegisterMemberModal/MembershipConfig";
import { editMemberMembershipSchema, type EditMemberMembershipFormData } from "../../validation/registrationSchema";
import { updateMember, updateMembership } from "../../api/memberRegistrationApi";
import type { Membership } from "../../../../types";
import styles from "./EditMemberModal.module.css";

export interface EditMemberModalProps {
  isOpen: boolean;
  membership: Membership | null;
  onClose: () => void;
  onSuccess: () => void;
}

/**
 * EditMemberModal - Modal for editing existing member and membership details
 */
export function EditMemberModal({ isOpen, membership, onClose, onSuccess }: EditMemberModalProps) {
  const semesterContext = useContext(SemesterContext);
  const { showToast } = useToast();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const form = useForm<EditMemberMembershipFormData>({
    resolver: zodResolver(editMemberMembershipSchema),
    defaultValues: {
      member: {
        firstName: "",
        lastName: "",
        email: "",
        faculty: "" as EditMemberMembershipFormData["member"]["faculty"],
        questId: "",
      },
      membership: {
        paid: false,
        discounted: false,
      },
    },
  });

  // Reset form with membership data when modal opens or membership changes
  useEffect(() => {
    if (isOpen && membership) {
      form.reset({
        member: {
          firstName: membership.user.firstName,
          lastName: membership.user.lastName,
          email: membership.user.email,
          faculty: (membership.user.faculty || "") as EditMemberMembershipFormData["member"]["faculty"],
          questId: membership.user.questId || "",
        },
        membership: {
          paid: membership.paid,
          discounted: membership.discounted,
        },
      });
      setSubmitError(null);
    }
  }, [isOpen, membership, form]);

  // Handle modal close
  const handleClose = useCallback(() => {
    form.reset();
    setSubmitError(null);
    onClose();
  }, [form, onClose]);

  // Handle form submit
  const handleSubmit = async (data: EditMemberMembershipFormData) => {
    if (!semesterContext?.currentSemester?.id || !membership) {
      setSubmitError("No semester or membership selected");
      return;
    }

    setIsSubmitting(true);
    setSubmitError(null);

    // Update member data
    const memberResult = await updateMember(String(membership.userId), {
      firstName: data.member.firstName,
      lastName: data.member.lastName,
      email: data.member.email,
      faculty: data.member.faculty,
      questId: data.member.questId || "",
    });

    if (!memberResult.success) {
      setIsSubmitting(false);
      setSubmitError(memberResult.error);
      showToast({
        message: memberResult.error,
        variant: "error",
        duration: 5000,
      });
      return;
    }

    // Update membership data
    const membershipResult = await updateMembership(semesterContext.currentSemester.id, membership.id, {
      paid: data.membership.paid,
      discounted: data.membership.discounted,
    });

    setIsSubmitting(false);

    if (membershipResult.success) {
      showToast({
        message: `${data.member.firstName} ${data.member.lastName} updated successfully!`,
        variant: "success",
        duration: 3000,
      });
      onSuccess();
      handleClose();
    } else {
      setSubmitError(membershipResult.error);
      showToast({
        message: membershipResult.error,
        variant: "error",
        duration: 5000,
      });
    }
  };

  // Footer with actions
  const footer = (
    <div className={styles.footer}>
      <Button data-qa="edit-cancel-btn" variant="tertiary" onClick={handleClose} disabled={isSubmitting}>
        Cancel
      </Button>
      <Button data-qa="edit-submit-btn" type="submit" form="edit-member-form" disabled={isSubmitting}>
        {isSubmitting ? "Saving..." : "Save Changes"}
      </Button>
    </div>
  );

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="Edit Member" size="lg" footer={footer}>
      <div className={styles.content} data-qa="edit-member-modal">
        {/* Error display */}
        {submitError && (
          <div className={styles.errorAlert} data-qa="edit-error-alert">
            {submitError}
          </div>
        )}

        <FormProvider {...form}>
          <form id="edit-member-form" onSubmit={form.handleSubmit(handleSubmit)} noValidate>
            <div className={styles.section}>
              <h3 className={styles.sectionTitle}>Member Details</h3>
              <EditMemberForm studentId={membership?.userId ? String(membership.userId) : ""} />
            </div>
            <MembershipConfig />
          </form>
        </FormProvider>
      </div>
    </Modal>
  );
}
