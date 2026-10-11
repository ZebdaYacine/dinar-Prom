package offers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dinar-dz/dinar/internal/auth"
	"github.com/dinar-dz/dinar/internal/storage"
	"github.com/google/uuid"
)

type Service struct {
	db        *sql.DB
	store     storage.ObjectStore
	mediaBase string
	now       func() time.Time
}

func New(db *sql.DB, store storage.ObjectStore, mediaBase string) *Service {
	return &Service{db: db, store: store, mediaBase: strings.TrimRight(mediaBase, "/"), now: func() time.Time { return time.Now().UTC() }}
}

type scanner interface{ Scan(...any) error }

const offerColumns = `id::text, merchant_id::text, title_ar, title_fr, description_ar, description_fr, category_slug, subcategory_slug, COALESCE(brand_slug,''), price_before_dzd, price_after_dzd, starts_at, ends_at, wilaya_code, city, address, whatsapp, discount_percent, status, rejection_reason, submitted_at, approved_at, created_at, updated_at`

func scanOffer(row scanner) (Offer, error) {
	var item Offer
	err := row.Scan(&item.ID, &item.MerchantID, &item.TitleAr, &item.TitleFr, &item.DescriptionAr, &item.DescriptionFr, &item.CategorySlug, &item.SubcategorySlug, &item.BrandSlug, &item.PriceBeforeDZD, &item.PriceAfterDZD, &item.StartsAt, &item.EndsAt, &item.WilayaCode, &item.City, &item.Address, &item.WhatsApp, &item.DiscountPercent, &item.Status, &item.RejectionReason, &item.SubmittedAt, &item.ApprovedAt, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (s *Service) merchantApproved(ctx context.Context, userID string) bool {
	var status string
	return s.db.QueryRowContext(ctx, `SELECT status FROM merchant_verifications WHERE user_id = $1::uuid`, userID).Scan(&status) == nil && status == "approved"
}

func (s *Service) Create(ctx context.Context, userID string) (Offer, error) {
	if !s.merchantApproved(ctx, userID) {
		return Offer{}, ErrMerchantApproval
	}
	var in Input
	_ = s.db.QueryRowContext(ctx, `SELECT category_id, wilaya_id, city, address, whatsapp FROM merchant_shops WHERE user_id = $1::uuid`, userID).Scan(&in.CategorySlug, &in.WilayaCode, &in.City, &in.Address, &in.WhatsApp)
	row := s.db.QueryRowContext(ctx, `INSERT INTO offers (merchant_id, category_slug, wilaya_code, city, address, whatsapp) VALUES ($1::uuid,$2,$3,$4,$5,$6) RETURNING `+offerColumns, userID, in.CategorySlug, in.WilayaCode, in.City, in.Address, in.WhatsApp)
	item, err := scanOffer(row)
	if err != nil {
		return Offer{}, err
	}
	_, _ = s.db.ExecContext(ctx, `INSERT INTO offer_status_history (offer_id, actor_user_id, to_status) VALUES ($1::uuid,$2::uuid,'draft')`, item.ID, userID)
	item.Media = []Media{}
	return item, nil
}

func (s *Service) List(ctx context.Context, userID string) ([]Offer, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+offerColumns+` FROM offers WHERE merchant_id = $1::uuid ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Offer{}
	for rows.Next() {
		item, err := scanOffer(rows)
		if err != nil {
			return nil, err
		}
		item.Media, err = s.media(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) Get(ctx context.Context, userID, offerID string) (Offer, error) {
	item, err := scanOffer(s.db.QueryRowContext(ctx, `SELECT `+offerColumns+` FROM offers WHERE id = $1::uuid`, offerID))
	if errors.Is(err, sql.ErrNoRows) {
		return Offer{}, ErrNotFound
	}
	if err != nil {
		return Offer{}, err
	}
	if item.MerchantID != userID {
		return Offer{}, ErrForbidden
	}
	item.Media, err = s.media(ctx, item.ID)
	return item, err
}

func (s *Service) media(ctx context.Context, offerID string) ([]Media, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, object_key_400, object_key_800, object_key_1600, source_width, source_height, position, alt_text, created_at FROM offer_media WHERE offer_id = $1::uuid ORDER BY position`, offerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Media{}
	for rows.Next() {
		var item Media
		if err := rows.Scan(&item.ID, &item.keys[0], &item.keys[1], &item.keys[2], &item.SourceWidth, &item.SourceHeight, &item.Position, &item.AltText, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.URL400 = s.mediaBase + "/" + item.keys[0]
		item.URL800 = s.mediaBase + "/" + item.keys[1]
		item.URL1600 = s.mediaBase + "/" + item.keys[2]
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) Update(ctx context.Context, userID, offerID string, input Input) (Offer, map[string]string, error) {
	input = normalizeInput(input)
	fields := validate(input, false, s.now())
	if len(fields) > 0 {
		return Offer{}, fields, nil
	}
	if input.WhatsApp != "" {
		normalized, err := auth.NormalizePhone(input.WhatsApp)
		if err != nil {
			fields["whatsapp"] = "invalid Algerian mobile number"
			return Offer{}, fields, nil
		}
		input.WhatsApp = normalized
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Offer{}, nil, err
	}
	defer tx.Rollback()
	current, err := scanOffer(tx.QueryRowContext(ctx, `SELECT `+offerColumns+` FROM offers WHERE id = $1::uuid FOR UPDATE`, offerID))
	if errors.Is(err, sql.ErrNoRows) {
		return Offer{}, nil, ErrNotFound
	}
	if err != nil {
		return Offer{}, nil, err
	}
	if current.MerchantID != userID {
		return Offer{}, nil, ErrForbidden
	}
	if current.Status == StatusPending || current.Status == StatusArchived || current.Status == StatusRemoved {
		return Offer{}, nil, ErrLocked
	}
	nextStatus, reason := current.Status, current.RejectionReason
	if current.Status == StatusApproved || current.Status == StatusPaused {
		nextStatus, reason = StatusPending, ""
	}
	if current.Status == StatusRejected {
		nextStatus, reason = StatusDraft, ""
	}
	value := 0
	if input.PriceBeforeDZD != nil {
		value = discount(*input.PriceBeforeDZD, *input.PriceAfterDZD)
	}
	brand := any(nil)
	if input.BrandSlug != "" {
		brand = input.BrandSlug
	}
	row := tx.QueryRowContext(ctx, `UPDATE offers SET title_ar=$3,title_fr=$4,description_ar=$5,description_fr=$6,category_slug=$7,subcategory_slug=$8,brand_slug=$9,price_before_dzd=$10,price_after_dzd=$11,discount_percent=$12,starts_at=$13,ends_at=$14,wilaya_code=$15,city=$16,address=$17,whatsapp=$18,status=$19,rejection_reason=$20,submitted_at=CASE WHEN $19='pending' THEN now() ELSE submitted_at END,updated_at=now() WHERE id=$1::uuid AND merchant_id=$2::uuid RETURNING `+offerColumns, offerID, userID, input.TitleAr, input.TitleFr, input.DescriptionAr, input.DescriptionFr, input.CategorySlug, input.SubcategorySlug, brand, input.PriceBeforeDZD, input.PriceAfterDZD, value, input.StartsAt, input.EndsAt, input.WilayaCode, input.City, input.Address, input.WhatsApp, nextStatus, reason)
	item, err := scanOffer(row)
	if err != nil {
		return Offer{}, nil, err
	}
	if nextStatus != current.Status {
		if _, err = tx.ExecContext(ctx, `INSERT INTO offer_status_history (offer_id,actor_user_id,from_status,to_status,reason) VALUES ($1::uuid,$2::uuid,$3,$4,'material edit')`, offerID, userID, current.Status, nextStatus); err != nil {
			return Offer{}, nil, err
		}
		item.RequiresReview = nextStatus == StatusPending
	}
	if err = tx.Commit(); err != nil {
		return Offer{}, nil, err
	}
	item.Media, err = s.media(ctx, item.ID)
	return item, nil, err
}

func (s *Service) Submit(ctx context.Context, userID, offerID string) (Offer, map[string]string, error) {
	if !s.merchantApproved(ctx, userID) {
		return Offer{}, nil, ErrMerchantApproval
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Offer{}, nil, err
	}
	defer tx.Rollback()
	item, err := scanOffer(tx.QueryRowContext(ctx, `SELECT `+offerColumns+` FROM offers WHERE id=$1::uuid FOR UPDATE`, offerID))
	if errors.Is(err, sql.ErrNoRows) {
		return Offer{}, nil, ErrNotFound
	}
	if err != nil {
		return Offer{}, nil, err
	}
	if item.MerchantID != userID {
		return Offer{}, nil, ErrForbidden
	}
	if item.Status != StatusDraft && item.Status != StatusRejected {
		return Offer{}, nil, ErrLocked
	}
	fields := validate(item.Input, true, s.now())
	var mediaCount int
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM offer_media WHERE offer_id=$1::uuid`, offerID).Scan(&mediaCount); err != nil {
		return Offer{}, nil, err
	}
	if mediaCount < 1 {
		fields["media"] = "at least one image is required"
	}
	var taxonomyOK bool
	if item.CategorySlug != "" && item.SubcategorySlug != "" {
		_ = tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM catalog_subcategories WHERE slug=$1 AND category_slug=$2 AND is_active)`, item.SubcategorySlug, item.CategorySlug).Scan(&taxonomyOK)
	}
	if !taxonomyOK {
		fields["subcategory_slug"] = "category and subcategory do not match"
	}
	if len(fields) > 0 {
		return Offer{}, fields, nil
	}
	var limit, active int
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE((value #>> '{}')::integer,20) FROM app_settings WHERE key='max_active_offers'`).Scan(&limit); err != nil {
		limit = 20
	}
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM offers WHERE merchant_id=$1::uuid AND id<>$2::uuid AND status IN ('pending','approved','paused')`, userID, offerID).Scan(&active); err != nil {
		return Offer{}, nil, err
	}
	if active >= limit {
		return Offer{}, nil, ErrActiveLimit
	}
	previous := item.Status
	if _, err = tx.ExecContext(ctx, `UPDATE offers SET status='pending',rejection_reason='',submitted_at=now(),updated_at=now() WHERE id=$1::uuid`, offerID); err != nil {
		return Offer{}, nil, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO offer_status_history (offer_id,actor_user_id,from_status,to_status) VALUES ($1::uuid,$2::uuid,$3,'pending')`, offerID, userID, previous); err != nil {
		return Offer{}, nil, err
	}
	if err = enqueueOfferStatus(ctx, tx, userID, offerID, "pending", ""); err != nil {
		return Offer{}, nil, err
	}
	if err = tx.Commit(); err != nil {
		return Offer{}, nil, err
	}
	item, err = s.Get(ctx, userID, offerID)
	return item, nil, err
}

func enqueueOfferStatus(ctx context.Context, tx *sql.Tx, userID, offerID, status, reason string) error {
	var email, name, title string
	if err := tx.QueryRowContext(ctx, `SELECT u.email,u.name,o.title_fr FROM users u JOIN offers o ON o.merchant_id=u.id WHERE u.id=$1::uuid AND o.id=$2::uuid`, userID, offerID).Scan(&email, &name, &title); err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string]string{"status": status, "name": name, "title": title, "reason": reason, "offer_id": offerID})
	_, err := tx.ExecContext(ctx, `INSERT INTO notification_outbox (id,kind,recipient,payload) VALUES ($1::uuid,'offer_status',$2,$3::jsonb)`, uuid.NewString(), email, string(payload))
	return err
}

