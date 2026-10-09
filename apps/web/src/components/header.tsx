import {getTranslations} from 'next-intl/server';
import {ThemeToggle} from './theme-toggle';
import {Link} from '@/i18n/navigation';
import {Logo} from './logo';
import {LanguageSwitcher} from './language-switcher';
import {NotificationCenter} from './notification-center';
import {HeaderFilterButton} from './header-filter-button';

export async function Header({locale}: {locale: string}) {
  const nav = await getTranslations({locale, namespace: 'nav'});

  return (
    <header className="site-header">
      <div className="header-start">
        <HeaderFilterButton locale={locale} />
        <Logo locale={locale} />
      </div>
      <nav className="desktop-nav" aria-label={locale === 'ar' ? 'التنقل' : 'Navigation'}>
        <Link href="/" locale={locale}>{nav('home')}</Link>
        <Link href="/offers" locale={locale}>{nav('offers')}</Link>
        <Link href="/login" locale={locale}>{nav('login')}</Link>
      </nav>
      <div className="header-actions">
        <NotificationCenter />
        <LanguageSwitcher locale={locale} />
        <ThemeToggle />
        <Link className="publish-link" href="/login" locale={locale}>{nav('publishOffer')}</Link>
      </div>
    </header>
  );
}
