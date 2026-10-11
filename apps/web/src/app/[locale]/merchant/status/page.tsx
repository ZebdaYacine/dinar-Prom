import {VerificationResult} from '@/components/merchant/verification-result';

export default async function MerchantStatusPage({params}: {params: Promise<{locale: string}>}) { const {locale} = await params; return <main className="auth-page"><VerificationResult locale={locale}/></main>; }
