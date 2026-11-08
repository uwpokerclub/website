import { useEffect, useState } from "react";
import { useToast } from "@uwpokerclub/components";
import { QueryErrorState } from "@/components";
import { useAuth, useCurrentSemester } from "@/hooks";
import { useSemesters } from "@/features/semesters/hooks/useSemesterQueries";
import { createExecutiveTeamActivationLink } from "../api/executiveTeamApi";
import { AddExecutiveMemberModal } from "./AddExecutiveMemberModal";
import { useExecutiveTeam } from "../hooks/useExecutiveTeamQueries";
import type { ExecutiveTeamMember } from "../types";
import { semesterRosterStatus } from "../utils/semesterRosterStatus";
import styles from "./ExecutiveTeamRoster.module.css";

const protectedRoles = new Set(["president", "vice_president", "secretary", "treasurer", "webmaster", "bot"]);

const roleLabels: Record<string, string> = {
  executive: "Executive",
  tournament_director: "Tournament Director",
  president: "President",
  vice_president: "Vice President",
  secretary: "Secretary",
  treasurer: "Treasurer",
  webmaster: "Webmaster",
  bot: "Bot",
};

const accountStatusLabels = {
  active: "Active",
  pending_activation: "Pending activation",
  disabled: "Disabled",
} as const;

function statusFor(member: ExecutiveTeamMember): string {
  if (!member.hasLogin) return "No account";
  return member.accountStatus ? (accountStatusLabels[member.accountStatus] ?? "Unknown") : "Unknown";
}

function ActivationLinkAction({ semesterId, member }: { semesterId: string; member: ExecutiveTeamMember }) {
  const { showToast } = useToast();
  const [isGenerating, setIsGenerating] = useState(false);
  const [generatedLink, setGeneratedLink] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const generate = async () => {
    if (!member.username || isGenerating) return;
    setIsGenerating(true);
    setError(null);
    try {
      const { activationToken } = await createExecutiveTeamActivationLink(semesterId, member.username);
      setGeneratedLink(`${window.location.origin}/activate#token=${encodeURIComponent(activationToken)}`);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Unable to generate an activation link.");
    } finally {
      setIsGenerating(false);
    }
  };

  const copy = async () => {
    if (!generatedLink) return;
    try {
      await navigator.clipboard.writeText(generatedLink);
      showToast({ message: "Activation link copied.", variant: "success", duration: 3000 });
    } catch {
      setError("Copy failed. Select and copy the link manually.");
    }
  };

  return (
    <div className={styles.activationAction}>
      {generatedLink ? (
        <>
          <label className={styles.visuallyHidden} htmlFor={`activation-link-${member.memberId}`}>
            Replacement activation link for {member.firstName} {member.lastName}
          </label>
          <input
            id={`activation-link-${member.memberId}`}
            className={styles.activationLink}
            value={generatedLink}
            readOnly
            onFocus={(event) => event.currentTarget.select()}
            data-qa={`executive-team-activation-link-${member.memberId}`}
          />
          <button type="button" className={styles.actionButton} onClick={() => void copy()}>
            Copy link
          </button>
        </>
      ) : (
        <button
          type="button"
          className={styles.actionButton}
          onClick={() => void generate()}
          disabled={isGenerating}
          data-qa={`executive-team-generate-link-${member.memberId}`}
        >
          {isGenerating ? "Generating…" : "Generate replacement link"}
        </button>
      )}
      {error && (
        <p className={styles.actionError} role="alert">
          {error}
        </p>
      )}
    </div>
  );
}

