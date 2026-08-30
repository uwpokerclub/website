export type FreeTrialFields = {
  paid: boolean;
  freeTrialAvailable: boolean;
};

// Compared against `false`, not negated: an absent membership must not flag.
export default function hasExhaustedFreeTrial(membership: FreeTrialFields | null | undefined): boolean {
  return membership?.paid === false && membership?.freeTrialAvailable === false;
}
