export interface Event {
  id: string;
  name: string;
  startDate: string;
  format: string;
  pointsMultiplier: number;
  additionalDetails: string;
  structureId: string;
  semesterId: string;
  state: number;
  rebuys: number;
}

export interface Semester {
  id: string;
  name: string;
  startDate: string;
  endDate: string;
}

export interface User {
  id: string;
  firstName: string;
  lastName: string;
  email: string;
  faculty: string;
  questId: string;
  createdAt: string;
}

export interface Structure {
  id: string;
  name: string;
  blinds: {
    small: number;
    big: number;
    ante: number;
    time: number;
  }[];
}

export interface Membership {
  id: string;
  userId: string;
  semesterId: string;
  paid: boolean;
  discounted: boolean;
}

export interface Transaction {
  id: string;
  description: string;
  amount: number;
  createdAt: string;
  semesterId: string;
}

export interface Ranking {
  membershipId: string;
  points: number;
  attendance: number;
}

export interface Participant {
  id: number;
  membershipId: string;
  eventId: number;
  placement: number;
  signedOutAt: string | null;
}