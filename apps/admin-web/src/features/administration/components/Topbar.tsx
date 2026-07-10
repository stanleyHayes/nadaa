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
import {
  Bell,
  ChevronDown,
  LogOut,
  Menu as MenuIcon,
  Settings,
  UserRound,
  type LucideIcon,
} from "lucide-react";
import { type AdminSession } from "@/app/session";
import type { NavItem } from "../navigation";
import type { SettingsTab } from "../account";
import { roleLabel } from "../utils";
import { initials } from "../account/primitives";
import { Eyebrow } from "./primitives";

export type AdminNotification = {
  id: string;
  title: string;
  detail: string;
  tone: "gold" | "red" | "green";
};

function MenuRow({
  icon: Icon,
  title,
  description,
}: {
  icon: LucideIcon;
  title: string;
  description: string;
}) {
  return (
    <>
      <Icon size={16} aria-hidden style={{ marginTop: 2, flex: "0 0 auto" }} />
      <span style={{ display: "flex", minWidth: 0, flexDirection: "column" }}>
        <span className="cc-menu__name" style={{ fontSize: "0.88rem" }}>
          {title}
        </span>
        <span className="cc-menu__sub">{description}</span>
      </span>
    </>
  );
}

export function Topbar({
  view,
  groupLabel,
  session,
  notifications,
  onSignOut,
  onOpenSettings,
  onOpenMobileNav,
}: {
  view: NavItem;
  groupLabel: string;
  session: AdminSession;
  notifications: AdminNotification[];
  onSignOut: () => void;
  onOpenSettings: (tab: SettingsTab) => void;
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

  return (
    <header className="cc-topbar">
      <Stack direction="row" spacing={1.25} alignItems="center" minWidth={0}>
        <IconButton
          className="cc-topbar__hamburger"
          onClick={onOpenMobileNav}
          aria-label="Open admin sections"
        >
          <MenuIcon size={20} />
        </IconButton>
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
            <span className="cc-user__role">{roleLabel(session.role)}</span>
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
          <Eyebrow>Governance notices</Eyebrow>
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
            <span className="cc-menu__sub">{roleLabel(session.role)}</span>
            <span className="cc-menu__sub">{session.agency}</span>
            <span className="cc-menu__id">{session.id}</span>
          </span>
        </li>
        <Divider component="li" />
        <MenuItem
          onClick={() => {
            setUserAnchor(null);
            onOpenSettings("profile");
          }}
          sx={{ alignItems: "flex-start", gap: 1.5, px: 2, py: 1.25 }}
        >
          <MenuRow
            icon={UserRound}
            title="Profile"
            description="View and edit your account details"
          />
        </MenuItem>
        <MenuItem
          onClick={() => {
            setUserAnchor(null);
            onOpenSettings("security");
          }}
          sx={{ alignItems: "flex-start", gap: 1.5, px: 2, py: 1.25 }}
        >
          <MenuRow
            icon={Settings}
            title="Settings"
            description="Security, notifications, and preferences"
          />
        </MenuItem>
        <Divider component="li" />
        <MenuItem
          onClick={() => {
            setUserAnchor(null);
            onSignOut();
          }}
          sx={{
            alignItems: "flex-start",
            gap: 1.5,
            px: 2,
            py: 1.25,
            color: "var(--nadaa-red, #e53935)",
            "& .cc-menu__name": { color: "inherit" },
          }}
        >
          <MenuRow
            icon={LogOut}
            title="Sign out"
            description="End your session on this device"
          />
        </MenuItem>
      </Menu>
    </header>
  );
}
