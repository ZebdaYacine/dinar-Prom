package auth

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestDatabasePersistenceSmoke(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	email := "db-smoke-" + time.Now().UTC().Format("20060102150405.000000000") + "@example.com"
	service, err := New(Config{Environment: "test", Database: db})
	if err != nil {
		t.Fatal(err)
	}
	user := service.LoginWithGoogleNames("google-id", email, "DB", "Smoke", "DB Smoke", "https://lh3.googleusercontent.com/avatar", true)
	defer db.Exec(`DELETE FROM users WHERE id = $1::uuid`, user.ID)
	defer db.Exec(`DELETE FROM notification_outbox WHERE recipient = $1`, email)
	if !service.Authorize(user.ID, "merchant_identity", "write") || service.Authorize(user.ID, "merchant_documents", "read") {
		t.Fatal("expected PostgreSQL Casbin policies to allow merchant onboarding and deny private document review")
	}

	if _, fields := service.SaveProfile(user.ID, Profile{FirstName: "DB", LastName: "Smoke", PhoneE164: "0551234567", WilayaID: "16", Address: "Algiers"}); len(fields) != 0 {
		t.Fatalf("profile was not persisted: %v", fields)
	}
	if _, fields := service.SaveShop(user.ID, Shop{Name: "DB Smoke Shop", CategoryID: "fashion", WilayaID: "16", RegisterNumber: "RC-DB-SMOKE", RegisterExpiresOn: "2099-01-01"}); len(fields) != 0 {
		t.Fatalf("shop was not persisted: %v", fields)
	}
	for _, typ := range []string{"id_front", "id_back", "commercial_register"} {
		if _, err := service.SaveDocument(user.ID, typ, []byte{0xff, 0xd8, 0xff, 0xe0, 0x00}); err != nil {
			t.Fatal(err)
		}
	}
	if _, fields := service.Submit(user.ID, true); len(fields) != 0 {
		t.Fatalf("application was not persisted: %v", fields)
	}

	secondLogin, err := New(Config{Environment: "test", Database: db})
	if err != nil {
		t.Fatal(err)
	}
	returning := secondLogin.LoginWithGoogleNames("google-id", email, "DB", "Smoke", "DB Smoke", "https://lh3.googleusercontent.com/avatar", true)
	if returning.ID != user.ID {
		t.Fatalf("expected returning Google account to reuse user %s, got %s", user.ID, returning.ID)
	}
	me, err := secondLogin.Current(returning.ID)
	if err != nil {
		t.Fatal(err)
	}
	if me.User.Picture == "" || me.Merchant.Status != StatusPending || me.Merchant.Profile == nil || me.Merchant.Shop == nil {
		t.Fatalf("expected persisted avatar, pending status, profile, and shop: %+v", me)
	}

	startApp := fiber.New()
	startApp.Post("/start", func(c *fiber.Ctx) error { return service.StartSession(user.ID, c) })
	startResponse, err := startApp.Test(httptest.NewRequest(http.MethodPost, "/start", nil))
	if err != nil {
		t.Fatal(err)
	}
	oldRefresh := responseCookie(startResponse, service.cfg.RefreshCookieName)
	csrf := responseCookie(startResponse, service.cfg.CSRFCookieName)
	if oldRefresh == nil || csrf == nil {
		t.Fatal("expected refresh and CSRF cookies")
	}
	refreshApp := fiber.New()
	Register(refreshApp.Group("/api/v1"), secondLogin)
	refreshRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	refreshRequest.AddCookie(oldRefresh)
	refreshRequest.AddCookie(csrf)
	refreshRequest.Header.Set(secondLogin.cfg.CSRFHeaderName, csrf.Value)
	refreshResponse, err := refreshApp.Test(refreshRequest)
	if err != nil || refreshResponse.StatusCode != fiber.StatusOK {
		t.Fatalf("expected persisted refresh rotation, status=%d err=%v", refreshResponse.StatusCode, err)
	}
	newRefresh := responseCookie(refreshResponse, secondLogin.cfg.RefreshCookieName)
	if newRefresh == nil || newRefresh.Value == oldRefresh.Value {
		t.Fatal("expected a rotated refresh token")
	}

	reuseRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	reuseRequest.AddCookie(oldRefresh)
	reuseRequest.AddCookie(csrf)
	reuseRequest.Header.Set(secondLogin.cfg.CSRFHeaderName, csrf.Value)
	reuseResponse, _ := refreshApp.Test(reuseRequest)
	if reuseResponse.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected refresh-token reuse to be rejected, got %d", reuseResponse.StatusCode)
	}
	newRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	newRequest.AddCookie(newRefresh)
	newRequest.AddCookie(csrf)
	newRequest.Header.Set(secondLogin.cfg.CSRFHeaderName, csrf.Value)
	newResponse, _ := refreshApp.Test(newRequest)
	if newResponse.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected token family revocation after reuse, got %d", newResponse.StatusCode)
	}
}

func responseCookie(response *http.Response, name string) *http.Cookie {
	for _, cookie := range response.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}
