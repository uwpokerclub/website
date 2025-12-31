import { useState } from "react";
import styles from "./SideNav.module.css";
import {
  FaBars,
  FaTimes,
  FaUsers,
  FaCalendarAlt,
  FaTrophy,
  FaChartLine,
  FaBoxOpen,
  FaMoneyBillWave,
  FaUserTie,
  FaChevronLeft,
  FaChevronRight,
} from "react-icons/fa";
import logoSvg from "@/assets/uwpsc_logo.svg";
import crestSvg from "@/assets/crest.svg";
import NavLink from "./NavLink";
import UserProfile from "./UserProfile";
import SemesterSelector from "./SemesterSelector";
import { useAuth } from "@/hooks";
import { ROLES } from "@/types/roles";

function SideNav() {
  const { hasRoles } = useAuth();

  // Initialize isExpanded based on screen size - expanded on desktop, collapsed on mobile
  const [isExpanded, setIsExpanded] = useState(() => {
    // Only run on client side
    if (typeof window !== "undefined") {
      // If desktop (window width > 768px), then true (expanded)
      // If mobile (window width <= 768px), then false (collapsed)
      return window.innerWidth > 768;
    }
    // Default for SSR
    return true;
  });

  // Main navigation items with Dashboard as the first item
  const mainNavItems = [
    { path: "/admin/dashboard", label: "Dashboard", icon: <FaChartLine /> },
    { path: "/admin/members", label: "Members", icon: <FaUsers /> },
    { path: "/admin/events", label: "Events", icon: <FaCalendarAlt /> },
    { path: "/admin/rankings", label: "Rankings", icon: <FaTrophy /> },
  ];

  // Additional navigation items in a separate section
  const additionalNavItems = [
    { path: "/admin/inventory", label: "Inventory", icon: <FaBoxOpen /> },
    { path: "/admin/finances", label: "Finances", icon: <FaMoneyBillWave /> },
    { path: "/admin/executive", label: "Executive Team", icon: <FaUserTie /> },
  ];

  // Toggle nav - on mobile this opens the sidebar, on desktop it collapses/expands
  const toggleNav = () => {
    setIsExpanded(!isExpanded);
  };

  // Only close the nav when clicking links on mobile
  const handleNavLinkClick = () => {
    if (window.innerWidth <= 768) {
      setIsExpanded(false);
    }
  };

  // Determine CSS classes based only on expanded state
  const sidenavClasses = `${styles.sidenav} ${!isExpanded ? styles.collapsed : ""}`;

  // Function to render navigation items using NavLink
  const renderNavItems = (items: typeof mainNavItems) => {
    return items.map((item) => (
      <NavLink
        key={item.path}
        icon={item.icon}
        label={item.label}
        path={item.path}
        isCollapsed={!isExpanded}
        onClick={handleNavLinkClick}
      />
    ));
  };

  return (
    <div className={styles.sidenavContainer}>
      {/* Main navigation */}
      <nav className={sidenavClasses} data-qa="sidenav">
        <div className={styles.sidenavHeader}>
          <div className={styles.logoContainer}>
            <img
              src={logoSvg}
              alt="UW Poker Club Logo"
              className={`${styles.logo} ${!isExpanded ? styles.hidden : ""}`}
            />
            <img
              src={crestSvg}
              alt="UW Poker Club Crest"
              className={`${styles.crest} ${isExpanded ? styles.hidden : ""}`}
            />
          </div>
        </div>

        {/* User profile section */}
        <UserProfile isExpanded={isExpanded} />

        {/* Semester selector */}
        <SemesterSelector isExpanded={isExpanded} onIconClick={toggleNav} />

        {/* Scrollable nav section */}
        <div className={styles.navScrollContainer}>
          {/* Main navigation links */}
          <ul className={styles.navLinks}>{renderNavItems(mainNavItems)}</ul>

          {hasRoles([ROLES.WEBMASTER, ROLES.PRESIDENT, ROLES.VICE_PRESIDENT, ROLES.TREASURER, ROLES.SECRETARY]) && (
            <>
              {/* Additional navigation section with separator */}
              <div className={styles.navSeparator} data-qa="sidenav-officers-section">
                <span>Officers</span>
              </div>

              {/* Additional navigation links */}
              <ul className={styles.navLinks}>{renderNavItems(additionalNavItems)}</ul>
            </>
          )}
        </div>

        {/* Desktop toggle button */}
        <div className={styles.navFooter}>
          <button
            className={styles.toggleBtn}
            onClick={toggleNav}
            aria-label="Toggle navigation"
            data-qa="sidenav-toggle"
          >
            {isExpanded ? <FaChevronLeft /> : <FaChevronRight />}
          </button>
        </div>
      </nav>

      {/* Mobile-only elements - CSS will hide/show these based on screen size */}
      <button
        className={styles.mobileToggle}
        onClick={() => setIsExpanded(true)}
        aria-label="Open navigation"
        data-qa="sidenav-mobile-open"
      >
        <FaBars />
      </button>

      {/* Overlay for clicking outside to close nav on mobile */}
      <div
        className={`${styles.overlay} ${isExpanded ? styles.visible : ""}`}
        onClick={() => setIsExpanded(false)}
        data-qa="sidenav-overlay"
      ></div>

      {/* Mobile close button */}
      <button
        className={styles.closeBtn}
        onClick={() => setIsExpanded(false)}
        aria-label="Close navigation"
        data-qa="sidenav-mobile-close"
      >
        <FaTimes />
      </button>
    </div>
  );
}

export default SideNav;
