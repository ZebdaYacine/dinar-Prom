'use client';

import {useEffect, useState} from 'react';
import {useTranslations} from 'next-intl';
import {useRouter} from '@/i18n/navigation';
import {OnboardingShell} from '@/components/merchant/onboarding-shell';
import {getMe, submitMerchant, type MeResponse} from '@/lib/auth-client';
import {useToast} from '@/components/ui/toast-provider';

export default function ReviewPage({params}: {params: Promise<{locale: string}>}) {
  const [locale, setLocale] = useState('ar');
  const [me, setMe] = useState<MeResponse['data'] | null>(null);
  const [consent, setConsent] = useState(false);
  const [error, setError] = useState('');
  const t = useTranslations('merchant');
  const router = useRouter();
  const {showToast} = useToast();

  useEffect(() => {
    void params.then((value) => setLocale(value.locale));
    void getMe().then((response) => setMe(response.data)).catch(() => setError(t('sessionExpired')));
  }, [params, t]);

  const submit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (me && ['pending', 'approved', 'suspended'].includes(me.merchant.verification_status)) {
      router.push('/merchant/status');
      return;
    }
    if (!consent) {
      setError(t('consentRequired'));
      showToast(t('consentRequired'), 'error');
      return;
    }
    const response = await submitMerchant();
    if (!response.ok) {
      setError(t('submitError'));
      showToast(t('submitError'), 'error');
      return;
    }
    showToast(t('submitted'), 'success');
    router.push('/onboarding/pending');
  };

  if (!me) return <OnboardingShell locale={locale} step={4}><p>{error || t('loadingStatus')}</p></OnboardingShell>;
  const locked = ['pending', 'approved', 'suspended'].includes(me.merchant.verification_status);
  const profile = me.merchant.profile;
  const shop = me.merchant.shop;
  return <OnboardingShell locale={locale} step={4}><form className="merchant-form" onSubmit={submit}>
    <h2>{t('reviewTitle')}</h2>
    <p>{t('reviewIntro')}</p>
    {locked && <p className="locked-notice">{t('applicationLocked')}</p>}
    <div className="review-summary">
      <section><h3>{t('reviewPersonal')}</h3><dl><div><dt>{t('firstName')}</dt><dd>{profile?.first_name || me.user.first_name}</dd></div><div><dt>{t('lastName')}</dt><dd>{profile?.last_name || me.user.last_name}</dd></div><div><dt>{t('phone')}</dt><dd>{profile?.phone_e164 || t('notProvided')}</dd></div><div><dt>{t('wilaya')}</dt><dd>{profile?.wilaya_id || t('notProvided')}</dd></div><div><dt>{t('address')}</dt><dd>{profile?.address || t('notProvided')}</dd></div></dl></section>
      <section><h3>{t('reviewShop')}</h3><dl><div><dt>{t('shopName')}</dt><dd>{shop?.name || t('notProvided')}</dd></div><div><dt>{t('category')}</dt><dd>{shop?.category_id || t('notProvided')}</dd></div><div><dt>{t('city')}</dt><dd>{shop?.city || t('notProvided')}</dd></div><div><dt>{t('registerNumber')}</dt><dd>{shop?.register_number || t('notProvided')}</dd></div><div><dt>{t('registerExpiry')}</dt><dd>{shop?.register_expires_on || t('notProvided')}</dd></div></dl></section>
      <section><h3>{t('reviewDocuments')}</h3><ul className="review-documents">{(me.merchant.documents ?? []).map((document) => <li key={document.id}><span>{t(`documents.${document.type}`)}</span><strong>✓</strong></li>)}</ul></section>
    </div>
    {!locked && <label className="consent"><input type="checkbox" checked={consent} onChange={(event) => setConsent(event.target.checked)} />{t('consent')}</label>}
    {error && <p className="form-error">{error}</p>}
    <button className="button button-primary">{locked ? t('viewStatus') : t('submit')}</button>
  </form></OnboardingShell>;
}
