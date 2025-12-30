import { FaHardHat } from "react-icons/fa";
import styles from "./ComingSoon.module.css";

export function ComingSoon() {
  return (
    <div className={styles.container} role="status" aria-live="polite">
      <div className={styles.card}>
        <div className={styles.badge}>NOT READY YET</div>

        <div className={styles.iconWrapper}>
          <FaHardHat className={styles.icon} />
        </div>

        <h1 className={styles.title}>Under Construction</h1>
        <p className={styles.message}>This page is currently being developed.</p>
      </div>
    </div>
  );
}
