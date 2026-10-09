import {getTranslations} from 'next-intl/server';
import {getBrands, getCategories, getOffers, getWilayas} from '@/lib/api';
import {CatalogBrowser} from '@/components/catalog/catalog-browser';

type SearchParams = Record<string, string | string[] | undefined>;

function one(value: string | string[] | undefined) {
  return Array.isArray(value) ? value[0] ?? '' : value ?? '';
}

export default async function OffersPage({params, searchParams}: {params: Promise<{locale: string}>; searchParams: Promise<SearchParams>}) {
  const {locale} = await params;
  const raw = await searchParams;
  const values = Object.fromEntries(['q', 'wilaya', 'city', 'category', 'subcategory', 'brand', 'gender', 'age_from', 'age_to', 'min_discount', 'ending_within', 'sort', 'page'].map((key) => [key, one(raw[key])])) as Record<string, string>;
  const [allOffers, categories, wilayas, brands, t] = await Promise.all([getOffers({per_page: 48, sort: 'newest'}, locale), getCategories(locale), getWilayas(locale), getBrands({}, locale), getTranslations({locale, namespace: 'offers'})]);
  return <section className="content-shell page-shell"><div className="page-heading"><div><p className="eyebrow">Dinar DZ</p><h1>{t('title')}</h1></div><p>{t('subtitle')}</p></div><CatalogBrowser locale={locale} categories={categories} wilayas={wilayas} brands={brands} offers={allOffers.data} initialValues={values} latestTitle={t('title')} endingSoonTitle={locale === 'ar' ? 'ينتهي قريباً' : 'Se terminent bientôt'} /></section>;
}
