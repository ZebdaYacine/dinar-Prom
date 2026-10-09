'use client';

import {useEffect, useMemo, useState, type CSSProperties} from 'react';
import type {OfferSummary, Wilaya} from '@/lib/api';
import {Icon} from '@/components/icon';

type FilterValues = Record<string, string>;

function FilterFields({locale, wilayas, values, onChange}: {locale: string; wilayas: Wilaya[]; values: FilterValues; onChange: (patch: Record<string, string>) => void}) {
  const selectedWilaya = values.wilaya ?? '';
  const selectedLocation = wilayas.find((item) => item.code === selectedWilaya);
  const ageFrom = Math.min(Math.max(Number(values.age_from || 0), 0), 120);
  const ageTo = Math.min(Math.max(Number(values.age_to || 120), 0), 120);
  const safeAgeFrom = Math.min(ageFrom, ageTo);
  const safeAgeTo = Math.max(ageFrom, ageTo);
  return <>
    <label className="sidebar-control">
      <span>{locale === 'ar' ? 'الولاية' : 'Wilaya'}</span>
      <select value={selectedWilaya} onChange={(event) => onChange({wilaya: event.target.value, city: ''})}>
        <option value="">{locale === 'ar' ? 'كل الولايات' : 'Toutes les wilayas'}</option>
        {wilayas.map((wilaya) => <option key={wilaya.code} value={wilaya.code}>{locale === 'ar' ? wilaya.name_ar : wilaya.name_fr}</option>)}
      </select>
    </label>
    <label className="sidebar-control">
      <span>{locale === 'ar' ? 'البلدية' : 'Commune'}</span>
      <select value={values.city ?? ''} onChange={(event) => onChange({city: event.target.value})} disabled={!selectedLocation}>
        <option value="">{selectedLocation ? (locale === 'ar' ? 'كل البلديات' : 'Toutes les communes') : (locale === 'ar' ? 'اختر الولاية أولاً' : 'Choisissez une wilaya')}</option>
        {selectedLocation?.cities.map((city) => <option key={city.id} value={city.id}>{locale === 'ar' ? city.name_ar : city.name_fr}</option>)}
      </select>
    </label>
    <label className="sidebar-control">
      <span>{locale === 'ar' ? 'جنس المستهلك' : 'Genre du consommateur'}</span>
      <select value={values.gender ?? ''} onChange={(event) => onChange({gender: event.target.value})}>
        <option value="">{locale === 'ar' ? 'كل المستهلكين' : 'Tous les consommateurs'}</option>
        <option value="men">{locale === 'ar' ? 'رجال' : 'Hommes'}</option>
        <option value="women">{locale === 'ar' ? 'نساء' : 'Femmes'}</option>
        <option value="unisex">{locale === 'ar' ? 'للجميع' : 'Unisexe'}</option>
      </select>
    </label>
    <div className="sidebar-age-control">
      <div className="filter-label-row"><span>{locale === 'ar' ? 'عمر المستهلك' : 'Âge du consommateur'}</span><strong dir="ltr">{safeAgeFrom} — {safeAgeTo}</strong></div>
      <div className="range-slider" dir="ltr" style={{'--range-start': String((safeAgeFrom / 120) * 100) + '%', '--range-end': String((safeAgeTo / 120) * 100) + '%'} as CSSProperties}>
        <span className="range-track"><i /></span>
        <input aria-label={locale === 'ar' ? 'العمر الأدنى' : 'Âge minimum'} type="range" min="0" max="120" value={safeAgeFrom} onChange={(event) => onChange({age_from: event.target.value})} />
        <input aria-label={locale === 'ar' ? 'العمر الأقصى' : 'Âge maximum'} type="range" min="0" max="120" value={safeAgeTo} onChange={(event) => onChange({age_to: event.target.value})} />
      </div>
      <div className="range-caption" dir="ltr"><span>0</span><span>120</span></div>
      {(values.age_from || values.age_to) && <button className="range-reset" type="button" onClick={() => onChange({age_from: '', age_to: ''})}>{locale === 'ar' ? 'كل الأعمار' : 'Tous les âges'}</button>}
    </div>
  </>;
}

