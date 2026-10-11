import {getTranslations} from 'next-intl/server';
import {getCategories, getWilayas} from '@/lib/api';
import {OnboardingShell} from '@/components/merchant/onboarding-shell';
import {ShopForm} from '@/components/merchant/shop-form';

export default async function ShopPage({params}: {params: Promise<{locale: string}>}) { const {locale}=await params; const [categories,wilayas]=await Promise.all([getCategories(locale),getWilayas(locale)]); await getTranslations({locale,namespace:'merchant'}); return <OnboardingShell locale={locale} step={2}><ShopForm locale={locale} categories={categories} wilayas={wilayas}/></OnboardingShell>; }
