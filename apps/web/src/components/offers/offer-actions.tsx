'use client';

import {useEffect, useState} from 'react';
import {Icon} from '@/components/icon';

export function OfferActions({id, initialWhatsAppURL, phone, locale}: {id: string; initialWhatsAppURL: string; phone?: string; locale: string}) {
  const [message, setMessage] = useState('');
  useEffect(() => {
    void fetch(`/api/v1/offers/${id}/view`, {method: 'POST', headers: {'X-Visitor-ID': window.crypto.randomUUID()}});
  }, [id]);

  async function contactWhatsApp() {
    try {
      const response = await fetch(`/api/v1/offers/${id}/whatsapp`, {method: 'POST'});
      const payload = (await response.json()) as {data?: {url: string}};
      window.open(payload.data?.url ?? initialWhatsAppURL, '_blank', 'noopener,noreferrer');
    } catch {
      window.open(initialWhatsAppURL, '_blank', 'noopener,noreferrer');
    }
  }

  async function share() {
    const shareData = {title: document.title, url: window.location.href};
    try {
      if (navigator.share) {
        await navigator.share(shareData);
        setMessage(locale === 'ar' ? 'تمت المشاركة' : 'Partagé');
      } else if (navigator.clipboard) {
        await navigator.clipboard.writeText(window.location.href);
        setMessage(locale === 'ar' ? 'تم نسخ الرابط' : 'Lien copié');
      } else {
        setMessage(locale === 'ar' ? 'انسخ الرابط من شريط المتصفح' : 'Copiez le lien depuis le navigateur');
      }
    } catch (error) {
      if (error instanceof DOMException && error.name === 'AbortError') return;
      setMessage(locale === 'ar' ? 'تعذر إتمام المشاركة' : 'Impossible de partager');
    }
  }

  return <div className="action-row">
    <button className="button button-primary" type="button" onClick={() => void contactWhatsApp()}><Icon name="whatsapp" /> {locale === 'ar' ? 'تواصل عبر واتساب' : 'Contacter sur WhatsApp'}</button>
    {phone && <a className="button button-secondary" href={`tel:${phone}`}><Icon name="phone" /> {locale === 'ar' ? 'اتصال هاتفي' : 'Appeler'}</a>}
    <button className="button button-secondary" type="button" onClick={() => void share()}><Icon name="share" /> {locale === 'ar' ? 'مشاركة' : 'Partager'}</button>
    {message && <span className="action-message" role="status">{message}</span>}
  </div>;
}