export function SearchBar({locale, wilayas, offers, values = {}, onChange, onClear}: {locale: string; wilayas: Wilaya[]; offers: OfferSummary[]; values?: FilterValues; onChange: (values: FilterValues) => void; onClear: () => void}) {
  const [isOpen, setIsOpen] = useState(false);
  const [isSearchFocused, setIsSearchFocused] = useState(false);
  const hasFilters = Boolean(values.q || values.wilaya || values.city || values.gender || values.age_from || values.age_to || values.category || values.subcategory || values.brand || (values.sort && values.sort !== 'newest'));
  const update = (patch: Record<string, string>) => onChange({...values, ...patch});

  const suggestions = useMemo(() => {
    const query = (values.q ?? '').trim();
    if (query.length < 2) return [];
    const normalizedQuery = query.toLocaleLowerCase();
    const seen = new Set<string>();
    const result: Array<{kind: 'product' | 'shop'; label: string; detail: string; value: string}> = [];
    const add = (suggestion: {kind: 'product' | 'shop'; label: string; detail: string; value: string}) => {
      const key = `${suggestion.kind}:${suggestion.value.toLocaleLowerCase()}`;
      if (seen.has(key) || result.length >= 8) return;
      seen.add(key);
      result.push(suggestion);
    };
    for (const offer of offers) {
      if (offer.title.toLocaleLowerCase().includes(normalizedQuery)) add({kind: 'product', label: offer.title, detail: locale === 'ar' ? 'منتج' : 'Produit', value: offer.title});
      if (offer.shop.name.toLocaleLowerCase().includes(normalizedQuery)) add({kind: 'shop', label: offer.shop.name, detail: locale === 'ar' ? 'محل' : 'Commerce', value: offer.shop.name});
      if (result.length >= 8) break;
    }
    return result;
  }, [locale, offers, values.q]);

  useEffect(() => {
    const closeOnEscape = (event: KeyboardEvent) => { if (event.key === 'Escape') setIsOpen(false); };
    const openFromHeader = () => setIsOpen(true);
    document.addEventListener('keydown', closeOnEscape);
    window.addEventListener('dinar:open-filters', openFromHeader);
    return () => {
      document.removeEventListener('keydown', closeOnEscape);
      window.removeEventListener('dinar:open-filters', openFromHeader);
    };
  }, []);

  useEffect(() => {
    const isMobile = window.matchMedia('(max-width: 800px)').matches;
    document.body.classList.toggle('catalog-sidebar-open', isOpen && isMobile);
    return () => document.body.classList.remove('catalog-sidebar-open');
  }, [isOpen]);

  const searchInput = <div className="search-autocomplete" onBlur={() => window.setTimeout(() => setIsSearchFocused(false), 120)}>
    <label className="sidebar-search-field">
      <span className="sr-only">{locale === 'ar' ? 'بحث' : 'Recherche'}</span>
      <Icon name="search" size={18} />
      <input name="q" value={values.q ?? ''} onFocus={() => setIsSearchFocused(true)} onChange={(event) => update({q: event.target.value})} placeholder={locale === 'ar' ? 'ابحث عن منتج أو محل...' : 'Rechercher un produit ou un commerce...'} autoComplete="off" />
      {values.q && <button type="button" onMouseDown={(event) => event.preventDefault()} onClick={() => update({q: ''})} aria-label={locale === 'ar' ? 'مسح البحث' : 'Effacer la recherche'}><Icon name="x" size={15} /></button>}
    </label>
    {isSearchFocused && suggestions.length > 0 && <div className="search-suggestions" role="listbox" aria-label={locale === 'ar' ? 'اقتراحات البحث' : 'Suggestions de recherche'}>{suggestions.map((suggestion) => <button key={`${suggestion.kind}-${suggestion.value}`} type="button" role="option" aria-selected="false" className="search-suggestion" onMouseDown={(event) => event.preventDefault()} onClick={() => { update({q: suggestion.value}); setIsSearchFocused(false); }}><span className="search-suggestion-kind">{suggestion.kind === 'product' ? '⌁' : '⌂'}</span><span><strong>{suggestion.label}</strong><small>{suggestion.detail}</small></span></button>)}</div>}
  </div>;
  const sortSelect = <label className="sidebar-control"><span>{locale === 'ar' ? 'الترتيب' : 'Trier'}</span><select name="sort" value={values.sort ?? 'newest'} onChange={(event) => update({sort: event.target.value})}><option value="newest">{locale === 'ar' ? 'الأحدث' : 'Plus récentes'}</option><option value="ending_soon">{locale === 'ar' ? 'ينتهي قريباً' : 'Se terminent bientôt'}</option><option value="discount">{locale === 'ar' ? 'أعلى تخفيض' : 'Meilleure réduction'}</option><option value="price_asc">{locale === 'ar' ? 'الأقل سعراً' : 'Prix croissant'}</option></select></label>;
  const toolbarSortSelect = <label className="toolbar-sort"><span className="sr-only">{locale === 'ar' ? 'الترتيب' : 'Trier'}</span><select name="sort" value={values.sort ?? 'newest'} onChange={(event) => update({sort: event.target.value})}><option value="newest">{locale === 'ar' ? 'الأحدث' : 'Plus récentes'}</option><option value="ending_soon">{locale === 'ar' ? 'ينتهي قريباً' : 'Se terminent bientôt'}</option><option value="discount">{locale === 'ar' ? 'أعلى تخفيض' : 'Meilleure réduction'}</option><option value="price_asc">{locale === 'ar' ? 'الأقل سعراً' : 'Prix croissant'}</option></select></label>;
  const filterButton = <button className="catalog-filter-button" type="button" onClick={() => setIsOpen((current) => !current)} aria-label={locale === 'ar' ? 'فتح الفلاتر' : 'Ouvrir les filtres'} aria-expanded={isOpen}><Icon name="filter" size={21} />{hasFilters && <span className="filter-count" aria-hidden="true">●</span>}</button>;
  return <div className="catalog-search-controls">
    <div className="desktop-search-toolbar">{searchInput}{toolbarSortSelect}{filterButton}{hasFilters && <button className="sidebar-clear" type="button" onClick={onClear}>{locale === 'ar' ? 'مسح الكل' : 'Tout effacer'}</button>}</div>
    {isOpen && <div className="desktop-filter-panel"><FilterFields locale={locale} wilayas={wilayas} values={values} onChange={update} /></div>}
    {isOpen && <button className="sidebar-backdrop" type="button" aria-label={locale === 'ar' ? 'إغلاق البحث والفلاتر' : 'Fermer la recherche et les filtres'} onClick={() => setIsOpen(false)} />}
    <aside className={`catalog-sidebar${isOpen ? ' is-open' : ''}`} aria-hidden={!isOpen} aria-label={locale === 'ar' ? 'الفلاتر' : 'Filtres'}>
      <div className="sidebar-head"><div><p className="eyebrow">Dinar DZ</p><h2>{locale === 'ar' ? 'تصفية العروض' : 'Filtrer les offres'}</h2></div><button className="sidebar-close" type="button" onClick={() => setIsOpen(false)} aria-label={locale === 'ar' ? 'إغلاق' : 'Fermer'}><Icon name="x" size={19} /></button></div>
      {sortSelect}
      <div className="sidebar-divider" />
      <div className="sidebar-filter-grid"><FilterFields locale={locale} wilayas={wilayas} values={values} onChange={update} /></div>
      {hasFilters && <button className="sidebar-clear" type="button" onClick={onClear}>{locale === 'ar' ? 'إعادة ضبط كل الفلاتر' : 'Réinitialiser les filtres'}</button>}
    </aside>
  </div>;
}
