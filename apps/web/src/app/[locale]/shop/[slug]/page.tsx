import {notFound} from 'next/navigation';
import {getShop} from '@/lib/api';
import {OfferCard} from '@/components/offers/offer-card';
import {Link} from '@/i18n/navigation';

export default async function ShopPage({params}: {params: Promise<{locale: string; slug: string}>}) {
  const {locale, slug} = await params;
  let result;
  try { result = await getShop(slug, locale); } catch { notFound(); }
  return <section className="content-shell page-shell"><Link className="back-link" href="/offers" locale={locale}>← {locale === 'ar' ? 'العروض' : 'Offres'}</Link><div className="shop-hero"><p className="eyebrow">{result.shop.wilaya}</p><h1>{result.shop.name} {result.shop.verified && '✓'}</h1><p>{result.shop.description}</p><span>{result.shop.address}</span><div className="action-row"><a className="button button-primary" href={`https://wa.me/${result.shop.whatsapp}`} target="_blank" rel="noopener noreferrer">{locale === 'ar' ? 'تواصل عبر واتساب' : 'Contacter sur WhatsApp'}</a>{result.shop.phone && <a className="button button-secondary" href={`tel:${result.shop.phone}`}>{locale === 'ar' ? 'اتصال' : 'Appeler'}</a>}</div></div><div className="section-heading"><h2>{locale === 'ar' ? 'العروض النشطة' : 'Offres actives'}</h2></div>{result.offers.length ? <div className="offer-grid">{result.offers.map((offer) => <OfferCard key={offer.id} offer={offer} locale={locale} />)}</div> : <p className="empty-state">{locale === 'ar' ? 'لا توجد عروض نشطة حالياً.' : 'Aucune offre active actuellement.'}</p>}</section>;
}
