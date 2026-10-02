import { useAuth, useCurrentSemester } from "@/hooks";
import { DashboardGrid } from "@/features/dashboard/components";
import { DashboardCardId, resolveDashboardLayout } from "@/features/dashboard/dashboardLayout";
import { assignLanes } from "@/features/dashboard/lanes";
import { DASHBOARD_CARDS } from "@/features/dashboard/dashboardCards";
import styles from "./Dashboard.module.css";

export function Dashboard() {
  const { user } = useAuth();
  const { currentSemester } = useCurrentSemester();

  if (!currentSemester) {
    return (
      <div className={styles.page} data-qa="dashboard-no-semester">
        <div className={styles.header}>
          <h1>Dashboard</h1>
        </div>
        <div className={styles.noSemester}>
          <p>Please select a semester to view the dashboard.</p>
        </div>
      </div>
    );
  }

  const { lead, wide, rail } = assignLanes(resolveDashboardLayout(user?.role));

  // DASHBOARD_CARDS is a total Record, so every id has a component and there is no
  // placeholder branch left to fall through to.
  const renderCard = (cardId: DashboardCardId) => {
    const CardComponent = DASHBOARD_CARDS[cardId];

    return <CardComponent key={cardId} semesterId={currentSemester.id} />;
  };

  return (
    <div className={styles.page} data-qa="dashboard-page">
      <div className={styles.header}>
        <h1>Dashboard</h1>
        <p className={styles.subtitle}>{currentSemester.name}</p>
      </div>
      <DashboardGrid
        data-qa="dashboard-grid"
        lead={renderCard(lead)}
        wide={wide.map((id) => renderCard(id))}
        rail={rail.map((id) => renderCard(id))}
      />
    </div>
  );
}
