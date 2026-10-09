package guest

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gofiber/fiber/v2"
)

const (
	defaultPageSize = 24
	maxPageSize     = 48
	visitorWindow   = time.Hour
)

type Catalog struct {
	mu             sync.RWMutex
	appURL         string
	offers         []offer
	shops          map[string]shop
	categories     []category
	brands         []brand
	wilayas        []wilaya
	views          map[string]time.Time
	whatsappClicks map[string]time.Time
	reports        map[string]time.Time
	reportWindows  map[string][]time.Time
	analytics      *analyticsStore
}

type offer struct {
	ID             string
	TitleAr        string
	TitleFr        string
	DescriptionAr  string
	DescriptionFr  string
	Category       category
	Subcategory    subcategory
	Brand          *brand
	ShopSlug       string
	Wilaya         string
	WilayaAr       string
	WilayaCode     string
	City           string
	CityAr         string
	CityID         string
	AudienceGender string
	AgeMin         int
	AgeMax         int
	BeforePrice    int
	Price          int
	StartsAt       time.Time
	EndsAt         time.Time
	Status         string
	AccessValid    bool
	Images         []image
	Address        string
	WhatsApp       string
	Phone          string
}

type category struct {
	ID            string        `json:"id"`
	Slug          string        `json:"slug"`
	NameAr        string        `json:"name_ar"`
	NameFr        string        `json:"name_fr"`
	Icon          string        `json:"icon"`
	Subcategories []subcategory `json:"subcategories"`
}

type shop struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Wilaya      string `json:"wilaya"`
	Address     string `json:"address"`
	WhatsApp    string `json:"whatsapp"`
	Phone       string `json:"phone"`
	Verified    bool   `json:"verified"`
}

type image struct {
	ID       string `json:"id"`
	URL400   string `json:"url_400"`
	URL800   string `json:"url_800"`
	URL1600  string `json:"url_1600"`
	Alt      string `json:"alt"`
	Position int    `json:"position"`
}

type OfferSummary struct {
	ID              string      `json:"id"`
	Title           string      `json:"title"`
	Category        category    `json:"category"`
	Subcategory     subcategory `json:"subcategory"`
	Brand           *brand      `json:"brand,omitempty"`
	Shop            shop        `json:"shop"`
	Wilaya          string      `json:"wilaya"`
	WilayaCode      string      `json:"wilaya_code"`
	City            string      `json:"city"`
	CityID          string      `json:"city_id"`
	AudienceGender  string      `json:"audience_gender"`
	AgeMin          int         `json:"age_min"`
	AgeMax          int         `json:"age_max"`
	CoverURL        string      `json:"cover_url"`
	BeforePrice     int         `json:"before_price"`
	Price           int         `json:"price"`
	DiscountPercent int         `json:"discount_percent"`
	StartsAt        time.Time   `json:"starts_at"`
	EndsAt          time.Time   `json:"ends_at"`
	Verified        bool        `json:"verified"`
}

type OfferDetails struct {
	OfferSummary
	Description string         `json:"description"`
	Address     string         `json:"address"`
	Phone       string         `json:"phone,omitempty"`
	WhatsAppURL string         `json:"whatsapp_url"`
	Images      []image        `json:"images"`
	Similar     []OfferSummary `json:"similar_offers"`
}

type pageMeta struct {
	Page    int  `json:"page"`
	PerPage int  `json:"per_page"`
	Total   int  `json:"total"`
	HasNext bool `json:"has_next"`
}

type listQuery struct {
	Query          string
	Wilaya         string
	Category       string
	Subcategory    string
	Brand          string
	City           string
	AudienceGender string
	AgeMin         int
	AgeMax         int
	MinDiscount    int
	EndingWithin   time.Duration
	Sort           string
	Page           int
	PerPage        int
}

