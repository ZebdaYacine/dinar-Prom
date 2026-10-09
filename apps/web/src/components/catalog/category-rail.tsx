import {Icon, type IconName} from '@/components/icon';
import type {Category} from '@/lib/api';

const iconNames = new Set<IconName>(['building', 'car', 'tire', 'mobile', 'laptop', 'shirt', 'sparkles', 'sofa', 'briefcase', 'wrench', 'utensils', 'plane', 'dumbbell']);

function categoryIcon(name: string): IconName {
  return iconNames.has(name as IconName) ? name as IconName : 'grid';
}

export function CategoryRail({categories, locale, selectedID = '', selectedSubcategory = '', onSelectCategory, onSelectSubcategory}: {categories: Category[]; locale: string; selectedID?: string; selectedSubcategory?: string; onSelectCategory: (categoryID: string) => void; onSelectSubcategory: (subcategoryID: string) => void}) {
  if (!categories.length) return null;
  const selected = categories.find((item) => item.id === selectedID);

  return <section className="catalog-explorer category-explorer" aria-labelledby="category-explorer-title">
    <div className="section-heading"><h2 id="category-explorer-title">{locale === 'ar' ? 'تصفح حسب الصنف' : 'Parcourir par catégorie'}</h2></div>
    <div className="category-rail" role="tablist" aria-label={locale === 'ar' ? 'الأصناف الرئيسية' : 'Catégories principales'}>
      {categories.map((category) => <button key={category.id} className={`category-tile${category.id === selectedID ? ' is-selected' : ''}`} type="button" role="tab" aria-selected={category.id === selectedID} onClick={() => onSelectCategory(category.id)}>
        <span className="category-icon"><Icon name={categoryIcon(category.icon)} size={28} /></span>
        <span className="category-name">{locale === 'ar' ? category.name_ar : category.name_fr}</span>
      </button>)}
    </div>
    {selected && <div className="subcategory-panel" role="tabpanel" aria-label={locale === 'ar' ? 'التصنيفات الفرعية' : 'Sous-catégories'}>
      <button className={`subcategory-chip${!selectedSubcategory ? ' is-selected' : ''}`} type="button" onClick={() => onSelectSubcategory('')}>{locale === 'ar' ? `كل ${selected.name_ar}` : `Tout ${selected.name_fr}`}</button>
      {selected.subcategories.map((item) => <button key={item.id} className={`subcategory-chip${item.id === selectedSubcategory ? ' is-selected' : ''}`} type="button" onClick={() => onSelectSubcategory(item.id)}>{locale === 'ar' ? item.name_ar : item.name_fr}</button>)}
    </div>}
  </section>;
}
