package offers

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
	"time"
)

func validInput(now time.Time) Input {
	before, after := 1000, 700
	start, end := now.Add(-time.Hour), now.Add(7*24*time.Hour)
	return Input{TitleAr: "تخفيض تجريبي", TitleFr: "Offre de test", DescriptionAr: "وصف", DescriptionFr: "Description", CategorySlug: "fashion", SubcategorySlug: "women-clothing", PriceBeforeDZD: &before, PriceAfterDZD: &after, StartsAt: &start, EndsAt: &end, WilayaCode: "16", City: "Alger", Address: "Centre", WhatsApp: "0551234567"}
}

func TestOfferValidationAndDiscount(t *testing.T) {
	now := time.Now().UTC()
	input := validInput(now)
	if fields := validate(input, true, now); len(fields) != 0 {
		t.Fatalf("expected valid offer, got %v", fields)
	}
	if got := discount(*input.PriceBeforeDZD, *input.PriceAfterDZD); got != 30 {
		t.Fatalf("expected 30 percent, got %d", got)
	}
	longEnd := input.StartsAt.Add(181 * 24 * time.Hour)
	input.EndsAt = &longEnd
	if fields := validate(input, true, now); fields["dates"] == "" {
		t.Fatal("expected 180-day duration validation")
	}
}

func TestDraftAllowsMissingFieldsButRejectsInvalidPrice(t *testing.T) {
	if fields := validate(Input{}, false, time.Now()); len(fields) != 0 {
		t.Fatalf("empty autosave draft should be valid: %v", fields)
	}
	before, after := 100, 120
	if fields := validate(Input{PriceBeforeDZD: &before, PriceAfterDZD: &after}, false, time.Now()); fields["price"] == "" {
		t.Fatal("expected invalid price")
	}
}

func TestImageProcessingValidatesDimensionsAndProducesWebP(t *testing.T) {
	small := image.NewNRGBA(image.Rect(0, 0, 399, 400))
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, small, nil); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := processImage(encoded.Bytes()); err == nil {
		t.Fatal("expected small image rejection")
	}
	large := image.NewNRGBA(image.Rect(0, 0, 600, 500))
	for y := 0; y < 500; y++ {
		for x := 0; x < 600; x++ {
			large.Set(x, y, color.NRGBA{R: uint8(x % 255), G: uint8(y % 255), B: 100, A: 255})
		}
	}
	encoded.Reset()
	if err := jpeg.Encode(&encoded, large, nil); err != nil {
		t.Fatal(err)
	}
	variants, width, height, err := processImage(encoded.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if width != 600 || height != 500 || len(variants) != 3 {
		t.Fatalf("unexpected variants: %dx%d %d", width, height, len(variants))
	}
	for _, variant := range variants {
		if len(variant.Data) < 12 || string(variant.Data[:4]) != "RIFF" || string(variant.Data[8:12]) != "WEBP" {
			t.Fatalf("variant %d is not WebP", variant.Size)
		}
	}
}