export function ExecutiveTeamRoster() {
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [addModalSemesterId, setAddModalSemesterId] = useState<string | null>(null);
  const { currentSemester, loading: semesterLoading } = useCurrentSemester();
  const semesters = useSemesters();
  const { hasPermission } = useAuth();
  const canView = hasPermission("get", "executive-team");
  const canCreate = hasPermission("create", "executive-team");
  const roster = useExecutiveTeam(currentSemester?.id, canView);
  const today = new Date();

  useEffect(() => {
    if (isAddModalOpen && currentSemester?.id !== addModalSemesterId) {
      setIsAddModalOpen(false);
      setAddModalSemesterId(null);
    }
  }, [addModalSemesterId, currentSemester?.id, isAddModalOpen]);

  if (semesterLoading || semesters.isLoading || (canView && currentSemester && roster.isPending)) {
    return (
      <p className={styles.message} role="status" aria-live="polite">
        Loading executive roster…
      </p>
    );
  }

  if (!canView) {
    return (
      <p className={styles.message} role="status">
        You do not have permission to view this roster.
      </p>
    );
  }

  if (!currentSemester) {
    if (semesters.isError) {
      return (
        <QueryErrorState
          data-qa="executive-team-semester-error"
          retryDataQa="executive-team-semester-retry"
          title="Failed to load semesters"
          message={semesters.error.message}
          onRetry={() => void semesters.refetch()}
        />
      );
    }
    return (
      <p className={styles.message} data-qa="executive-team-no-semester">
        Please select a semester to view its roster.
      </p>
    );
  }

  if (roster.isError) {
    return (
      <QueryErrorState
        data-qa="executive-team-error"
        retryDataQa="executive-team-retry"
        title="Failed to load executive roster"
        message={roster.error.message}
        onRetry={() => void roster.refetch()}
      />
    );
  }

  const dateStatus = semesterRosterStatus(currentSemester, today);
  const isReadOnly = dateStatus === "past" || dateStatus === "unknown";
  const members = roster.data?.members ?? [];

  return (
    <div className={styles.roster} data-qa="executive-team-roster">
      {canCreate && !isReadOnly && (
        <div className={styles.rosterActions}>
          <button
            type="button"
            className={styles.actionButton}
            onClick={() => {
              setAddModalSemesterId(currentSemester.id);
              setIsAddModalOpen(true);
            }}
            data-qa="executive-team-add-member"
          >
            Add team member
          </button>
        </div>
      )}
      {dateStatus === "past" && (
        <p className={styles.readOnlyNotice} data-qa="executive-team-read-only">
          Past semester — this roster is read-only.
        </p>
      )}
      {dateStatus === "unknown" && (
        <p className={styles.readOnlyNotice} data-qa="executive-team-read-only">
          Semester dates are unavailable or invalid — this roster is read-only.
        </p>
      )}
      {members.length === 0 ? (
        <p className={styles.message} data-qa="executive-team-empty">
          No executive members are recorded for this semester.
        </p>
      ) : (
        <div className={styles.tableWrapper}>
          <table className={styles.table}>
            <caption className={styles.visuallyHidden}>{currentSemester.name} executive roster</caption>
            <thead>
              <tr>
                <th scope="col">Name</th>
                <th scope="col">Quest ID</th>
                <th scope="col">Role</th>
                <th scope="col">Account status</th>
                <th scope="col">Activation link</th>
              </tr>
            </thead>
            <tbody>
              {members.map((member) => {
                const protectedRole = member.role ? protectedRoles.has(member.role) : false;
                const canGenerateLink =
                  !isReadOnly &&
                  member.hasExecutiveMembership &&
                  member.hasLogin &&
                  member.accountStatus === "pending_activation" &&
                  !!member.username &&
                  (member.role === "executive" || member.role === "tournament_director");
                return (
                  <tr key={member.memberId} data-qa={`executive-team-row-${member.memberId}`}>
                    <th scope="row">{`${member.firstName} ${member.lastName}`.trim() || "Name unavailable"}</th>
                    <td>{member.questId || "Quest ID unavailable"}</td>
                    <td>
                      {member.role ? (roleLabels[member.role] ?? "Unknown role") : "Not assigned (no account)"}
                      {protectedRole && <span className={styles.guidance}>Managed through officer transition.</span>}
                    </td>
                    <td>
                      <span
                        className={`${styles.status} ${member.hasLogin ? styles[member.accountStatus ?? "unknown"] : styles.noAccount}`}
                      >
                        {statusFor(member)}
                      </span>
                    </td>
                    <td>
                      {canGenerateLink ? (
                        <ActivationLinkAction
                          key={`${currentSemester.id}:${member.username}`}
                          semesterId={currentSemester.id}
                          member={member}
                        />
                      ) : member.accountStatus === "pending_activation" ? (
                        <span>
                          {isReadOnly
                            ? "Unavailable for read-only semester"
                            : protectedRole
                              ? "Use officer transition to replace this link"
                              : "No activation action is available for this account"}
                        </span>
                      ) : (
                        <span>—</span>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
      {canCreate && !isReadOnly && isAddModalOpen && (
        <AddExecutiveMemberModal
          isOpen={isAddModalOpen}
          semester={currentSemester}
          onClose={() => {
            setIsAddModalOpen(false);
            setAddModalSemesterId(null);
          }}
        />
      )}
    </div>
  );
}