func NewCatalog(appURL string, databases ...*sql.DB) *Catalog {
	if appURL == "" {
		appURL = "http://localhost:8180"
	}
	now := time.Now().UTC()
	categories := defaultCategories()
	categoryByID := make(map[string]category, len(categories))
	for _, item := range categories {
		categoryByID[item.ID] = item
	}
	catFashion := categoryByID["fashion"]
	catFood := categoryByID["food"]
	catFurniture := categoryByID["furniture"]
	findSubcategory := func(item category, id string) subcategory {
		for _, child := range item.Subcategories {
			if child.ID == id {
				return child
			}
		}
		return subcategory{}
	}
	fashionWomen := findSubcategory(catFashion, "fashion-women")
	foodGrocery := findSubcategory(catFood, "food-grocery")
	furnitureLiving := findSubcategory(catFurniture, "furniture-living")
	brands := defaultBrands()
	brandBySlug := func(slug string) *brand {
		for index := range brands {
			if brands[index].Slug == slug {
				value := brands[index]
				return &value
			}
		}
		return nil
	}
	shops := map[string]shop{
		"el-bahia-fashion": {Slug: "el-bahia-fashion", Name: "El Bahia Fashion", Description: "ملابس عصرية بأسعار مناسبة", Wilaya: "Oran", Address: "Hai El Yasmine, Oran", WhatsApp: "213550000001", Phone: "041000001", Verified: true},
		"dar-dz":           {Slug: "dar-dz", Name: "Dar DZ", Description: "منتجات منزلية جزائرية", Wilaya: "Algiers", Address: "Bab Ezzouar, Algiers", WhatsApp: "213550000002", Phone: "021000002", Verified: true},
		"cafe-constantine": {Slug: "cafe-constantine", Name: "Café Constantine", Description: "نكهات قسنطينية أصيلة", Wilaya: "Constantine", Address: "Centre Ville, Constantine", WhatsApp: "213550000003", Phone: "031000003", Verified: true},
		"demo-marketplace": {Slug: "demo-marketplace", Name: "Dinar DZ Demo Market", Description: "Fake seed products for local development", Wilaya: "Algeria", Address: "Demo catalogue", WhatsApp: "213550000099", Phone: "021000099", Verified: true},
	}
	imageFor := func(id, title string) []image {
		base := "https://images.unsplash.com/photo-1441986300917-64674bd600d8"
		return []image{{ID: id + "-cover", URL400: base + "?w=400", URL800: base + "?w=800", URL1600: base + "?w=1600", Alt: title, Position: 1}}
	}
	// Public product photo from Ouedkniss used for the featured fashion offer.
	fashionImage := []image{{ID: "offer-001-cover", URL400: "https://cdn7.ouedkniss.com/500/medias/announcements/images/jlmW/tVpl4xQj4oXFsrlTbxjh2lQspcGAw5S9ECv0FZEb.webp", URL800: "https://cdn7.ouedkniss.com/500/medias/announcements/images/jlmW/tVpl4xQj4oXFsrlTbxjh2lQspcGAw5S9ECv0FZEb.webp", URL1600: "https://cdn7.ouedkniss.com/500/medias/announcements/images/jlmW/tVpl4xQj4oXFsrlTbxjh2lQspcGAw5S9ECv0FZEb.webp", Alt: "تخفيضات الصيف على الأزياء", Position: 1}}
	offers := []offer{
		{ID: "offer-001", TitleAr: "تخفيضات الصيف على الأزياء", TitleFr: "Soldes d'été sur la mode", DescriptionAr: "استفد من تخفيضات تصل إلى 50٪ على التشكيلة الجديدة.", DescriptionFr: "Profitez de réductions jusqu'à 50 % sur la nouvelle collection.", Category: catFashion, Subcategory: fashionWomen, Brand: brandBySlug("zara"), ShopSlug: "el-bahia-fashion", Wilaya: "Oran", WilayaAr: "وهران", WilayaCode: "31", City: "Oran", CityAr: "وهران", CityID: "1131", AudienceGender: "women", AgeMin: 18, AgeMax: 45, BeforePrice: 8000, Price: 4000, StartsAt: now.Add(-30 * time.Minute), EndsAt: now.Add(12 * 24 * time.Hour), Status: "approved", AccessValid: true, Images: fashionImage, Address: shops["el-bahia-fashion"].Address, WhatsApp: shops["el-bahia-fashion"].WhatsApp, Phone: shops["el-bahia-fashion"].Phone},
		{ID: "offer-002", TitleAr: "وجبة عائلية لشخصين", TitleFr: "Menu familial pour deux", DescriptionAr: "وجبة كاملة مع مشروبين وحلوى بسعر خاص.", DescriptionFr: "Menu complet avec deux boissons et dessert à prix spécial.", Category: catFood, Subcategory: foodGrocery, ShopSlug: "cafe-constantine", Wilaya: "Constantine", WilayaAr: "قسنطينة", WilayaCode: "25", City: "Constantine", CityAr: "قسنطينة", CityID: "891", AudienceGender: "unisex", AgeMin: 0, AgeMax: 120, BeforePrice: 2400, Price: 1800, StartsAt: now.Add(-2 * time.Hour), EndsAt: now.Add(36 * time.Hour), Status: "approved", AccessValid: true, Images: imageFor("offer-002", "وجبة عائلية لشخصين"), Address: shops["cafe-constantine"].Address, WhatsApp: shops["cafe-constantine"].WhatsApp, Phone: shops["cafe-constantine"].Phone},
		{ID: "offer-003", TitleAr: "طقم منزلي أنيق", TitleFr: "Ensemble maison élégant", DescriptionAr: "أضف لمسة دافئة إلى منزلك مع هذا الطقم الجزائري.", DescriptionFr: "Ajoutez une touche chaleureuse à votre maison avec cet ensemble algérien.", Category: catFurniture, Subcategory: furnitureLiving, ShopSlug: "dar-dz", Wilaya: "Algiers", WilayaAr: "الجزائر", WilayaCode: "16", City: "Alger Centre", CityAr: "الجزائر الوسطى", CityID: "556", AudienceGender: "unisex", AgeMin: 18, AgeMax: 80, BeforePrice: 6500, Price: 4550, StartsAt: now.Add(-5 * 24 * time.Hour), EndsAt: now.Add(20 * 24 * time.Hour), Status: "approved", AccessValid: true, Images: imageFor("offer-003", "طقم منزلي أنيق"), Address: shops["dar-dz"].Address, WhatsApp: shops["dar-dz"].WhatsApp, Phone: shops["dar-dz"].Phone},
		{ID: "offer-004", TitleAr: "قهوة وحلوى بعد الظهر", TitleFr: "Café et douceur de l'après-midi", DescriptionAr: "عرض يومي من الساعة 14 إلى 18.", DescriptionFr: "Offre quotidienne de 14 h à 18 h.", Category: catFood, Subcategory: foodGrocery, ShopSlug: "cafe-constantine", Wilaya: "Constantine", WilayaAr: "قسنطينة", WilayaCode: "25", City: "Constantine", CityAr: "قسنطينة", CityID: "891", AudienceGender: "unisex", AgeMin: 0, AgeMax: 120, BeforePrice: 900, Price: 540, StartsAt: now.Add(-time.Hour), EndsAt: now.Add(5 * 24 * time.Hour), Status: "approved", AccessValid: true, Images: imageFor("offer-004", "قهوة وحلوى بعد الظهر"), Address: shops["cafe-constantine"].Address, WhatsApp: shops["cafe-constantine"].WhatsApp, Phone: shops["cafe-constantine"].Phone},
	}
	wilayas := loadWilayas()
	offers = append(offers, loadDemoOffers(categories, wilayas, brands, now)...)
	var analytics *analyticsStore
	if len(databases) > 0 && databases[0] != nil {
		analytics = newAnalyticsStore(databases[0])
	}
	return &Catalog{appURL: strings.TrimRight(appURL, "/"), offers: offers, shops: shops, categories: categories, brands: brands, wilayas: wilayas, views: make(map[string]time.Time), whatsappClicks: make(map[string]time.Time), reports: make(map[string]time.Time), reportWindows: make(map[string][]time.Time), analytics: analytics}
}

