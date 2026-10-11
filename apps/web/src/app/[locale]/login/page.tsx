import {getTranslations} from 'next-intl/server';
import {Link} from '@/i18n/navigation';
import {googleLoginURL} from '@/lib/auth-client';

export default async function LoginPage({params}: {params: Promise<{locale: string}>}) { const {locale} = await params; const t = await getTranslations({locale, namespace: 'auth'}); return <main className="auth-page"><section className="auth-card"><p className="eyebrow">Dinar DZ</p><h1>{t('title')}</h1><p>{t('subtitle')}</p><a className="button button-primary google-button" href={googleLoginURL(locale)}>{t('google')}</a><p className="auth-note">{t('privacy')}</p><Link className="button button-secondary" href="/" locale={locale}>{t('back')}</Link></section></main>; }
