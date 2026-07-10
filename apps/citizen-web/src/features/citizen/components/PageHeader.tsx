import type { ReactNode } from "react";
import type { LucideIcon } from "lucide-react";

type PageHeaderTone = "gold" | "navy" | "green" | "red";

type PageHeaderProps = {
  icon: LucideIcon;
  title: string;
  subtitle?: ReactNode;
  /** Optional right-aligned control (e.g. a Refresh button). */
  action?: ReactNode;
  /** Plaque tint. Public surfaces default to gold. */
  tone?: PageHeaderTone;
  /** Optional heading level for the title element. Defaults to h2. */
  as?: "h1" | "h2" | "h3";
};

/**
 * Shared public page/section header: a tinted icon plaque, title, optional
 * subtitle and action, plus a faint watermark of the same icon. Gives every
 * citizen panel one consistent, welcoming heading treatment.
 */
export function PageHeader({
  icon: Icon,
  title,
  subtitle,
  action,
  tone = "gold",
  as: Heading = "h2",
}: PageHeaderProps) {
  return (
    <div className="page-header" data-tone={tone}>
      <Icon className="page-header__watermark" size={116} aria-hidden="true" />
      <span className="page-header__plaque" aria-hidden="true">
        <Icon size={22} strokeWidth={2.2} />
      </span>
      <div className="page-header__text">
        <Heading className="page-header__title">{title}</Heading>
        {subtitle ? <p className="page-header__subtitle">{subtitle}</p> : null}
      </div>
      {action ? <div className="page-header__action">{action}</div> : null}
    </div>
  );
}

export default PageHeader;
