import {getCategories,getWilayas} from '@/lib/api';
import {OfferEditor} from '@/components/merchant/offers/offer-editor';

export default async function EditPromotionPage({params}:{params:Promise<{locale:string;id:string}>}){const {locale,id}=await params;const [categories,wilayas]=await Promise.all([getCategories(locale),getWilayas(locale)]);return <main className="merchant-page"><OfferEditor locale={locale} offerId={id} categories={categories} wilayas={wilayas}/></main>}
