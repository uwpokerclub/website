import { ComingSoon } from "@/components";
import { SemesterSetupPrompt } from "@/features/semesters";

export function Dashboard() {
  return (
    <>
      <SemesterSetupPrompt />
      <ComingSoon />
    </>
  );
}
