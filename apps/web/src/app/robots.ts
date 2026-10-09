import type {MetadataRoute} from 'next';

export default function robots(): MetadataRoute.Robots {
  const base = process.env.NEXT_PUBLIC_APP_URL ?? 'http://localhost:8180';
  return {rules: [{userAgent: '*', allow: ['/ar', '/fr'], disallow: ['/api/', '/login']}], sitemap: `${base}/sitemap.xml`};
}
