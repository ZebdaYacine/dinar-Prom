import {getCategories,getWilayas} from '@/lib/api';
import {OfferEditor} from '@/components/merchant/offers/offer-editor';

export default async function NewPromotionPage({params}: {params: Promise<{locale: string}>}) {const {locale}=await params;const [categories,wilayas]=await Promise.all([getCategories(locale),getWilayas(locale)]);return <main className="merchant-page"><OfferEditor locale={locale} categories={categories} wilayas={wilayas}/></main>}
