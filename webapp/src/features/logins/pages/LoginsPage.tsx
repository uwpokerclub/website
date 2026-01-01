import { LoginsList } from "../components/LoginsList";
import styles from "./LoginsPage.module.css";

export function LoginsPage() {
  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <h1>Manage Logins</h1>
        <p className={styles.subtitle}>Create and manage login accounts for club members and officers</p>
      </div>
      <LoginsList />
    </div>
  );
}
