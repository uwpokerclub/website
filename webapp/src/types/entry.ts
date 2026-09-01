export type Entry = {
  id: string;
  entryId: number;
  membershipId: string | null;
  eventId: string;
  firstName: string;
  lastName: string;
  signedOutAt: Date;
  points?: number;
};
