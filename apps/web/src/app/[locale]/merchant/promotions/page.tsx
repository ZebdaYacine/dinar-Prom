import {OfferDashboard} from '@/components/merchant/offers/offer-dashboard';
export default async function PromotionsPage({params}:{params:Promise<{locale:string}>}){const {locale}=await params;return <main className="merchant-page"><OfferDashboard locale={locale}/></main>}
