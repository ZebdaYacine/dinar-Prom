import {OfferStats} from '@/components/merchant/offers/offer-stats';
export default async function PromotionStatsPage({params}:{params:Promise<{locale:string}>}){const {locale}=await params;return <main className="merchant-page"><OfferStats locale={locale}/></main>}
