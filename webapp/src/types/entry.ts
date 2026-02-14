export type Entry = {
  id: string;
  membershipId: string | null;
  eventId: string;
  firstName: string;
  lastName: string;
  signedOutAt: Date;
  placement?: number;
};
