import { NavLink } from "react-router-dom";
import { accountNavItems } from "../data";

type AccountSidenavProps = {
  /** Unread notification count, shown as a badge on the Notifications item. */
  unread: number;
};

const linkClass = ({ isActive }: { isActive: boolean }) =>
  isActive ? "account-nav__link is-active" : "account-nav__link";

/**
 * Account sub-navigation. On desktop it is a sticky vertical rail; below the
 * layout breakpoint it collapses to a horizontally scrollable row of the same
 * pills (see `.account-nav` in global.css).
 */
export function AccountSidenav({ unread }: AccountSidenavProps) {
  return (
    <nav aria-label="Account sections" className="account-nav">
      {accountNavItems.map(({ to, label, icon: Icon, end }) => {
        const isNotifications = to === "/account/notifications";
        return (
          <NavLink className={linkClass} end={end} key={to} to={to}>
            <Icon aria-hidden="true" size={18} />
            <span className="account-nav__label">{label}</span>
            {isNotifications && unread > 0 ? (
              <span className="account-nav__badge" aria-hidden="true">
                {unread}
              </span>
            ) : null}
            {isNotifications && unread > 0 ? (
              <span className="visually-hidden">{unread} unread</span>
            ) : null}
          </NavLink>
        );
      })}
    </nav>
  );
}

export default AccountSidenav;
