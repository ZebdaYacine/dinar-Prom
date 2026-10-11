package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *Service) applicationLocked(userID string) bool {
	status := StatusIncomplete
	if s.db != nil {
		if err := s.db.QueryRow(`SELECT status FROM merchant_verifications WHERE user_id = $1::uuid`, userID).Scan(&status); err != nil {
			return false
		}
	} else {
		s.mu.RLock()
		if merchant := s.merchants[userID]; merchant != nil {
			status = merchant.Status
		}
		s.mu.RUnlock()
	}
	return status == StatusPending || status == StatusApproved || status == StatusSuspended
}

// loginWithGoogleDB upserts the Google identity and creates its verification
// record. The picture URL is refreshed on every login so changes made in the
// Google account are reflected in the Dinar header on the next session.
func (s *Service) loginWithGoogleDB(providerID, email, firstName, lastName, name, picture string, verified bool) (User, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()

	var user User
	var currentSubject sql.NullString
	err = tx.QueryRow(`SELECT id::text, email, name, first_name, last_name, picture_url, role, email_verified, google_subject FROM users WHERE google_subject = $1 OR lower(email) = lower($2) ORDER BY (google_subject = $1) DESC LIMIT 1`, providerID, email).Scan(
		&user.ID, &user.Email, &user.Name, &user.FirstName, &user.LastName, &user.Picture, &user.Role, &user.EmailVerified, &currentSubject,
	)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		user = User{ID: uuid.NewString(), Email: strings.ToLower(strings.TrimSpace(email)), Role: RoleMerchant}
		if _, err = tx.Exec(`INSERT INTO users (id, email, name, first_name, last_name, picture_url, email_verified, role, google_subject) VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9)`, user.ID, user.Email, name, firstName, lastName, picture, verified, RoleMerchant, providerID); err != nil {
			return User{}, err
		}
	case err != nil:
		return User{}, err
	default:
		if currentSubject.Valid && currentSubject.String != "" && currentSubject.String != providerID {
			return User{}, errors.New("Google identity does not match this account")
		}
		if _, err = tx.Exec(`UPDATE users SET email = lower($2), name = $3, first_name = $4, last_name = $5, picture_url = $6, email_verified = $7, google_subject = $8, updated_at = now() WHERE id = $1::uuid`, user.ID, email, name, firstName, lastName, picture, verified, providerID); err != nil {
			return User{}, err
		}
	}
	user.Email = strings.ToLower(strings.TrimSpace(email))
	user.Name, user.FirstName, user.LastName, user.Picture, user.EmailVerified = name, firstName, lastName, picture, verified
	if user.Role == "" {
		user.Role = RoleMerchant
	}
	if _, err = tx.Exec(`INSERT INTO merchant_verifications (user_id, status) VALUES ($1::uuid, $2) ON CONFLICT (user_id) DO NOTHING`, user.ID, StatusIncomplete); err != nil {
		return User{}, err
	}
	if err = tx.Commit(); err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *Service) currentDB(userID string) (Me, error) {
	var me Me
	err := s.db.QueryRow(`SELECT id::text, email, name, first_name, last_name, picture_url, role, email_verified FROM users WHERE id = $1::uuid`, userID).Scan(
		&me.User.ID, &me.User.Email, &me.User.Name, &me.User.FirstName, &me.User.LastName, &me.User.Picture, &me.User.Role, &me.User.EmailVerified,
	)
	if err != nil {
		return Me{}, err
	}
	me.Merchant.Status = StatusIncomplete
	_ = s.db.QueryRow(`SELECT status, consent_accepted_at, rejection_reason, rejection_document FROM merchant_verifications WHERE user_id = $1::uuid`, userID).Scan(
		&me.Merchant.Status, &me.Merchant.ConsentAcceptedAt, &me.Merchant.RejectionReason, &me.Merchant.RejectionDocument,
	)

	var profile Profile
	err = s.db.QueryRow(`SELECT phone_e164, wilaya_id, address FROM merchant_profiles WHERE user_id = $1::uuid`, userID).Scan(&profile.PhoneE164, &profile.WilayaID, &profile.Address)
	if err == nil {
		profile.FirstName, profile.LastName = me.User.FirstName, me.User.LastName
		me.Merchant.Profile = &profile
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Me{}, err
	}

	var shop Shop
	err = s.db.QueryRow(`SELECT name, slug, category_id, description, wilaya_id, city, address, phone, whatsapp, register_number, register_expires_on::text, slug_edits FROM merchant_shops WHERE user_id = $1::uuid`, userID).Scan(
		&shop.Name, &shop.Slug, &shop.CategoryID, &shop.Description, &shop.WilayaID, &shop.City, &shop.Address, &shop.Phone, &shop.WhatsApp, &shop.RegisterNumber, &shop.RegisterExpiresOn, &shop.SlugEdits,
	)
	if err == nil {
		me.Merchant.Shop = &shop
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Me{}, err
	}

	rows, err := s.db.Query(`SELECT id::text, document_type, content_type, sha256, byte_size, created_at, storage_key FROM merchant_documents WHERE user_id = $1::uuid ORDER BY created_at`, userID)
	if err != nil {
		return Me{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var document Document
		if err := rows.Scan(&document.ID, &document.Type, &document.ContentType, &document.SHA256, &document.Size, &document.CreatedAt, &document.ObjectKey); err != nil {
			return Me{}, err
		}
		me.Merchant.Documents = append(me.Merchant.Documents, document)
	}
	if err := rows.Err(); err != nil {
		return Me{}, err
	}
	return me, nil
}

func (s *Service) submitDB(userID string, accepted bool) (Merchant, map[string]string) {
	fields := map[string]string{}
	tx, err := s.db.Begin()
	if err != nil {
		return Merchant{}, map[string]string{"database": "could not start submission"}
	}
	defer tx.Rollback()
	var status string
	if err := tx.QueryRow(`SELECT status FROM merchant_verifications WHERE user_id = $1::uuid FOR UPDATE`, userID).Scan(&status); err == nil && (status == StatusPending || status == StatusApproved || status == StatusSuspended) {
		me, currentErr := s.currentDB(userID)
		if currentErr != nil {
			return Merchant{}, map[string]string{"application": "submitted application data is locked"}
		}
		return me.Merchant, map[string]string{"application": "submitted application data is locked"}
	}
	var hasProfile, hasShop bool
	var documentCount int
	if err := tx.QueryRow(`SELECT EXISTS (SELECT 1 FROM merchant_profiles WHERE user_id = $1::uuid), EXISTS (SELECT 1 FROM merchant_shops WHERE user_id = $1::uuid)`, userID).Scan(&hasProfile, &hasShop); err != nil {
		fields["database"] = "could not load onboarding data"
	}
	if err := tx.QueryRow(`SELECT count(DISTINCT document_type) FROM merchant_documents WHERE user_id = $1::uuid AND document_type IN ('id_front', 'id_back', 'commercial_register')`, userID).Scan(&documentCount); err != nil {
		fields["database"] = "could not load verification documents"
	}
	if !hasProfile {
		fields["profile"] = "personal information is incomplete"
	}
	if !hasShop {
		fields["shop"] = "shop information is incomplete"
	}
	if documentCount < 3 {
		fields["documents"] = "all three verification documents are required"
	}
	if !accepted {
		fields["consent_accepted"] = "privacy consent is required"
	}
	if len(fields) > 0 {
		me, err := s.currentDB(userID)
		if err != nil {
			return Merchant{}, fields
		}
		return me.Merchant, fields
	}

	now := time.Now().UTC()
	_, err = tx.Exec(`INSERT INTO merchant_verifications (user_id, status, consent_accepted_at, updated_at) VALUES ($1::uuid, $2, $3, now()) ON CONFLICT (user_id) DO UPDATE SET status = EXCLUDED.status, consent_accepted_at = EXCLUDED.consent_accepted_at, rejection_reason = '', rejection_document = '', updated_at = now()`, userID, StatusPending, now)
	if err != nil {
		return Merchant{}, map[string]string{"submission": "could not save the application"}
	}
	var email, name string
	if err = tx.QueryRow(`SELECT email, name FROM users WHERE id = $1::uuid`, userID).Scan(&email, &name); err != nil {
		return Merchant{}, map[string]string{"submission": "could not queue the status email"}
	}
	payload, _ := json.Marshal(map[string]string{"status": StatusPending, "name": name})
	if _, err = tx.Exec(`INSERT INTO notification_outbox (id, kind, recipient, payload) VALUES ($1::uuid, 'merchant_verification_status', $2, $3::jsonb)`, uuid.NewString(), email, string(payload)); err != nil {
		return Merchant{}, map[string]string{"submission": "could not queue the status email"}
	}
	if err = tx.Commit(); err != nil {
		return Merchant{}, map[string]string{"submission": "could not save the application"}
	}
	me, err := s.currentDB(userID)
	if err != nil {
		return Merchant{}, map[string]string{"submission": "application was saved but status could not be loaded"}
	}
	s.mu.Lock()
	if merchant := s.merchants[userID]; merchant != nil {
		merchant.Status = StatusPending
		merchant.ConsentAcceptedAt = &now
	}
	s.mu.Unlock()
	return me.Merchant, nil
}
