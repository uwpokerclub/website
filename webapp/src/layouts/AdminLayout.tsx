import { ReactNode } from "react";
import SideNav from "@/components/SideNav/SideNav";
import { ErrorBoundary } from "@/components";
import styles from "./AdminLayout.module.css";

interface AdminLayoutProps {
  children: ReactNode;
}

const AdminLayout = ({ children }: AdminLayoutProps) => {
  return (
    <div className={styles.layoutContainer}>
      <SideNav />
      <main className={styles.mainContent}>
        <ErrorBoundary>{children}</ErrorBoundary>
      </main>
    </div>
  );
};

export default AdminLayout;
