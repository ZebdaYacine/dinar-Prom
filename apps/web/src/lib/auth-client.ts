let accessToken: string | null = null;
let refreshPromise: Promise<string | null> | null = null;

function csrfToken() { if (typeof document === 'undefined') return ''; return document.cookie.split('; ').find((item) => item.startsWith('dinar_csrf='))?.split('=').slice(1).join('=') ?? ''; }

export async function refreshAccessToken(): Promise<string | null> {
  if (refreshPromise) return refreshPromise;
  refreshPromise = fetch('/api/v1/auth/refresh', {method: 'POST', credentials: 'include', headers: {'X-CSRF-Token': csrfToken()}})
    .then(async (response) => { if (!response.ok) return null; const payload = await response.json() as {data?: {access_token?: string}}; accessToken = payload.data?.access_token ?? null; return accessToken; })
    .catch(() => null)
    .finally(() => { refreshPromise = null; });
  return refreshPromise;
}

async function request(path: string, init: RequestInit = {}) {
  if (!accessToken) await refreshAccessToken();
  const headers = new Headers(init.headers); if (accessToken) headers.set('Authorization', `Bearer ${accessToken}`); if (init.body && !(init.body instanceof FormData)) headers.set('Content-Type', 'application/json');
  let response = await fetch(`/api/v1${path}`, {...init, headers, credentials: 'include'});
  if (response.status === 401) { accessToken = null; if (await refreshAccessToken()) { headers.set('Authorization', `Bearer ${accessToken}`); response = await fetch(`/api/v1${path}`, {...init, headers, credentials: 'include'}); } }
  return response;
}

export interface MeResponse {data: {user: {id: string; email: string; name: string; first_name: string; last_name: string; picture: string; email_verified: boolean}; merchant: {verification_status: string; rejection_reason?: string; rejection_document?: string; profile?: {first_name: string; last_name: string; phone_e164: string; wilaya_id: string; address: string}; shop?: {name: string; slug: string; category_id: string; description: string; wilaya_id: string; city: string; address: string; phone: string; whatsapp: string; register_number: string; register_expires_on: string}; documents?: {id: string; type: string; content_type: string; size: number; created_at: string}[]}}}
export async function getMe() { const response = await request('/me'); if (!response.ok) throw new Error('unauthorized'); return response.json() as Promise<MeResponse>; }
export async function logout() { const response = await request('/auth/logout', {method: 'POST', headers: {'X-CSRF-Token': csrfToken()}}); accessToken = null; return response; }
export async function saveProfile(payload: {first_name: string; last_name: string; phone_e164: string; wilaya_id: string; address: string}) { return request('/merchant/profile', {method: 'PUT', body: JSON.stringify(payload)}); }
export async function saveShop(payload: Record<string, string>) { return request('/merchant/shop', {method: 'PUT', body: JSON.stringify(payload)}); }
export async function uploadDocument(type: string, file: File) { const body = new FormData(); body.set('type', type); body.set('file', file); return request('/merchant/documents', {method: 'POST', body}); }
export async function submitMerchant() { return request('/merchant/submit', {method: 'POST', body: JSON.stringify({consent_accepted: true})}); }
export function googleLoginURL(locale: string) { return `/api/v1/auth/google/login?next=${encodeURIComponent(`/${locale}/onboarding/personal`)}`; }

export interface MerchantOfferInput {title_ar: string; title_fr: string; description_ar: string; description_fr: string; category_slug: string; subcategory_slug: string; brand_slug: string; price_before_dzd: number | null; price_after_dzd: number | null; starts_at: string | null; ends_at: string | null; wilaya_code: string; city: string; address: string; whatsapp: string}
export interface OfferMedia {id: string; url_400: string; url_800: string; url_1600: string; source_width: number; source_height: number; position: number; alt_text: string; created_at: string}
export interface MerchantOffer extends MerchantOfferInput {id: string; discount_percent: number; status: string; rejection_reason?: string; submitted_at?: string; approved_at?: string; created_at: string; updated_at: string; media: OfferMedia[]; requires_review?: boolean}
export interface MerchantStats {totals: {offers: number; views: number; whatsapp_clicks: number}; series: {date: string; views: number; whatsapp_clicks: number}[]; offers: {id: string; title: string; status: string; views: number; whatsapp_clicks: number}[]}
export async function listMerchantOffers() { const response=await request('/merchant/offers'); if(!response.ok) throw new Error('offers'); return response.json() as Promise<{data: MerchantOffer[]}>; }
export async function createMerchantOffer() { return request('/merchant/offers', {method:'POST'}); }
export async function getMerchantOffer(id:string) { const response=await request(`/merchant/offers/${id}`); if(!response.ok) throw new Error('offer'); return response.json() as Promise<{data: MerchantOffer}>; }
export async function saveMerchantOffer(id:string,payload:MerchantOfferInput) { return request(`/merchant/offers/${id}`, {method:'PUT',body:JSON.stringify(payload)}); }
export async function submitOffer(id:string) { return request(`/merchant/offers/${id}/submit`, {method:'POST'}); }
export async function offerAction(id:string,action:'pause'|'resume'|'archive'|'duplicate') { return request(`/merchant/offers/${id}/${action}`, {method:'POST'}); }
export async function deleteOffer(id:string) { return request(`/merchant/offers/${id}`, {method:'DELETE'}); }
export async function uploadOfferMedia(id:string,file:File,altText:string) { const body=new FormData();body.set('file',file);body.set('alt_text',altText);return request(`/merchant/offers/${id}/media`,{method:'POST',body}); }
export async function deleteOfferMedia(id:string,mediaID:string) { return request(`/merchant/offers/${id}/media/${mediaID}`,{method:'DELETE'}); }
export async function reorderOfferMedia(id:string,ids:string[]) { return request(`/merchant/offers/${id}/media/order`,{method:'PUT',body:JSON.stringify({ids})}); }
export async function getMerchantStats() { const response=await request('/merchant/stats');if(!response.ok)throw new Error('stats');return response.json() as Promise<{data:MerchantStats}>; }
