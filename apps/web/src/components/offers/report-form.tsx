'use client';

import {useEffect, useState} from 'react';
import {Icon} from '@/components/icon';

export function ReportForm({id, locale}: {id: string; locale: string}) {
  const [isOpen, setIsOpen] = useState(false);
  const [state, setState] = useState<'idle' | 'sent' | 'error'>('idle');
  const ar = locale === 'ar';

  useEffect(() => {
    if (!isOpen) return;
    const closeOnEscape = (event: KeyboardEvent) => { if (event.key === 'Escape') setIsOpen(false); };
    document.addEventListener('keydown', closeOnEscape);
    return () => document.removeEventListener('keydown', closeOnEscape);
  }, [isOpen]);

  async function submit(form: FormData) {
    const response = await fetch(`/api/v1/offers/${id}/report`, {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({reason: form.get('reason'), message: form.get('message')})});
    setState(response.ok ? 'sent' : 'error');
  }

  return <>
    <button className="report-trigger" type="button" onClick={() => {setState('idle'); setIsOpen(true);}}><span aria-hidden="true">⚠</span> {ar ? 'هل يوجد خطأ في هذا العرض؟' : 'Un problème avec cette offre ?'}</button>
    {isOpen && <div className="report-modal-backdrop" role="presentation" onMouseDown={(event) => {if (event.target === event.currentTarget) setIsOpen(false);}}>
      <section className="report-modal" role="alertdialog" aria-modal="true" aria-labelledby="report-modal-title">
        <div className="report-modal-head"><div><p className="eyebrow">{ar ? 'تنبيه ومساعدة' : 'Alerte et assistance'}</p><h2 id="report-modal-title">{ar ? 'هل يوجد خطأ في هذا العرض؟' : 'Un problème avec cette offre ?'}</h2></div><button className="modal-close" type="button" onClick={() => setIsOpen(false)} aria-label={ar ? 'إغلاق' : 'Fermer'}><Icon name="x" size={19} /></button></div>
        <p className="report-warning">{ar ? 'ساعدنا على إبقاء العروض موثوقة. اختر السبب وأرسل ملاحظة قصيرة، وسيراجعها فريق Dinar DZ.' : 'Aidez-nous à garder les offres fiables. Choisissez un motif et envoyez une courte note ; notre équipe Dinar DZ vérifiera le signalement.'}</p>
        {state === 'sent' ? <p className="notice success" role="status">{ar ? 'تم استلام البلاغ، شكرًا لمساعدتنا.' : 'Signalement reçu, merci de votre aide.'}</p> : <form className="report-form" action={submit}>
          <label className="sidebar-control"><span>{ar ? 'سبب البلاغ' : 'Motif'}</span><select name="reason" defaultValue="misleading"><option value="misleading">{ar ? 'معلومات مضللة' : 'Informations trompeuses'}</option><option value="expired">{ar ? 'العرض منتهٍ' : 'Offre expirée'}</option><option value="wrong_price">{ar ? 'السعر غير صحيح' : 'Prix incorrect'}</option><option value="inappropriate">{ar ? 'محتوى غير مناسب' : 'Contenu inapproprié'}</option><option value="other">{ar ? 'سبب آخر' : 'Autre motif'}</option></select></label>
          <textarea name="message" maxLength={1000} placeholder={ar ? 'تفاصيل اختيارية...' : 'Détails facultatifs...'} />
          <button className="button button-primary" type="submit">{ar ? 'إرسال البلاغ' : 'Envoyer le signalement'}</button>
          {state === 'error' && <span className="notice error" role="alert">{ar ? 'تعذر إرسال البلاغ.' : 'Impossible d’envoyer le signalement.'}</span>}
        </form>}
      </section>
    </div>}
  </>;
}
