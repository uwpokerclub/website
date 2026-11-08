import { useEffect, useRef, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { zodResolver } from "@hookform/resolvers/zod";
import { FormProvider, useForm } from "react-hook-form";
import { Button, FormField, Input, Modal, Select, useToast } from "@uwpokerclub/components";
import { ApiError } from "@/lib/apiClient";
import type { Semester, User } from "@/types";
import type { Membership } from "@/types/membership";
import { memberKeys } from "@/features/members/hooks/useMemberQueries";
import { createMember, fetchMemberships, updateMember } from "@/features/members/api/memberRegistrationApi";
import {
  executiveAddMemberFormSchema,
  isValidQuestId,
  normalizeQuestIdInput,
  type ExecutiveAddMemberFormData,
} from "@/features/members/validation/registrationSchema";
import { NewMemberForm } from "@/features/members/components/RegisterMemberModal/NewMemberForm";
import { MemberSearch, type SelectedMemberData } from "@/features/members/components/RegisterMemberModal/MemberSearch";
import { semesterKeys } from "@/features/semesters/hooks/useSemesterQueries";
import {
  addExecutiveTeamMember,
  type AddExecutiveTeamMemberResponse,
  type ExecutiveTeamRole,
} from "../api/executiveTeamApi";
import { executiveTeamKeys } from "../hooks/useExecutiveTeamQueries";
import { executiveMembershipRefund } from "../utils/executiveRefund";
import styles from "./AddExecutiveMemberModal.module.css";

type Mode = "search" | "create";
type Stage = "details" | "confirm" | "success";

const roleOptions = [
  { value: "executive", label: "Executive" },
  { value: "tournament_director", label: "Tournament Director" },
];

function addErrorMessage(error: unknown): string {
  if (!(error instanceof ApiError)) return "The executive team could not be updated. Please try again.";
  if (error.status === 400) return "Enter a valid Quest ID containing only letters and numbers.";
  if (error.status === 403) return "This account has a protected role and cannot be added through this flow.";
  if (error.status === 404) return "The member or semester could not be found. Refresh and try again.";
  if (error.status === 409) {
    if (/staged|transition/i.test(error.message))
      return "This account is part of an officer transition and cannot be changed here.";
    if (/already/i.test(error.message)) return "This member is already on the selected semester’s executive team.";
    if (/ambiguous|identity/i.test(error.message))
      return "This Quest ID matches conflicting member or account records. Ask an administrator to resolve it.";
    return "The member or account changed while adding them. Refresh the roster and try again.";
  }
  return "The executive team could not be updated. Please try again.";
}

function money(amount: number): string {
  return `$${amount.toFixed(2)}`;
}

export interface AddExecutiveMemberModalProps {
  isOpen: boolean;
  semester: Semester;
  onClose: () => void;
}

export function AddExecutiveMemberModal({ isOpen, semester, onClose }: AddExecutiveMemberModalProps) {
  const queryClient = useQueryClient();
  const { showToast } = useToast();
  const initialSemesterId = useRef(semester.id);
  const [mode, setMode] = useState<Mode>("search");
  const [stage, setStage] = useState<Stage>("details");
  const [role, setRole] = useState<ExecutiveTeamRole>("executive");
  const [selectedMember, setSelectedMember] = useState<SelectedMemberData | null>(null);
  const [questIdInput, setQuestIdInput] = useState("");
  const [createData, setCreateData] = useState<ExecutiveAddMemberFormData | null>(null);
  const [savedMember, setSavedMember] = useState<User | null>(null);
  const [questIdSaved, setQuestIdSaved] = useState(false);
  const [questIdSaveUncertain, setQuestIdSaveUncertain] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [receipt, setReceipt] = useState<AddExecutiveTeamMemberResponse | null>(null);
  const [activationLink, setActivationLink] = useState<string | null>(null);
  const [copyError, setCopyError] = useState(false);

  const createForm = useForm<{ newMember: ExecutiveAddMemberFormData }>({
    resolver: zodResolver(executiveAddMemberFormSchema),
    defaultValues: {
      newMember: {
        id: "",
        questId: "",
        firstName: "",
        lastName: "",
        email: "",
        faculty: "" as ExecutiveAddMemberFormData["faculty"],
      },
    },
  });

  const membershipQuery = useQuery({
    queryKey: [...memberKeys.bySemester(semester.id), "executive-add-preview", selectedMember?.id ?? ""],
    queryFn: () => fetchMemberships(semester.id, { limit: 1, offset: 0, studentId: selectedMember!.id }),
    enabled: isOpen && mode === "search" && !!selectedMember,
    retry: false,
  });
  const existingMembership: Membership | undefined = membershipQuery.data?.data[0];
  const refund = executiveMembershipRefund(existingMembership, semester);

  useEffect(() => {
    if (isOpen && semester.id !== initialSemesterId.current) onClose();
  }, [isOpen, onClose, semester.id]);

  const reset = () => {
    createForm.reset();
    setMode("search");
    setStage("details");
    setRole("executive");
    setSelectedMember(null);
    setQuestIdInput("");
    setCreateData(null);
    setSavedMember(null);
    setQuestIdSaved(false);
    setQuestIdSaveUncertain(false);
    setSubmitting(false);
    setSubmitError(null);
    setReceipt(null);
    setActivationLink(null);
    setCopyError(false);
  };

  const close = () => {
    if (submitting) return;
    reset();
    onClose();
  };

  const selectMember = (member: SelectedMemberData | null) => {
    setSelectedMember(member);
    setQuestIdInput("");
    setSubmitError(null);
  };

  const reviewExisting = () => {
    if (!selectedMember || membershipQuery.isFetching || membershipQuery.isError) return;
    if (!selectedMember.questId && !isValidQuestId(normalizeQuestIdInput(questIdInput))) {
      setSubmitError("Quest ID must contain only letters and numbers.");
      return;
    }
    setSubmitError(null);
    setStage("confirm");
  };

  const reviewNew = createForm.handleSubmit((data) => {
    setCreateData(data.newMember);
    setSubmitError(null);
    setStage("confirm");
  });

  const confirm = async () => {
    if (submitting || semester.id !== initialSemesterId.current) return;
    setSubmitError(null);
    setSubmitting(true);
    try {
      let member: User;
      let questId: string;

      if (mode === "create") {
        if (savedMember) {
          member = savedMember;
          questId = savedMember.questId;
        } else {
          if (!createData) return;
          questId = normalizeQuestIdInput(createData.questId);
          member = await createMember({ ...createData, questId });
          member = { ...member, questId };
          setSavedMember(member);
        }
      } else {
        if (!selectedMember) return;
        questId = selectedMember.questId || normalizeQuestIdInput(questIdInput);
        if (!questId || (!selectedMember.questId && !isValidQuestId(questId))) {
          setSubmitError("Enter a Quest ID containing only letters and numbers.");
          setStage("details");
          return;
        }
        member = {
          id: selectedMember.id,
          firstName: selectedMember.firstName,
          lastName: selectedMember.lastName,
          email: selectedMember.email ?? "",
          faculty: selectedMember.faculty ?? "",
          questId,
          createdAt: "",
        };
        if (!selectedMember.questId && !questIdSaved) {
          try {
            await updateMember(selectedMember.id, {
              firstName: member.firstName,
              lastName: member.lastName,
              email: member.email,
              faculty: member.faculty,
              questId,
            });
            setSelectedMember({ ...selectedMember, questId });
            setQuestIdSaved(true);
          } catch {
            setQuestIdSaveUncertain(true);
            setSubmitError(
              "We could not confirm whether the Quest ID was saved. Close this dialog, search for the member again, and verify their record before retrying.",
            );
            return;
          }
        }
      }

      let response: AddExecutiveTeamMemberResponse;
      try {
        response = await addExecutiveTeamMember(semester.id, { questId, role });
      } catch (error) {
        const partialSuccess =
          mode === "create"
            ? ` The member was saved, but the team addition failed. Retry only the team add; do not create the member again.`
            : questIdSaved
              ? ` The Quest ID was saved, but the team addition failed. Retry only the team add.`
              : "";
        setSubmitError(`${addErrorMessage(error)}${partialSuccess}`);
        return;
      }

      setReceipt(response);
      if (response.accountStatus === "pending_activation" && response.activationToken) {
        setActivationLink(`${window.location.origin}/activate#token=${encodeURIComponent(response.activationToken)}`);
      }
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: executiveTeamKeys.bySemester(semester.id) }),
        queryClient.invalidateQueries({ queryKey: memberKeys.bySemester(semester.id) }),
        queryClient.invalidateQueries({ queryKey: semesterKeys.list() }),
      ]);
      showToast({
        message: `${member.firstName} ${member.lastName} added to the executive team.`,
        variant: "success",
        duration: 3000,
      });
      setStage("success");
    } catch {
      setSubmitError(
        "The member could not be saved. Check whether the member already exists, then search for them before trying again.",
      );
    } finally {
      setSubmitting(false);
    }
  };

  const copyActivationLink = async () => {
    if (!activationLink) return;
    try {
      await navigator.clipboard.writeText(activationLink);
      setCopyError(false);
      showToast({ message: "Activation link copied.", variant: "success", duration: 3000 });
    } catch {
      setCopyError(true);
    }
  };

  const personName =
    mode === "create"
      ? `${savedMember?.firstName ?? createData?.firstName ?? ""} ${savedMember?.lastName ?? createData?.lastName ?? ""}`.trim()
      : `${selectedMember?.firstName ?? ""} ${selectedMember?.lastName ?? ""}`.trim();
  const accountUsername =
    mode === "create"
      ? (savedMember?.questId ?? normalizeQuestIdInput(createData?.questId))
      : selectedMember?.questId || normalizeQuestIdInput(questIdInput);
  const refundForConfirmation = mode === "create" ? 0 : refund;

  return (
    <Modal
      isOpen={isOpen}
      onClose={close}
      title={
        stage === "success"
          ? "Executive team member added"
          : stage === "confirm"
            ? "Confirm executive team addition"
            : "Add executive team member"
      }
      size="lg"
      footer={
        <div className={styles.footer}>
          {stage === "details" ? (
            <>
              <Button variant="tertiary" onClick={close} disabled={submitting} data-qa="executive-add-cancel">
                Cancel
              </Button>
              {mode === "search" ? (
                <Button
                  onClick={reviewExisting}
                  disabled={
                    !selectedMember ||
                    membershipQuery.isFetching ||
                    membershipQuery.isError ||
                    (!selectedMember.questId && !normalizeQuestIdInput(questIdInput))
                  }
                  data-qa="executive-add-review"
                >
                  {membershipQuery.isFetching ? "Checking membership…" : "Review addition"}
                </Button>
              ) : (
                <Button onClick={() => void reviewNew()} data-qa="executive-add-review">
                  Review addition
                </Button>
              )}
            </>
          ) : stage === "confirm" ? (
            <>
              <Button
                variant="tertiary"
                onClick={() => setStage("details")}
                disabled={submitting || !!savedMember || questIdSaved || questIdSaveUncertain}
                data-qa="executive-add-back"
              >
                Back
              </Button>
              <Button
                onClick={() => void confirm()}
                disabled={submitting || questIdSaveUncertain}
                data-qa="executive-add-confirm"
              >
                {submitting ? "Adding…" : savedMember || questIdSaved ? "Retry team add" : "Confirm and add"}
              </Button>
            </>
          ) : (
            <Button onClick={close} data-qa="executive-add-done">
              Done
            </Button>
          )}
        </div>
      }
    >
      <div className={styles.content} data-qa="executive-add-modal">
        {stage === "details" && (
          <>
            <div className={styles.modeToggle}>
              <button
                type="button"
                onClick={() => {
                  setMode("search");
                  setSubmitError(null);
                }}
                data-qa="executive-add-search-mode"
              >
                Find existing member
              </button>
              <button
                type="button"
                onClick={() => {
                  setMode("create");
                  setSubmitError(null);
                }}
                data-qa="executive-add-create-mode"
              >
                Create new member
              </button>
            </div>
            {mode === "search" ? (
              <>
                <FormField label="Search member by name" htmlFor="executive-member-search">
                  {() => (
                    <MemberSearch value={selectedMember?.id ?? null} onSelect={selectMember} includeIdentityDetails />
                  )}
                </FormField>
                {selectedMember && !selectedMember.questId && (
                  <FormField label="Quest ID" htmlFor="executive-add-quest-id" required>
                    {(props) => (
                      <Input
                        {...props}
                        value={questIdInput}
                        onChange={(event) => setQuestIdInput(event.target.value)}
                        data-qa="executive-add-quest-id"
                        autoComplete="off"
                      />
                    )}
                  </FormField>
                )}
                {submitError && stage === "details" && (
                  <p role="alert" className={styles.error}>
                    {submitError}
                  </p>
                )}
                {selectedMember && membershipQuery.isError && (
                  <div role="alert" className={styles.error}>
                    Membership details could not be loaded, so the refund cannot be confirmed.{" "}
                    <button
                      type="button"
                      onClick={() => void membershipQuery.refetch()}
                      data-qa="executive-add-membership-retry"
                    >
                      Retry lookup
                    </button>
                  </div>
                )}
                {selectedMember && !membershipQuery.isFetching && !membershipQuery.isError && (
                  <p className={styles.help} data-qa="executive-add-membership-loaded">
                    {existingMembership
                      ? "Selected semester membership loaded."
                      : "No membership exists for this member in the selected semester."}
                  </p>
                )}
              </>
            ) : (
              <FormProvider {...createForm}>
                <form
                  onSubmit={(event) => {
                    event.preventDefault();
                    void reviewNew();
                  }}
                  className={styles.form}
                >
                  <NewMemberForm requiredQuestId />
                  <div className={styles.visuallyHidden}>
                    <button type="submit" data-qa="executive-add-create-submit">
                      Validate member
                    </button>
                  </div>
                </form>
              </FormProvider>
            )}
            <FormField label="Executive role" htmlFor="executive-add-role" required>
              {(props) => (
                <Select
                  {...props}
                  value={role}
                  onChange={(event) => setRole(event.target.value as ExecutiveTeamRole)}
                  options={roleOptions}
                  data-qa="executive-add-role"
                />
              )}
            </FormField>
          </>
        )}

        {stage === "confirm" && (
          <div className={styles.confirmation} data-qa="executive-add-confirmation">
            <p>
              Add <strong>{personName}</strong> ({accountUsername}) as{" "}
              <strong>{role === "executive" ? "Executive" : "Tournament Director"}</strong> for{" "}
              <strong>{semester.name}</strong>?
            </p>
            <p>
              {refundForConfirmation > 0
                ? `Refund ${money(refundForConfirmation)} cash for the ${existingMembership?.discounted ? "discounted " : ""}membership fee and reduce the semester budget by the same amount.`
                : "No paid membership fee refund or semester budget reduction applies."}
            </p>
            <p>
              The account will be created or activated if needed. If an account already exists, it will be reused or
              re-enabled with its existing password.
            </p>
            {submitError && (
              <p role="alert" className={styles.error} data-qa="executive-add-error">
                {submitError}
              </p>
            )}
          </div>
        )}

        {stage === "success" && receipt && (
          <div className={styles.confirmation} data-qa="executive-add-success">
            <p>
              <strong>{personName}</strong> was added to {semester.name} as{" "}
              {receipt.role === "executive" ? "Executive" : "Tournament Director"}.
            </p>
            {receipt.accountStatus === "pending_activation" && activationLink ? (
              <div className={styles.activation}>
                <p>Send this activation link to the member to set up their account.</p>
                <label htmlFor="executive-add-activation-link" className={styles.visuallyHidden}>
                  Activation link
                </label>
                <input
                  id="executive-add-activation-link"
                  value={activationLink}
                  readOnly
                  onFocus={(event) => event.currentTarget.select()}
                  data-qa="executive-add-activation-link"
                />
                <Button onClick={() => void copyActivationLink()} data-qa="executive-add-copy-link">
                  Copy link
                </Button>
                {copyError && <p role="alert">Copy failed. Select and copy the link manually.</p>}
              </div>
            ) : receipt.accountStatus === "pending_activation" ? (
              <p role="alert">
                The account is pending activation, but no activation token was returned. Use the roster’s
                replacement-link action.
              </p>
            ) : (
              <p>The existing account password is unchanged. The member may need to reauthenticate.</p>
            )}
          </div>
        )}
      </div>
    </Modal>
  );
}
