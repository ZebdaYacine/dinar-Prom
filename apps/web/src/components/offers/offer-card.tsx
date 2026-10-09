import {Link} from '@/i18n/navigation';
import type {OfferSummary} from '@/lib/api';
import {formatDZD} from '@/lib/api';
import {Countdown} from './countdown';
import {Icon} from '@/components/icon';

export function OfferCard({offer, locale}: {offer: OfferSummary; locale: string}) {
  return (
    <article className="offer-card">
      <Link href={`/offers/${offer.id}`} locale={locale} className="offer-image-link">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img src={offer.cover_url} alt={offer.title} className="offer-image" loading="lazy" />
        <span className="discount-badge">-{offer.discount_percent}%</span>
      </Link>
      <div className="offer-card-body">
        <div className="offer-card-meta"><span>{offer.wilaya}</span><span>{offer.category[locale === 'ar' ? 'name_ar' : 'name_fr']}</span></div>
        <h3><Link href={`/offers/${offer.id}`} locale={locale}>{offer.title}</Link></h3>
        <Link className="shop-link" href={`/shop/${offer.shop.slug}`} locale={locale}>{offer.shop.name}{offer.verified && <span className="verified-badge"><Icon name="check" size={14} />{locale === 'ar' ? 'موثق' : 'Vérifié'}</span>}</Link>
        <div className="price-row"><strong>{formatDZD(offer.price, locale)}</strong><del>{formatDZD(offer.before_price, locale)}</del></div>
        <div className="countdown-row"><Countdown endsAt={offer.ends_at} locale={locale} /></div>
      </div>
    </article>
  );
}
