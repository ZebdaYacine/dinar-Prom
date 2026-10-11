package notifications

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/dinar-dz/dinar/internal/config"
)

type Message struct {
	ID, Recipient, Kind string
	Payload             []byte
	Attempts            int
}

type Mailer interface {
	Send(context.Context, string, string, string) error
}

type SMTPMailer struct{ cfg config.Config }

func NewSMTPMailer(cfg config.Config) *SMTPMailer { return &SMTPMailer{cfg: cfg} }

func (m *SMTPMailer) Send(_ context.Context, recipient, subject, body string) error {
	if m.cfg.SMTPHost == "" {
		return errors.New("SMTP_HOST is not configured")
	}
	to, err := mail.ParseAddress(recipient)
	if err != nil {
		return err
	}
	from, err := mail.ParseAddress(m.cfg.MailFrom)
	if err != nil {
		return err
	}
	if strings.ContainsAny(to.Address+from.Address, "\r\n") {
		return errors.New("invalid email address")
	}
	headers := []string{
		"From: " + m.cfg.MailFrom,
		"To: " + to.Address,
		"Subject: " + mime.QEncoding.Encode("UTF-8", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
	}
	var auth smtp.Auth
	if m.cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", m.cfg.SMTPUser, m.cfg.SMTPPass, m.cfg.SMTPHost)
	}
	address := net.JoinHostPort(m.cfg.SMTPHost, strconv.Itoa(m.cfg.SMTPPort))
	return smtp.SendMail(address, auth, from.Address, []string{to.Address}, []byte(strings.Join(headers, "\r\n")+"\r\n\r\n"+body))
}

func ProcessOne(ctx context.Context, db *sql.DB, mailer Mailer) (bool, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	var message Message
	err = tx.QueryRowContext(ctx, `
		SELECT id::text, recipient, kind, payload::text, attempts
		FROM notification_outbox
		WHERE available_at <= now()
		  AND (status IN ('pending', 'failed') OR (status = 'processing' AND updated_at < now() - interval '10 minutes'))
		ORDER BY created_at
		FOR UPDATE SKIP LOCKED
		LIMIT 1`).Scan(&message.ID, &message.Recipient, &message.Kind, &message.Payload, &message.Attempts)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE notification_outbox SET status = 'processing', attempts = attempts + 1, updated_at = now() WHERE id = $1::uuid`, message.ID); err != nil {
		return false, err
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	subject, body, err := render(message)
	if err == nil {
		err = mailer.Send(ctx, message.Recipient, subject, body)
	}
	if err != nil {
		delayMinutes := 1 << min(message.Attempts, 6)
		_, updateErr := db.ExecContext(ctx, `UPDATE notification_outbox SET status = 'failed', last_error = $2, available_at = now() + ($3 * interval '1 minute'), updated_at = now() WHERE id = $1::uuid`, message.ID, truncate(err.Error(), 1000), delayMinutes)
		if updateErr != nil {
			return true, updateErr
		}
		return true, err
	}
	_, err = db.ExecContext(ctx, `UPDATE notification_outbox SET status = 'sent', sent_at = now(), last_error = '', updated_at = now() WHERE id = $1::uuid`, message.ID)
	return true, err
}

func render(message Message) (string, string, error) {
	var payload struct {
		Status       string `json:"status"`
		Name         string `json:"name"`
		Reason       string `json:"reason"`
		DocumentType string `json:"document_type"`
		Title        string `json:"title"`
	}
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return "", "", err
	}
	if message.Kind == "offer_status" {
		subjects := map[string]string{"pending": "Dinar DZ — offre en révision / العرض قيد المراجعة", "approved": "Dinar DZ — offre approuvée / تمت الموافقة على العرض", "rejected": "Dinar DZ — offre à corriger / يجب تصحيح العرض", "removed": "Dinar DZ — offre retirée / تم سحب العرض"}
		subject, ok := subjects[payload.Status]
		if !ok {
			return "", "", errors.New("unsupported offer status")
		}
		body := fmt.Sprintf("Bonjour %s,\n\nLe statut de votre offre « %s » est : %s.\n\nمرحباً %s،\nحالة عرضك « %s » هي: %s.\n", fallbackName(payload.Name), payload.Title, payload.Status, fallbackName(payload.Name), payload.Title, payload.Status)
		if payload.Reason != "" {
			body += "\nMotif / السبب: " + payload.Reason + "\n"
		}
		return subject, body, nil
	}
	if message.Kind != "merchant_verification_status" {
		return "", "", errors.New("unsupported notification kind")
	}
	subjects := map[string]string{
		"pending":   "Dinar DZ — demande reçue / تم استلام الطلب",
		"approved":  "Dinar DZ — compte vérifié / تم توثيق الحساب",
		"rejected":  "Dinar DZ — mise à jour requise / يجب تحديث الطلب",
		"suspended": "Dinar DZ — compte suspendu / تم تعليق الحساب",
	}
	subject, ok := subjects[payload.Status]
	if !ok {
		return "", "", errors.New("unsupported verification status")
	}
	name := fallbackName(payload.Name)
	body := fmt.Sprintf("Bonjour %s,\n\nLe statut de votre demande Dinar DZ est : %s.\n\nمرحباً %s،\nحالة طلبك على Dinar DZ هي: %s.\n", name, payload.Status, name, payload.Status)
	if payload.Reason != "" {
		body += "\nMotif / السبب: " + payload.Reason + "\n"
	}
	if payload.DocumentType != "" {
		body += "Document / الوثيقة: " + payload.DocumentType + "\n"
	}
	return subject, body, nil
}

func fallbackName(value string) string {
	if name := strings.TrimSpace(value); name != "" {
		return name
	}
	return "marchand / التاجر"
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
