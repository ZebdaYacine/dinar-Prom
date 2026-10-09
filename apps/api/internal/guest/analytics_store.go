package guest

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

const (
	reportAccepted = iota
	reportDuplicate
	reportRateLimited
)

type analyticsStore struct {
	db *sql.DB
}

func newAnalyticsStore(db *sql.DB) *analyticsStore {
	return &analyticsStore{db: db}
}

func (s *analyticsStore) recordEvent(ctx context.Context, eventType, offerID, visitorHash string, now time.Time) (bool, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO guest_offer_events (offer_id, event_type, visitor_hash, bucket_start, created_at)
		VALUES ($1, $2, $3, date_trunc('hour', $4::timestamptz), $4)
		ON CONFLICT (offer_id, event_type, visitor_hash, bucket_start) DO NOTHING
		RETURNING id`, offerID, eventType, visitorHash, now).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (s *analyticsStore) recordReport(ctx context.Context, offerID, visitorHash, reason, message string, now time.Time) (int, error) {
	var duplicate bool
	if err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM guest_offer_reports
			WHERE offer_id = $1 AND visitor_hash = $2 AND created_at >= $3::timestamptz - interval '24 hours'
		)`, offerID, visitorHash, now).Scan(&duplicate); err != nil {
		return reportAccepted, err
	}
	if duplicate {
		return reportDuplicate, nil
	}
	var recent int
	if err := s.db.QueryRowContext(ctx, `
		SELECT count(*) FROM guest_offer_reports
		WHERE visitor_hash = $1 AND created_at >= $2::timestamptz - interval '1 hour'`, visitorHash, now).Scan(&recent); err != nil {
		return reportAccepted, err
	}
	if recent >= 5 {
		return reportRateLimited, nil
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO guest_offer_reports (offer_id, visitor_hash, reason, message, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (offer_id, visitor_hash) DO NOTHING`, offerID, visitorHash, reason, message, now)
	if err != nil {
		return reportAccepted, err
	}
	return reportAccepted, nil
}
