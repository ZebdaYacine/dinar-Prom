'use client';

import {useEffect, useMemo, useState} from 'react';
import type {Brand, Category, OfferSummary, Wilaya} from '@/lib/api';
import {OfferCard} from '@/components/offers/offer-card';
import {SearchBar} from '@/components/offers/search-bar';
import {BrandRail} from '@/components/catalog/brand-rail';
import {CategoryRail} from '@/components/catalog/category-rail';

type Filters = Record<string, string>;

function normalize(value: string) {
  return value.toLocaleLowerCase().replace(/[إأآا]/g, 'ا').replace(/ى/g, 'ي').replace(/ة/g, 'ه').trim();
}

function initialFilters(values: Record<string, string | undefined>): Filters {
  return {
    q: values.q ?? '',
    wilaya: values.wilaya ?? '',
    city: values.city ?? '',
    category: values.category ?? '',
    subcategory: values.subcategory ?? '',
    brand: values.brand ?? '',
    gender: values.gender ?? '',
    age_from: values.age_from ?? '',
    age_to: values.age_to ?? '',
    min_discount: values.min_discount ?? '',
    ending_within: values.ending_within ?? '',
    sort: values.sort ?? 'newest',
    page: values.page ?? '1'
  };
}

export function CatalogBrowser({locale, categories, wilayas, brands, offers, initialValues = {}, latestTitle, endingSoonTitle}: {locale: string; categories: Category[]; wilayas: Wilaya[]; brands: Brand[]; offers: OfferSummary[]; initialValues?: Record<string, string | undefined>; latestTitle: string; endingSoonTitle: string}) {
  const [filters, setFilters] = useState<Filters>(() => initialFilters(initialValues));
  const availableBrands = useMemo(() => brands.filter((brand) => (!filters.category || brand.category_id === filters.category) && (!filters.subcategory || brand.subcategory_id === filters.subcategory)), [brands, filters.category, filters.subcategory]);

  useEffect(() => {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([key, value]) => {
      if (value && !(key === 'sort' && value === 'newest')) params.set(key, value);
    });
    const nextURL = `${window.location.pathname}${params.toString() ? `?${params.toString()}` : ''}`;
    if (`${window.location.pathname}${window.location.search}` !== nextURL) {
      window.history.replaceState({}, '', nextURL);
    }
  }, [filters]);

  const filteredOffers = useMemo(() => {
    const query = normalize(filters.q ?? '');
    const result = offers.filter((offer) => {
      if (filters.category && offer.category.id !== filters.category) return false;
      if (filters.subcategory && offer.subcategory.id !== filters.subcategory) return false;
      if (filters.brand && offer.brand?.slug !== filters.brand) return false;
      if (filters.wilaya && offer.wilaya_code !== filters.wilaya) return false;
      if (filters.city && offer.city_id !== filters.city) return false;
      if (filters.gender && filters.gender !== 'all' && offer.audience_gender !== filters.gender && offer.audience_gender !== 'unisex') return false;
      const enteredAgeFrom = Number(filters.age_from || 0);
      const enteredAgeTo = Number(filters.age_to || 0);
      const ageFrom = enteredAgeTo > 0 && enteredAgeFrom > enteredAgeTo ? enteredAgeTo : enteredAgeFrom;
      const ageTo = enteredAgeTo > 0 && enteredAgeFrom > enteredAgeTo ? enteredAgeFrom : enteredAgeTo;
      if (ageFrom > 0 && offer.age_max < ageFrom) return false;
      if (ageTo > 0 && offer.age_min > ageTo) return false;
      if (Number(filters.min_discount || 0) > 0 && offer.discount_percent < Number(filters.min_discount)) return false;
      if (filters.ending_within) {
        const hours = filters.ending_within === '24h' ? 24 : filters.ending_within === '7d' ? 24 * 7 : 0;
        if (hours > 0 && new Date(offer.ends_at).getTime() > Date.now() + hours * 60 * 60 * 1000) return false;
      }
      if (query) {
        const searchable = normalize([offer.title, offer.shop.name, offer.wilaya, offer.city, offer.category.name_ar, offer.category.name_fr, offer.subcategory.name_ar, offer.subcategory.name_fr, offer.brand?.name_ar ?? '', offer.brand?.name_fr ?? ''].join(' '));
        if (!searchable.includes(query)) return false;
      }
      return true;
    });
    return result.sort((left, right) => {
      if (filters.sort === 'ending_soon') return new Date(left.ends_at).getTime() - new Date(right.ends_at).getTime();
      if (filters.sort === 'discount') return right.discount_percent - left.discount_percent;
      if (filters.sort === 'price_asc') return left.price - right.price;
      return new Date(right.starts_at).getTime() - new Date(left.starts_at).getTime();
    });
  }, [filters, offers]);

  const page = Math.max(1, Number(filters.page || 1));
  const pageSize = 24;
  const totalPages = Math.max(1, Math.ceil(filteredOffers.length / pageSize));
  const currentPage = Math.min(page, totalPages);
  const pageOffers = filteredOffers.slice((currentPage - 1) * pageSize, currentPage * pageSize);
  const latest = pageOffers.slice(0, 8);
  const endingSoon = pageOffers.filter((offer) => new Date(offer.ends_at).getTime() <= Date.now() + 7 * 24 * 60 * 60 * 1000).slice(0, 4);

  const updateFilters = (next: Filters) => setFilters({...next, page: '1'});
  const selectCategory = (category: string) => setFilters((current) => ({...current, category, subcategory: '', brand: '', page: '1'}));
  const selectSubcategory = (subcategory: string) => setFilters((current) => ({...current, subcategory, brand: '', page: '1'}));
  const selectBrand = (brand: string) => setFilters((current) => ({...current, brand, page: '1'}));
  const clear = () => setFilters(initialFilters({}));

  return <div className="catalog-layout">
    <SearchBar locale={locale} wilayas={wilayas} offers={offers} values={filters} onChange={updateFilters} onClear={clear} />
    <div className="catalog-main">
      <CategoryRail categories={categories} locale={locale} selectedID={filters.category} selectedSubcategory={filters.subcategory} onSelectCategory={selectCategory} onSelectSubcategory={selectSubcategory} />
      {filters.subcategory && <BrandRail brands={availableBrands} locale={locale} category={filters.category ?? ''} subcategory={filters.subcategory} selectedBrand={filters.brand} onSelectBrand={selectBrand} />}
      <OfferSection title={endingSoonTitle} offers={endingSoon} locale={locale} />
      <OfferSection title={latestTitle} offers={latest} locale={locale} />
      <Pagination locale={locale} page={currentPage} totalPages={totalPages} onSelect={(nextPage) => setFilters((current) => ({...current, page: String(nextPage)}))} />
    </div>
  </div>;
}

function Pagination({locale, page, totalPages, onSelect}: {locale: string; page: number; totalPages: number; onSelect: (page: number) => void}) {
  if (totalPages <= 1) return null;
  return <nav className="catalog-pagination" aria-label={locale === 'ar' ? 'صفحات العروض' : 'Pages des offres'}>
    <button type="button" disabled={page === 1} onClick={() => onSelect(page - 1)}>{locale === 'ar' ? 'السابق' : 'Précédent'}</button>
    <span dir="ltr">{page} / {totalPages}</span>
    <button type="button" disabled={page === totalPages} onClick={() => onSelect(page + 1)}>{locale === 'ar' ? 'التالي' : 'Suivant'}</button>
  </nav>;
}

function OfferSection({title, offers, locale}: {title: string; offers: OfferSummary[]; locale: string}) {
  return <section className="offer-section"><div className="section-heading"><h2>{title}</h2><span className="location-count">{offers.length}</span></div>{offers.length ? <div className="offer-grid">{offers.map((offer) => <OfferCard key={offer.id} offer={offer} locale={locale} />)}</div> : <p className="empty-state">{locale === 'ar' ? 'لا توجد عروض متاحة حالياً.' : 'Aucune offre disponible pour le moment.'}</p>}</section>;
}
