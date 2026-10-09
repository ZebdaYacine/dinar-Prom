import {getTranslations} from 'next-intl/server';
import {Link} from '@/i18n/navigation';
import {getBrands, getCategories, getOffers, getWilayas} from '@/lib/api';
import {CatalogBrowser} from '@/components/catalog/catalog-browser';

type SearchParams = Record<string, string | string[] | undefined>;

function one(value: string | string[] | undefined) {
  return Array.isArray(value) ? value[0] ?? '' : value ?? '';
}

export default async function HomePage({params, searchParams}: {params: Promise<{locale: string}>; searchParams: Promise<SearchParams>}) {
  const {locale} = await params;
  const raw = await searchParams;
  const values = Object.fromEntries(['q', 'wilaya', 'city', 'category', 'subcategory', 'brand', 'gender', 'age_from', 'age_to', 'sort'].map((key) => [key, one(raw[key])])) as Record<string, string>;
  const t = await getTranslations({locale, namespace: 'home'});
  const [allOffers, categories, wilayas, brands] = await Promise.all([
    getOffers({per_page: 48, sort: 'newest'}, locale),
    getCategories(locale),
    getWilayas(locale),
    getBrands({}, locale)
  ]);
  const featured = allOffers.data[0];

  return (
    <>
      <section className="announcement">{t('announcement')}</section>
      <section className="hero-shell">
        <div className="hero-copy">
          <p className="eyebrow">Dinar DZ</p>
          <h1>{t('title')}</h1>
          <p className="lead">{t('description')}</p>
          <div className="hero-actions">
            <Link className="button button-primary" href="/offers" locale={locale}>{t('browseOffers')}</Link>
            <Link className="button button-secondary" href="/login" locale={locale}>{t('merchantLogin')}</Link>
          </div>
        </div>
        {featured ? <Link className="hero-visual" href={`/offers/${featured.id}`} locale={locale} aria-label={featured.title}>
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src={featured.cover_url} alt={featured.title} />
          <span className="hero-offer-label"><span>{locale === 'ar' ? 'عرض مميز' : 'À la une'}</span><strong>{featured.title}</strong></span>
        </Link> : <div className="hero-visual" aria-label={t('cardLabel')}><div className="hero-offer-label"><span>{t('cardTitle')}</span><strong>{t('cardText')}</strong></div></div>}
      </section>
      <section className="content-shell">
        <CatalogBrowser locale={locale} categories={categories} wilayas={wilayas} brands={brands} offers={allOffers.data} initialValues={values} latestTitle={t('latest')} endingSoonTitle={t('endingSoon')} />
      </section>
    </>
  );
}
