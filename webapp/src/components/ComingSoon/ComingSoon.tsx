import { FaHardHat } from "react-icons/fa";
import styles from "./ComingSoon.module.css";

export function ComingSoon({ headingLevel: Heading = "h1" }: { headingLevel?: "h1" | "h2" | "h3" }) {
  return (
    <div className={styles.container} role="status" aria-live="polite">
      <div className={styles.card}>
        <div className={styles.badge}>NOT READY YET</div>

        <div className={styles.iconWrapper}>
          <FaHardHat className={styles.icon} />
        </div>

        <Heading className={styles.title}>Under Construction</Heading>
        <p className={styles.message}>This page is currently being developed.</p>
      </div>
    </div>
  );
}
