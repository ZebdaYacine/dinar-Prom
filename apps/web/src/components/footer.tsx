import {getTranslations} from 'next-intl/server';
import {Link} from '@/i18n/navigation';
import {Logo} from './logo';
import {Icon} from './icon';

export async function Footer({locale}: {locale: string}) {
  const t = await getTranslations({locale, namespace: 'footer'});
  const year = new Date().getFullYear();
  const ar = locale === 'ar';
  return <footer className="site-footer">
    <div className="footer-inner">
      <div className="footer-brand">
        <div className="footer-brand-head"><Logo locale={locale} /><span className="footer-country">🇩🇿 {ar ? 'من الجزائر' : 'Depuis l’Algérie'}</span></div>
        <p>{ar ? 'منصة جزائرية لاكتشاف التخفيضات والعروض القريبة منك.' : 'Une plateforme algérienne pour découvrir les offres près de chez vous.'}</p>
        <div className="footer-contact-list">
          <a className="footer-contact" href="mailto:hello@dinardz.example"><span className="footer-contact-icon"><Icon name="share" size={16} /></span><span><small>{ar ? 'راسلنا' : 'Écrivez-nous'}</small><strong>hello@dinardz.example</strong></span></a>
          <a className="footer-contact" href="tel:+213560000000"><span className="footer-contact-icon"><Icon name="phone" size={16} /></span><span><small>{ar ? 'اتصل بنا' : 'Appelez-nous'}</small><strong dir="ltr">+213 560 000 000</strong></span></a>
        </div>
      </div>
      <div className="footer-column">
        <h3>{ar ? 'اكتشف' : 'Découvrir'}</h3>
        <Link href="/" locale={locale}>{ar ? 'الرئيسية' : 'Accueil'}</Link>
        <Link href="/offers" locale={locale}>{ar ? 'كل العروض' : 'Toutes les offres'}</Link>
        <Link href="/offers" locale={locale}>{ar ? 'حسب الولاية' : 'Par wilaya'}</Link>
      </div>
      <div className="footer-column">
        <h3>{ar ? 'للتجار' : 'Pour les commerçants'}</h3>
        <Link href="/login" locale={locale}>{ar ? 'تسجيل الدخول' : 'Se connecter'}</Link>
        <Link href="/login" locale={locale}>{ar ? 'نشر عرض' : 'Publier une offre'}</Link>
        <a href="mailto:partners@dinardz.example">{ar ? 'تواصل مع الفريق' : 'Contacter l’équipe'}</a>
      </div>
      <div className="footer-column">
        <h3>{ar ? 'المساعدة' : 'Aide'}</h3>
        <Link href="/" locale={locale}>{t('privacy')}</Link>
        <Link href="/" locale={locale}>{t('terms')}</Link>
        <a href="mailto:support@dinardz.example">{t('contact')}</a>
      </div>
      <div className="footer-cta-card">
        <span className="footer-cta-kicker">Dinar DZ</span>
        <h3>{ar ? 'اكتشف عرضك القادم' : 'Trouvez votre prochaine offre'}</h3>
        <p>{ar ? 'عروض محلية، أسعار أفضل، واكتشاف أسهل.' : 'Des offres locales, de meilleurs prix et une découverte simple.'}</p>
        <Link className="footer-cta-link" href="/offers" locale={locale}>{ar ? 'تصفح العروض' : 'Voir les offres'} <Icon name="arrow" size={15} /></Link>
      </div>
    </div>
    <div className="footer-bottom">
      <span>{t('copyright', {year})}</span>
      <div className="footer-bottom-links"><Link href="/" locale={locale}>{t('privacy')}</Link><span>{ar ? 'صنع بحب في الجزائر 🇩🇿' : 'Fait avec soin en Algérie 🇩🇿'}</span></div>
    </div>
  </footer>;
}
