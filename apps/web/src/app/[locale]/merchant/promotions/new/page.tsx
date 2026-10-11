import {getTranslations} from 'next-intl/server';
import {Link} from '@/i18n/navigation';

export default async function NewPromotionPage({params}: {params: Promise<{locale: string}>}) { const {locale} = await params; const t = await getTranslations({locale, namespace: 'merchant'}); return <main className="auth-page"><section className="auth-card"><p className="eyebrow">Dinar DZ</p><h1>{t('promotionCreateTitle')}</h1><p>{t('promotionCreateText')}</p><Link className="button button-primary" href="/" locale={locale}>{t('backHome')}</Link></section></main>; }
