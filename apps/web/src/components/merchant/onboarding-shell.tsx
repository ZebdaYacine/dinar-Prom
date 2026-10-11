'use client';

import {useTranslations} from 'next-intl';
import {Link} from '@/i18n/navigation';

export function OnboardingShell({locale, step, children}: {locale: string; step: 1 | 2 | 3 | 4; children: React.ReactNode}) {
  const t = useTranslations('merchant');
  const labels = [t('steps.personal'), t('steps.shop'), t('steps.documents'), t('steps.review')];
  return <main className="merchant-page"><div className="merchant-card"><div className="merchant-heading"><Link href="/" locale={locale} className="eyebrow">Dinar DZ</Link><span className="status-pill">{t('secure')}</span><h1>{t('onboardingTitle')}</h1><p>{t('onboardingSubtitle')}</p></div><ol className="progress-steps" aria-label={t('progressLabel')}>{labels.map((label, index) => <li key={label} className={index + 1 <= step ? 'active' : ''}><span>{index + 1}</span>{label}</li>)}</ol>{children}</div></main>;
}
