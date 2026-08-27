/**
 * The subset of a membership needed to decide whether to flag it. Both fields are
 * optional because the entries endpoint returns a partially-populated membership
 * (see ParticipantResponse in features/entries/api/entriesApi.ts).
 */
export type FreeTrialFields = {
  paid?: boolean;
  freeTrialAvailable?: boolean;
};

/**
 * True when a member should be flagged as having used up their free trial.
 *
 * Both fields are compared against `false` explicitly rather than negated: an
 * absent field (an older cached response, or an API that predates the field)
 * must mean "don't flag," but `!undefined` is `true` and would flag everyone.
 *
 * `paid` is part of the check because freeTrialAvailable is a display-only cache
 * that can go stale, and paid memberships are never restricted by the free trial
 * — a stale `false` on a paid member would be a false alarm.
 */
export default function hasExhaustedFreeTrial(membership: FreeTrialFields | null | undefined): boolean {
  if (!membership) {
    return false;
  }
  return membership.paid === false && membership.freeTrialAvailable === false;
}
