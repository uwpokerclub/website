import { ReactNode } from "react";
import { NavLink as RRNavLink } from "react-router-dom";

import styles from "./NavLink.module.css";

type NavLinkProps = {
  icon: ReactNode;
  label: string;
  path: string;
  isCollapsed: boolean;
  onClick: () => void;
};

function NavLink({ icon, label, path, isCollapsed, onClick }: NavLinkProps) {
  return (
    <li className={styles.navItem}>
      <RRNavLink
        to={path}
        onClick={onClick}
        className={({ isActive }: { isActive: boolean }) => (isActive ? styles.active : undefined)}
      >
        <span className={styles.icon}>{icon}</span>
        <span className={`${styles.label} ${isCollapsed ? styles.collapsed : ""}`}>{label}</span>
      </RRNavLink>
    </li>
  );
}

export default NavLink;
