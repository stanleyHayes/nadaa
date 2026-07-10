import { useEffect, useState } from "react";
import { Box, Tooltip, Typography } from "@mui/material";
import { ChevronDown, PanelLeftClose, PanelLeftOpen } from "lucide-react";
import {
  groupIdForView,
  navGroups,
  type BadgeKey,
  type NavGroup,
  type NavItem,
  type ViewId,
} from "../navigation";

type SidebarProps = {
  /** The active section; "settings" leaves every nav item unhighlighted. */
  activeView: ViewId | "settings";
  onSelect: (id: ViewId) => void;
  badges: Record<BadgeKey, number>;
  variant?: "rail" | "drawer";
  collapsed?: boolean;
  onToggleCollapse?: () => void;
};

const GROUPS_KEY = "nadaa.agency.nav.groups";

function readOpenGroups(): Record<string, boolean> {
  const base: Record<string, boolean> = {};
  for (const group of navGroups) {
    base[group.id] = true;
  }
  if (typeof window === "undefined") {
    return base;
  }
  try {
    const raw = window.localStorage.getItem(GROUPS_KEY);
    if (!raw) {
      return base;
    }
    const parsed = JSON.parse(raw) as Record<string, boolean>;
    return { ...base, ...parsed };
  } catch {
    return base;
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
  open,
  onToggle,
  onSelect,
}: {
  group: NavGroup;
  activeView: ViewId | "settings";
  badges: Record<BadgeKey, number>;
  open: boolean;
  onToggle: () => void;
  onSelect: (id: ViewId) => void;
}) {
  const GroupIcon = group.icon;
  const panelId = `cc-nav-group-${group.id}`;
  return (
    <div
      className={`cc-nav__group cc-nav__group--${group.accent}${open ? " is-open" : ""}`}
    >
      <button
        type="button"
        className="cc-nav__group-head"
        aria-expanded={open}
        aria-controls={panelId}
        onClick={onToggle}
      >
        <span className="cc-chip cc-nav__group-chip" aria-hidden>
          <GroupIcon size={14} />
        </span>
        <span className="cc-nav__group-label">{group.label}</span>
        <ChevronDown size={15} className="cc-nav__group-chevron" aria-hidden />
      </button>
      <ul className="cc-nav__list" id={panelId} hidden={!open}>
        {group.items.map((item) => (
          <li key={item.id}>
            <NavButton
              item={item}
              active={item.id === activeView}
              compact={false}
              badge={item.badgeKey ? badges[item.badgeKey] : 0}
              onSelect={onSelect}
            />
          </li>
        ))}
      </ul>
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
  const [openGroups, setOpenGroups] =
    useState<Record<string, boolean>>(readOpenGroups);

  // The group containing the active route always auto-opens. The settings view
  // sits outside the nav tree, so it leaves the open groups untouched.
  useEffect(() => {
    if (activeView === "settings") {
      return;
    }
    const activeGroup = groupIdForView(activeView);
    setOpenGroups((current) =>
      current[activeGroup] ? current : { ...current, [activeGroup]: true },
    );
  }, [activeView]);

  const toggleGroup = (id: string) => {
    setOpenGroups((current) => {
      const next = { ...current, [id]: !current[id] };
      try {
        window.localStorage.setItem(GROUPS_KEY, JSON.stringify(next));
      } catch {
        /* storage unavailable */
      }
      return next;
    });
  };

  return (
    <div className={`cc-rail${compact ? " is-compact" : ""}`}>
      <span className="cc-rail__hairline" aria-hidden />

      <div className="cc-rail__brand">
        <span className="cc-chip cc-rail__brand-chip" aria-hidden>
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
            <span className="cc-rail__brand-sub">Agency</span>
          </div>
        ) : null}
      </div>

      <nav className="cc-nav" aria-label="Agency sections">
        {compact
          ? navGroups.map((group) => (
              <div className="cc-nav__group is-compact" key={group.id}>
                <span className="cc-nav__group-rule" aria-hidden />
                <ul className="cc-nav__list">
                  {group.items.map((item) => (
                    <li key={item.id}>
                      <NavButton
                        item={item}
                        active={item.id === activeView}
                        compact
                        badge={item.badgeKey ? badges[item.badgeKey] : 0}
                        onSelect={onSelect}
                      />
                    </li>
                  ))}
                </ul>
              </div>
            ))
          : navGroups.map((group) => (
              <NavGroupSection
                key={group.id}
                group={group}
                activeView={activeView}
                badges={badges}
                open={openGroups[group.id] ?? true}
                onToggle={() => toggleGroup(group.id)}
                onSelect={onSelect}
              />
            ))}
      </nav>

      <div className="cc-rail__foot">
        {!compact ? (
          <div className="cc-rail__status">
            <span className="cc-rail__pulse" aria-hidden />
            <span>Greater Accra · field desk</span>
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
