import { useEffect, useRef, useState } from "react";
import { FaEllipsisV } from "react-icons/fa";
import { Spinner } from "@uwpokerclub/components";
import styles from "./DropdownMenu.module.css";

export interface DropdownMenuItem {
  key: string;
  label: string;
  icon?: React.ReactNode;
  onClick: () => void;
  disabled?: boolean;
}

interface DropdownMenuProps {
  items: DropdownMenuItem[];
  isLoading?: boolean;
}

export function DropdownMenu({ items, isLoading = false }: DropdownMenuProps) {
  const [isOpen, setIsOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  // Close menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Don't render if no items
  if (items.length === 0) {
    return null;
  }

  return (
    <div className={styles.menuWrapper} ref={menuRef}>
      <button
        type="button"
        className={styles.iconButton}
        onClick={() => setIsOpen(!isOpen)}
        title="Actions"
        aria-label="Actions"
        aria-expanded={isOpen}
        aria-haspopup="menu"
        disabled={isLoading}
      >
        {isLoading ? <Spinner size="sm" /> : <FaEllipsisV />}
      </button>

      {isOpen && (
        <div className={styles.dropdownMenu} role="menu">
          {items.map((item) =>
            item.disabled ? (
              <span
                key={item.key}
                className={`${styles.menuItem} ${styles.menuItemDisabled}`}
                role="menuitem"
                aria-disabled="true"
              >
                {item.icon} {item.label}
              </span>
            ) : (
              <button
                key={item.key}
                type="button"
                className={styles.menuItem}
                onClick={() => {
                  setIsOpen(false);
                  item.onClick();
                }}
                role="menuitem"
              >
                {item.icon} {item.label}
              </button>
            ),
          )}
        </div>
      )}
    </div>
  );
}
