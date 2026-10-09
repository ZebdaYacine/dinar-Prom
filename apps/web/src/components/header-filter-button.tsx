'use client';

import {Icon} from './icon';

export function HeaderFilterButton({locale}: {locale: string}) {
  function openFilters() {
    window.dispatchEvent(new CustomEvent('dinar:open-filters'));
  }

  return <button className="header-filter-button" type="button" onClick={openFilters} aria-label={locale === 'ar' ? 'فتح البحث والفلاتر' : 'Ouvrir la recherche et les filtres'} title={locale === 'ar' ? 'البحث والفلاتر' : 'Recherche et filtres'}><Icon name="filter" size={20} /></button>;
}
