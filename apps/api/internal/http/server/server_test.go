package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dinar-dz/dinar/internal/config"
	"github.com/gofiber/fiber/v2"
)

func TestHealthz(t *testing.T) {
	app := New(config.Config{Addr: ":8080", Environment: "test"}, slog.Default())
	response, err := app.Test(httptest.NewRequest("GET", "/healthz", nil))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", response.StatusCode)
	}
}

func TestOffersUsesPaginationEnvelope(t *testing.T) {
	app := New(config.Config{Addr: ":8080", Environment: "test"}, slog.Default())
	response, err := app.Test(httptest.NewRequest("GET", "/api/v1/offers?page=2&per_page=100", nil))
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Data []fiber.Map `json:"data"`
		Meta struct {
			Page    int `json:"page"`
			PerPage int `json:"per_page"`
			Total   int `json:"total"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Meta.Page != 2 || payload.Meta.PerPage != 48 {
		t.Fatalf("unexpected pagination: %+v", payload.Meta)
	}
	if payload.Meta.Total == 0 {
		t.Fatal("expected seeded public offers")
	}
}

func TestGuestSearchNormalizesArabic(t *testing.T) {
	app := New(config.Config{Addr: ":8080", Environment: "test"}, slog.Default())
	response, err := app.Test(httptest.NewRequest("GET", "/api/v1/offers?q=ازياء", nil))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", response.StatusCode)
	}
	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Data) != 1 || payload.Data[0].ID != "offer-001" {
		t.Fatalf("unexpected search results: %+v", payload.Data)
	}
}

func TestGuestAudienceFiltersUseGenderAndAgeOverlap(t *testing.T) {
	app := New(config.Config{Addr: ":8080", Environment: "test"}, slog.Default())
	response, err := app.Test(httptest.NewRequest("GET", "/api/v1/offers?category=beauty&gender=women&age_from=56&age_to=58", nil))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", response.StatusCode)
	}
	var payload struct {
		Data []struct {
			ID             string `json:"id"`
			AudienceGender string `json:"audience_gender"`
			AgeMin         int    `json:"age_min"`
			AgeMax         int    `json:"age_max"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Data) != 1 || payload.Data[0].ID != "demo-beauty-fragrance" {
		t.Fatalf("expected fragrance to overlap the requested range, got %+v", payload.Data)
	}
	item := payload.Data[0]
	if item.AudienceGender != "women" || item.AgeMin > 58 || item.AgeMax < 56 {
		t.Fatalf("unexpected audience metadata: %+v", item)
	}
}