func (s *Service) transition(ctx context.Context, userID, offerID, action string) (Offer, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Offer{}, err
	}
	defer tx.Rollback()
	item, err := scanOffer(tx.QueryRowContext(ctx, `SELECT `+offerColumns+` FROM offers WHERE id=$1::uuid FOR UPDATE`, offerID))
	if errors.Is(err, sql.ErrNoRows) {
		return Offer{}, ErrNotFound
	}
	if err != nil {
		return Offer{}, err
	}
	if item.MerchantID != userID {
		return Offer{}, ErrForbidden
	}
	next := ""
	switch action {
	case "pause":
		if item.Status == StatusApproved {
			next = StatusPaused
		}
	case "resume":
		if item.Status == StatusPaused && item.EndsAt != nil && item.EndsAt.After(s.now()) {
			next = StatusApproved
		}
	case "archive":
		if item.Status != StatusRemoved {
			next = StatusArchived
		}
	}
	if next == "" {
		return Offer{}, ErrLocked
	}
	if _, err = tx.ExecContext(ctx, `UPDATE offers SET status=$2,updated_at=now() WHERE id=$1::uuid`, offerID, next); err != nil {
		return Offer{}, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO offer_status_history (offer_id,actor_user_id,from_status,to_status) VALUES ($1::uuid,$2::uuid,$3,$4)`, offerID, userID, item.Status, next); err != nil {
		return Offer{}, err
	}
	if err = tx.Commit(); err != nil {
		return Offer{}, err
	}
	return s.Get(ctx, userID, offerID)
}
func (s *Service) Pause(ctx context.Context, u, id string) (Offer, error) {
	return s.transition(ctx, u, id, "pause")
}
func (s *Service) Resume(ctx context.Context, u, id string) (Offer, error) {
	return s.transition(ctx, u, id, "resume")
}
func (s *Service) Archive(ctx context.Context, u, id string) (Offer, error) {
	return s.transition(ctx, u, id, "archive")
}

func (s *Service) Delete(ctx context.Context, userID, offerID string) error {
	item, err := s.Get(ctx, userID, offerID)
	if err != nil {
		return err
	}
	if item.Status != StatusDraft {
		return ErrLocked
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `DELETE FROM offers WHERE id=$1::uuid AND merchant_id=$2::uuid AND status='draft'`, offerID, userID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected != 1 {
		return ErrLocked
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	s.deleteMediaObjects(ctx, item.Media)
	return nil
}

func (s *Service) AddMedia(ctx context.Context, userID, offerID, alt string, data []byte) (Media, error) {
	variants, width, height, err := processImage(data)
	if err != nil {
		return Media{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Media{}, err
	}
	defer tx.Rollback()
	current, err := scanOffer(tx.QueryRowContext(ctx, `SELECT `+offerColumns+` FROM offers WHERE id=$1::uuid FOR UPDATE`, offerID))
	if errors.Is(err, sql.ErrNoRows) {
		return Media{}, ErrNotFound
	}
	if err != nil {
		return Media{}, err
	}
	if current.MerchantID != userID {
		return Media{}, ErrForbidden
	}
	if current.Status == StatusPending || current.Status == StatusArchived || current.Status == StatusRemoved {
		return Media{}, ErrLocked
	}
	var count, limit int
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM offer_media WHERE offer_id=$1::uuid`, offerID).Scan(&count); err != nil {
		return Media{}, err
	}
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE((value #>> '{}')::integer,10) FROM app_settings WHERE key='max_images_per_offer'`).Scan(&limit); err != nil {
		limit = 10
	}
	if count >= limit {
		return Media{}, errors.New("offer already has the maximum number of images")
	}
	mediaID := uuid.NewString()
	keys := [3]string{}
	uploaded := []string{}
	for index, variant := range variants {
		key := fmt.Sprintf("offers/%s/%s_%d.webp", offerID, mediaID, variant.Size)
		if err = s.store.Put(ctx, key, "image/webp", variant.Data); err != nil {
			for _, value := range uploaded {
				_ = s.store.Delete(ctx, value)
			}
			return Media{}, err
		}
		keys[index] = key
		uploaded = append(uploaded, key)
	}
	position := count + 1
	_, err = tx.ExecContext(ctx, `INSERT INTO offer_media (id,offer_id,object_key_400,object_key_800,object_key_1600,source_width,source_height,position,alt_text) VALUES ($1::uuid,$2::uuid,$3,$4,$5,$6,$7,$8,$9)`, mediaID, offerID, keys[0], keys[1], keys[2], width, height, position, strings.TrimSpace(alt))
	if err != nil {
		for _, value := range uploaded {
			_ = s.store.Delete(ctx, value)
		}
		return Media{}, err
	}
	if current.Status == StatusApproved || current.Status == StatusPaused {
		if _, err = tx.ExecContext(ctx, `UPDATE offers SET status='pending',submitted_at=now(),updated_at=now() WHERE id=$1::uuid`, offerID); err != nil {
			return Media{}, err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO offer_status_history (offer_id,actor_user_id,from_status,to_status,reason) VALUES ($1::uuid,$2::uuid,$3,'pending','media edit')`, offerID, userID, current.Status); err != nil {
			return Media{}, err
		}
	}
	if current.Status == StatusRejected {
		_, err = tx.ExecContext(ctx, `UPDATE offers SET status='draft',rejection_reason='',updated_at=now() WHERE id=$1::uuid`, offerID)
		if err != nil {
			return Media{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		for _, value := range uploaded {
			_ = s.store.Delete(ctx, value)
		}
		return Media{}, err
	}
	return Media{ID: mediaID, URL400: s.mediaBase + "/" + keys[0], URL800: s.mediaBase + "/" + keys[1], URL1600: s.mediaBase + "/" + keys[2], SourceWidth: width, SourceHeight: height, Position: position, AltText: strings.TrimSpace(alt), CreatedAt: s.now(), keys: keys}, nil
}

func (s *Service) DeleteMedia(ctx context.Context, userID, offerID, mediaID string) error {
	item, err := s.Get(ctx, userID, offerID)
	if err != nil {
		return err
	}
	if item.Status == StatusPending || item.Status == StatusArchived || item.Status == StatusRemoved {
		return ErrLocked
	}
	var target *Media
	for i := range item.Media {
		if item.Media[i].ID == mediaID {
			target = &item.Media[i]
			break
		}
	}
	if target == nil {
		return ErrNotFound
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM offer_media WHERE id=$1::uuid AND offer_id=$2::uuid`, mediaID, offerID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `WITH ranked AS (SELECT id,row_number() OVER (ORDER BY position,created_at)::smallint AS next FROM offer_media WHERE offer_id=$1::uuid) UPDATE offer_media m SET position=ranked.next FROM ranked WHERE m.id=ranked.id`, offerID); err != nil {
		return err
	}
	if item.Status == StatusApproved || item.Status == StatusPaused {
		if _, err = tx.ExecContext(ctx, `UPDATE offers SET status='pending',submitted_at=now(),updated_at=now() WHERE id=$1::uuid`, offerID); err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	for _, key := range target.keys {
		_ = s.store.Delete(ctx, key)
	}
	return nil
}

func (s *Service) ReorderMedia(ctx context.Context, userID, offerID string, ids []string) error {
	item, err := s.Get(ctx, userID, offerID)
	if err != nil {
		return err
	}
	if item.Status == StatusPending || item.Status == StatusArchived || item.Status == StatusRemoved {
		return ErrLocked
	}
	if len(ids) != len(item.Media) {
		return errors.New("media order must include every image")
	}
	known := map[string]bool{}
	for _, m := range item.Media {
		known[m.ID] = true
	}
	for _, id := range ids {
		if !known[id] {
			return errors.New("media order contains an unknown image")
		}
		delete(known, id)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SET CONSTRAINTS offer_media_position_unique DEFERRED`); err != nil {
		return err
	}
	for index, id := range ids {
		if _, err = tx.ExecContext(ctx, `UPDATE offer_media SET position=$3 WHERE id=$1::uuid AND offer_id=$2::uuid`, id, offerID, index+1); err != nil {
			return err
		}
	}
	if item.Status == StatusApproved || item.Status == StatusPaused {
		if _, err = tx.ExecContext(ctx, `UPDATE offers SET status='pending',submitted_at=now(),updated_at=now() WHERE id=$1::uuid`, offerID); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO offer_status_history (offer_id,actor_user_id,from_status,to_status,reason) VALUES ($1::uuid,$2::uuid,$3,'pending','media reorder')`, offerID, userID, item.Status); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Service) Duplicate(ctx context.Context, userID, offerID string) (Offer, error) {
	source, err := s.Get(ctx, userID, offerID)
	if err != nil {
		return Offer{}, err
	}
	if source.Status == StatusDraft || source.Status == StatusPending {
		return Offer{}, ErrLocked
	}
	copyOffer, err := scanOffer(s.db.QueryRowContext(ctx, `INSERT INTO offers (merchant_id,title_ar,title_fr,description_ar,description_fr,category_slug,subcategory_slug,brand_slug,price_before_dzd,price_after_dzd,discount_percent,wilaya_code,city,address,whatsapp,status) SELECT merchant_id,title_ar,title_fr,description_ar,description_fr,category_slug,subcategory_slug,brand_slug,price_before_dzd,price_after_dzd,discount_percent,wilaya_code,city,address,whatsapp,'draft' FROM offers WHERE id=$1::uuid AND merchant_id=$2::uuid RETURNING `+offerColumns, offerID, userID))
	if err != nil {
		return Offer{}, err
	}
	cleanup := []string{}
	for _, media := range source.Media {
		newID := uuid.NewString()
		keys := [3]string{}
		for index, oldKey := range media.keys {
			data, getErr := s.store.Get(ctx, oldKey)
			if getErr != nil {
				err = getErr
				break
			}
			key := fmt.Sprintf("offers/%s/%s_%d.webp", copyOffer.ID, newID, []int{400, 800, 1600}[index])
			if putErr := s.store.Put(ctx, key, "image/webp", data); putErr != nil {
				err = putErr
				break
			}
			keys[index] = key
			cleanup = append(cleanup, key)
		}
		if err != nil {
			break
		}
		_, err = s.db.ExecContext(ctx, `INSERT INTO offer_media (id,offer_id,object_key_400,object_key_800,object_key_1600,source_width,source_height,position,alt_text) VALUES ($1::uuid,$2::uuid,$3,$4,$5,$6,$7,$8,$9)`, newID, copyOffer.ID, keys[0], keys[1], keys[2], media.SourceWidth, media.SourceHeight, media.Position, media.AltText)
		if err != nil {
			break
		}
	}
	if err != nil {
		for _, key := range cleanup {
			_ = s.store.Delete(ctx, key)
		}
		_, _ = s.db.ExecContext(ctx, `DELETE FROM offers WHERE id=$1::uuid`, copyOffer.ID)
		return Offer{}, err
	}
	_, _ = s.db.ExecContext(ctx, `INSERT INTO offer_status_history (offer_id,actor_user_id,to_status,reason) VALUES ($1::uuid,$2::uuid,'draft','duplicated')`, copyOffer.ID, userID)
	return s.Get(ctx, userID, copyOffer.ID)
}

func (s *Service) deleteMediaObjects(ctx context.Context, media []Media) {
	for _, item := range media {
		for _, key := range item.keys {
			_ = s.store.Delete(ctx, key)
		}
	}
}

func (s *Service) Stats(ctx context.Context, userID string) (Stats, error) {
	var result Stats
	result.Series = make([]DayStat, 30)
	result.Offers = []OfferStat{}
	days := map[string]int{}
	today := s.now().Truncate(24 * time.Hour)
	for index := 0; index < 30; index++ {
		day := today.AddDate(0, 0, index-29).Format("2006-01-02")
		result.Series[index] = DayStat{Day: day}
		days[day] = index
	}
	rows, err := s.db.QueryContext(ctx, `SELECT day::text,views,whatsapp_clicks FROM merchant_offer_stats_daily WHERE merchant_id=$1::uuid AND day>=CURRENT_DATE-29 ORDER BY day`, userID)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var day string
		var views, clicks int
		if err = rows.Scan(&day, &views, &clicks); err != nil {
			rows.Close()
			return result, err
		}
		if index, ok := days[day]; ok {
			result.Series[index].Views = views
			result.Series[index].WhatsAppClicks = clicks
		}
		result.Totals.Views += views
		result.Totals.WhatsAppClicks += clicks
	}
	rows.Close()
	offerRows, err := s.db.QueryContext(ctx, `SELECT o.id::text,COALESCE(NULLIF(o.title_fr,''),o.title_ar),o.status,COALESCE(sum(s.views),0)::integer,COALESCE(sum(s.whatsapp_clicks),0)::integer FROM offers o LEFT JOIN merchant_offer_stats_daily s ON s.offer_id=o.id WHERE o.merchant_id=$1::uuid GROUP BY o.id ORDER BY o.updated_at DESC`, userID)
	if err != nil {
		return result, err
	}
	defer offerRows.Close()
	for offerRows.Next() {
		var item OfferStat
		if err = offerRows.Scan(&item.ID, &item.Title, &item.Status, &item.Views, &item.WhatsAppClicks); err != nil {
			return result, err
		}
		result.Offers = append(result.Offers, item)
	}
	result.Totals.Offers = len(result.Offers)
	return result, offerRows.Err()
}

func (s *Service) AdminQueue(ctx context.Context) ([]Offer, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+offerColumns+` FROM offers WHERE status='pending' ORDER BY submitted_at,created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Offer{}
	for rows.Next() {
		item, err := scanOffer(rows)
		if err != nil {
			return nil, err
		}
		item.Media, err = s.media(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) Review(ctx context.Context, actorID, offerID, status, reason string) (Offer, error) {
	if status != StatusApproved && status != StatusRejected && status != StatusRemoved {
		return Offer{}, errors.New("invalid review status")
	}
	if status == StatusRejected && strings.TrimSpace(reason) == "" {
		return Offer{}, errors.New("rejection reason is required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Offer{}, err
	}
	defer tx.Rollback()
	item, err := scanOffer(tx.QueryRowContext(ctx, `SELECT `+offerColumns+` FROM offers WHERE id=$1::uuid FOR UPDATE`, offerID))
	if errors.Is(err, sql.ErrNoRows) {
		return Offer{}, ErrNotFound
	}
	if err != nil {
		return Offer{}, err
	}
	if item.Status != StatusPending && status != StatusRemoved {
		return Offer{}, ErrLocked
	}
	approvedAt := any(nil)
	if status == StatusApproved {
		approvedAt = s.now()
	}
	if _, err = tx.ExecContext(ctx, `UPDATE offers SET status=$2,rejection_reason=$3,approved_at=COALESCE($4,approved_at),updated_at=now() WHERE id=$1::uuid`, offerID, status, strings.TrimSpace(reason), approvedAt); err != nil {
		return Offer{}, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO offer_status_history (offer_id,actor_user_id,from_status,to_status,reason) VALUES ($1::uuid,$2::uuid,$3,$4,$5)`, offerID, actorID, item.Status, status, strings.TrimSpace(reason)); err != nil {
		return Offer{}, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO audit_log (id,actor_user_id,action,resource_type,resource_id,metadata) VALUES ($1::uuid,$2::uuid,'review','offer',$3,$4::jsonb)`, uuid.NewString(), actorID, offerID, fmt.Sprintf(`{"from":%q,"to":%q,"reason":%q}`, item.Status, status, reason)); err != nil {
		return Offer{}, err
	}
	if err = enqueueOfferStatus(ctx, tx, item.MerchantID, offerID, status, reason); err != nil {
		return Offer{}, err
	}
	if err = tx.Commit(); err != nil {
		return Offer{}, err
	}
	reviewed, err := scanOffer(s.db.QueryRowContext(ctx, `SELECT `+offerColumns+` FROM offers WHERE id=$1::uuid`, offerID))
	if err != nil {
		return Offer{}, err
	}
	reviewed.Media, err = s.media(ctx, offerID)
	return reviewed, err
}
