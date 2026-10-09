'use client';

import {usePathname, useRouter} from '@/i18n/navigation';
import {Icon} from './icon';

export function LanguageSwitcher({locale}: {locale: string}) {
  const pathname = usePathname();
  const router = useRouter();
  const nextLocale = locale === 'ar' ? 'fr' : 'ar';

  function switchLanguage() {
    router.replace(pathname, {locale: nextLocale});
  }

  return <button className="language-switcher" type="button" onClick={switchLanguage} aria-label={locale === 'ar' ? 'Passer au français' : 'التبديل إلى العربية'}>
    <Icon name={locale === 'ar' ? 'flag-dz' : 'flag-fr'} size={18} />
    <span>{locale === 'ar' ? 'العربية' : 'Français'}</span>
  </button>;
}
