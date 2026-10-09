INSERT INTO brands (slug, name_ar, name_fr)
VALUES
    ('lumberjack', 'لامبرجاك', 'Lumberjack'),
    ('zara', 'زارا', 'Zara'),
    ('lc-waikiki', 'إل سي وايكيكي', 'LC Waikiki'),
    ('nike', 'نايكي', 'Nike'),
    ('adidas', 'أديداس', 'Adidas'),
    ('puma', 'بوما', 'Puma'),
    ('skechers', 'سكتشرز', 'Skechers'),
    ('new-balance', 'نيو بالانس', 'New Balance'),
    ('suzuki', 'سوزوكي', 'Suzuki'),
    ('peugeot', 'بيجو', 'Peugeot'),
    ('seat', 'سيات', 'Seat'),
    ('chevrolet', 'شيفروليه', 'Chevrolet'),
    ('renault', 'رونو', 'Renault'),
    ('dacia', 'داسيا', 'Dacia'),
    ('volkswagen', 'فولكسفاغن', 'Volkswagen'),
    ('geely', 'جيلي', 'Geely'),
    ('livan', 'ليفان', 'Livan'),
    ('jac', 'جاك', 'JAC'),
    ('apple', 'آبل', 'Apple'),
    ('condor', 'كوندور', 'Condor'),
    ('lenovo', 'لينوفو', 'Lenovo'),
    ('logitech', 'لوجيتيك', 'Logitech')
ON CONFLICT (slug) DO UPDATE SET name_ar = EXCLUDED.name_ar, name_fr = EXCLUDED.name_fr, updated_at = now();

INSERT INTO subcategory_brands (brand_id, category_slug, subcategory_slug, sort_order)
SELECT brands.id, mapping.category_slug, mapping.subcategory_slug, mapping.sort_order
FROM brands
JOIN (VALUES
    ('lumberjack', 'fashion', 'fashion-clothing', 1), ('zara', 'fashion', 'fashion-clothing', 2), ('lc-waikiki', 'fashion', 'fashion-clothing', 3),
    ('nike', 'fashion', 'fashion-shoes', 1), ('adidas', 'fashion', 'fashion-shoes', 2), ('puma', 'fashion', 'fashion-shoes', 3), ('skechers', 'fashion', 'fashion-shoes', 4),
    ('nike', 'fashion', 'fashion-sneakers', 1), ('adidas', 'fashion', 'fashion-sneakers', 2), ('new-balance', 'fashion', 'fashion-sneakers', 3), ('skechers', 'fashion', 'fashion-sneakers', 4),
    ('suzuki', 'cars', 'cars-used', 1), ('peugeot', 'cars', 'cars-used', 2), ('seat', 'cars', 'cars-used', 3), ('dacia', 'cars', 'cars-used', 4), ('chevrolet', 'cars', 'cars-used', 5),
    ('renault', 'cars', 'cars-new', 1), ('livan', 'cars', 'cars-new', 2), ('jac', 'cars', 'cars-new', 3), ('volkswagen', 'cars', 'cars-new', 4), ('geely', 'cars', 'cars-new', 5),
    ('apple', 'phones', 'phones-smartphones', 1), ('condor', 'electronics', 'electronics-appliances', 1), ('lenovo', 'electronics', 'electronics-computers', 1), ('logitech', 'electronics', 'electronics-computers', 2)
) AS mapping(slug, category_slug, subcategory_slug, sort_order) ON mapping.slug = brands.slug
ON CONFLICT (brand_id, category_slug, subcategory_slug) DO UPDATE SET sort_order = EXCLUDED.sort_order;
