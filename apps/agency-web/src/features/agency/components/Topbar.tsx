import { useState, type MouseEvent } from "react";
import {
  Avatar,
  Badge,
  Box,
  Divider,
  IconButton,
  Menu,
  MenuItem,
  Stack,
  Typography,
} from "@mui/material";
import { Bell, ChevronDown, LogOut, Menu as MenuIcon } from "lucide-react";
import { roleLabels, type AgencySession } from "@/app/session";
import type { NavItem } from "../navigation";
import { Eyebrow } from "./primitives";

export type AgencyNotification = {
  id: string;
  title: string;
  detail: string;
  tone: "gold" | "red" | "green";
};

function initials(name: string) {
  const parts = name.trim().split(/\s+/).slice(0, 2);
  return parts.map((part) => part[0]?.toUpperCase() ?? "").join("") || "NA";
}

export function Topbar({
  view,
  groupLabel,
  session,
  notifications,
  onSignOut,
  onOpenMobileNav,
}: {
  view: NavItem;
  groupLabel: string;
  session: AgencySession;
  notifications: AgencyNotification[];
  onSignOut: () => void;
  onOpenMobileNav: () => void;
}) {
  const [userAnchor, setUserAnchor] = useState<null | HTMLElement>(null);
  const [bellAnchor, setBellAnchor] = useState<null | HTMLElement>(null);
  const userOpen = Boolean(userAnchor);
  const bellOpen = Boolean(bellAnchor);

  const openUser = (event: MouseEvent<HTMLElement>) =>
    setUserAnchor(event.currentTarget);
  const openBell = (event: MouseEvent<HTMLElement>) =>
    setBellAnchor(event.currentTarget);

  const ViewIcon = view.icon;

  return (
    <header className="cc-topbar">
      <span className="cc-topbar__hairline" aria-hidden />
      <Stack direction="row" spacing={1.25} alignItems="center" minWidth={0}>
        <IconButton
          className="cc-topbar__hamburger"
          onClick={onOpenMobileNav}
          aria-label="Open agency sections"
        >
          <MenuIcon size={20} />
        </IconButton>
        <span className="cc-chip cc-topbar__chip" aria-hidden>
          <ViewIcon size={18} />
        </span>
        <Box minWidth={0}>
          <Eyebrow>{groupLabel}</Eyebrow>
          <Typography variant="h5" className="cc-topbar__title" noWrap>
            {view.label}
          </Typography>
          <Typography variant="caption" className="cc-topbar__desc" noWrap>
            {view.description}
          </Typography>
        </Box>
      </Stack>

      <Stack direction="row" spacing={1} alignItems="center">
        <IconButton
          onClick={openBell}
          aria-label={
            notifications.length
              ? `Notifications, ${notifications.length} active`
              : "Notifications"
          }
          aria-haspopup="true"
          aria-expanded={bellOpen}
          className="cc-topbar__bell"
        >
          <Badge
            color="error"
            badgeContent={notifications.length}
            invisible={notifications.length === 0}
            overlap="circular"
            className="cc-topbar__bell-badge"
          >
            <Bell size={19} />
          </Badge>
        </IconButton>

        <button
          type="button"
          className="cc-user"
          onClick={openUser}
          aria-haspopup="true"
          aria-expanded={userOpen}
        >
          <Avatar className="cc-user__avatar">{initials(session.name)}</Avatar>
          <span className="cc-user__meta">
            <span className="cc-user__name">{session.name}</span>
            <span className="cc-user__role">{roleLabels[session.role]}</span>
          </span>
          <ChevronDown size={16} className="cc-user__chevron" aria-hidden />
        </button>
      </Stack>

      <Menu
        anchorEl={bellAnchor}
        open={bellOpen}
        onClose={() => setBellAnchor(null)}
        anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
        transformOrigin={{ vertical: "top", horizontal: "right" }}
        slotProps={{ paper: { className: "cc-menu cc-menu--notifications" } }}
      >
        <li className="cc-menu__head">
          <Eyebrow>Operational notices</Eyebrow>
        </li>
        {notifications.length === 0 ? (
          <li className="cc-menu__empty">You are all caught up.</li>
        ) : (
          notifications.map((notice) => (
            <li className="cc-notice" key={notice.id}>
              <span className={`cc-notice__dot cc-notice__dot--${notice.tone}`} />
              <span className="cc-notice__body">
                <span className="cc-notice__title">{notice.title}</span>
                <span className="cc-notice__detail">{notice.detail}</span>
              </span>
            </li>
          ))
        )}
      </Menu>

      <Menu
        anchorEl={userAnchor}
        open={userOpen}
        onClose={() => setUserAnchor(null)}
        anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
        transformOrigin={{ vertical: "top", horizontal: "right" }}
        slotProps={{ paper: { className: "cc-menu cc-menu--user" } }}
      >
        <li className="cc-menu__identity">
          <Avatar className="cc-user__avatar cc-user__avatar--lg">
            {initials(session.name)}
          </Avatar>
          <span>
            <span className="cc-menu__name">{session.name}</span>
            <span className="cc-menu__sub">{roleLabels[session.role]}</span>
            <span className="cc-menu__sub">{session.agency}</span>
            <span className="cc-menu__id">{session.id}</span>
          </span>
        </li>
        <Divider component="li" />
        <MenuItem
          onClick={() => {
            setUserAnchor(null);
            onSignOut();
          }}
          className="cc-menu__signout"
        >
          <LogOut size={16} />
          Sign out
        </MenuItem>
      </Menu>
    </header>
  );
}
