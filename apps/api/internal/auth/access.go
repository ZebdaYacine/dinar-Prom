package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"time"

	casbin "github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/google/uuid"
)

var ErrForbidden = errors.New("forbidden")

type SignedDocument struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

const authorizationModel = `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`

// Authorize applies Casbin policies loaded from PostgreSQL. The in-memory
// rules mirror the migration solely for isolated unit tests.
func (s *Service) Authorize(userID, resource, action string) bool {
	if s.db != nil {
		var role string
		if err := s.db.QueryRow(`SELECT role FROM users WHERE id = $1::uuid`, userID).Scan(&role); err != nil {
			return false
		}
		rows, err := s.db.Query(`SELECT role, resource, action FROM auth_policies ORDER BY role, resource, action`)
		if err != nil {
			return false
		}
		defer rows.Close()
		var policies [][]string
		for rows.Next() {
			var policyRole, policyResource, policyAction string
			if err := rows.Scan(&policyRole, &policyResource, &policyAction); err != nil {
				return false
			}
			policies = append(policies, []string{policyRole, policyResource, policyAction})
		}
		return rows.Err() == nil && enforcePolicies(role, resource, action, policies)
	}
	s.mu.RLock()
	user, ok := s.users[userID]
	s.mu.RUnlock()
	if !ok {
		return false
	}
	policies := [][]string{{RoleMerchant, "merchant_identity", "write"}, {"admin", "merchant_identity", "write"}, {"admin", "merchant_documents", "read"}}
	return enforcePolicies(user.Role, resource, action, policies)
}

func enforcePolicies(subject, resource, action string, policies [][]string) bool {
	m, err := model.NewModelFromString(authorizationModel)
	if err != nil {
		return false
	}
	enforcer, err := casbin.NewEnforcer(m)
	if err != nil {
		return false
	}
	for _, policy := range policies {
		if len(policy) != 3 {
			return false
		}
		if _, err := enforcer.AddPolicy(policy[0], policy[1], policy[2]); err != nil {
			return false
		}
	}
	allowed, err := enforcer.Enforce(subject, resource, action)
	return err == nil && allowed
}

func (s *Service) SignedDocumentURL(ctx context.Context, actorID, documentID, requestID, ip string) (SignedDocument, error) {
	if !s.Authorize(actorID, "merchant_documents", "read") {
		return SignedDocument{}, ErrForbidden
	}
	if s.db == nil {
		return SignedDocument{}, errors.New("signed document access requires persistent storage")
	}
	var objectKey string
	if err := s.db.QueryRowContext(ctx, `SELECT storage_key FROM merchant_documents WHERE id = $1::uuid`, documentID).Scan(&objectKey); err != nil {
		return SignedDocument{}, err
	}
	ttl := 5 * time.Minute
	signedURL, err := s.store.PresignGet(ctx, objectKey, ttl)
	if err != nil {
		return SignedDocument{}, err
	}
	var ipValue any
	if parsed := net.ParseIP(ip); parsed != nil {
		ipValue = parsed.String()
	}
	metadata, _ := json.Marshal(map[string]string{"storage_key": objectKey})
	if _, err := s.db.ExecContext(ctx, `INSERT INTO audit_log (id, actor_user_id, action, resource_type, resource_id, request_id, ip_address, metadata) VALUES ($1::uuid, $2::uuid, 'view', 'merchant_document', $3, $4, $5::inet, $6::jsonb)`, uuid.NewString(), actorID, documentID, requestID, ipValue, string(metadata)); err != nil {
		return SignedDocument{}, err
	}
	return SignedDocument{URL: signedURL, ExpiresAt: time.Now().UTC().Add(ttl)}, nil
}

// UpdateVerificationStatus is the transaction boundary used by the review
// workflow. Persisting a decision and enqueueing its email either both happen
// or neither happens.
func (s *Service) UpdateVerificationStatus(ctx context.Context, userID, status, reason, documentType string) error {
	valid := status == StatusPending || status == StatusApproved || status == StatusRejected || status == StatusSuspended
	if !valid {
		return errors.New("invalid verification status")
	}
	if status == StatusRejected && reason == "" {
		return errors.New("rejection reason is required")
	}
	if documentType != "" {
		if _, ok := allowedDocumentTypes[documentType]; !ok {
			return errors.New("invalid rejection document")
		}
	}
	if s.db == nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		merchant := s.merchants[userID]
		if merchant == nil {
			return sql.ErrNoRows
		}
		merchant.Status, merchant.RejectionReason, merchant.RejectionDocument = status, reason, documentType
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE merchant_verifications SET status = $2, rejection_reason = $3, rejection_document = $4, updated_at = now() WHERE user_id = $1::uuid`, userID, status, reason, documentType)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return sql.ErrNoRows
	}
	var email, name string
	if err := tx.QueryRowContext(ctx, `SELECT email, name FROM users WHERE id = $1::uuid`, userID).Scan(&email, &name); err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string]string{"status": status, "name": name, "reason": reason, "document_type": documentType})
	if _, err := tx.ExecContext(ctx, `INSERT INTO notification_outbox (id, kind, recipient, payload) VALUES ($1::uuid, 'merchant_verification_status', $2, $3::jsonb)`, uuid.NewString(), email, string(payload)); err != nil {
		return err
	}
	return tx.Commit()
}
