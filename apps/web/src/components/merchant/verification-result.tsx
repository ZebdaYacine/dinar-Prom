'use client';

import {useEffect, useState} from 'react';
import {useTranslations} from 'next-intl';
import {Link} from '@/i18n/navigation';
import {getMe, type MeResponse} from '@/lib/auth-client';
import {useToast} from '@/components/ui/toast-provider';

export function VerificationResult({locale}: {locale: string}) {
  const t = useTranslations('merchant');
  const {showToast} = useToast();
  const [me, setMe] = useState<MeResponse['data'] | null>(null);
  const [loading, setLoading] = useState(true);
  useEffect(() => { void getMe().then((response) => setMe(response.data)).catch(() => showToast(t('sessionExpired'), 'error')).finally(() => setLoading(false)); }, [showToast, t]);
  if (loading) return <section className="status-card"><div className="status-loader" aria-hidden="true" /><p>{t('loadingStatus')}</p></section>;
  const status = me?.merchant.verification_status ?? 'incomplete';
  if (status === 'pending') return <section className="status-card status-pending"><span className="status-icon" aria-hidden="true">⌛</span><p className="eyebrow">Dinar DZ</p><h1>{t('statusPendingTitle')}</h1><p>{t('statusPendingText')}</p><Link className="button button-primary" href="/" locale={locale}>{t('backHome')}</Link></section>;
  if (status === 'rejected') return <section className="status-card status-rejected"><span className="status-icon" aria-hidden="true">!</span><p className="eyebrow">Dinar DZ</p><h1>{t('statusRejectedTitle')}</h1><p>{t('statusRejectedText')}</p>{me?.merchant.rejection_reason && <p className="locked-notice"><strong>{t('rejectionReason')}:</strong> {me.merchant.rejection_reason}</p>}{me?.merchant.rejection_document && <p>{t('rejectedDocument', {document: t(`documents.${me.merchant.rejection_document}`)})}</p>}<Link className="button button-primary" href="/onboarding/personal" locale={locale} onClick={() => showToast(t('retryStarted'), 'info')}>{t('tryAgain')}</Link></section>;
  if (status === 'approved') return <section className="status-card status-approved"><span className="status-icon" aria-hidden="true">✓</span><p className="eyebrow">Dinar DZ</p><h1>{t('statusApprovedTitle')}</h1><p>{t('statusApprovedText')}</p><div className="empty-promotion"><strong>{t('noPromotionsTitle')}</strong><span>{t('noPromotionsText')}</span></div><Link className="button button-primary" href="/merchant/promotions/new" locale={locale}>{t('createPromotion')}</Link></section>;
  return <section className="status-card"><p className="eyebrow">Dinar DZ</p><h1>{t('continueOnboardingTitle')}</h1><p>{t('continueOnboardingText')}</p><Link className="button button-primary" href="/onboarding/personal" locale={locale}>{t('continue')}</Link></section>;
}
