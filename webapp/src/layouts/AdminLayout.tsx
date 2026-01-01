import { ReactNode } from "react";
import SideNav from "@/components/SideNav/SideNav";
import styles from "./AdminLayout.module.css";

interface AdminLayoutProps {
  children: ReactNode;
}

const AdminLayout = ({ children }: AdminLayoutProps) => {
  return (
    <div className={styles.layoutContainer}>
      <SideNav />
      <main className={styles.mainContent}>{children}</main>
    </div>
  );
};

export default AdminLayout;
