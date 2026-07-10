import type { ReactNode } from "react";

type PageBannerProps = {
  eyebrow: string;
  title: string;
  subtitle?: string;
  children?: ReactNode;
};

/** uposa-style navy banner band with a gold hairline and radar watermark. */
export function PageBanner({
  eyebrow,
  title,
  subtitle,
  children,
}: PageBannerProps) {
  return (
    <section className="page-banner">
      <div className="page-banner-inner">
        <p className="eyebrow">{eyebrow}</p>
        <h1>{title}</h1>
        {subtitle ? <p className="page-banner-sub">{subtitle}</p> : null}
        {children}
      </div>
    </section>
  );
}