func (c *Catalog) Register(router fiber.Router) {
	router.Get("/offers", c.ListOffers)
	router.Get("/offers/:id", c.GetOffer)
	router.Get("/offers/:id/similar", c.GetSimilar)
	router.Post("/offers/:id/view", c.RecordView)
	router.Post("/offers/:id/whatsapp", c.RecordWhatsApp)
	router.Post("/offers/:id/report", c.ReportOffer)
	router.Get("/shops/:slug", c.GetShop)
	router.Get("/categories", c.ListCategories)
	router.Get("/brands", c.ListBrands)
	router.Get("/wilayas", c.ListWilayas)
	router.Get("/wilayas/:code/cities", c.ListCities)
	router.Get("/ads", c.ListAds)
}

func (c *Catalog) ListOffers(ctx *fiber.Ctx) error {
	query := parseQuery(ctx)
	c.mu.RLock()
	defer c.mu.RUnlock()
	items := make([]OfferSummary, 0, len(c.offers))
	for _, item := range c.offers {
		if !c.isPublic(item, time.Now().UTC()) || !matches(item, query) {
			continue
		}
		items = append(items, c.summary(item, ctx.Get("Accept-Language")))
	}
	sortSummaries(items, query.Sort)
	return pageResponse(ctx, items, query.Page, query.PerPage)
}

