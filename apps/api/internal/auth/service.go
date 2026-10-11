package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dinar-dz/dinar/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	RoleMerchant     = "merchant"
	StatusIncomplete = "incomplete"
	StatusPending    = "pending"
	StatusApproved   = "approved"
	StatusRejected   = "rejected"
	StatusSuspended  = "suspended"
)

var phonePattern = regexp.MustCompile(`^(05|06|07)[0-9]{8}$`)
var slugPattern = regexp.MustCompile(`[^a-z0-9]+`)
var allowedDocumentTypes = map[string]struct{}{"id_front": {}, "id_back": {}, "commercial_register": {}}

type Config struct {
	AppURL, Environment, GoogleClientID, GoogleClientSecret, GoogleRedirectURL string
	Database                                                                   *sql.DB
	Issuer, Audience, RefreshCookieName, CSRFCookieName, CSRFHeaderName        string
	JWTPrivateKeyFile, JWTPublicKeyFile                                        string
	AccessTTL, RefreshTTL                                                      time.Duration
	CookieSecure                                                               bool
	IPHashSalt                                                                 string
	PrivateStore                                                               storage.PrivateStore
}

type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Picture       string `json:"picture"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
}
type Profile struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	PhoneE164 string `json:"phone_e164"`
	WilayaID  string `json:"wilaya_id"`
	Address   string `json:"address"`
}
type Shop struct {
	Name              string `json:"name"`
	Slug              string `json:"slug"`
	CategoryID        string `json:"category_id"`
	Description       string `json:"description"`
	WilayaID          string `json:"wilaya_id"`
	City              string `json:"city"`
	Address           string `json:"address"`
	Phone             string `json:"phone"`
	WhatsApp          string `json:"whatsapp"`
	RegisterNumber    string `json:"register_number"`
	RegisterExpiresOn string `json:"register_expires_on"`
	SlugEdits         int    `json:"-"`
}
type Document struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	ContentType string    `json:"content_type"`
	SHA256      string    `json:"sha256"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	ObjectKey   string    `json:"-"`
	Bytes       []byte    `json:"-"`
}
type Merchant struct {
	Status            string     `json:"verification_status"`
	Profile           *Profile   `json:"profile,omitempty"`
	Shop              *Shop      `json:"shop,omitempty"`
	Documents         []Document `json:"documents"`
	ConsentAcceptedAt *time.Time `json:"consent_accepted_at,omitempty"`
	RejectionReason   string     `json:"rejection_reason,omitempty"`
	RejectionDocument string     `json:"rejection_document,omitempty"`
}
type Me struct {
	User     User     `json:"user"`
	Merchant Merchant `json:"merchant"`
}
type googleProfile struct {
	ID            string `json:"sub"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	FirstName     string `json:"given_name"`
	LastName      string `json:"family_name"`
	Picture       string `json:"picture"`
	EmailVerified bool   `json:"email_verified"`
}
type oauthState struct {
	Verifier, Next string
	ExpiresAt      time.Time
}
type session struct {
	UserID, FamilyID, TokenHash string
	ExpiresAt                   time.Time
	Revoked                     bool
}

type Service struct {
	cfg        Config
	db         *sql.DB
	mu         sync.RWMutex
	users      map[string]User
	merchants  map[string]*Merchant
	sessions   map[string]*session
	oauth      map[string]oauthState
	store      storage.PrivateStore
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func New(cfg Config) (*Service, error) {
	if cfg.Issuer == "" {
		cfg.Issuer = "dinar-api"
	}
	if cfg.Audience == "" {
		cfg.Audience = "dinar-web"
	}
	if cfg.RefreshCookieName == "" {
		cfg.RefreshCookieName = "dinar_refresh"
	}
	if cfg.CSRFCookieName == "" {
		cfg.CSRFCookieName = "dinar_csrf"
	}
	if cfg.CSRFHeaderName == "" {
		cfg.CSRFHeaderName = "X-CSRF-Token"
	}
	if cfg.AccessTTL == 0 {
		cfg.AccessTTL = 10 * time.Minute
	}
	if cfg.RefreshTTL == 0 {
		cfg.RefreshTTL = 30 * 24 * time.Hour
	}
	priv, pub, err := loadKeys(cfg)
	if err != nil {
		return nil, err
	}
	if cfg.PrivateStore == nil {
		cfg.PrivateStore = storage.NewMemoryStore()
	}
	return &Service{cfg: cfg, db: cfg.Database, users: map[string]User{}, merchants: map[string]*Merchant{}, sessions: map[string]*session{}, oauth: map[string]oauthState{}, store: cfg.PrivateStore, privateKey: priv, publicKey: pub}, nil
}

func loadKeys(cfg Config) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	if cfg.JWTPrivateKeyFile != "" {
		data, err := os.ReadFile(cfg.JWTPrivateKeyFile)
		if err != nil && cfg.Environment == "production" {
			return nil, nil, fmt.Errorf("read jwt private key: %w", err)
		}
		if err == nil {
			block, _ := pem.Decode(data)
			if block == nil {
				return nil, nil, errors.New("invalid jwt private key")
			}
			if key, e := x509.ParsePKCS8PrivateKey(block.Bytes); e == nil {
				if rsaKey, ok := key.(*rsa.PrivateKey); ok {
					return rsaKey, &rsaKey.PublicKey, nil
				}
			}
			if rsaKey, e := x509.ParsePKCS1PrivateKey(block.Bytes); e == nil {
				return rsaKey, &rsaKey.PublicKey, nil
			}
			return nil, nil, errors.New("invalid jwt private key")
		}
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return key, &key.PublicKey, nil
}
func randomString(size int) (string, error) {
	b := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func hash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (s *Service) BeginOAuth(next string) (string, string, error) {
	if !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") {
		next = "/ar/onboarding/personal"
	}
	state, err := randomString(32)
	if err != nil {
		return "", "", err
	}
	verifier, err := randomString(48)
	if err != nil {
		return "", "", err
	}
	challenge := sha256.Sum256([]byte(verifier))
	s.mu.Lock()
	s.oauth[state] = oauthState{Verifier: verifier, Next: next, ExpiresAt: time.Now().Add(10 * time.Minute)}
	s.mu.Unlock()
	if s.cfg.GoogleClientID == "" {
		return "", state, errors.New("google oauth is not configured")
	}
	values := url.Values{"client_id": {s.cfg.GoogleClientID}, "redirect_uri": {s.cfg.GoogleRedirectURL}, "response_type": {"code"}, "scope": {"openid email profile"}, "state": {state}, "code_challenge": {base64.RawURLEncoding.EncodeToString(challenge[:])}, "code_challenge_method": {"S256"}, "access_type": {"offline"}, "prompt": {"select_account"}}
	return "https://accounts.google.com/o/oauth2/v2/auth?" + values.Encode(), state, nil
}

func (s *Service) finishOAuth(state, code string) (User, string, error) {
	s.mu.Lock()
	flow, ok := s.oauth[state]
	if ok {
		delete(s.oauth, state)
	}
	s.mu.Unlock()
	if !ok || time.Now().After(flow.ExpiresAt) || code == "" {
		return User{}, "", errors.New("invalid oauth state")
	}
	if s.cfg.GoogleClientID == "" || s.cfg.GoogleClientSecret == "" {
		return User{}, "", errors.New("google oauth is not configured")
	}
	form := url.Values{"code": {code}, "client_id": {s.cfg.GoogleClientID}, "client_secret": {s.cfg.GoogleClientSecret}, "redirect_uri": {s.cfg.GoogleRedirectURL}, "grant_type": {"authorization_code"}, "code_verifier": {flow.Verifier}}
	req, _ := http.NewRequest(http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return User{}, "", errors.New("oauth token exchange failed")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return User{}, "", errors.New("oauth token exchange rejected")
	}
	var token struct {
		AccessToken string `json:"access_token"`
	}
	if json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&token) != nil || token.AccessToken == "" {
		return User{}, "", errors.New("oauth token response invalid")
	}
	profileReq, _ := http.NewRequest(http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	profileReq.Header.Set("Authorization", "Bearer "+token.AccessToken)
	profileRes, err := http.DefaultClient.Do(profileReq)
	if err != nil {
		return User{}, "", errors.New("google profile request failed")
	}
	defer profileRes.Body.Close()
	var profile googleProfile
	if profileRes.StatusCode != http.StatusOK || json.NewDecoder(io.LimitReader(profileRes.Body, 1<<20)).Decode(&profile) != nil || !profile.EmailVerified || profile.Email == "" {
		return User{}, "", errors.New("google email is not verified")
	}
	if s.db != nil {
		user, err := s.loginWithGoogleDB(profile.ID, profile.Email, profile.FirstName, profile.LastName, profile.Name, profile.Picture, profile.EmailVerified)
		if err != nil {
			return User{}, "", errors.New("could not persist Google identity")
		}
		s.mu.Lock()
		s.users[user.ID] = user
		s.merchants[user.ID] = &Merchant{Status: StatusIncomplete}
		s.mu.Unlock()
		return user, flow.Next, nil
	}
	return s.LoginWithGoogleNames(profile.ID, profile.Email, profile.FirstName, profile.LastName, profile.Name, profile.Picture, profile.EmailVerified), flow.Next, nil
}
func (s *Service) LoginWithGoogle(providerID, email, name, picture string, verified bool) User {
	firstName, lastName := splitName(name)
	return s.LoginWithGoogleNames(providerID, email, firstName, lastName, name, picture, verified)
}
func splitName(name string) (string, string) {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}
func (s *Service) LoginWithGoogleNames(providerID, email, firstName, lastName, name, picture string, verified bool) User {
	_ = providerID
	if s.db != nil {
		if user, err := s.loginWithGoogleDB(providerID, email, firstName, lastName, name, picture, verified); err == nil {
			s.mu.Lock()
			s.users[user.ID] = user
			if _, exists := s.merchants[user.ID]; !exists {
				s.merchants[user.ID] = &Merchant{Status: StatusIncomplete}
			}
			s.mu.Unlock()
			return user
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, user := range s.users {
		if strings.EqualFold(user.Email, email) {
			user.Name = name
			user.FirstName = firstName
			user.LastName = lastName
			user.Picture = picture
			user.EmailVerified = verified
			s.users[id] = user
			return user
		}
	}
	user := User{ID: uuid.NewString(), Email: strings.ToLower(strings.TrimSpace(email)), Name: name, FirstName: firstName, LastName: lastName, Picture: picture, Role: RoleMerchant, EmailVerified: verified}
	s.users[user.ID] = user
	s.merchants[user.ID] = &Merchant{Status: StatusIncomplete}
	return user
}

func (s *Service) StartSession(userID string, c *fiber.Ctx) error {
	token, err := randomString(48)
	if err != nil {
		return err
	}
	family, err := randomString(24)
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(s.cfg.RefreshTTL)
	if s.db != nil {
		_, err = s.db.Exec(`INSERT INTO refresh_tokens (token_hash, user_id, family_id, jti, expires_at, user_agent, ip_hash) VALUES ($1, $2::uuid, $3, $4, $5, $6, $7)`, hash(token), userID, family, uuid.NewString(), expiresAt, c.Get("User-Agent"), hash(s.cfg.IPHashSalt+c.IP()))
		if err != nil {
			return err
		}
	} else {
		s.mu.Lock()
		s.sessions[hash(token)] = &session{UserID: userID, FamilyID: family, TokenHash: hash(token), ExpiresAt: expiresAt}
		s.mu.Unlock()
	}
	csrf, err := randomString(24)
	if err != nil {
		return err
	}
	secure := s.cfg.CookieSecure || s.cfg.Environment == "production"
	c.Cookie(&fiber.Cookie{Name: s.cfg.RefreshCookieName, Value: token, HTTPOnly: true, Secure: secure, SameSite: "Strict", Path: "/", MaxAge: int(s.cfg.RefreshTTL.Seconds())})
	c.Cookie(&fiber.Cookie{Name: s.cfg.CSRFCookieName, Value: csrf, HTTPOnly: false, Secure: secure, SameSite: "Strict", Path: "/", MaxAge: int(s.cfg.RefreshTTL.Seconds())})
	return nil
}

type claims struct {
	Issuer    string `json:"iss"`
	Audience  string `json:"aud"`
	Subject   string `json:"sub"`
	JTI       string `json:"jti"`
	IssuedAt  int64  `json:"iat"`
	NotBefore int64  `json:"nbf"`
	ExpiresAt int64  `json:"exp"`
	TokenUse  string `json:"token_use"`
}

func (s *Service) issueAccess(userID string) (string, error) {
	now := time.Now()
	c := claims{Issuer: s.cfg.Issuer, Audience: s.cfg.Audience, Subject: userID, JTI: uuid.NewString(), IssuedAt: now.Unix(), NotBefore: now.Unix(), ExpiresAt: now.Add(s.cfg.AccessTTL).Unix(), TokenUse: "access"}
	head, _ := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT"})
	body, _ := json.Marshal(c)
	enc := base64.RawURLEncoding
	input := enc.EncodeToString(head) + "." + enc.EncodeToString(body)
	sum := sha256.Sum256([]byte(input))
	sig, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, sum[:])
	if err != nil {
		return "", err
	}
	return input + "." + enc.EncodeToString(sig), nil
}
func (s *Service) VerifyAccess(raw string) (string, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return "", errors.New("invalid token")
	}
	enc := base64.RawURLEncoding
	hb, err := enc.DecodeString(parts[0])
	if err != nil {
		return "", errors.New("invalid token")
	}
	var header map[string]any
	if json.Unmarshal(hb, &header) != nil || header["alg"] != "RS256" {
		return "", errors.New("invalid token")
	}
	b, err := enc.DecodeString(parts[1])
	if err != nil {
		return "", errors.New("invalid token")
	}
	var c claims
	if json.Unmarshal(b, &c) != nil || c.Issuer != s.cfg.Issuer || c.Audience != s.cfg.Audience || c.TokenUse != "access" {
		return "", errors.New("invalid token")
	}
	if _, err := uuid.Parse(c.Subject); err != nil {
		return "", errors.New("invalid subject")
	}
	now := time.Now().Unix()
	if c.ExpiresAt < now || c.NotBefore > now+30 {
		return "", errors.New("expired token")
	}
	sig, err := enc.DecodeString(parts[2])
	if err != nil {
		return "", errors.New("invalid token")
	}
	sum := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if rsa.VerifyPKCS1v15(s.publicKey, crypto.SHA256, sum[:], sig) != nil {
		return "", errors.New("invalid signature")
	}
	return c.Subject, nil
}
func (s *Service) Refresh(c *fiber.Ctx) (string, error) {
	raw := c.Cookies(s.cfg.RefreshCookieName)
	if raw == "" {
		return "", errors.New("unauthorized")
	}
	h := hash(raw)
	if s.db != nil {
		return s.refreshDB(c, h)
	}
	s.mu.Lock()
	old, ok := s.sessions[h]
	if !ok || old.Revoked || time.Now().After(old.ExpiresAt) {
		s.mu.Unlock()
		return "", errors.New("unauthorized")
	}
	old.Revoked = true
	token, err := randomString(48)
	if err != nil {
		s.mu.Unlock()
		return "", err
	}
	s.sessions[hash(token)] = &session{UserID: old.UserID, FamilyID: old.FamilyID, TokenHash: hash(token), ExpiresAt: time.Now().Add(s.cfg.RefreshTTL)}
	s.mu.Unlock()
	secure := s.cfg.CookieSecure || s.cfg.Environment == "production"
	c.Cookie(&fiber.Cookie{Name: s.cfg.RefreshCookieName, Value: token, HTTPOnly: true, Secure: secure, SameSite: "Strict", Path: "/", MaxAge: int(s.cfg.RefreshTTL.Seconds())})
	return s.issueAccess(old.UserID)
}

func (s *Service) refreshDB(c *fiber.Ctx, tokenHash string) (string, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var userID, familyID string
	var expiresAt time.Time
	var revokedAt sql.NullTime
	err = tx.QueryRow(`SELECT user_id::text, family_id, expires_at, revoked_at FROM refresh_tokens WHERE token_hash = $1 FOR UPDATE`, tokenHash).Scan(&userID, &familyID, &expiresAt, &revokedAt)
	if err != nil {
		return "", errors.New("unauthorized")
	}
	if revokedAt.Valid || time.Now().After(expiresAt) {
		_, _ = tx.Exec(`UPDATE refresh_tokens SET revoked_at = COALESCE(revoked_at, now()) WHERE family_id = $1`, familyID)
		_ = tx.Commit()
		return "", errors.New("unauthorized")
	}
	newToken, err := randomString(48)
	if err != nil {
		return "", err
	}
	newHash := hash(newToken)
	if _, err = tx.Exec(`UPDATE refresh_tokens SET revoked_at = now(), replaced_by_hash = $2 WHERE token_hash = $1`, tokenHash, newHash); err != nil {
		return "", err
	}
	if _, err = tx.Exec(`INSERT INTO refresh_tokens (token_hash, user_id, family_id, jti, expires_at, user_agent, ip_hash) VALUES ($1, $2::uuid, $3, $4, $5, $6, $7)`, newHash, userID, familyID, uuid.NewString(), time.Now().Add(s.cfg.RefreshTTL), c.Get("User-Agent"), hash(s.cfg.IPHashSalt+c.IP())); err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	secure := s.cfg.CookieSecure || s.cfg.Environment == "production"
	c.Cookie(&fiber.Cookie{Name: s.cfg.RefreshCookieName, Value: newToken, HTTPOnly: true, Secure: secure, SameSite: "Strict", Path: "/", MaxAge: int(s.cfg.RefreshTTL.Seconds())})
	return s.issueAccess(userID)
}

func (s *Service) Logout(c *fiber.Ctx) {
	raw := c.Cookies(s.cfg.RefreshCookieName)
	if s.db != nil {
		_, _ = s.db.Exec(`UPDATE refresh_tokens SET revoked_at = COALESCE(revoked_at, now()) WHERE family_id = (SELECT family_id FROM refresh_tokens WHERE token_hash = $1)`, hash(raw))
	} else {
		s.mu.Lock()
		if current := s.sessions[hash(raw)]; current != nil {
			for _, item := range s.sessions {
				if item.FamilyID == current.FamilyID {
					item.Revoked = true
				}
			}
		}
		s.mu.Unlock()
	}
	c.ClearCookie(s.cfg.RefreshCookieName, s.cfg.CSRFCookieName)
}

func (s *Service) Current(userID string) (Me, error) {
	if s.db != nil {
		return s.currentDB(userID)
	}
	s.mu.RLock()
	user, ok := s.users[userID]
	merchant := s.merchants[userID]
	var clone *Merchant
	if merchant != nil {
		copy := *merchant
		if merchant.Profile != nil {
			p := *merchant.Profile
			copy.Profile = &p
		}
		if merchant.Shop != nil {
			sh := *merchant.Shop
			copy.Shop = &sh
		}
		copy.Documents = append([]Document(nil), merchant.Documents...)
		clone = &copy
	}
	s.mu.RUnlock()
	if !ok || clone == nil {
		return Me{}, errors.New("not found")
	}
	return Me{User: user, Merchant: *clone}, nil
}

// LoginRedirect keeps the OAuth callback decision server-side so a returning
// merchant cannot be sent back to the first onboarding step accidentally.
func (s *Service) LoginRedirect(userID, requested string) string {
	locale := "ar"
	if strings.HasPrefix(requested, "/fr/") {
		locale = "fr"
	}
	status := StatusIncomplete
	if s.db != nil {
		_ = s.db.QueryRow(`SELECT status FROM merchant_verifications WHERE user_id = $1::uuid`, userID).Scan(&status)
	} else {
		s.mu.RLock()
		merchant := s.merchants[userID]
		if merchant != nil {
			status = merchant.Status
		}
		s.mu.RUnlock()
	}
	switch status {
	case StatusPending, StatusRejected, StatusApproved, StatusSuspended:
		return "/" + locale + "/merchant/status"
	default:
		me, err := s.Current(userID)
		if err != nil || me.Merchant.Profile == nil {
			return "/" + locale + "/onboarding/personal"
		}
		if me.Merchant.Shop == nil {
			return "/" + locale + "/onboarding/shop"
		}
		found := map[string]bool{}
		for _, document := range me.Merchant.Documents {
			found[document.Type] = true
		}
		if !found["id_front"] || !found["id_back"] || !found["commercial_register"] {
			return "/" + locale + "/onboarding/documents"
		}
		return "/" + locale + "/onboarding/review"
	}
}
func NormalizePhone(value string) (string, error) {
	value = strings.ReplaceAll(strings.TrimSpace(value), " ", "")
	if strings.HasPrefix(value, "+213") {
		value = "0" + strings.TrimPrefix(value, "+213")
	}
	if !phonePattern.MatchString(value) {
		return "", errors.New("phone must be a valid Algerian mobile number")
	}
	return "+213" + value[1:], nil
}
func (s *Service) SaveProfile(userID string, in Profile) (Profile, map[string]string) {
	if s.applicationLocked(userID) {
		return in, map[string]string{"application": "submitted application data is locked"}
	}
	phone, err := NormalizePhone(in.PhoneE164)
	fields := map[string]string{}
	if err != nil {
		fields["phone_e164"] = err.Error()
	}
	if strings.TrimSpace(in.WilayaID) == "" {
		fields["wilaya_id"] = "wilaya is required"
	}
	if len(strings.TrimSpace(in.Address)) < 5 {
		fields["address"] = "address is required"
	}
	if len(fields) > 0 {
		return in, fields
	}
	saved := Profile{FirstName: strings.TrimSpace(in.FirstName), LastName: strings.TrimSpace(in.LastName), PhoneE164: phone, WilayaID: strings.TrimSpace(in.WilayaID), Address: strings.TrimSpace(in.Address)}
	if s.db != nil {
		if _, err := s.db.Exec(`INSERT INTO merchant_profiles (user_id, phone_e164, wilaya_id, address) VALUES ($1::uuid, $2, $3, $4) ON CONFLICT (user_id) DO UPDATE SET phone_e164 = EXCLUDED.phone_e164, wilaya_id = EXCLUDED.wilaya_id, address = EXCLUDED.address, updated_at = now()`, userID, saved.PhoneE164, saved.WilayaID, saved.Address); err != nil {
			return in, map[string]string{"profile": "could not save personal information"}
		}
		name := strings.TrimSpace(strings.TrimSpace(saved.FirstName) + " " + strings.TrimSpace(saved.LastName))
		if _, err := s.db.Exec(`UPDATE users SET first_name = $2, last_name = $3, name = CASE WHEN $4 = '' THEN name ELSE $4 END, updated_at = now() WHERE id = $1::uuid`, userID, saved.FirstName, saved.LastName, name); err != nil {
			return in, map[string]string{"profile": "could not save personal information"}
		}
	}
	s.mu.Lock()
	if m := s.merchants[userID]; m != nil {
		m.Profile = &saved
	}
	if user, ok := s.users[userID]; ok {
		if saved.FirstName != "" {
			user.FirstName = saved.FirstName
		}
		if saved.LastName != "" {
			user.LastName = saved.LastName
		}
		user.Name = strings.TrimSpace(strings.TrimSpace(user.FirstName) + " " + strings.TrimSpace(user.LastName))
		s.users[userID] = user
	}
	s.mu.Unlock()
	return saved, nil
}
func slugify(value, fallback string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = slugPattern.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return fallback
	}
	return value
}
func (s *Service) SaveShop(userID string, in Shop) (Shop, map[string]string) {
	if s.applicationLocked(userID) {
		return in, map[string]string{"application": "submitted application data is locked"}
	}
	fields := map[string]string{}
	if len([]rune(strings.TrimSpace(in.Name))) < 2 {
		fields["name"] = "shop name is required"
	}
	if strings.TrimSpace(in.CategoryID) == "" {
		fields["category_id"] = "category is required"
	}
	if strings.TrimSpace(in.WilayaID) == "" {
		fields["wilaya_id"] = "wilaya is required"
	}
	if strings.TrimSpace(in.RegisterNumber) == "" {
		fields["register_number"] = "commercial register is required"
	}
	if in.RegisterExpiresOn == "" {
		fields["register_expires_on"] = "register expiry is required"
	} else if expires, err := time.Parse("2006-01-02", in.RegisterExpiresOn); err != nil || expires.Before(time.Now().UTC().Truncate(24*time.Hour)) {
		fields["register_expires_on"] = "register expiry must be a valid future date"
	}
	if len(fields) > 0 {
		return in, fields
	}
	slug := slugify(in.Slug, slugify(in.Name, "shop-"+userID[:8]))
	if s.db != nil {
		var existingSlug string
		var edits int
		err := s.db.QueryRow(`SELECT slug, slug_edits FROM merchant_shops WHERE user_id = $1::uuid`, userID).Scan(&existingSlug, &edits)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return in, map[string]string{"shop": "could not load shop information"}
		}
		if err == nil && existingSlug != slug {
			if edits >= 1 {
				return in, map[string]string{"slug": "slug can only be edited once before verification"}
			}
			edits++
		}
		var conflict bool
		if err := s.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM merchant_shops WHERE slug = $1 AND user_id <> $2::uuid)`, slug, userID).Scan(&conflict); err != nil {
			return in, map[string]string{"shop": "could not validate shop slug"}
		}
		if conflict {
			slug += "-" + userID[:6]
		}
		in.Slug, in.SlugEdits = slug, edits
		_, err = s.db.Exec(`INSERT INTO merchant_shops (user_id, name, slug, category_id, description, wilaya_id, city, address, phone, whatsapp, register_number, register_expires_on, slug_edits) VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) ON CONFLICT (user_id) DO UPDATE SET name = EXCLUDED.name, slug = EXCLUDED.slug, category_id = EXCLUDED.category_id, description = EXCLUDED.description, wilaya_id = EXCLUDED.wilaya_id, city = EXCLUDED.city, address = EXCLUDED.address, phone = EXCLUDED.phone, whatsapp = EXCLUDED.whatsapp, register_number = EXCLUDED.register_number, register_expires_on = EXCLUDED.register_expires_on, slug_edits = EXCLUDED.slug_edits, updated_at = now()`, userID, strings.TrimSpace(in.Name), in.Slug, strings.TrimSpace(in.CategoryID), strings.TrimSpace(in.Description), strings.TrimSpace(in.WilayaID), strings.TrimSpace(in.City), strings.TrimSpace(in.Address), strings.TrimSpace(in.Phone), strings.TrimSpace(in.WhatsApp), strings.TrimSpace(in.RegisterNumber), in.RegisterExpiresOn, in.SlugEdits)
		if err != nil {
			return in, map[string]string{"shop": "could not save shop information"}
		}
		s.mu.Lock()
		if m := s.merchants[userID]; m != nil {
			m.Shop = &in
		}
		s.mu.Unlock()
		return in, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.merchants[userID]
	if m == nil {
		return in, map[string]string{"shop": "merchant not found"}
	}
	if m.Shop != nil && m.Shop.Slug != slug && m.Shop.SlugEdits >= 1 {
		return in, map[string]string{"slug": "slug can only be edited once before verification"}
	}
	for id, other := range s.merchants {
		if id != userID && other.Shop != nil && other.Shop.Slug == slug {
			slug += "-" + userID[:6]
			break
		}
	}
	edits := 0
	if m.Shop != nil {
		edits = m.Shop.SlugEdits
		if m.Shop.Slug != slug {
			edits++
		}
	}
	in.Slug, in.SlugEdits = slug, edits
	m.Shop = &in
	return in, nil
}
func contentType(data []byte) (string, string, bool) {
	switch typ := http.DetectContentType(data); typ {
	case "image/jpeg":
		return typ, "jpg", true
	case "image/png":
		return typ, "png", true
	case "image/webp":
		return typ, "webp", true
	case "application/pdf":
		return typ, "pdf", true
	}
	return http.DetectContentType(data), "", false
}
func (s *Service) SaveDocument(userID, typ string, data []byte) (Document, error) {
	if s.applicationLocked(userID) {
		return Document{}, errors.New("submitted application data is locked")
	}
	if _, ok := allowedDocumentTypes[typ]; !ok {
		return Document{}, errors.New("unsupported document type")
	}
	if len(data) == 0 || int64(len(data)) > 8<<20 {
		return Document{}, errors.New("document must be between 1 byte and 8MB")
	}
	ct, ext, ok := contentType(data)
	if !ok {
		return Document{}, errors.New("document content type is not allowed")
	}
	sum := sha256.Sum256(data)
	doc := Document{ID: uuid.NewString(), Type: typ, ContentType: ct, SHA256: hex.EncodeToString(sum[:]), Size: int64(len(data)), CreatedAt: time.Now().UTC(), ObjectKey: "merchants/" + userID + "/" + typ + "/" + uuid.NewString() + "." + ext, Bytes: append([]byte(nil), data...)}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.store.Put(ctx, doc.ObjectKey, doc.ContentType, data); err != nil {
		return Document{}, fmt.Errorf("store private document: %w", err)
	}
	if s.db != nil {
		tx, err := s.db.Begin()
		if err != nil {
			_ = s.store.Delete(ctx, doc.ObjectKey)
			return Document{}, err
		}
		defer tx.Rollback()
		var oldKey string
		queryErr := tx.QueryRow(`SELECT storage_key FROM merchant_documents WHERE user_id = $1::uuid AND document_type = $2 FOR UPDATE`, userID, typ).Scan(&oldKey)
		if queryErr != nil && !errors.Is(queryErr, sql.ErrNoRows) {
			_ = s.store.Delete(ctx, doc.ObjectKey)
			return Document{}, queryErr
		}
		if _, err = tx.Exec(`INSERT INTO merchant_documents (id, user_id, document_type, content_type, byte_size, sha256, storage_key, created_at) VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8) ON CONFLICT (user_id, document_type) DO UPDATE SET id = EXCLUDED.id, content_type = EXCLUDED.content_type, byte_size = EXCLUDED.byte_size, sha256 = EXCLUDED.sha256, storage_key = EXCLUDED.storage_key, created_at = EXCLUDED.created_at`, doc.ID, userID, doc.Type, doc.ContentType, doc.Size, doc.SHA256, doc.ObjectKey, doc.CreatedAt); err != nil {
			_ = s.store.Delete(ctx, doc.ObjectKey)
			return Document{}, err
		}
		if err = tx.Commit(); err != nil {
			_ = s.store.Delete(ctx, doc.ObjectKey)
			return Document{}, err
		}
		if oldKey != "" && oldKey != doc.ObjectKey {
			_ = s.store.Delete(ctx, oldKey)
		}
		doc.Bytes = nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.merchants[userID]
	if m == nil {
		if s.db != nil {
			return doc, nil
		}
		return Document{}, errors.New("merchant not found")
	}
	filtered := m.Documents[:0]
	for _, old := range m.Documents {
		if old.Type != typ {
			filtered = append(filtered, old)
		}
	}
	m.Documents = append(filtered, doc)
	return doc, nil
}
func (s *Service) Submit(userID string, accepted bool) (Merchant, map[string]string) {
	fields := map[string]string{}
	if s.db != nil {
		return s.submitDB(userID, accepted)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.merchants[userID]
	if m.Status == StatusPending || m.Status == StatusApproved || m.Status == StatusSuspended {
		return *m, map[string]string{"application": "submitted application data is locked"}
	}
	if m.Profile == nil {
		fields["profile"] = "personal information is incomplete"
	}
	if m.Shop == nil {
		fields["shop"] = "shop information is incomplete"
	}
	found := map[string]bool{}
	for _, d := range m.Documents {
		found[d.Type] = true
	}
	for _, typ := range []string{"id_front", "id_back", "commercial_register"} {
		if !found[typ] {
			fields["documents"] = "all three verification documents are required"
			break
		}
	}
	if !accepted {
		fields["consent_accepted"] = "privacy consent is required"
	}
	if len(fields) > 0 {
		return *m, fields
	}
	now := time.Now().UTC()
	m.ConsentAcceptedAt = &now
	m.Status = StatusPending
	return *m, nil
}
func (s *Service) SeedTestUser(email string) (User, string, error) {
	user := s.LoginWithGoogle("test", email, "Test Merchant", "", true)
	token, err := s.issueAccess(user.ID)
	return user, token, err
}
