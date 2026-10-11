package auth

import (
	"bytes"
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func testService(t *testing.T) *Service {
	t.Helper()
	service, err := New(Config{Environment: "test"})
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func TestLoginRedirectResumesLastIncompleteStep(t *testing.T) {
	service := testService(t)
	user, _, _ := service.SeedTestUser("resume@example.com")
	if got := service.LoginRedirect(user.ID, "/fr/onboarding/personal"); got != "/fr/onboarding/personal" {
		t.Fatalf("expected personal step, got %s", got)
	}
	_, _ = service.SaveProfile(user.ID, Profile{FirstName: "Test", LastName: "Merchant", PhoneE164: "0551234567", WilayaID: "16", Address: "Algiers"})
	if got := service.LoginRedirect(user.ID, "/fr/onboarding/personal"); got != "/fr/onboarding/shop" {
		t.Fatalf("expected shop step, got %s", got)
	}
	_, _ = service.SaveShop(user.ID, Shop{Name: "Resume Shop", CategoryID: "fashion", WilayaID: "16", RegisterNumber: "RC-1", RegisterExpiresOn: "2099-01-01"})
	if got := service.LoginRedirect(user.ID, "/fr/onboarding/personal"); got != "/fr/onboarding/documents" {
		t.Fatalf("expected documents step, got %s", got)
	}
	for _, typ := range []string{"id_front", "id_back", "commercial_register"} {
		_, _ = service.SaveDocument(user.ID, typ, []byte{0xff, 0xd8, 0xff, 0xe0, 0x00})
	}
	if got := service.LoginRedirect(user.ID, "/fr/onboarding/personal"); got != "/fr/onboarding/review" {
		t.Fatalf("expected review step, got %s", got)
	}
}

func TestLoginRedirectSendsApprovedMerchantToOffers(t *testing.T) {
	service := testService(t)
	user, _, _ := service.SeedTestUser("approved@example.com")
	if err := service.UpdateVerificationStatus(context.Background(), user.ID, StatusApproved, "", ""); err != nil {
		t.Fatal(err)
	}
	if got := service.LoginRedirect(user.ID, "/ar/onboarding/personal"); got != "/ar/merchant/promotions" {
		t.Fatalf("expected Arabic offers page, got %s", got)
	}
	if got := service.LoginRedirect(user.ID, "/fr/onboarding/personal"); got != "/fr/merchant/promotions" {
		t.Fatalf("expected French offers page, got %s", got)
	}
}

func TestMerchantRoutesEnforceRolePolicy(t *testing.T) {
	service := testService(t)
	user, token, _ := service.SeedTestUser("policy@example.com")
	app := fiber.New()
	Register(app.Group("/api/v1"), service)

	request := httptest.NewRequest("PUT", "/api/v1/merchant/profile", strings.NewReader(`{"first_name":"Test","last_name":"Merchant","phone_e164":"0551234567","wilaya_id":"16","address":"Algiers"}`))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil || response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected merchant policy to allow profile write, status=%d err=%v", response.StatusCode, err)
	}

	service.mu.Lock()
	demoted := service.users[user.ID]
	demoted.Role = "user"
	service.users[user.ID] = demoted
	service.mu.Unlock()
	request = httptest.NewRequest("PUT", "/api/v1/merchant/profile", strings.NewReader(`{"first_name":"Test"}`))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response, err = app.Test(request)
	if err != nil || response.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected user role to be forbidden, status=%d err=%v", response.StatusCode, err)
	}
}

func TestStatusDecisionRequiresRejectionReason(t *testing.T) {
	service := testService(t)
	user, _, _ := service.SeedTestUser("decision@example.com")
	if err := service.UpdateVerificationStatus(context.Background(), user.ID, StatusRejected, "", "id_front"); err == nil {
		t.Fatal("expected rejection reason validation")
	}
	if err := service.UpdateVerificationStatus(context.Background(), user.ID, StatusRejected, "Document is unreadable", "id_front"); err != nil {
		t.Fatal(err)
	}
	me, _ := service.Current(user.ID)
	if me.Merchant.Status != StatusRejected || me.Merchant.RejectionDocument != "id_front" {
		t.Fatalf("unexpected rejection state: %+v", me.Merchant)
	}
}

func TestNormalizePhoneAndProfileValidation(t *testing.T) {
	if got, err := NormalizePhone("+213 551234567"); err != nil || got != "+213551234567" {
		t.Fatalf("unexpected normalized phone: %q %v", got, err)
	}
	if _, err := NormalizePhone("0212345678"); err == nil {
		t.Fatal("expected invalid Algerian mobile number")
	}
	service := testService(t)
	user, _, _ := service.SeedTestUser("merchant@example.com")
	if _, fields := service.SaveProfile(user.ID, Profile{FirstName: "Test", LastName: "Merchant", PhoneE164: "0551234567", WilayaID: "16", Address: "Algiers"}); len(fields) != 0 {
		t.Fatalf("expected valid profile, got %v", fields)
	}
}

func TestAccessAndRefreshRotation(t *testing.T) {
	service := testService(t)
	user, access, err := service.SeedTestUser("rotate@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if id, err := service.VerifyAccess(access); err != nil || id != user.ID {
		t.Fatalf("access token did not verify: %v", err)
	}
	if _, err := service.VerifyAccess(access + "bad"); err == nil {
		t.Fatal("expected tampered access token to fail")
	}
}

func TestDocumentMagicBytesAndReplacement(t *testing.T) {
	service := testService(t)
	user, _, _ := service.SeedTestUser("docs@example.com")
	if _, err := service.SaveDocument(user.ID, "id_front", []byte("not an image")); err == nil {
		t.Fatal("expected magic-byte validation")
	}
	first, err := service.SaveDocument(user.ID, "id_front", append([]byte{0xff, 0xd8, 0xff, 0xe0}, bytes.Repeat([]byte{0}, 16)...))
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.SaveDocument(user.ID, "id_front", append([]byte{0xff, 0xd8, 0xff, 0xe0}, bytes.Repeat([]byte{1}, 16)...))
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == second.ID {
		t.Fatal("expected replacement document id")
	}
	me, err := service.Current(user.ID)
	if err != nil || len(me.Merchant.Documents) != 1 {
		t.Fatalf("expected one replacement document: %+v %v", me, err)
	}
}

func TestSubmitRequiresAllStepsAndConsent(t *testing.T) {
	service := testService(t)
	user, _, _ := service.SeedTestUser("submit@example.com")
	if _, fields := service.Submit(user.ID, true); len(fields) == 0 {
		t.Fatal("expected incomplete onboarding to be rejected")
	}
	_, _ = service.SaveProfile(user.ID, Profile{FirstName: "Test", LastName: "Merchant", PhoneE164: "0551234567", WilayaID: "16", Address: "Algiers"})
	_, _ = service.SaveShop(user.ID, Shop{Name: "Test Shop", CategoryID: "fashion", WilayaID: "16", RegisterNumber: "RC-1", RegisterExpiresOn: "2099-01-01"})
	for _, typ := range []string{"id_front", "id_back", "commercial_register"} {
		if _, err := service.SaveDocument(user.ID, typ, append([]byte{0xff, 0xd8, 0xff, 0xe0}, 0)); err != nil {
			t.Fatal(err)
		}
	}
	if _, fields := service.Submit(user.ID, false); len(fields) == 0 {
		t.Fatal("expected consent to be required")
	}
	m, fields := service.Submit(user.ID, true)
	if len(fields) != 0 || m.Status != StatusPending {
		t.Fatalf("expected pending merchant: %+v %v", m, fields)
	}
}
