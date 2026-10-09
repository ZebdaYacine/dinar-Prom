import {getTranslations} from 'next-intl/server';
import {Link} from '@/i18n/navigation';
import {Icon} from './icon';

export async function BottomNav({locale}: {locale: string}) {
  const t = await getTranslations({locale, namespace: 'nav'});
  return <nav className="bottom-nav" aria-label={locale === 'ar' ? 'التنقل السفلي' : 'Navigation mobile'}>
    <Link href="/" locale={locale}><Icon name="house" /><span>{t('home')}</span></Link>
    <Link href="/offers" locale={locale}><Icon name="grid" /><span>{t('offers')}</span></Link>
    <Link href="/offers" locale={locale}><Icon name="map-pin" /><span>{t('wilaya')}</span></Link>
    <Link href="/login" locale={locale}><Icon name="user" /><span>{t('account')}</span></Link>
  </nav>;
}
