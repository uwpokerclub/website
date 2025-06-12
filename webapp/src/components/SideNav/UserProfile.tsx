import { FaSignOutAlt, FaUserCircle } from "react-icons/fa";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks";

import styles from "./UserProfile.module.css";
import prettyPrintRole from "@/utils/prettyPrintRole";

type UserProfileProps = {
  isExpanded: boolean;
};

function UserProfile({ isExpanded }: UserProfileProps) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout(() => {
      navigate("/admin/login");
    });
  };

  if (!isExpanded) {
    return (
      <div className={styles.userProfileCollapsed}>
        <div className={styles.userAvatarCollapsed}>
          <FaUserCircle />
        </div>
        <button className={styles.logoutBtnCollapsed} onClick={handleLogout} aria-label="Logout">
          <FaSignOutAlt />
        </button>
      </div>
    );
  }

  return (
    <div className={styles.userProfile}>
      <div className={styles.userAvatar}>
        <FaUserCircle />
      </div>
      <div className={styles.userInfo}>
        <span className={styles.userName}>{user!.username}</span>
        <span className={styles.userRole}>{prettyPrintRole(user!.role)}</span>
      </div>
      <button className={styles.logoutBtn} onClick={handleLogout} aria-label="Logout">
        <FaSignOutAlt />
      </button>
    </div>
  );
}

export default UserProfile;