func (c *Catalog) GetOffer(ctx *fiber.Ctx) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.findPublic(ctx.Params("id"))
	if !ok {
		return errorResponse(ctx, fiber.StatusNotFound, "not_found", "Offer not found")
	}
	return success(ctx, c.details(item, ctx.Get("Accept-Language")))
}

func (c *Catalog) GetSimilar(ctx *fiber.Ctx) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.findPublic(ctx.Params("id"))
	if !ok {
		return errorResponse(ctx, fiber.StatusNotFound, "not_found", "Offer not found")
	}
	items := make([]OfferSummary, 0, 3)
	for _, candidate := range c.offers {
		if candidate.ID == item.ID || !c.isPublic(candidate, time.Now().UTC()) || candidate.Category.ID != item.Category.ID {
			continue
		}
		items = append(items, c.summary(candidate, ctx.Get("Accept-Language")))
		if len(items) == 3 {
			break
		}
	}
	return success(ctx, items)
}

func (c *Catalog) RecordView(ctx *fiber.Ctx) error {
	c.mu.RLock()
	item, ok := c.findPublic(ctx.Params("id"))
	c.mu.RUnlock()
	if !ok {
		return errorResponse(ctx, fiber.StatusNotFound, "not_found", "Offer not found")
	}
	if c.analytics != nil {
		recorded, err := c.analytics.recordEvent(ctx.Context(), "view", item.ID, visitorKey(ctx), time.Now().UTC())
		if err == nil {
			return success(ctx, fiber.Map{"recorded": recorded})
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	key := "view:" + item.ID + ":" + visitorKey(ctx)
	now := time.Now().UTC()
	recorded := false
	if previous, exists := c.views[key]; !exists || now.Sub(previous) >= visitorWindow {
		c.views[key] = now
		recorded = true
	}
	return success(ctx, fiber.Map{"recorded": recorded})
}

func (c *Catalog) RecordWhatsApp(ctx *fiber.Ctx) error {
	c.mu.RLock()
	item, ok := c.findPublic(ctx.Params("id"))
	c.mu.RUnlock()
	if !ok {
		return errorResponse(ctx, fiber.StatusNotFound, "not_found", "Offer not found")
	}
	if c.analytics != nil {
		recorded, err := c.analytics.recordEvent(ctx.Context(), "whatsapp", item.ID, visitorKey(ctx), time.Now().UTC())
		if err == nil {
			return success(ctx, fiber.Map{"url": c.whatsappURL(item), "recorded": recorded})
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now().UTC()
	key := "whatsapp:" + item.ID + ":" + visitorKey(ctx)
	if previous, exists := c.whatsappClicks[key]; exists && now.Sub(previous) < visitorWindow {
		return success(ctx, fiber.Map{"url": c.whatsappURL(item), "recorded": false})
	}
	c.whatsappClicks[key] = now
	return success(ctx, fiber.Map{"url": c.whatsappURL(item), "recorded": true})
}

func (c *Catalog) ReportOffer(ctx *fiber.Ctx) error {
	c.mu.RLock()
	item, ok := c.findPublic(ctx.Params("id"))
	c.mu.RUnlock()
	if !ok {
		return errorResponse(ctx, fiber.StatusNotFound, "not_found", "Offer not found")
	}
	var body struct {
		Reason  string `json:"reason"`
		Message string `json:"message"`
	}
	if err := ctx.BodyParser(&body); err != nil {
		return errorResponse(ctx, fiber.StatusUnprocessableEntity, "validation_failed", "Invalid JSON body")
	}
	allowed := map[string]bool{"misleading": true, "expired": true, "wrong_price": true, "inappropriate": true, "other": true}
	if !allowed[body.Reason] || len([]rune(body.Message)) > 1000 {
		return errorResponse(ctx, fiber.StatusUnprocessableEntity, "validation_failed", "Choose a valid reason and keep the message under 1000 characters")
	}
	visitor := visitorKey(ctx)
	now := time.Now().UTC()
	if c.analytics != nil {
		result, err := c.analytics.recordReport(ctx.Context(), item.ID, visitor, body.Reason, body.Message, now)
		if err == nil {
			switch result {
			case reportDuplicate:
				return errorResponse(ctx, fiber.StatusConflict, "already_reported", "This offer was already reported")
			case reportRateLimited:
				return errorResponse(ctx, fiber.StatusTooManyRequests, "rate_limited", "Too many reports; try again later")
			default:
				return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"data": fiber.Map{"accepted": true, "offer_id": item.ID}})
			}
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	duplicateKey := item.ID + ":" + visitor
	if previous, exists := c.reports[duplicateKey]; exists && now.Sub(previous) < 24*time.Hour {
		return errorResponse(ctx, fiber.StatusConflict, "already_reported", "This offer was already reported")
	}
	windowKey := "hour:" + visitor
	window := c.reportWindows[windowKey]
	filtered := window[:0]
	for _, timestamp := range window {
		if now.Sub(timestamp) < visitorWindow {
			filtered = append(filtered, timestamp)
		}
	}
	if len(filtered) >= 5 {
		return errorResponse(ctx, fiber.StatusTooManyRequests, "rate_limited", "Too many reports; try again later")
	}
	c.reportWindows[windowKey] = append(filtered, now)
	c.reports[duplicateKey] = now
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"data": fiber.Map{"accepted": true, "offer_id": item.ID}})
}

func (c *Catalog) GetShop(ctx *fiber.Ctx) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.shops[ctx.Params("slug")]
	if !ok {
		return errorResponse(ctx, fiber.StatusNotFound, "not_found", "Shop not found")
	}
	items := make([]OfferSummary, 0)
	for _, item := range c.offers {
		if item.ShopSlug == value.Slug && c.isPublic(item, time.Now().UTC()) {
			items = append(items, c.summary(item, ctx.Get("Accept-Language")))
		}
	}
	return success(ctx, fiber.Map{"shop": value, "offers": items})
}

func (c *Catalog) ListCategories(ctx *fiber.Ctx) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return successWithMeta(ctx, c.categories, fiber.Map{"total": len(c.categories)})
}

