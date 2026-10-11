package offers

import (
	"errors"
	"strings"
	"time"
)

const (
	StatusDraft    = "draft"
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
	StatusPaused   = "paused"
	StatusArchived = "archived"
	StatusRemoved  = "removed"
)

var (
	ErrNotFound         = errors.New("offer not found")
	ErrForbidden        = errors.New("offer does not belong to this merchant")
	ErrMerchantApproval = errors.New("merchant verification must be approved")
	ErrLocked           = errors.New("offer is locked in its current state")
	ErrActiveLimit      = errors.New("active offer limit reached")
)

type Input struct {
	TitleAr         string     `json:"title_ar"`
	TitleFr         string     `json:"title_fr"`
	DescriptionAr   string     `json:"description_ar"`
	DescriptionFr   string     `json:"description_fr"`
	CategorySlug    string     `json:"category_slug"`
	SubcategorySlug string     `json:"subcategory_slug"`
	BrandSlug       string     `json:"brand_slug"`
	PriceBeforeDZD  *int       `json:"price_before_dzd"`
	PriceAfterDZD   *int       `json:"price_after_dzd"`
	StartsAt        *time.Time `json:"starts_at"`
	EndsAt          *time.Time `json:"ends_at"`
	WilayaCode      string     `json:"wilaya_code"`
	City            string     `json:"city"`
	Address         string     `json:"address"`
	WhatsApp        string     `json:"whatsapp"`
}

type Media struct {
	ID           string    `json:"id"`
	URL400       string    `json:"url_400"`
	URL800       string    `json:"url_800"`
	URL1600      string    `json:"url_1600"`
	SourceWidth  int       `json:"source_width"`
	SourceHeight int       `json:"source_height"`
	Position     int       `json:"position"`
	AltText      string    `json:"alt_text"`
	CreatedAt    time.Time `json:"created_at"`
	keys         [3]string
}

type Offer struct {
	ID         string `json:"id"`
	MerchantID string `json:"-"`
	Input
	DiscountPercent int        `json:"discount_percent"`
	Status          string     `json:"status"`
	RejectionReason string     `json:"rejection_reason,omitempty"`
	SubmittedAt     *time.Time `json:"submitted_at,omitempty"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Media           []Media    `json:"media"`
	RequiresReview  bool       `json:"requires_review,omitempty"`
}

type DayStat struct {
	Day            string `json:"date"`
	Views          int    `json:"views"`
	WhatsAppClicks int    `json:"whatsapp_clicks"`
}
type OfferStat struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Status         string `json:"status"`
	Views          int    `json:"views"`
	WhatsAppClicks int    `json:"whatsapp_clicks"`
}
type Stats struct {
	Totals struct {
		Offers         int `json:"offers"`
		Views          int `json:"views"`
		WhatsAppClicks int `json:"whatsapp_clicks"`
	} `json:"totals"`
	Series []DayStat   `json:"series"`
	Offers []OfferStat `json:"offers"`
}

func normalizeInput(in Input) Input {
	in.TitleAr = strings.TrimSpace(in.TitleAr)
	in.TitleFr = strings.TrimSpace(in.TitleFr)
	in.DescriptionAr = strings.TrimSpace(in.DescriptionAr)
	in.DescriptionFr = strings.TrimSpace(in.DescriptionFr)
	in.CategorySlug = strings.TrimSpace(in.CategorySlug)
	in.SubcategorySlug = strings.TrimSpace(in.SubcategorySlug)
	in.BrandSlug = strings.TrimSpace(in.BrandSlug)
	in.WilayaCode = strings.TrimSpace(in.WilayaCode)
	in.City = strings.TrimSpace(in.City)
	in.Address = strings.TrimSpace(in.Address)
	in.WhatsApp = strings.TrimSpace(in.WhatsApp)
	return in
}

func discount(before, after int) int { return (before - after) * 100 / before }

func validate(in Input, complete bool, now time.Time) map[string]string {
	fields := map[string]string{}
	checkText := func(name, value string, minimum, maximum int) {
		length := len([]rune(value))
		if complete && length < minimum {
			fields[name] = "is required"
		} else if length > maximum {
			fields[name] = "is too long"
		}
	}
	checkText("title_ar", in.TitleAr, 5, 120)
	checkText("title_fr", in.TitleFr, 5, 120)
	checkText("description_ar", in.DescriptionAr, 1, 2000)
	checkText("description_fr", in.DescriptionFr, 1, 2000)
	for name, value := range map[string]string{"category_slug": in.CategorySlug, "subcategory_slug": in.SubcategorySlug, "wilaya_code": in.WilayaCode, "city": in.City, "address": in.Address, "whatsapp": in.WhatsApp} {
		if complete && value == "" {
			fields[name] = "is required"
		}
	}
	if (in.PriceBeforeDZD == nil) != (in.PriceAfterDZD == nil) {
		fields["price"] = "both prices are required"
	} else if in.PriceBeforeDZD != nil {
		if *in.PriceBeforeDZD <= 0 || *in.PriceAfterDZD <= 0 || *in.PriceAfterDZD >= *in.PriceBeforeDZD {
			fields["price"] = "discounted price must be lower than the original price"
		} else if value := discount(*in.PriceBeforeDZD, *in.PriceAfterDZD); value < 1 || value > 99 {
			fields["price"] = "discount must be between 1 and 99 percent"
		}
	} else if complete {
		fields["price"] = "both prices are required"
	}
	if (in.StartsAt == nil) != (in.EndsAt == nil) {
		fields["dates"] = "both dates are required"
	} else if in.StartsAt != nil {
		if !in.EndsAt.After(*in.StartsAt) {
			fields["dates"] = "end date must be after start date"
		} else if in.EndsAt.Sub(*in.StartsAt) > 180*24*time.Hour {
			fields["dates"] = "offer duration cannot exceed 180 days"
		} else if complete && !in.EndsAt.After(now) {
			fields["dates"] = "end date must be in the future"
		}
	} else if complete {
		fields["dates"] = "both dates are required"
	}
	return fields
}
