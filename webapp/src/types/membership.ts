import type { User } from "./user";

export type Membership = {
  id: string;
  userId: number;
  user: User;
  semesterId: string;
  paid: boolean;
  discounted: boolean;
  attendance: number;
};
