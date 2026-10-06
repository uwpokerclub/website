import { DashboardGrid } from "@/features/dashboard/components";
import { DASHBOARD_CARDS } from "@/features/dashboard/dashboardCards";
import { resolveDashboardLayout } from "@/features/dashboard/dashboardLayout";
import { placeDashboardCards } from "@/features/dashboard/dashboardRows";
import { SemesterSetupPrompt } from "@/features/semesters";
import { useAuth, useCurrentSemester } from "@/hooks";
import styles from "./Dashboard.module.css";

export function Dashboard() {
  const { user } = useAuth();
  const { currentSemester } = useCurrentSemester();

  if (!currentSemester) {
    return (
      <>
        <SemesterSetupPrompt />
        <div className={styles.page} data-qa="dashboard-no-semester">
          <div className={styles.header}>
            <h1>Dashboard</h1>
          </div>
          <div className={styles.noSemester}>
            <p>Please select a semester to view the dashboard.</p>
          </div>
        </div>
      </>
    );
  }

  // DASHBOARD_CARDS is a total Record, so every id has a component and there is no
  // placeholder branch left to fall through to.
  const items = placeDashboardCards(resolveDashboardLayout(user?.role)).map((placed) => {
    const CardComponent = DASHBOARD_CARDS[placed.id];

    return { ...placed, content: <CardComponent semesterId={currentSemester.id} /> };
  });

  return (
    <>
      <SemesterSetupPrompt />
      <div className={styles.page} data-qa="dashboard-page">
        <div className={styles.header}>
          <h1>Dashboard</h1>
          <p className={styles.subtitle}>{currentSemester.name}</p>
        </div>
        <DashboardGrid data-qa="dashboard-grid" items={items} />
      </div>
    </>
  );
}
