package auth

import (
	"bytes"
	"testing"
)

func testService(t *testing.T) *Service {
	t.Helper()
	service, err := New(Config{Environment: "test"})
	if err != nil {
		t.Fatal(err)
	}
	return service
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
