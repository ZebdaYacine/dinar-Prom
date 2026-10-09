import {hasLocale, NextIntlClientProvider} from 'next-intl';
import {getMessages} from 'next-intl/server';
import {notFound} from 'next/navigation';
import {Header} from '@/components/header';
import {BottomNav} from '@/components/bottom-nav';
import {Footer} from '@/components/footer';
import {SplashScreen} from '@/components/splash-screen';
import {routing} from '@/i18n/routing';

export default async function LocaleLayout({
  children,
  params
}: Readonly<{
  children: React.ReactNode;
  params: Promise<{locale: string}>;
}>) {
  const {locale} = await params;
  if (!hasLocale(routing.locales, locale)) {
    notFound();
  }

  const messages = await getMessages({locale});
  const direction = locale === 'ar' ? 'rtl' : 'ltr';

  return (
    <div lang={locale} dir={direction}>
      <NextIntlClientProvider messages={messages}>
        <Header locale={locale} />
        <SplashScreen locale={locale} />
        <main>{children}</main>
        <Footer locale={locale} />
        <BottomNav locale={locale} />
      </NextIntlClientProvider>
    </div>
  );
}
