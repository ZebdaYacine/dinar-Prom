package offers

import (
	"bytes"
	"context"
	"database/sql"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"testing"
	"time"

	"github.com/dinar-dz/dinar/internal/storage"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestDatabaseOfferLifecycle(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not configured")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	merchantID, adminID := uuid.NewString(), uuid.NewString()
	email := "offer-smoke-" + uuid.NewString() + "@example.com"
	adminEmail := "offer-admin-" + uuid.NewString() + "@example.com"
	defer func() {
		_, _ = db.Exec(`DELETE FROM notification_outbox WHERE recipient IN ($1,$2)`, email, adminEmail)
		_, _ = db.Exec(`DELETE FROM audit_log WHERE actor_user_id=$1::uuid`, adminID)
		_, _ = db.Exec(`DELETE FROM users WHERE id IN ($1::uuid,$2::uuid)`, merchantID, adminID)
	}()
	if _, err = db.Exec(`INSERT INTO users (id,email,name,role,email_verified) VALUES ($1::uuid,$2,'Offer Merchant','merchant',true),($3::uuid,$4,'Offer Admin','admin',true)`, merchantID, email, adminID, adminEmail); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO merchant_verifications (user_id,status) VALUES ($1::uuid,'approved')`, merchantID); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO merchant_shops (user_id,name,slug,category_id,wilaya_id,city,address,whatsapp,register_number,register_expires_on) VALUES ($1::uuid,'Smoke Shop',$2,'fashion','16','Alger','Centre','0551234567','SMOKE','2099-01-01')`, merchantID, "smoke-"+merchantID); err != nil {
		t.Fatal(err)
	}
	service := New(db, storage.NewMemoryStore(), "http://media.test/dinar-public")
	item, err := service.Create(ctx, merchantID)
	if err != nil {
		t.Fatal(err)
	}
	input := validInput(time.Now().UTC())
	item, fields, err := service.Update(ctx, merchantID, item.ID, input)
	if err != nil || len(fields) > 0 {
		t.Fatalf("update failed: %v %v", err, fields)
	}
	img := image.NewNRGBA(image.Rect(0, 0, 400, 400))
	for y := 0; y < 400; y++ {
		for x := 0; x < 400; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 50, G: 120, B: 80, A: 255})
		}
	}
	var encoded bytes.Buffer
	if err = jpeg.Encode(&encoded, img, nil); err != nil {
		t.Fatal(err)
	}
	if _, err = service.AddMedia(ctx, merchantID, item.ID, "Smoke image", encoded.Bytes()); err != nil {
		t.Fatal(err)
	}
	item, fields, err = service.Submit(ctx, merchantID, item.ID)
	if err != nil || len(fields) > 0 || item.Status != StatusPending {
		t.Fatalf("submit failed: %v %v %+v", err, fields, item)
	}
	item, err = service.Review(ctx, adminID, item.ID, StatusApproved, "")
	if err != nil || item.Status != StatusApproved {
		t.Fatalf("review failed: %v %+v", err, item)
	}
	var publicCount int
	if err = db.QueryRow(`SELECT count(*) FROM public_offers WHERE id=$1::uuid`, item.ID).Scan(&publicCount); err != nil || publicCount != 1 {
		t.Fatalf("approved offer is not public: count=%d err=%v", publicCount, err)
	}
	item, fields, err = service.Update(ctx, merchantID, item.ID, input)
	if err != nil || len(fields) > 0 || item.Status != StatusPending || !item.RequiresReview {
		t.Fatalf("approved edit did not return to review: %v %v %+v", err, fields, item)
	}
}
