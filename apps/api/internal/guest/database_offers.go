package guest

import (
	"context"
	"strings"
)

func (c *Catalog) currentOffers(ctx context.Context) []offer {
	items := append([]offer(nil), c.offers...)
	if c.database == nil {
		return items
	}
	databaseItems, err := c.databaseOffers(ctx)
	if err != nil {
		return items
	}
	return append(items, databaseItems...)
}

func (c *Catalog) databaseOffers(ctx context.Context) ([]offer, error) {
	rows, err := c.database.QueryContext(ctx, `
		SELECT o.id::text,o.title_ar,o.title_fr,o.description_ar,o.description_fr,
		       o.category_slug,o.subcategory_slug,COALESCE(o.brand_slug,''),
		       o.wilaya_code,o.city,o.price_before_dzd,o.price_after_dzd,o.starts_at,o.ends_at,
		       o.address,o.whatsapp,ms.slug,ms.name,ms.description,ms.address,ms.phone,ms.whatsapp
		FROM public_offers o JOIN merchant_shops ms ON ms.user_id=o.merchant_id
		ORDER BY o.starts_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []offer{}
	for rows.Next() {
		var item offer
		var categorySlug, subcategorySlug, brandSlug string
		var shopAddress, shopPhone, shopWhatsApp string
		if err := rows.Scan(&item.ID, &item.TitleAr, &item.TitleFr, &item.DescriptionAr, &item.DescriptionFr, &categorySlug, &subcategorySlug, &brandSlug, &item.WilayaCode, &item.City, &item.BeforePrice, &item.Price, &item.StartsAt, &item.EndsAt, &item.Address, &item.WhatsApp, &item.Shop.Slug, &item.Shop.Name, &item.Shop.Description, &shopAddress, &shopPhone, &shopWhatsApp); err != nil {
			return nil, err
		}
		item.ShopSlug = item.Shop.Slug
		item.Shop.Address = shopAddress
		item.Shop.Phone = shopPhone
		item.Shop.WhatsApp = shopWhatsApp
		item.Shop.Verified = true
		item.Status = "approved"
		item.AccessValid = true
		item.AudienceGender = "unisex"
		item.AgeMax = 120
		for _, categoryValue := range c.categories {
			if categoryValue.Slug == categorySlug || categoryValue.ID == categorySlug {
				item.Category = categoryValue
				for _, child := range categoryValue.Subcategories {
					if child.Slug == subcategorySlug || child.ID == subcategorySlug {
						item.Subcategory = child
						break
					}
				}
				break
			}
		}
		for _, brandValue := range c.brands {
			if brandValue.Slug == brandSlug {
				value := brandValue
				item.Brand = &value
				break
			}
		}
		for _, wilayaValue := range c.wilayas {
			if wilayaValue.Code == item.WilayaCode {
				item.Wilaya = wilayaValue.NameFr
				item.WilayaAr = wilayaValue.NameAr
				item.Shop.Wilaya = wilayaValue.NameFr
				for _, city := range wilayaValue.Cities {
					if strings.EqualFold(city.NameFr, item.City) || strings.EqualFold(city.NameAr, item.City) {
						item.CityID = city.ID
						item.CityAr = city.NameAr
						break
					}
				}
				break
			}
		}
		mediaRows, mediaErr := c.database.QueryContext(ctx, `SELECT id::text,object_key_400,object_key_800,object_key_1600,alt_text,position FROM offer_media WHERE offer_id=$1::uuid ORDER BY position`, item.ID)
		if mediaErr != nil {
			return nil, mediaErr
		}
		for mediaRows.Next() {
			var media image
			var key400, key800, key1600 string
			if err := mediaRows.Scan(&media.ID, &key400, &key800, &key1600, &media.Alt, &media.Position); err != nil {
				mediaRows.Close()
				return nil, err
			}
			media.URL400 = c.mediaBase + "/" + key400
			media.URL800 = c.mediaBase + "/" + key800
			media.URL1600 = c.mediaBase + "/" + key1600
			item.Images = append(item.Images, media)
		}
		mediaRows.Close()
		if len(item.Images) > 0 {
			items = append(items, item)
		}
	}
	return items, rows.Err()
}
