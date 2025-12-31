import { Event, Membership, Participant, Ranking, User } from "./types";

export const STRUCTURE = {
  id: 1,
  name: "Structure A",
};

export const SEMESTER = {
  id: "84f026be-53e0-4759-ab89-131c4a66d649",
  name: "Winter 2025",
  startDate: "2025-01-01T00:00:00.000Z",
  endDate: "2025-06-30T00:00:000Z",
  startingBudget: 100,
  currentBudget: 100,
  membershipFee: 10,
  memershipDiscountFee: 7,
  rebuyFee: 2,
  meta: "Seed semester",
};

export const EVENT: Event = {
  id: "1",
  name: "Winter 2025 Event #1",
  format: "No Limit Hold'em",
  semesterId: SEMESTER.id,
  state: 0,
  structureId: "1",
  rebuys: 0,
  pointsMultiplier: 1.0,
  startDate: "2025-01-03T19:00:00.000Z",
  additionalDetails: "Seed event",
};

export const ENDED_EVENT: Event = {
  id: "2",
  name: "Winter 2025 Event #2",
  format: "No Limit Hold'em",
  semesterId: SEMESTER.id,
  state: 1,
  structureId: "1",
  rebuys: 0,
  pointsMultiplier: 1.0,
  startDate: "2025-01-10T19:00:00.000Z",
  additionalDetails: "Completed event",
};

export const MEMBERS: Membership[] = [
  {
    id: "5d312426-ad56-4231-bb12-241acbfb91e2",
    userId: "62958169",
    semesterId: "84f026be-53e0-4759-ab89-131c4a66d649",
    paid: false,
    discounted: false,
  },
  {
    id: "c0f1b2a4-3d5e-4b8c-8f7d-6a9e0f3b1c5d",
    userId: "20141158",
    semesterId: "84f026be-53e0-4759-ab89-131c4a66d649",
    paid: false,
    discounted: false,
  },
  {
    id: "b2a7e2b6-5c3f-4a1e-a0d5-8f3e1b2c3d4e",
    userId: "85018940",
    semesterId: "84f026be-53e0-4759-ab89-131c4a66d649",
    paid: false,
    discounted: false,
  },
  {
    id: "d4e5f6a7-b8c9-4d0e-a1b2-c3d4e5f6a7b8",
    userId: "77679767",
    semesterId: "84f026be-53e0-4759-ab89-131c4a66d649",
    paid: true,
    discounted: false,
  },
  {
    id: "65f65311-cc74-4c76-9cdf-29c5d674d40a",
    userId: "70492884",
    semesterId: "84f026be-53e0-4759-ab89-131c4a66d649",
    paid: true,
    discounted: true,
  },
  {
    id: "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d",
    userId: "39166759",
    semesterId: "84f026be-53e0-4759-ab89-131c4a66d649",
    paid: true,
    discounted: false,
  },
  {
    id: "b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e",
    userId: "55686346",
    semesterId: "84f026be-53e0-4759-ab89-131c4a66d649",
    paid: true,
    discounted: false,
  },
  {
    id: "c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f",
    userId: "81085720",
    semesterId: "84f026be-53e0-4759-ab89-131c4a66d649",
    paid: true,
    discounted: true,
  },
  {
    id: "d4e5f6a7-b8c9-4d0e-1f2a-3b4c5d6e7f8a",
    userId: "52873146",
    semesterId: "84f026be-53e0-4759-ab89-131c4a66d649",
    paid: false,
    discounted: false,
  },
  {
    id: "e5f6a7b8-c9d0-4e1f-2a3b-4c5d6e7f8a9b",
    userId: "75969632",
    semesterId: "84f026be-53e0-4759-ab89-131c4a66d649",
    paid: true,
    discounted: false,
  },
];