func (c *Catalog) ListBrands(ctx *fiber.Ctx) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	categoryID := strings.TrimSpace(ctx.Query("category"))
	subcategoryID := strings.TrimSpace(ctx.Query("subcategory"))
	items := make([]brand, 0)
	for _, item := range c.brands {
		if categoryID != "" && item.CategoryID != categoryID {
			continue
		}
		if subcategoryID != "" && item.SubcategoryID != subcategoryID {
			continue
		}
		items = append(items, item)
	}
	return successWithMeta(ctx, items, fiber.Map{"total": len(items), "category": categoryID, "subcategory": subcategoryID})
}

func (c *Catalog) ListWilayas(ctx *fiber.Ctx) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return successWithMeta(ctx, c.wilayas, fiber.Map{"total": len(c.wilayas), "cities": countCities(c.wilayas)})
}

func (c *Catalog) ListCities(ctx *fiber.Ctx) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, item := range c.wilayas {
		if item.Code == ctx.Params("code") {
			return successWithMeta(ctx, item.Cities, fiber.Map{"wilaya": item.Code, "total": len(item.Cities)})
		}
	}
	return errorResponse(ctx, fiber.StatusNotFound, "wilaya_not_found", "Wilaya not found")
}

func (c *Catalog) ListAds(ctx *fiber.Ctx) error {
	return success(ctx, []fiber.Map{})
}

func countCities(items []wilaya) int {
	total := 0
	for _, item := range items {
		total += len(item.Cities)
	}
	return total
}

func (c *Catalog) isPublic(item offer, now time.Time) bool {
	return item.Status == "approved" && item.AccessValid && !item.StartsAt.After(now) && item.EndsAt.After(now)
}

func (c *Catalog) findPublic(id string) (offer, bool) {
	now := time.Now().UTC()
	for _, item := range c.offers {
		if item.ID == id && c.isPublic(item, now) {
			return item, true
		}
	}
	return offer{}, false
}

