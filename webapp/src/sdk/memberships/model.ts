import type { User } from "../../types/user";

export interface Membership {
  id: string;
  userId: number;
  user: User;
  semesterId: string;
  paid: boolean;
  discounted: boolean;
  attendance: number;
}
