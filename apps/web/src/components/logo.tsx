import {Link} from '@/i18n/navigation';

/* eslint-disable @next/next/no-img-element */

export function Logo({href = '/', compact = false, locale}: {href?: string; compact?: boolean; locale?: string}) {
  return <Link className={`brand-logo${compact ? ' brand-logo-compact' : ''}`} href={href} locale={locale} aria-label="Dinar DZ">
    <img src="/brand/dinar-dz-mark.svg" alt="Dinar DZ" width={compact ? 42 : 48} height={compact ? 42 : 48} />
    {!compact && <span className="brand-wordmark"><strong>Dinar</strong><small>DZ</small><em>دينار</em></span>}
  </Link>;
}
