import { RankingsPage } from "../features/rankings";
import { RequirePermission } from "@/components";

export function Rankings() {
  return (
    <RequirePermission resource="semester" subResource="rankings" action="list">
      <RankingsPage />
    </RequirePermission>
  );
}
