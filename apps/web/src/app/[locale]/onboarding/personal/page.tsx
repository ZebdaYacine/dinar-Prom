'use client';

import {useEffect, useState} from 'react';
import {useTranslations} from 'next-intl';
import {useRouter} from '@/i18n/navigation';
import {OnboardingShell} from '@/components/merchant/onboarding-shell';
import {getMe, saveProfile} from '@/lib/auth-client';
import {useToast} from '@/components/ui/toast-provider';

export default function PersonalPage({params}: {params: Promise<{locale: string}>}) {
  const [locale, setLocale] = useState('ar');
  const [form, setForm] = useState({first_name: '', last_name: '', phone_e164: '', wilaya_id: '', address: ''});
  const [message, setMessage] = useState('');
  const [saving, setSaving] = useState(false);
  const [locked, setLocked] = useState(false);
  const router = useRouter();
  const t = useTranslations('merchant');
  const {showToast} = useToast();
  useEffect(() => { void params.then((value) => setLocale(value.locale)); void getMe().then((response) => { const user = response.data.user; const profile = response.data.merchant.profile; const status = response.data.merchant.verification_status; setLocked(status === 'pending' || status === 'approved' || status === 'suspended'); setForm({first_name: user.first_name || '', last_name: user.last_name || '', phone_e164: profile?.phone_e164 || '', wilaya_id: profile?.wilaya_id || '', address: profile?.address || ''}); }).catch(() => setMessage(t('sessionExpired'))); }, [params, t]);
  const update = (key: keyof typeof form, value: string) => setForm((current) => ({...current, [key]: value}));
  const submit = async (event: React.FormEvent) => { event.preventDefault(); if (locked) { router.push('/onboarding/shop'); return; } setSaving(true); const response = await saveProfile(form); setSaving(false); if (!response.ok) { setMessage(t('fixFields')); showToast(t('fixFields'), 'error'); return; } showToast(t('saved'), 'success'); router.push('/onboarding/shop'); };
  return <OnboardingShell locale={locale} step={1}><form className="merchant-form" onSubmit={submit}>{locked && <p className="locked-notice">{t('applicationLocked')}</p>}<div className="form-grid"><label>{t('firstName')}<input disabled={locked} required autoComplete="given-name" value={form.first_name} onChange={(event) => update('first_name', event.target.value)} /></label><label>{t('lastName')}<input disabled={locked} required autoComplete="family-name" value={form.last_name} onChange={(event) => update('last_name', event.target.value)} /></label></div><label>{t('phone')}<input disabled={locked} required type="tel" inputMode="tel" autoComplete="tel" value={form.phone_e164} onChange={(event) => update('phone_e164', event.target.value)} placeholder="0550 00 00 00" /></label><label>{t('wilaya')}<input disabled={locked} required value={form.wilaya_id} onChange={(event) => update('wilaya_id', event.target.value)} placeholder="16" /></label><label>{t('address')}<textarea disabled={locked} required minLength={5} value={form.address} onChange={(event) => update('address', event.target.value)} /></label><p className="autosave">{locked ? t('readOnlySubmitted') : t('autosave')}</p>{message && <p className="form-error">{message}</p>}<button className="button button-primary" disabled={saving}>{saving ? t('saving') : t('next')}</button></form></OnboardingShell>;
}
