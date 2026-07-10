import { useState } from "react";
import { Box, Tooltip, Typography } from "@mui/material";
import { ChevronDown, PanelLeftClose, PanelLeftOpen } from "lucide-react";
import {
  navGroups,
  type BadgeKey,
  type NavGroup,
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

const GROUP_STATE_KEY = "nadaa.admin.nav.groups";

function readGroupState(): Record<string, boolean> {
  if (typeof window === "undefined") {
    return {};
  }
  try {
    const raw = window.localStorage.getItem(GROUP_STATE_KEY);
    return raw ? (JSON.parse(raw) as Record<string, boolean>) : {};
  } catch {
    return {};
  }
}

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

function NavGroupSection({
  group,
  activeView,
  badges,
  compact,
  open,
  onToggle,
  onSelect,
}: {
  group: NavGroup;
  activeView: ViewId | "settings" | "guide";
  badges: Record<BadgeKey, number>;
  compact: boolean;
  open: boolean;
  onToggle: () => void;
  onSelect: (id: ViewId) => void;
}) {
  const GroupIcon = group.icon;
  const list = (
    <ul className="cc-nav__list">
      {group.items.map((item) => (
        <li key={item.id} className={item.id === activeView ? "is-active" : ""}>
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
  );

  if (compact) {
    return (
      <div className="cc-nav__group">
        <span className="cc-nav__group-rule" aria-hidden />
        {list}
      </div>
    );
  }

  return (
    <div className={`cc-nav__group${open ? " is-open" : ""}`}>
      <button
        type="button"
        className="cc-nav__group-head"
        aria-expanded={open}
        onClick={onToggle}
      >
        <span className="cc-nav__group-chip" aria-hidden>
          <GroupIcon size={15} />
        </span>
        <span className="cc-nav__group-label">{group.label}</span>
        <ChevronDown size={15} className="cc-nav__group-caret" aria-hidden />
      </button>
      {open ? list : null}
    </div>
  );
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
  const [groupState, setGroupState] = useState<Record<string, boolean>>(
    readGroupState,
  );

  const toggleGroup = (id: string, isOpen: boolean) => {
    setGroupState((current) => {
      const next = { ...current, [id]: !isOpen };
      try {
        window.localStorage.setItem(GROUP_STATE_KEY, JSON.stringify(next));
      } catch {
        /* storage unavailable */
      }
      return next;
    });
  };

  return (
    <div className={`cc-rail${compact ? " is-compact" : ""}`}>
      <div className="cc-rail__brand">
        <span className="cc-rail__brand-chip" aria-hidden>
          <Box
            component="img"
            src="/brand/nadaa-logo.png"
            alt=""
            className="cc-rail__logo"
          />
        </span>
        {!compact ? (
          <div className="cc-rail__brand-text">
            <Typography component="span" className="cc-rail__wordmark">
              NADAA
            </Typography>
            <span className="cc-rail__brand-sub">Governance</span>
          </div>
        ) : null}
      </div>

      <nav className="cc-nav" aria-label="Admin sections">
        {navGroups.map((group) => {
          const hasActive = group.items.some((item) => item.id === activeView);
          // Active group always opens; otherwise honour the stored preference
          // (default open). Persistence lives in localStorage.
          const open = hasActive || groupState[group.id] !== false;
          return (
            <NavGroupSection
              key={group.id}
              group={group}
              activeView={activeView}
              badges={badges}
              compact={compact}
              open={open}
              onToggle={() => toggleGroup(group.id, open)}
              onSelect={onSelect}
            />
          );
        })}
      </nav>

      <div className="cc-rail__foot">
        {!compact ? (
          <div className="cc-rail__status">
            <span className="cc-rail__pulse" aria-hidden />
            <span>National desk · governance</span>
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
