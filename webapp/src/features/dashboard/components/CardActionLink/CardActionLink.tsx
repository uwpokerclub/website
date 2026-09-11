import { ReactNode } from "react";
import { Link } from "react-router-dom";
import styles from "./CardActionLink.module.css";

type CardActionLinkProps = {
  to?: string;
  href?: string;
  children: ReactNode;
};

export function CardActionLink({ to, href, children }: CardActionLinkProps) {
  if (to) {
    return (
      <Link to={to} className={styles.link}>
        {children}
      </Link>
    );
  }

  return (
    <a href={href} className={styles.link}>
      {children}
    </a>
  );
}
