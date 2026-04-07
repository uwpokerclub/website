const FACULTIES = {
  AHS: "AHS",
  ARTS: "Arts",
  ENGINEERING: "Engineering",
  ENVIRONMENT: "Environment",
  MATH: "Math",
  SCIENCE: "Science",
} as const;

export type Faculty = (typeof FACULTIES)[keyof typeof FACULTIES];

export const FACULTY_VALUES = Object.values(FACULTIES) as [Faculty, ...Faculty[]];
