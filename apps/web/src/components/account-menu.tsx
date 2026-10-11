'use client';

import {useEffect, useRef, useState} from 'react';
import {useTranslations} from 'next-intl';
import {Link} from '@/i18n/navigation';
import {getMe, logout, type MeResponse} from '@/lib/auth-client';
import {useToast} from '@/components/ui/toast-provider';

/* eslint-disable @next/next/no-img-element */
export function AccountMenu({locale}: {locale: string}) {
  const t = useTranslations('account');
  const {showToast} = useToast();
  const [me, setMe] = useState<MeResponse['data'] | null>(null);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [avatarFailed, setAvatarFailed] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => { void getMe().then((response) => setMe(response.data)).catch(() => setMe(null)); }, []);
  useEffect(() => { const close = (event: MouseEvent) => { if (ref.current && !ref.current.contains(event.target as Node)) setOpen(false); }; document.addEventListener('mousedown', close); return () => document.removeEventListener('mousedown', close); }, []);
  if (!me) return <Link className="header-login-link" href="/login" locale={locale}>{t('login')}</Link>;
  const initials = (me.user.first_name || me.user.name || 'D').slice(0, 1).toUpperCase();
  const displayName = me.user.first_name || me.user.name;
  const handleLogout = async () => { setLoading(true); const response = await logout(); setLoading(false); if (response.ok || response.status === 204) { showToast(t('loggedOut'), 'success'); window.location.href = `/${locale}/login`; } else showToast(t('logoutError'), 'error'); };
  return <div className="account-menu" ref={ref}><button type="button" className="account-trigger" aria-label={displayName || t('avatar')} title={displayName || t('avatar')} aria-expanded={open} aria-haspopup="menu" onClick={() => setOpen((current) => !current)}>{me.user.picture && !avatarFailed ? <img src={me.user.picture} alt="" onError={() => setAvatarFailed(true)} /> : <span className="account-initials">{initials}</span>}<span className="account-name">{displayName}</span><span className="account-chevron" aria-hidden="true">⌄</span></button>{open && <div className="account-dropdown" role="menu"><div className="account-summary"><strong>{me.user.name}</strong><small>{me.user.email}</small></div><Link href="/onboarding/personal" locale={locale} role="menuitem" onClick={() => setOpen(false)}>{t('onboarding')}</Link><button type="button" role="menuitem" disabled={loading} onClick={() => void handleLogout()}>{loading ? t('loggingOut') : t('logout')}</button></div>}</div>;
}
