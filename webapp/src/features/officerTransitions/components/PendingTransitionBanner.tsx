import { useQueries } from "@tanstack/react-query";
import { Button, useToast } from "@uwpokerclub/components";
import { useState } from "react";
import { useAuth } from "@/hooks";
import { resolveQuestId } from "../api/officerTransitionsApi";
import { useCancelOfficerTransition, useReissueOfficerTransitionLink } from "../hooks/useOfficerTransitionQueries";
import { OFFICER_ROLES, type OfficerRole, type OfficerTransition } from "../types";
import { activationLabel, roleLabels, usernameForRole } from "../transitionPresentation";
import styles from "./OfficerTransition.module.css";

export function PendingTransitionBanner({ transition }: { transition: OfficerTransition }) {
  const { hasPermission } = useAuth();
  const { showToast } = useToast();
  const canManage = hasPermission("cancel", "officer-transition");
  const cancel = useCancelOfficerTransition();
  const reissue = useReissueOfficerTransitionLink();
  const [links, setLinks] = useState<Partial<Record<OfficerRole, string>>>({});
  const nameQueries = useQueries({
    queries: OFFICER_ROLES.map((role) => {
      const username = usernameForRole(transition, role);
      return {
        queryKey: ["officer-transition", "nominee", username],
        queryFn: () => resolveQuestId(username!),
        enabled: !!username,
        retry: false,
      };
    }),
  });

  const nameFor = (role: OfficerRole) => {
    const nominee = transition.nominees?.[role] ?? nameQueries[OFFICER_ROLES.indexOf(role)]?.data?.nominee;
    return nominee
      ? `${nominee.firstName} ${nominee.lastName}`
      : (usernameForRole(transition, role) ?? "Unknown member");
  };
  const copy = async (role: OfficerRole) => {
    const link = links[role];
    if (!link) return;
    try {
      await navigator.clipboard.writeText(link);
      showToast({ message: "Activation link copied.", variant: "success", duration: 3000 });
    } catch {
      showToast({ message: "Copy failed. Select and copy the link manually.", variant: "error", duration: 5000 });
    }
  };
  const cancelTransition = () => {
    if (
      !window.confirm(
        "Cancel this staged transition? The staged accounts and activation links will be discarded. The current team's access will not change.",
      )
    )
      return;
    cancel.mutate(transition.id, {
      onError: (error) => showToast({ message: error.message, variant: "error", duration: 5000 }),
    });
  };
  const reissueLink = (role: OfficerRole) => {
    reissue.mutate(
      { id: transition.id, role },
      {
        onSuccess: ({ activationToken }) => {
          setLinks((current) => ({
            ...current,
            [role]: `${window.location.origin}/activate#token=${encodeURIComponent(activationToken)}`,
          }));
        },
        onError: (error) => showToast({ message: error.message, variant: "error", duration: 5000 }),
      },
    );
  };

  return (
    <section
      className={styles.pendingBanner}
      aria-labelledby="pending-transition-heading"
      data-qa="pending-transition-banner"
    >
      <div className={styles.pendingHeader}>
        <div>
          <p className={styles.eyebrow}>Pending officer transition</p>
          <h2 id="pending-transition-heading">Incoming executive team</h2>
        </div>
        <p className={styles.accessNotice}>
          Current access transfers when <strong>{nameFor("president")}</strong> completes their transition activation.
        </p>
      </div>
      <dl className={styles.pendingNomineeList} data-qa="pending-transition-nominees">
        {OFFICER_ROLES.map((role) => {
          const activated = transition.activated?.[role] ?? false;
          const link = links[role];
          return (
            <div key={role} className={styles.nomineeItem} data-qa={`pending-transition-${role}`}>
              <dt>{roleLabels[role]}</dt>
              <dd>{nameFor(role)}</dd>
              <p className={activated ? styles.activated : styles.awaiting}>{activationLabel(activated)}</p>
              {canManage && (
                <div className={styles.reissueControl}>
                  <Button
                    variant="secondary"
                    onClick={() => reissueLink(role)}
                    disabled={reissue.isPending}
                    data-qa={`pending-transition-reissue-${role}`}
                  >
                    Re-issue link
                  </Button>
                  {link && (
                    <div className={styles.linkControl}>
                      <input
                        readOnly
                        value={link}
                        aria-label={`${nameFor(role)} replacement activation link`}
                        data-qa={`pending-transition-link-${role}`}
                      />
                      <Button
                        variant="secondary"
                        onClick={() => void copy(role)}
                        data-qa={`pending-transition-copy-${role}`}
                      >
                        Copy link
                      </Button>
                    </div>
                  )}
                </div>
              )}
            </div>
          );
        })}
      </dl>
      {canManage && (
        <div className={styles.pendingActions}>
          <p>Re-issuing invalidates the previous link. The original link cannot be redisplayed.</p>
          <Button
            variant="destructive"
            onClick={cancelTransition}
            disabled={cancel.isPending}
            data-qa="pending-transition-cancel"
          >
            Cancel transition
          </Button>
        </div>
      )}
    </section>
  );
}
