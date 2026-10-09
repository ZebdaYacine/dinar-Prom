import type {Metadata} from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'Dinar DZ',
  description: 'Local discounts from verified Algerian merchants.'
};

export default function RootLayout({
  children
}: Readonly<{children: React.ReactNode}>) {
  return (
    <html lang="ar" suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{__html: `(()=>{try{const s=localStorage.getItem('dinar_theme');const d=s==='light'?'light':'dark';const fr=location.pathname.split('/')[1]==='fr';document.documentElement.dataset.theme=d;document.documentElement.lang=fr?'fr':'ar';document.documentElement.dir=fr?'ltr':'rtl'}catch(e){document.documentElement.dataset.theme='dark'}})()`}} />
      </head>
      <body>{children}</body>
    </html>
  );
}
