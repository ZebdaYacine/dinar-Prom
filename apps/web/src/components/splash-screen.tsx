'use client';

import {useEffect, useState} from 'react';
import {useTranslations} from 'next-intl';
import {Icon} from './icon';

/* eslint-disable @next/next/no-img-element */

type Branding = {proverb: string; logo_url: string; logo_alt: string};
const fallbackBranding: Branding = {proverb: 'كل تخفيضات مدينتك في مكان واحد', logo_url: '/brand/dinar-dz-mark.svg', logo_alt: 'Dinar DZ'};
const fallbackBrandingFr: Branding = {proverb: 'Toutes les réductions de votre ville au même endroit', logo_url: '/brand/dinar-dz-mark.svg', logo_alt: 'Logo Dinar DZ'};

export function SplashScreen({locale}: {locale: string}) {
  const t = useTranslations('splash');
  const initialBranding = locale === 'ar' ? fallbackBranding : fallbackBrandingFr;
  const [branding, setBranding] = useState(initialBranding);
  const [progress, setProgress] = useState(5);
  const [stage, setStage] = useState(t('branding'));
  const [visible, setVisible] = useState(true);
  const [fading, setFading] = useState(false);

  useEffect(() => {
    document.documentElement.lang = locale;
    document.documentElement.dir = locale === 'ar' ? 'rtl' : 'ltr';
    let alive = true;
    let removeTimer: number | undefined;

    async function load() {
      setProgress(25);
      try {
        const response = await fetch(`/api/v1/site/branding?site=dinar_dz&locale=${locale === 'ar' ? 'ar-DZ' : 'fr-FR'}`, {cache: 'no-store'});
        const payload = (await response.json()) as {data?: Branding};
        if (alive && payload.data && payload.data.logo_url.startsWith('/brand/')) setBranding(payload.data);
      } catch {
        // The local logo and proverb are the intentional offline fallback.
      }
      if (!alive) return;
      setStage(t('page'));
      setProgress(45);
      await new Promise((resolve) => window.setTimeout(resolve, 120));
      if (!alive) return;
      setProgress(75);
      await new Promise((resolve) => requestAnimationFrame(() => resolve(undefined)));
      if (!alive) return;
      setProgress(88);
      setStage(t('ready'));
      setProgress(100);
      removeTimer = window.setTimeout(() => {
        if (!alive) return;
        setFading(true);
        window.setTimeout(() => alive && setVisible(false), 200);
	  }, 2000);
    }

    void load();
    return () => { alive = false; if (removeTimer) window.clearTimeout(removeTimer); };
  }, [locale, t]);

  if (!visible) return null;
  return <div className={`splash-screen${fading ? ' is-fading' : ''}`} role="status" aria-live="polite">
    <div className="splash-content">
      <div className="splash-brand-lockup">
        <img className="splash-logo" src={branding.logo_url} alt={branding.logo_alt} width="112" height="112" />
        <div className="splash-name"><strong>Dinar DZ</strong><span>دينار</span></div>
      </div>
      <p className="splash-proverb">{branding.proverb}</p>
      <div className="splash-status"><span>{t('loading')}</span><span>{progress}%</span></div>
      <div className="splash-progress" role="progressbar" aria-valuemin={0} aria-valuemax={100} aria-valuenow={progress}><span style={{width: `${progress}%`}} /></div>
      <p className="splash-stage"><Icon name="check" size={16} /> {stage}</p>
    </div>
  </div>;
}
