export interface Category {
  id: string;
  slug: string;
  name_ar: string;
  name_fr: string;
  icon: string;
  subcategories: Subcategory[];
}

export interface Subcategory {
  id: string;
  slug: string;
  name_ar: string;
  name_fr: string;
}

export interface Brand {
  id: string;
  slug: string;
  name_ar: string;
  name_fr: string;
  origin?: 'global' | 'local';
  category_id?: string;
  subcategory_id?: string;
}

export interface City {
  id: string;
  name_ar: string;
  name_fr: string;
  daira_ar: string;
  daira_fr: string;
}

export interface Wilaya {
  code: string;
  name_ar: string;
  name_fr: string;
  cities: City[];
}

export interface Shop {
  slug: string;
  name: string;
  description: string;
  wilaya: string;
  address: string;
  whatsapp: string;
  phone: string;
  verified: boolean;
}

export interface OfferImage {
  id: string;
  url_400: string;
  url_800: string;
  url_1600: string;
  alt: string;
  position: number;
}

export interface OfferSummary {
  id: string;
  title: string;
  category: Category;
  shop: Shop;
  wilaya: string;
  wilaya_code: string;
  city: string;
  city_id: string;
  subcategory: Subcategory;
  brand?: Brand | null;
  audience_gender: 'men' | 'women' | 'unisex';
  age_min: number;
  age_max: number;
  cover_url: string;
  before_price: number;
  price: number;
  discount_percent: number;
  starts_at: string;
  ends_at: string;
  verified: boolean;
}

export interface OfferDetails extends OfferSummary {
  description: string;
  address: string;
  phone?: string;
  whatsapp_url: string;
  images: OfferImage[];
  similar_offers: OfferSummary[];
}

export interface OffersMeta {
  page: number;
  per_page: number;
  total: number;
  has_next: boolean;
}

export interface OffersResult {
  data: OfferSummary[];
  meta: OffersMeta;
}

export interface ShopResult {
  shop: Shop;
  offers: OfferSummary[];
}

const apiBase = process.env.API_INTERNAL_URL ?? 'http://localhost:8188';

async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${apiBase}/api/v1${path}`, {
    ...options,
    headers: {'Accept-Language': 'ar', ...options.headers},
    next: {revalidate: 60}
  } as RequestInit & {next: {revalidate: number}});

  if (!response.ok) {
    const error = new Error(`API request failed: ${response.status}`);
    Object.assign(error, {status: response.status});
    throw error;
  }
  const payload = (await response.json()) as {data: T};
  return payload.data;
}

async function apiFetchEnvelope<T>(path: string, options: RequestInit = {}): Promise<{data: T; meta?: OffersMeta}> {
  const response = await fetch(`${apiBase}/api/v1${path}`, {
    ...options,
    headers: {'Accept-Language': 'ar', ...options.headers},
    next: {revalidate: 60}
  } as RequestInit & {next: {revalidate: number}});

  if (!response.ok) {
    const error = new Error(`API request failed: ${response.status}`);
    Object.assign(error, {status: response.status});
    throw error;
  }
  return (await response.json()) as {data: T; meta?: OffersMeta};
}

export async function getOffers(params: Record<string, string | number | undefined> = {}, locale = 'ar') {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '') search.set(key, String(value));
  }
  const suffix = search.toString() ? `?${search.toString()}` : '';
  return apiFetchEnvelope<OfferSummary[]>(`/offers${suffix}`, {headers: {'Accept-Language': locale}}) as Promise<OffersResult>;
}

export async function getOffer(id: string, locale = 'ar') {
  return apiFetch<OfferDetails>(`/offers/${encodeURIComponent(id)}`, {headers: {'Accept-Language': locale}});
}

export async function getShop(slug: string, locale = 'ar') {
  return apiFetch<ShopResult>(`/shops/${encodeURIComponent(slug)}`, {headers: {'Accept-Language': locale}});
}

export async function getCategories(locale = 'ar') {
  return apiFetch<Category[]>('/categories', {headers: {'Accept-Language': locale}});
}

export async function getBrands(params: {category?: string; subcategory?: string} = {}, locale = 'ar') {
  const search = new URLSearchParams();
  if (params.category) search.set('category', params.category);
  if (params.subcategory) search.set('subcategory', params.subcategory);
  const suffix = search.toString() ? `?${search.toString()}` : '';
  return apiFetch<Brand[]>(`/brands${suffix}`, {headers: {'Accept-Language': locale}});
}

export async function getWilayas(locale = 'ar') {
  return apiFetch<Wilaya[]>('/wilayas', {headers: {'Accept-Language': locale}});
}

export function formatDZD(value: number, locale: string) {
  return new Intl.NumberFormat(locale === 'ar' ? 'ar-DZ' : 'fr-DZ', {
    style: 'currency',
    currency: 'DZD',
    maximumFractionDigits: 0
  }).format(value);
}

export function formatDate(value: string, locale: string) {
  return new Intl.DateTimeFormat(locale === 'ar' ? 'ar-DZ' : 'fr-DZ', {
    dateStyle: 'medium',
    timeStyle: 'short',
    timeZone: 'Africa/Algiers'
  }).format(new Date(value));
}
