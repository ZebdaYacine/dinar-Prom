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

export interface MeResponse {data: {user: {id: string; email: string; name: string; first_name: string; last_name: string; picture: string; email_verified: boolean}; merchant: {verification_status: string; rejection_reason?: string; profile?: {first_name: string; last_name: string; phone_e164: string; wilaya_id: string; address: string}; shop?: {name: string; slug: string; category_id: string; description: string; wilaya_id: string; city: string; address: string; phone: string; whatsapp: string; register_number: string; register_expires_on: string}; documents?: {id: string; type: string; content_type: string; size: number; created_at: string}[]}}}
export async function getMe() { const response = await request('/me'); if (!response.ok) throw new Error('unauthorized'); return response.json() as Promise<MeResponse>; }
export async function logout() { const response = await request('/auth/logout', {method: 'POST', headers: {'X-CSRF-Token': csrfToken()}}); accessToken = null; return response; }
export async function saveProfile(payload: {first_name: string; last_name: string; phone_e164: string; wilaya_id: string; address: string}) { return request('/merchant/profile', {method: 'PUT', body: JSON.stringify(payload)}); }
export async function saveShop(payload: Record<string, string>) { return request('/merchant/shop', {method: 'PUT', body: JSON.stringify(payload)}); }
export async function uploadDocument(type: string, file: File) { const body = new FormData(); body.set('type', type); body.set('file', file); return request('/merchant/documents', {method: 'POST', body}); }
export async function submitMerchant() { return request('/merchant/submit', {method: 'POST', body: JSON.stringify({consent_accepted: true})}); }
export function googleLoginURL(locale: string) { return `/api/v1/auth/google/login?next=${encodeURIComponent(`/${locale}/onboarding/personal`)}`; }
