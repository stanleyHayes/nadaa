import type { ReactNode } from "react";

type PageBannerProps = {
  eyebrow: string;
  title: string;
  subtitle?: string;
  children?: ReactNode;
};

/**
 * uposa-style navy banner band: a navy-to-green gradient, a gold hairline top
 * strip and a faint radar watermark, with an eyebrow, h1 and optional subtitle.
 * Mirrors marketing-web's PageBanner so inner citizen pages share the look.
 */
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

export default PageBanner;
