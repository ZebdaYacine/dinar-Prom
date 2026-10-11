package notifications

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type recordingMailer struct{ recipient, subject, body string }

func (mailer *recordingMailer) Send(_ context.Context, recipient, subject, body string) error {
	mailer.recipient, mailer.subject, mailer.body = recipient, subject, body
	return nil
}

func TestRenderVerificationStatusEmail(t *testing.T) {
	subject, body, err := render(Message{Kind: "merchant_verification_status", Payload: []byte(`{"status":"rejected","name":"Amina","reason":"Unreadable","document_type":"id_front"}`)})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(subject, "Dinar DZ") || !strings.Contains(body, "Unreadable") || !strings.Contains(body, "id_front") {
		t.Fatalf("unexpected email: %q %q", subject, body)
	}
}

func TestNotificationOutboxDatabaseSmoke(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	id := uuid.NewString()
	_, err = db.Exec(`INSERT INTO notification_outbox (id, kind, recipient, payload, available_at, created_at) VALUES ($1::uuid, 'merchant_verification_status', 'merchant@example.com', '{"status":"approved","name":"Merchant"}'::jsonb, $2, '2000-01-01T00:00:00Z')`, id, time.Now().Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Exec(`DELETE FROM notification_outbox WHERE id = $1::uuid`, id)
	mailer := &recordingMailer{}
	processed, err := ProcessOne(context.Background(), db, mailer)
	if err != nil || !processed {
		t.Fatalf("expected outbox item to be delivered: processed=%v err=%v", processed, err)
	}
	var status string
	var sentAt sql.NullTime
	if err := db.QueryRow(`SELECT status, sent_at FROM notification_outbox WHERE id = $1::uuid`, id).Scan(&status, &sentAt); err != nil {
		t.Fatal(err)
	}
	if status != "sent" || !sentAt.Valid || mailer.recipient != "merchant@example.com" {
		t.Fatalf("unexpected delivery state: status=%s sent=%v recipient=%s", status, sentAt.Valid, mailer.recipient)
	}
}
