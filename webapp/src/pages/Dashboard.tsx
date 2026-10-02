import { useAuth, useCurrentSemester } from "@/hooks";
import { DashboardCard, DashboardGrid } from "@/features/dashboard/components";
import { CARD_TITLES, DashboardCardId, resolveDashboardLayout } from "@/features/dashboard/dashboardLayout";
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

  const renderCard = (cardId: DashboardCardId) => {
    const CardComponent = DASHBOARD_CARDS[cardId];

    if (!CardComponent) {
      return (
        <DashboardCard key={cardId} title={CARD_TITLES[cardId]} status="ready" data-qa={`dashboard-card-${cardId}`}>
          {() => <p className={styles.placeholder}>Coming soon.</p>}
        </DashboardCard>
      );
    }

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
