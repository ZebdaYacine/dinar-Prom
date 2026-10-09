import type {Brand} from '@/lib/api';

export function BrandRail({brands, locale, category, subcategory, selectedBrand = '', onSelectBrand}: {brands: Brand[]; locale: string; category: string; subcategory?: string; selectedBrand?: string; onSelectBrand: (brand: string) => void}) {
  if (!category || !brands.length) return null;

  return <section className="brand-explorer" aria-labelledby="brand-explorer-title">
    <div className="section-heading"><h2 id="brand-explorer-title">{locale === 'ar' ? 'العلامات التجارية' : 'Marques disponibles'}</h2><span className="location-count">{brands.length} {locale === 'ar' ? 'علامة' : 'marques'}</span></div>
    <div className="brand-rail" role="list" aria-label={locale === 'ar' ? 'العلامات التجارية المتاحة' : 'Marques disponibles'}>
      <button className={`brand-chip${!selectedBrand ? ' is-selected' : ''}`} type="button" onClick={() => onSelectBrand('')}><span className="brand-monogram">All</span><span>{locale === 'ar' ? 'كل العلامات' : 'Toutes'}</span></button>
      {brands.map((brand) => <button key={brand.id} className={`brand-chip${brand.slug === selectedBrand ? ' is-selected' : ''}`} type="button" onClick={() => onSelectBrand(brand.slug)} role="listitem"><span className="brand-monogram">{brand.name_fr.slice(0, 2)}</span><span>{brand.name_fr}</span></button>)}
    </div>
  </section>;
}