func TestGuestTaxonomyLocationsAndDemoMedia(t *testing.T) {
	app := New(config.Config{Addr: ":8080", Environment: "test"}, slog.Default())

	categoriesResponse, err := app.Test(httptest.NewRequest("GET", "/api/v1/categories", nil))
	if err != nil {
		t.Fatal(err)
	}
	var categories struct {
		Data []struct {
			Icon          string     `json:"icon"`
			Subcategories []struct{} `json:"subcategories"`
		} `json:"data"`
	}
	if err := json.NewDecoder(categoriesResponse.Body).Decode(&categories); err != nil {
		t.Fatal(err)
	}
	if len(categories.Data) != 8 || categories.Data[0].Icon == "" || len(categories.Data[0].Subcategories) == 0 {
		t.Fatalf("unexpected category catalogue: %+v", categories.Data)
	}

	brandsResponse, err := app.Test(httptest.NewRequest("GET", "/api/v1/brands?category=fashion&subcategory=fashion-women", nil))
	if err != nil {
		t.Fatal(err)
	}
	var brands struct {
		Data []struct {
			Slug   string `json:"slug"`
			Origin string `json:"origin"`
		} `json:"data"`
	}
	if err := json.NewDecoder(brandsResponse.Body).Decode(&brands); err != nil {
		t.Fatal(err)
	}
	if len(brands.Data) != 5 || brands.Data[0].Slug != "zara" || brands.Data[0].Origin != "global" || brands.Data[3].Origin != "local" {
		t.Fatalf("unexpected fashion shoe brands: %+v", brands.Data)
	}

	brandOffersResponse, err := app.Test(httptest.NewRequest("GET", "/api/v1/offers?category=fashion&subcategory=fashion-women&brand=zara", nil))
	if err != nil {
		t.Fatal(err)
	}
	var brandOffers struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(brandOffersResponse.Body).Decode(&brandOffers); err != nil {
		t.Fatal(err)
	}
	if len(brandOffers.Data) < 1 {
		t.Fatal("expected lumberjack offers")
	}

	locationsResponse, err := app.Test(httptest.NewRequest("GET", "/api/v1/wilayas", nil))
	if err != nil {
		t.Fatal(err)
	}
	var locations struct {
		Meta struct {
			Total  int `json:"total"`
			Cities int `json:"cities"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(locationsResponse.Body).Decode(&locations); err != nil {
		t.Fatal(err)
	}
	if locations.Meta.Total != 58 || locations.Meta.Cities != 1541 {
		t.Fatalf("unexpected location catalogue: %+v", locations.Meta)
	}

	offersResponse, err := app.Test(httptest.NewRequest("GET", "/api/v1/offers?q=Smartphone", nil))
	if err != nil {
		t.Fatal(err)
	}
	var offers struct {
		Data []struct {
			ID       string `json:"id"`
			CoverURL string `json:"cover_url"`
		} `json:"data"`
	}
	if err := json.NewDecoder(offersResponse.Body).Decode(&offers); err != nil {
		t.Fatal(err)
	}
	if len(offers.Data) != 1 || offers.Data[0].ID != "demo-phone-smartphone" || !strings.Contains(offers.Data[0].CoverURL, "demo-products/demo-phone-smartphone.jpg") {
		t.Fatalf("unexpected demo offer: %+v", offers.Data)
	}
}

func TestWhatsAppAndReportAreDeduplicated(t *testing.T) {
	app := New(config.Config{Addr: ":8080", Environment: "test", AppURL: "http://test.local"}, slog.Default())
	req := httptest.NewRequest("POST", "/api/v1/offers/offer-001/whatsapp", nil)
	req.Header.Set("X-Visitor-ID", "visitor-1")
	first, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if first.StatusCode != fiber.StatusOK {
		t.Fatalf("expected first WhatsApp click to succeed, got %d", first.StatusCode)
	}
	secondReq := httptest.NewRequest("POST", "/api/v1/offers/offer-001/whatsapp", nil)
	secondReq.Header.Set("X-Visitor-ID", "visitor-1")
	second, err := app.Test(secondReq)
	if err != nil {
		t.Fatal(err)
	}
	if second.StatusCode != fiber.StatusOK {
		t.Fatalf("expected duplicate WhatsApp click to be accepted, got %d", second.StatusCode)
	}

	reportReq := httptest.NewRequest("POST", "/api/v1/offers/offer-001/report", strings.NewReader(`{"reason":"misleading","message":"The price is unclear"}`))
	reportReq.Header.Set("Content-Type", "application/json")
	reportReq.Header.Set("X-Visitor-ID", "visitor-2")
	report, err := app.Test(reportReq)
	if err != nil {
		t.Fatal(err)
	}
	if report.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected report creation, got %d", report.StatusCode)
	}
	duplicateReq := httptest.NewRequest("POST", "/api/v1/offers/offer-001/report", strings.NewReader(`{"reason":"misleading"}`))
	duplicateReq.Header.Set("Content-Type", "application/json")
	duplicateReq.Header.Set("X-Visitor-ID", "visitor-2")
	duplicate, err := app.Test(duplicateReq)
	if err != nil {
		t.Fatal(err)
	}
	if duplicate.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected duplicate report conflict, got %d", duplicate.StatusCode)
	}
}
