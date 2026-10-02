import { Link } from "react-router-dom";
import { useAuth } from "@/hooks";
import type { Actions, Resources, SubResources } from "@/interfaces/responses";
import { DashboardCard } from "../DashboardCard";
import { CARD_TITLES } from "../../dashboardLayout";
import styles from "./QuickActionsCard.module.css";

type Action = {
  key: string;
  label: string;
  to: string;
  action: Actions;
  resource: Resources;
  subResource?: SubResources;
};

/**
 * Labelled by destination, not by verb.
 *
 * Creating an event or adding a member happens in a modal on the relevant list page,
 * and no URL opens one — Events routes only "/" and "/:eventId". A tile reading
 * "Create event" that lands on a list would name something the link does not do, so
 * these name where they go. If the create modals ever gain deep links, the labels
 * become the verbs.
 */
const ACTIONS: Action[] = [
  { key: "events", label: "Events", to: "/admin/events", action: "list", resource: "event" },
  { key: "members", label: "Members", to: "/admin/members", action: "list", resource: "membership" },
  {
    key: "rankings",
    label: "Rankings",
    to: "/admin/rankings",
    action: "list",
    resource: "semester",
    subResource: "rankings",
  },
  { key: "logins", label: "Logins", to: "/admin/logins", action: "list", resource: "login" },
];

/**
 * The only card with no data. It shows none — no figure, no chart, no delta — which
 * is what distinguishes it on a page where everything else is a measurement.
 *
 * It takes no props: a parameterless component is assignable to
 * ComponentType<DashboardCardComponentProps>, so it registers like the others
 * without carrying a semesterId it never reads.
 */
export function QuickActionsCard() {
  const { hasPermission } = useAuth();
  const available = ACTIONS.filter((item) => hasPermission(item.action, item.resource, item.subResource));

  return (
    <DashboardCard
      title={CARD_TITLES.quickActions}
      status={available.length === 0 ? "empty" : "ready"}
      emptyMessage="No actions available for your role."
      data-qa="quick-actions-card"
    >
      {() => (
        <div className={styles.grid}>
          {available.map((item) => (
            <Link key={item.key} to={item.to} className={styles.action}>
              {item.label}
            </Link>
          ))}
        </div>
      )}
    </DashboardCard>
  );
}
