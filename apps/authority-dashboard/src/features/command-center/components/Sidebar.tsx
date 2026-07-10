import { Box, Tooltip, Typography } from "@mui/material";
import { PanelLeftClose, PanelLeftOpen } from "lucide-react";
import {
  navGroups,
  type BadgeKey,
  type NavItem,
  type ViewId,
} from "../navigation";

type SidebarProps = {
  /** Active section; "settings"/"guide" leave every rail item unhighlighted. */
  activeView: ViewId | "settings" | "guide";
  onSelect: (id: ViewId) => void;
  badges: Record<BadgeKey, number>;
  variant?: "rail" | "drawer";
  collapsed?: boolean;
  onToggleCollapse?: () => void;
};

function NavButton({
  item,
  active,
  compact,
  badge,
  onSelect,
}: {
  item: NavItem;
  active: boolean;
  compact: boolean;
  badge: number;
  onSelect: (id: ViewId) => void;
}) {
  const Icon = item.icon;
  const button = (
    <button
      type="button"
      className={`cc-nav__item${active ? " is-active" : ""}`}
      aria-current={active ? "page" : undefined}
      aria-label={compact ? item.label : undefined}
      onClick={() => onSelect(item.id)}
    >
      <span className="cc-nav__marker" aria-hidden />
      <span className="cc-nav__icon" aria-hidden>
        <Icon size={19} />
      </span>
      {!compact ? <span className="cc-nav__label">{item.label}</span> : null}
      {badge > 0 ? (
        <span
          className={`cc-nav__badge cc-nav__badge--${item.badgeTone ?? "gold"}${
            compact ? " cc-nav__badge--dot" : ""
          }`}
          aria-label={`${badge} pending`}
        >
          {compact ? "" : badge}
        </span>
      ) : null}
    </button>
  );

  if (compact) {
    return (
      <Tooltip title={item.label} placement="right" arrow>
        {button}
      </Tooltip>
    );
  }
  return button;
}

export function Sidebar({
  activeView,
  onSelect,
  badges,
  variant = "rail",
  collapsed = false,
  onToggleCollapse,
}: SidebarProps) {
  const compact = variant === "rail" && collapsed;

  return (
    <div className={`cc-rail${compact ? " is-compact" : ""}`}>
      <div className="cc-rail__brand">
        <Box
          component="img"
          src="/brand/nadaa-logo.png"
          alt=""
          className="cc-rail__logo"
        />
        {!compact ? (
          <div className="cc-rail__brand-text">
            <Typography component="span" className="cc-rail__wordmark">
              NADAA
            </Typography>
            <span className="cc-rail__brand-sub">Command</span>
          </div>
        ) : null}
      </div>

      <nav className="cc-nav" aria-label="Command sections">
        {navGroups.map((group) => (
          <div className="cc-nav__group" key={group.id}>
            {!compact ? (
              <p className="cc-nav__group-label">{group.label}</p>
            ) : (
              <span className="cc-nav__group-rule" aria-hidden />
            )}
            <ul className="cc-nav__list">
              {group.items.map((item) => (
                <li key={item.id}>
                  <NavButton
                    item={item}
                    active={item.id === activeView}
                    compact={compact}
                    badge={item.badgeKey ? badges[item.badgeKey] : 0}
                    onSelect={onSelect}
                  />
                </li>
              ))}
            </ul>
          </div>
        ))}
      </nav>

      <div className="cc-rail__foot">
        {!compact ? (
          <div className="cc-rail__status">
            <span className="cc-rail__pulse" aria-hidden />
            <span>Greater Accra · live desk</span>
          </div>
        ) : null}
        {variant === "rail" && onToggleCollapse ? (
          <button
            type="button"
            className="cc-rail__collapse"
            onClick={onToggleCollapse}
            aria-label={compact ? "Expand sidebar" : "Collapse sidebar"}
            aria-expanded={!compact}
          >
            {compact ? (
              <PanelLeftOpen size={18} />
            ) : (
              <>
                <PanelLeftClose size={18} />
                <span>Collapse</span>
              </>
            )}
          </button>
        ) : null}
      </div>
    </div>
  );
}
