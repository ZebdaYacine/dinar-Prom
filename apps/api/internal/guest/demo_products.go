package guest

import (
	_ "embed"
	"encoding/json"
	"os"
	"time"
)

//go:embed data/demo-products.json
var demoProductsJSON []byte

type demoProduct struct {
	ID             string `json:"id"`
	TitleAr        string `json:"title_ar"`
	TitleFr        string `json:"title_fr"`
	DescriptionAr  string `json:"description_ar"`
	DescriptionFr  string `json:"description_fr"`
	Category       string `json:"category"`
	Subcategory    string `json:"subcategory"`
	Brand          string `json:"brand"`
	WilayaCode     string `json:"wilaya_code"`
	City           string `json:"city"`
	CityAr         string `json:"city_ar"`
	ObjectKey      string `json:"object_key"`
	AudienceGender string `json:"audience_gender"`
	AgeMin         int    `json:"age_min"`
	AgeMax         int    `json:"age_max"`
}

func loadDemoOffers(categories []category, wilayas []wilaya, brands []brand, now time.Time) []offer {
	var products []demoProduct
	if err := json.Unmarshal(demoProductsJSON, &products); err != nil {
		return nil
	}
	categoryByID := make(map[string]category, len(categories))
	for _, item := range categories {
		categoryByID[item.ID] = item
	}
	mediaBase := os.Getenv("DEMO_MEDIA_BASE_URL")
	if mediaBase == "" {
		mediaBase = "http://localhost:9100/dinar-public"
	}
	mediaBase = trimRightSlash(mediaBase)
	items := make([]offer, 0, len(products))
	for index, product := range products {
		cat, ok := categoryByID[product.Category]
		if !ok {
			continue
		}
		child := findSubcategory(cat, product.Subcategory)
		if child.ID == "" {
			continue
		}
		wilayaName := product.WilayaCode
		wilayaNameAr := product.WilayaCode
		cityID := ""
		for _, location := range wilayas {
			if location.Code != product.WilayaCode {
				continue
			}
			wilayaName = location.NameFr
			wilayaNameAr = location.NameAr
			for _, city := range location.Cities {
				if city.NameFr == product.City {
					cityID = city.ID
					break
				}
			}
			break
		}
		publicURL := mediaBase + "/" + product.ObjectKey
		productBrand := findBrand(brands, product.Brand)
		items = append(items, offer{
			ID: product.ID, TitleAr: product.TitleAr, TitleFr: product.TitleFr,
			DescriptionAr: product.DescriptionAr, DescriptionFr: product.DescriptionFr, Brand: productBrand,
			Category: cat, Subcategory: child, ShopSlug: "demo-marketplace",
			Wilaya: wilayaName, WilayaAr: wilayaNameAr, WilayaCode: product.WilayaCode,
			City: product.City, CityAr: product.CityAr, CityID: cityID,
			AudienceGender: product.AudienceGender, AgeMin: product.AgeMin, AgeMax: product.AgeMax,
			BeforePrice: 1000 + (index+1)*500, Price: 700 + (index+1)*300,
			StartsAt: now.Add(-time.Duration(index+1) * time.Hour), EndsAt: now.Add(21 * 24 * time.Hour),
			Status: "approved", AccessValid: true,
			Images:  []image{{ID: product.ID + "-cover", URL400: publicURL, URL800: publicURL, URL1600: publicURL, Alt: product.TitleAr, Position: 1}},
			Address: product.City, WhatsApp: "213550000099", Phone: "021000099",
		})
	}
	return items
}

func findBrand(items []brand, slug string) *brand {
	for _, item := range items {
		if item.Slug == slug {
			value := item
			return &value
		}
	}
	return nil
}

func findSubcategory(item category, id string) subcategory {
	for _, child := range item.Subcategories {
		if child.ID == id {
			return child
		}
	}
	return subcategory{}
}

func trimRightSlash(value string) string {
	for len(value) > 0 && value[len(value)-1] == '/' {
		value = value[:len(value)-1]
	}
	return value
}
