package auth

import (
	"database/sql"
	"os"
	"testing"
	"time"

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
}
