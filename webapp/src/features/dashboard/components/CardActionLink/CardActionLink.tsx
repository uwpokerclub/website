import { ReactNode } from "react";
import { Link } from "react-router-dom";
import styles from "./CardActionLink.module.css";

type CardActionLinkProps = { children: ReactNode } & ({ to: string; href?: never } | { href: string; to?: never });

export function CardActionLink({ children, ...props }: CardActionLinkProps) {
  if (props.to) {
    return (
      <Link to={props.to} className={styles.actionLink}>
        {children}
      </Link>
    );
  }

  return (
    <a href={props.href} className={styles.actionLink}>
      {children}
    </a>
  );
}
