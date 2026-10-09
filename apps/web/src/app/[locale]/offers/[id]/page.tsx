/* eslint-disable @next/next/no-img-element */
import type {Metadata} from 'next';
import {getTranslations} from 'next-intl/server';
import {notFound} from 'next/navigation';
import {formatDate, formatDZD, getOffer} from '@/lib/api';
import {Link} from '@/i18n/navigation';
import {OfferActions} from '@/components/offers/offer-actions';
import {OfferCard} from '@/components/offers/offer-card';
import {ReportForm} from '@/components/offers/report-form';

type Props = {params: Promise<{locale: string; id: string}>};

export async function generateMetadata({params}: Props): Promise<Metadata> {
  const {locale, id} = await params;
  try {
    const offer = await getOffer(id, locale);
    const canonical = `/${locale}/offers/${encodeURIComponent(id)}`;
    return {
      title: `${offer.title} · Dinar DZ`,
      description: offer.description,
      alternates: {canonical, languages: {ar: `/ar/offers/${encodeURIComponent(id)}`, fr: `/fr/offers/${encodeURIComponent(id)}`}},
      openGraph: {title: offer.title, description: offer.description, url: canonical, type: 'article', images: [{url: offer.cover_url, alt: offer.title}]},
      twitter: {card: 'summary_large_image', title: offer.title, description: offer.description, images: [offer.cover_url]}
    };
  } catch {
    return {title: 'Dinar DZ'};
  }
}

export default async function OfferPage({params}: Props) {
  const {locale, id} = await params;
  const t = await getTranslations({locale, namespace: 'offers'});
  let offer;
  try { offer = await getOffer(id, locale); } catch { notFound(); }
  const similarOffers = offer.similar_offers ?? [];
  const structuredData = { '@context': 'https://schema.org', '@type': 'Offer', name: offer.title, description: offer.description, url: `${process.env.NEXT_PUBLIC_APP_URL ?? 'http://localhost:8180'}/${locale}/offers/${encodeURIComponent(offer.id)}`, priceCurrency: 'DZD', price: offer.price, availability: 'https://schema.org/InStock', validFrom: offer.starts_at, priceValidUntil: offer.ends_at, image: offer.images.map((image) => image.url_800), seller: {'@type': 'Organization', name: offer.shop.name} };
  return <article className="content-shell detail-page"><script type="application/ld+json" dangerouslySetInnerHTML={{__html: JSON.stringify(structuredData).replace(/</g, '\\u003c')}} /> <Link className="back-link" href="/offers" locale={locale}>← {t('back')}</Link><div className="detail-grid"><div><img className="detail-image" src={offer.cover_url} alt={offer.title} />{offer.images.length > 1 && <div className="thumbnail-row">{offer.images.map((image) => <img key={image.id} src={image.url_400} alt={image.alt} />)}</div>}</div><div className="detail-copy"><div className="offer-card-meta"><span>{offer.wilaya}</span><span>{offer.category[locale === 'ar' ? 'name_ar' : 'name_fr']}</span></div><h1>{offer.title}</h1><Link className="shop-link large" href={`/shop/${offer.shop.slug}`} locale={locale}>{offer.shop.name} {offer.verified && '✓'}</Link><div className="price-row detail-price"><strong>{formatDZD(offer.price, locale)}</strong><del>{formatDZD(offer.before_price, locale)}</del><span className="discount-badge">-{offer.discount_percent}%</span></div><p className="detail-description">{offer.description}</p><dl className="detail-facts"><div><dt>{t('validFrom')}</dt><dd>{formatDate(offer.starts_at, locale)}</dd></div><div><dt>{t('validUntil')}</dt><dd>{formatDate(offer.ends_at, locale)}</dd></div><div><dt>{t('address')}</dt><dd>{offer.address}</dd></div></dl><OfferActions id={offer.id} initialWhatsAppURL={offer.whatsapp_url} phone={offer.phone} locale={locale} /></div></div><section className="similar-section"><div className="section-heading"><h2>{t('similar')}</h2></div><div className="offer-grid">{similarOffers.map((similar) => <OfferCard key={similar.id} offer={similar} locale={locale} />)}</div></section><ReportForm id={offer.id} locale={locale} /></article>;
}