func (c *Catalog) summary(item offer, language string) OfferSummary {
	title := item.TitleFr
	wilaya := item.Wilaya
	city := item.City
	if strings.HasPrefix(strings.ToLower(language), "ar") || language == "" {
		title = item.TitleAr
		wilaya = item.WilayaAr
		city = item.CityAr
	}
	return OfferSummary{ID: item.ID, Title: title, Category: item.Category, Subcategory: item.Subcategory, Brand: item.Brand, Shop: c.shops[item.ShopSlug], Wilaya: wilaya, WilayaCode: item.WilayaCode, City: city, CityID: item.CityID, AudienceGender: item.AudienceGender, AgeMin: item.AgeMin, AgeMax: item.AgeMax, CoverURL: item.Images[0].URL800, BeforePrice: item.BeforePrice, Price: item.Price, DiscountPercent: discount(item.BeforePrice, item.Price), StartsAt: item.StartsAt, EndsAt: item.EndsAt, Verified: c.shops[item.ShopSlug].Verified}
}

func (c *Catalog) details(item offer, language string) OfferDetails {
	title := c.summary(item, language)
	description := item.DescriptionFr
	if strings.HasPrefix(strings.ToLower(language), "ar") || language == "" {
		description = item.DescriptionAr
	}
	detail := OfferDetails{OfferSummary: title, Description: description, Address: item.Address, Phone: item.Phone, WhatsAppURL: c.whatsappURL(item), Images: item.Images, Similar: make([]OfferSummary, 0)}
	for _, candidate := range c.offers {
		if candidate.ID != item.ID && c.isPublic(candidate, time.Now().UTC()) && candidate.Category.ID == item.Category.ID {
			detail.Similar = append(detail.Similar, c.summary(candidate, language))
		}
	}
	if len(detail.Similar) > 3 {
		detail.Similar = detail.Similar[:3]
	}
	return detail
}

func (c *Catalog) whatsappURL(item offer) string {
	message := fmt.Sprintf("%s — %s/offers/%s", item.TitleFr, c.appURL, item.ID)
	return "https://wa.me/" + item.WhatsApp + "?text=" + url.QueryEscape(message)
}

func parseQuery(ctx *fiber.Ctx) listQuery {
	query := listQuery{Query: normalize(ctx.Query("q")), Wilaya: strings.TrimSpace(ctx.Query("wilaya")), Category: strings.TrimSpace(ctx.Query("category")), Subcategory: strings.TrimSpace(ctx.Query("subcategory")), Brand: strings.TrimSpace(ctx.Query("brand")), City: strings.TrimSpace(ctx.Query("city")), AudienceGender: strings.TrimSpace(ctx.Query("gender", ctx.Query("audience_gender", ""))), Sort: ctx.Query("sort", "newest"), Page: positiveInt(ctx.Query("page"), 1), PerPage: positiveInt(ctx.Query("per_page"), defaultPageSize)}
	if query.PerPage > maxPageSize {
		query.PerPage = maxPageSize
	}
	query.MinDiscount, _ = strconv.Atoi(ctx.Query("min_discount", "0"))
	query.AgeMin = clampAge(ctx.Query("age_from", ctx.Query("age_min", "0")))
	query.AgeMax = clampAge(ctx.Query("age_to", ctx.Query("age_max", "0")))
	if query.AgeMax > 0 && query.AgeMin > query.AgeMax {
		query.AgeMin, query.AgeMax = query.AgeMax, query.AgeMin
	}
	ending := ctx.Query("ending_within", ctx.Query("ending", ""))
	if strings.HasSuffix(ending, "h") {
		hours, _ := strconv.Atoi(strings.TrimSuffix(ending, "h"))
		query.EndingWithin = time.Duration(hours) * time.Hour
	} else if days, err := strconv.Atoi(strings.TrimSuffix(ending, "d")); err == nil && days > 0 {
		query.EndingWithin = time.Duration(days) * 24 * time.Hour
	}
	return query
}

func clampAge(value string) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0
	}
	if parsed > 120 {
		return 120
	}
	return parsed
}