export const USERS: User[] = [
  {
    id: "62958169",
    firstName: "Heinrik",
    lastName: "Drust",
    email: "hdrust0@merriam-webster.com",
    faculty: "AHS",
    questId: "hdrust0",
    createdAt: "2025-04-18",
  },
  {
    id: "20141158",
    firstName: "Doretta",
    lastName: "Housegoe",
    email: "dhousegoe1@xinhuanet.com",
    faculty: "AHS",
    questId: "dhousegoe1",
    createdAt: "2025-04-19",
  },
  {
    id: "85018940",
    firstName: "Elita",
    lastName: "Aucock",
    email: "eaucock2@si.edu",
    faculty: "Science",
    questId: "eaucock2",
    createdAt: "2024-10-07",
  },
  {
    id: "77679767",
    firstName: "Khalil",
    lastName: "Duckham",
    email: "kduckham3@wsj.com",
    faculty: "Engineering",
    questId: "kduckham3",
    createdAt: "2025-03-22",
  },
  {
    id: "70492884",
    firstName: "Amandie",
    lastName: "Libbis",
    email: "alibbis4@google.co.jp",
    faculty: "Engineering",
    questId: "alibbis4",
    createdAt: "2025-03-31",
  },
  {
    id: "39166759",
    firstName: "Wald",
    lastName: "Sundin",
    email: "wsundin5@prweb.com",
    faculty: "Science",
    questId: "wsundin5",
    createdAt: "2024-06-30",
  },
  {
    id: "55686346",
    firstName: "Germayne",
    lastName: "Croom",
    email: "gcroom6@drupal.org",
    faculty: "Science",
    questId: "gcroom6",
    createdAt: "2024-05-28",
  },
  {
    id: "81085720",
    firstName: "Oralie",
    lastName: "Bunten",
    email: "obunten7@dropbox.com",
    faculty: "Science",
    questId: "obunten7",
    createdAt: "2024-07-19",
  },
  {
    id: "52873146",
    firstName: "Kristel",
    lastName: "Callan",
    email: "kcallan8@arizona.edu",
    faculty: "Math",
    questId: "kcallan8",
    createdAt: "2024-06-27",
  },
  {
    id: "75969632",
    firstName: "Winslow",
    lastName: "Josey",
    email: "wjosey9@blogger.com",
    faculty: "Environment",
    questId: "wjosey9",
    createdAt: "2025-01-30",
  },
];

// Users that exist in the database but do NOT have memberships for the current semester
// Use these for testing member registration flows
export const USERS_WITHOUT_MEMBERSHIPS: User[] = [
  {
    id: "11111111",
    firstName: "Unregistered",
    lastName: "TestUser",
    email: "unregistered@test.com",
    faculty: "Math",
    questId: "unreg1",
    createdAt: "2025-01-01",
  },
  {
    id: "22222222",
    firstName: "Another",
    lastName: "Unregistered",
    email: "another.unreg@test.com",
    faculty: "Science",
    questId: "unreg2",
    createdAt: "2025-01-01",
  },
];

export const ACTIVE_EVENT_PARTICIPANTS: Participant[] = [
  {
    id: 1,
    membershipId: "5d312426-ad56-4231-bb12-241acbfb91e2",
    eventId: 1,
  },
  {
    id: 2,
    membershipId: "c0f1b2a4-3d5e-4b8c-8f7d-6a9e0f3b1c5d",
    eventId: 1,
  },
  {
    id: 3,
    membershipId: "b2a7e2b6-5c3f-4a1e-a0d5-8f3e1b2c3d4e",
    eventId: 1,
  },
];

export const ENDED_EVENT_PARTICIPANTS: Participant[] = [
  {
    id: 4,
    membershipId: "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d",
    eventId: 2,
    placement: 1,
    signedOutAt: "2025-01-10T23:30:00.000Z",
  },
  {
    id: 5,
    membershipId: "b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e",
    eventId: 2,
    placement: 2,
    signedOutAt: "2025-01-10T23:25:00.000Z",
  },
  {
    id: 6,
    membershipId: "c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f",
    eventId: 2,
    placement: 3,
    signedOutAt: "2025-01-10T23:00:00.000Z",
  },
  {
    id: 7,
    membershipId: "d4e5f6a7-b8c9-4d0e-1f2a-3b4c5d6e7f8a",
    eventId: 2,
    placement: 4,
    signedOutAt: "2025-01-10T22:30:00.000Z",
  },
  {
    id: 8,
    membershipId: "e5f6a7b8-c9d0-4e1f-2a3b-4c5d6e7f8a9b",
    eventId: 2,
    placement: 5,
    signedOutAt: "2025-01-10T22:00:00.000Z",
  },
];

export const RANKINGS: Ranking[] = [
  {
    membershipId: "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d",
    points: 4,
    attendance: 1,
  },
  {
    membershipId: "b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e",
    points: 3,
    attendance: 1,
  },
  {
    membershipId: "c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f",
    points: 3,
    attendance: 1,
  },
  {
    membershipId: "d4e5f6a7-b8c9-4d0e-1f2a-3b4c5d6e7f8a",
    points: 3,
    attendance: 1,
  },
  {
    membershipId: "e5f6a7b8-c9d0-4e1f-2a3b-4c5d6e7f8a9b",
    points: 2,
    attendance: 1,
  },
];
