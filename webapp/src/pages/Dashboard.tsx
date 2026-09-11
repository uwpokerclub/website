import { useAuth, useCurrentSemester } from "@/hooks";
import { DashboardCard, DashboardGrid } from "@/features/dashboard/components";
import { CARD_TITLES, resolveDashboardLayout } from "@/features/dashboard/dashboardLayout";
import { DASHBOARD_CARDS } from "@/features/dashboard/dashboardCards";
import styles from "./Dashboard.module.css";

export function Dashboard() {
  const { user } = useAuth();
  const { currentSemester } = useCurrentSemester();

  if (!currentSemester) {
    return (
      <div className={styles.page} data-qa="dashboard-no-semester">
        <div className={styles.noSemester}>
          <p>Please select a semester to view the dashboard.</p>
        </div>
      </div>
    );
  }

  const layout = resolveDashboardLayout(user?.role);

  return (
    <div className={styles.page} data-qa="dashboard-page">
      <DashboardGrid data-qa="dashboard-grid">
        {layout.map((cardId) => {
          const CardComponent = DASHBOARD_CARDS[cardId];

          if (CardComponent) {
            return <CardComponent key={cardId} semesterId={currentSemester.id} />;
          }

          return (
            <DashboardCard key={cardId} title={CARD_TITLES[cardId]} status="ready" data-qa={`dashboard-card-${cardId}`}>
              {() => <p className={styles.placeholder}>Coming soon.</p>}
            </DashboardCard>
          );
        })}
      </DashboardGrid>
    </div>
  );
}
