import type {MetadataRoute} from 'next';
import {getOffers} from '@/lib/api';

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const base = process.env.NEXT_PUBLIC_APP_URL ?? 'http://localhost:8180';
  const entries: MetadataRoute.Sitemap = ['ar', 'fr'].flatMap((locale) => [
    {url: `${base}/${locale}`, changeFrequency: 'hourly' as const, priority: 1},
    {url: `${base}/${locale}/offers`, changeFrequency: 'hourly' as const, priority: 0.9}
  ]);
  try {
    const offers = await getOffers({per_page: 48, sort: 'newest'}, 'ar');
    for (const offer of offers.data) {
      for (const locale of ['ar', 'fr']) entries.push({url: `${base}/${locale}/offers/${offer.id}`, lastModified: offer.starts_at, changeFrequency: 'daily', priority: 0.8});
    }
  } catch {
    // The site remains buildable when the API is unavailable; static routes stay indexed.
  }
  return entries;
}
