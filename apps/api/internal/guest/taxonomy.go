package guest

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
)

// The product contract currently targets Algeria's 58-wilaya structure.
// Keep the source data embedded so development and production return the same
// complete location catalogue without an external dependency at request time.
//
//go:embed data/algeria_58_wilayas.json
var algeriaLocationsFS embed.FS

// Consumer catalogue is embedded for deterministic API responses in local and
// production environments. The same catalogue is mirrored into PostgreSQL by
// the seed migration used for reporting and future merchant/admin workflows.
//
//go:embed data/consumer-catalog.json
var consumerCatalogFS embed.FS

type consumerCatalog struct {
	Categories []category `json:"categories"`
	Brands     []brand    `json:"brands"`
}

type subcategory struct {
	ID     string `json:"id"`
	Slug   string `json:"slug"`
	NameAr string `json:"name_ar"`
	NameFr string `json:"name_fr"`
}

type brand struct {
	ID            string `json:"id"`
	Slug          string `json:"slug"`
	NameAr        string `json:"name_ar"`
	NameFr        string `json:"name_fr"`
	Origin        string `json:"origin,omitempty"`
	CategoryID    string `json:"category_id,omitempty"`
	SubcategoryID string `json:"subcategory_id,omitempty"`
}

type locationCity struct {
	ID      string `json:"id"`
	NameAr  string `json:"name_ar"`
	NameFr  string `json:"name_fr"`
	DairaAr string `json:"daira_ar"`
	DairaFr string `json:"daira_fr"`
}

type wilaya struct {
	Code   string         `json:"code"`
	NameAr string         `json:"name_ar"`
	NameFr string         `json:"name_fr"`
	Cities []locationCity `json:"cities"`
}

type locationRecord struct {
	ID               int    `json:"id"`
	CommuneNameASCII string `json:"commune_name_ascii"`
	CommuneName      string `json:"commune_name"`
	DairaNameASCII   string `json:"daira_name_ascii"`
	DairaName        string `json:"daira_name"`
	WilayaCode       string `json:"wilaya_code"`
	WilayaNameASCII  string `json:"wilaya_name_ascii"`
	WilayaName       string `json:"wilaya_name"`
}

func defaultCategories() []category {
	data, err := consumerCatalogFS.ReadFile("data/consumer-catalog.json")
	if err != nil {
		return nil
	}
	var catalog consumerCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil
	}
	return catalog.Categories
}

func defaultBrands() []brand {
	data, err := consumerCatalogFS.ReadFile("data/consumer-catalog.json")
	if err != nil {
		return nil
	}
	var catalog consumerCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil
	}
	for index := range catalog.Brands {
		if catalog.Brands[index].Origin == "" {
			catalog.Brands[index].Origin = "global"
		}
	}
	return catalog.Brands
}

func loadWilayas() []wilaya {
	raw, err := algeriaLocationsFS.ReadFile("data/algeria_58_wilayas.json")
	if err != nil {
		return nil
	}
	var records []locationRecord
	if err := json.Unmarshal(raw, &records); err != nil {
		return nil
	}
	byCode := make(map[string]*wilaya)
	for _, record := range records {
		item, ok := byCode[record.WilayaCode]
		if !ok {
			item = &wilaya{Code: record.WilayaCode, NameAr: record.WilayaName, NameFr: record.WilayaNameASCII}
			byCode[record.WilayaCode] = item
		}
		item.Cities = append(item.Cities, locationCity{ID: stringID(record.ID), NameAr: record.CommuneName, NameFr: record.CommuneNameASCII, DairaAr: record.DairaName, DairaFr: record.DairaNameASCII})
	}
	items := make([]wilaya, 0, len(byCode))
	for _, item := range byCode {
		sort.Slice(item.Cities, func(i, j int) bool { return item.Cities[i].NameFr < item.Cities[j].NameFr })
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Code < items[j].Code })
	return items
}

func stringID(value int) string {
	if value == 0 {
		return ""
	}
	return fmt.Sprintf("%d", value)
}
