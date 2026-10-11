'use client';

import {useEffect, useMemo, useState} from 'react';
import {useTranslations} from 'next-intl';
import {useRouter} from '@/i18n/navigation';
import type {Category, Wilaya} from '@/lib/api';
import {getMe, saveShop} from '@/lib/auth-client';
import {Combobox} from '@/components/ui/combobox';
import {useToast} from '@/components/ui/toast-provider';
import {useDraftAutosave} from '@/hooks/use-draft-autosave';

const initialForm = {name: '', slug: '', category_id: '', description: '', wilaya_id: '', city: '', address: '', phone: '', whatsapp: '', register_number: '', register_expires_on: ''};

export function ShopForm({locale, categories, wilayas}: {locale: string; categories: Category[]; wilayas: Wilaya[]}) {
  const t = useTranslations('merchant');
  const router = useRouter();
  const {showToast} = useToast();
  const [form, setForm] = useState(initialForm);
  const [error, setError] = useState('');
  const [locked, setLocked] = useState(false);
  const [hydrated, setHydrated] = useState(false);
  const update = (key: keyof typeof form, value: string) => setForm((current) => ({...current, [key]: value}));
  const selectedWilaya = useMemo(() => wilayas.find((wilaya) => wilaya.code === form.wilaya_id), [wilayas, form.wilaya_id]);
  const categoryOptions = categories.map((category) => ({value: category.id, label: locale === 'ar' ? category.name_ar : category.name_fr}));
  const wilayaOptions = wilayas.map((wilaya) => ({value: wilaya.code, label: `${wilaya.code} — ${locale === 'ar' ? wilaya.name_ar : wilaya.name_fr}`}));
  const cityOptions = (selectedWilaya?.cities ?? []).map((city) => ({value: city.id, label: locale === 'ar' ? city.name_ar : city.name_fr}));

  useEffect(() => {
    void getMe().then(({data}) => {
      const status = data.merchant.verification_status;
      setLocked(status === 'pending' || status === 'approved' || status === 'suspended');
      if (data.merchant.shop) setForm({...initialForm, ...data.merchant.shop});
      setHydrated(true);
    }).catch(() => undefined);
  }, []);

  const autosave = useDraftAutosave(form, hydrated && !locked, async (draft) => (await saveShop(draft)).ok);

  const submit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (locked) {
      router.push('/onboarding/documents');
      return;
    }
    const response = await saveShop(form);
    if (!response.ok) {
      setError(t('fixFields'));
      showToast(t('fixFields'), 'error');
      return;
    }
    showToast(t('saved'), 'success');
    router.push('/onboarding/documents');
  };

  return <form className="merchant-form" onSubmit={submit}>
    {locked && <p className="locked-notice">{t('applicationLocked')}</p>}
    <label>{t('shopName')}<input disabled={locked} required value={form.name} onChange={(event) => update('name', event.target.value)} /></label>
    <label>{t('slug')}<input disabled={locked} value={form.slug} onChange={(event) => update('slug', event.target.value)} placeholder="ma-boutique" /></label>
    <Combobox disabled={locked} label={t('category')} placeholder={t('chooseCategory')} required value={form.category_id} options={categoryOptions} onChange={(value) => update('category_id', value)} />
    <label>{t('description')}<textarea disabled={locked} value={form.description} onChange={(event) => update('description', event.target.value)} /></label>
    <div className="form-grid"><Combobox disabled={locked} label={t('wilaya')} placeholder={t('chooseWilaya')} required value={form.wilaya_id} options={wilayaOptions} onChange={(value) => { update('wilaya_id', value); update('city', ''); }} /><Combobox disabled={locked} label={t('city')} placeholder={selectedWilaya ? t('chooseCity') : t('chooseWilayaFirst')} value={form.city} options={cityOptions} onChange={(value) => update('city', value)} /></div>
    <label>{t('address')}<input disabled={locked} value={form.address} onChange={(event) => update('address', event.target.value)} /></label>
    <div className="form-grid"><label>{t('phone')}<input disabled={locked} type="tel" value={form.phone} onChange={(event) => update('phone', event.target.value)} /></label><label>{t('whatsapp')}<input disabled={locked} type="tel" value={form.whatsapp} onChange={(event) => update('whatsapp', event.target.value)} /></label></div>
    <div className="form-grid"><label>{t('registerNumber')}<input disabled={locked} required value={form.register_number} onChange={(event) => update('register_number', event.target.value)} /></label><label>{t('registerExpiry')}<input disabled={locked} required type="date" value={form.register_expires_on} onChange={(event) => update('register_expires_on', event.target.value)} /></label></div>
    {!locked && <p className="autosave">{autosave === 'saving' ? t('saving') : autosave === 'saved' ? t('draftSaved') : t('autosave')}</p>}
    {error && <p className="form-error">{error}</p>}
    <button className="button button-primary">{t('next')}</button>
  </form>;
}