func matches(item offer, query listQuery) bool {
	if query.Wilaya != "" && !strings.EqualFold(item.Wilaya, query.Wilaya) && !strings.EqualFold(item.WilayaAr, query.Wilaya) && !strings.EqualFold(item.WilayaCode, query.Wilaya) {
		return false
	}
	if query.Category != "" && !strings.EqualFold(item.Category.ID, query.Category) && !strings.EqualFold(item.Category.Slug, query.Category) {
		return false
	}
	if query.Subcategory != "" && !strings.EqualFold(item.Subcategory.ID, query.Subcategory) && !strings.EqualFold(item.Subcategory.Slug, query.Subcategory) {
		return false
	}
	if query.Brand != "" && (item.Brand == nil || (!strings.EqualFold(item.Brand.ID, query.Brand) && !strings.EqualFold(item.Brand.Slug, query.Brand))) {
		return false
	}
	if query.City != "" && !strings.EqualFold(item.CityID, query.City) && !strings.EqualFold(item.City, query.City) && !strings.EqualFold(item.CityAr, query.City) {
		return false
	}
	if query.AudienceGender != "" && query.AudienceGender != "all" && item.AudienceGender != query.AudienceGender && item.AudienceGender != "unisex" {
		return false
	}
	if query.AgeMin > 0 && item.AgeMax < query.AgeMin {
		return false
	}
	if query.AgeMax > 0 && item.AgeMin > query.AgeMax {
		return false
	}
	if query.MinDiscount > 0 && discount(item.BeforePrice, item.Price) < query.MinDiscount {
		return false
	}
	if query.EndingWithin > 0 && item.EndsAt.Sub(time.Now().UTC()) > query.EndingWithin {
		return false
	}
	if query.Query != "" {
		brandText := ""
		if item.Brand != nil {
			brandText = " " + item.Brand.NameAr + " " + item.Brand.NameFr
		}
		searchable := normalize(item.TitleAr + " " + item.TitleFr + " " + item.DescriptionAr + " " + item.DescriptionFr + " " + item.Wilaya + " " + item.WilayaAr + " " + item.City + " " + item.CityAr + " " + item.Category.NameAr + " " + item.Category.NameFr + " " + item.Subcategory.NameAr + " " + item.Subcategory.NameFr + brandText)
		return strings.Contains(searchable, query.Query)
	}
	return true
}

func sortSummaries(items []OfferSummary, sortBy string) {
	sort.SliceStable(items, func(i, j int) bool {
		switch sortBy {
		case "ending_soon", "endingSoon":
			return items[i].EndsAt.Before(items[j].EndsAt)
		case "discount":
			return items[i].DiscountPercent > items[j].DiscountPercent
		case "price_asc", "priceAsc":
			return items[i].Price < items[j].Price
		default:
			return items[i].StartsAt.After(items[j].StartsAt)
		}
	})
}

func pageResponse(ctx *fiber.Ctx, items []OfferSummary, page, perPage int) error {
	total := len(items)
	start := (page - 1) * perPage
	if start > total {
		start = total
	}
	end := start + perPage
	if end > total {
		end = total
	}
	data := items[start:end]
	return successWithMeta(ctx, data, pageMeta{Page: page, PerPage: perPage, Total: total, HasNext: end < total})
}

func discount(before, current int) int {
	if before <= 0 || current >= before {
		return 0
	}
	return int(float64(before-current) / float64(before) * 100)
}

func normalize(value string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(value) {
		switch r {
		case 'أ', 'إ', 'آ', 'ٱ':
			r = 'ا'
		case 'ى':
			r = 'ي'
		case 'ة':
			r = 'ه'
		case 'ـ':
			continue
		}
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		builder.WriteRune(r)
	}
	return strings.Join(strings.Fields(builder.String()), " ")
}

func visitorKey(ctx *fiber.Ctx) string {
	value := ctx.Get("X-Visitor-ID")
	if value == "" {
		value = ctx.IP()
	}
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

func success(ctx *fiber.Ctx, data any) error {
	return ctx.JSON(fiber.Map{"data": data})
}

func successWithMeta(ctx *fiber.Ctx, data any, meta any) error {
	return ctx.JSON(fiber.Map{"data": data, "meta": meta})
}

func errorResponse(ctx *fiber.Ctx, status int, code, message string) error {
	return ctx.Status(status).JSON(fiber.Map{"error": fiber.Map{"code": code, "message": message, "request_id": ctx.Locals("requestid")}})
}

func positiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}
