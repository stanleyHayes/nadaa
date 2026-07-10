import type { ReactNode } from "react";
import type { LucideIcon } from "lucide-react";
import { Link } from "react-router-dom";

/** Citizen initials for the avatar, e.g. "Ama Boateng" -> "AB". */
export function initialsOf(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) {
    return "?";
  }
  const first = parts[0][0] ?? "";
  const last = parts.length > 1 ? (parts[parts.length - 1][0] ?? "") : "";
  return (first + last).toUpperCase();
}

type StatCardProps = {
  icon: LucideIcon;
  label: string;
  value: ReactNode;
  hint?: string;
  tone?: "gold" | "navy" | "green" | "red";
};

/** Compact quick-stat tile for the account overview. */
export function StatCard({
  icon: Icon,
  label,
  value,
  hint,
  tone = "navy",
}: StatCardProps) {
  return (
    <div className="account-stat" data-tone={tone}>
      <span className="account-stat__icon" aria-hidden="true">
        <Icon size={20} strokeWidth={2.2} />
      </span>
      <div className="account-stat__body">
        <span className="account-stat__value">{value}</span>
        <span className="account-stat__label">{label}</span>
        {hint ? <span className="account-stat__hint">{hint}</span> : null}
      </div>
    </div>
  );
}

type ShortcutCardProps = {
  icon: LucideIcon;
  title: string;
  description: string;
  to: string;
  /** External-style href (e.g. tel:) instead of a router link. */
  href?: string;
};

/** Action shortcut used in the overview grid. */
export function ShortcutCard({
  icon: Icon,
  title,
  description,
  to,
  href,
}: ShortcutCardProps) {
  const inner = (
    <>
      <span className="account-shortcut__icon" aria-hidden="true">
        <Icon size={20} strokeWidth={2.2} />
      </span>
      <span className="account-shortcut__text">
        <strong>{title}</strong>
        <small>{description}</small>
      </span>
    </>
  );

  if (href) {
    return (
      <a className="account-shortcut" href={href}>
        {inner}
      </a>
    );
  }

  return (
    <Link className="account-shortcut" to={to}>
      {inner}
    </Link>
  );
}
