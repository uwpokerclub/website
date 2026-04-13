import { useState, useContext, useCallback } from "react";
import { useForm, FormProvider } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Modal, Button, useToast } from "@uwpokerclub/components";
import { SemesterContext } from "../../../../contexts";
import { MemberSearch, type SelectedMemberData } from "./MemberSearch";
import { NewMemberForm } from "./NewMemberForm";
import { MembershipConfig } from "./MembershipConfig";
import {
  searchModeSchema,
  createModeSchema,
  type SearchModeFormData,
  type CreateModeFormData,
} from "../../validation/registrationSchema";
import { createMembership, registerNewMemberWithMembership } from "../../api/memberRegistrationApi";
import styles from "./RegisterMemberModal.module.css";

type Mode = "search" | "create";

/**
 * Data returned on successful member registration
 */
export interface RegistrationSuccessData {
  membershipId: string;
  userId: number;
  firstName: string;
  lastName: string;
}

export interface RegisterMemberModalProps {
  isOpen: boolean;
  onClose: () => void;
  /** Called on success. Optionally receives the created membership data. */
  onSuccess: (data?: RegistrationSuccessData) => void;
}

/**
 * RegisterMemberModal - Modal for registering new members
 *
 * Supports two modes:
 * - Search mode: Find existing member and create membership
 * - Create mode: Create new member and membership together
 */
export function RegisterMemberModal({ isOpen, onClose, onSuccess }: RegisterMemberModalProps) {
  const semesterContext = useContext(SemesterContext);
  const { showToast } = useToast();
  const [mode, setMode] = useState<Mode>("search");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [selectedMemberName, setSelectedMemberName] = useState<{ firstName: string; lastName: string } | null>(null);

  // Search mode form
  const searchForm = useForm<SearchModeFormData>({
    resolver: zodResolver(searchModeSchema),
    defaultValues: {
      selectedMemberId: "",
      membership: {
        paid: false,
        discounted: false,
      },
    },
  });

  // Create mode form
  const createForm = useForm<CreateModeFormData>({
    resolver: zodResolver(createModeSchema),
    defaultValues: {
      newMember: {
        id: "",
        questId: "",
        firstName: "",
        lastName: "",
        email: "",
        faculty: "" as CreateModeFormData["newMember"]["faculty"],
      },
      membership: {
        paid: false,
        discounted: false,
      },
    },
  });

  // Reset forms when modal closes or mode changes
  const resetForms = useCallback(() => {
    searchForm.reset();
    createForm.reset();
    setSubmitError(null);
    setSelectedMemberName(null);
  }, [searchForm, createForm]);

  // Handle mode toggle
  const handleModeToggle = () => {
    setMode((prev) => (prev === "search" ? "create" : "search"));
    setSubmitError(null);
  };

  // Handle modal close
  const handleClose = () => {
    resetForms();
    setMode("search");
    onClose();
  };

  // Handle search mode selection
  const handleMemberSelect = (member: SelectedMemberData | null) => {
    searchForm.setValue("selectedMemberId", member?.id ?? "", { shouldValidate: true });
    if (member) {
      setSelectedMemberName({ firstName: member.firstName, lastName: member.lastName });
    } else {
      setSelectedMemberName(null);
    }
  };

  // Submit handler for search mode
  const handleSearchSubmit = async (data: SearchModeFormData) => {
    if (!semesterContext?.currentSemester?.id) {
      setSubmitError("No semester selected");
      return;
    }

    setIsSubmitting(true);
    setSubmitError(null);

    try {
      const membership = await createMembership(
        semesterContext.currentSemester.id,
        data.selectedMemberId,
        data.membership.paid,
        data.membership.discounted,
      );

      showToast({
        message: "Member registered successfully!",
        variant: "success",
        duration: 3000,
      });
      searchForm.reset();
      onSuccess({
        membershipId: membership.id,
        userId: membership.userId,
        firstName: selectedMemberName?.firstName ?? membership.user?.firstName ?? "",
        lastName: selectedMemberName?.lastName ?? membership.user?.lastName ?? "",
      });
      setSelectedMemberName(null);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to create membership";
      setSubmitError(message);
      showToast({
        message,
        variant: "error",
        duration: 5000,
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  // Submit handler for create mode
  const handleCreateSubmit = async (data: CreateModeFormData) => {
    if (!semesterContext?.currentSemester?.id) {
      setSubmitError("No semester selected");
      return;
    }

    setIsSubmitting(true);
    setSubmitError(null);

    try {
      const { member, membership } = await registerNewMemberWithMembership(
        data.newMember,
        semesterContext.currentSemester.id,
        data.membership.paid,
        data.membership.discounted,
      );

      showToast({
        message: `${member.firstName} ${member.lastName} registered successfully!`,
        variant: "success",
        duration: 3000,
      });
      createForm.reset();
      onSuccess({
        membershipId: membership.id,
        userId: membership.userId,
        firstName: member.firstName,
        lastName: member.lastName,
      });
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to register member";
      setSubmitError(message);
      showToast({
        message,
        variant: "error",
        duration: 5000,
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  // Footer with actions
  const footer = (
    <div className={styles.footer}>
      <Button data-qa="register-cancel-btn" variant="tertiary" onClick={handleClose} disabled={isSubmitting}>
        Cancel
      </Button>
      <Button
        data-qa="register-submit-btn"
        type="submit"
        form={mode === "search" ? "search-form" : "create-form"}
        disabled={isSubmitting}
      >
        {isSubmitting ? "Registering..." : "Register Member"}
      </Button>
    </div>
  );

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="Semester Registration" size="lg" footer={footer}>
      <div className={styles.content} data-qa="register-member-modal">
        {/* Mode toggle */}
        <div className={styles.modeToggle}>
          {mode === "search" ? (
            <button
              type="button"
              onClick={handleModeToggle}
              className={styles.modeLink}
              data-qa="toggle-new-member-btn"
            >
              Can&apos;t find the member you are looking for? Create a new member
            </button>
          ) : (
            <button type="button" onClick={handleModeToggle} className={styles.modeLink} data-qa="toggle-search-btn">
              &larr; Back to search
            </button>
          )}
        </div>

        {/* Error display */}
        {submitError && (
          <div className={styles.errorAlert} data-qa="register-error-alert">
            {submitError}
          </div>
        )}

        {/* Search mode */}
        {mode === "search" && (
          <FormProvider {...searchForm}>
            <form id="search-form" onSubmit={searchForm.handleSubmit(handleSearchSubmit)} noValidate>
              <div className={styles.section}>
                <h3 className={styles.sectionTitle}>Find an existing member</h3>
                <MemberSearch
                  value={searchForm.watch("selectedMemberId") || null}
                  onSelect={handleMemberSelect}
                  error={searchForm.formState.errors.selectedMemberId?.message}
                />
              </div>
              <MembershipConfig />
            </form>
          </FormProvider>
        )}

        {/* Create mode */}
        {mode === "create" && (
          <FormProvider {...createForm}>
            <form id="create-form" onSubmit={createForm.handleSubmit(handleCreateSubmit)} noValidate>
              <div className={styles.section}>
                <h3 className={styles.sectionTitle}>New Member Details</h3>
                <NewMemberForm />
              </div>
              <MembershipConfig />
            </form>
          </FormProvider>
        )}
      </div>
    </Modal>
  );
}
