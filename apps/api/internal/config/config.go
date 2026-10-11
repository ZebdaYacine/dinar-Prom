package config

import (
	"os"
	"strconv"
	"time"
)

// Config contains the runtime values needed by the API bootstrap. Feature
// modules will extend this without reading process environment directly.
type Config struct {
	Addr               string
	Environment        string
	AppURL             string
	DatabaseURL        string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	JWTPrivateKeyFile  string
	JWTPublicKeyFile   string
	JWTIssuer          string
	JWTAudience        string
	RefreshCookieName  string
	CSRFCookieName     string
	CSRFHeaderName     string
	CookieSecure       bool
	JWTAccessTTL       time.Duration
	JWTRefreshTTL      time.Duration
	AllowedOrigins     string
	IPHashSalt         string
	MinIOEndpoint      string
	MinIOAccessKey     string
	MinIOSecretKey     string
	MinIOBucketPrivate string
	MinIOBucketPublic  string
	PublicMediaURL     string
	MinIORegion        string
	MinIOUseSSL        bool
	SMTPHost           string
	SMTPPort           int
	SMTPUser           string
	SMTPPass           string
	MailFrom           string
}

func Load() Config {
	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	environment := os.Getenv("APP_ENV")
	if environment == "" {
		environment = "development"
	}
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:8180"
	}
	accessTTL := durationEnv("JWT_ACCESS_TTL_SECONDS", 600)
	refreshTTL := durationEnv("JWT_REFRESH_TTL_SECONDS", 2592000)
	secure, _ := strconv.ParseBool(os.Getenv("COOKIE_SECURE"))
	minioSSL, _ := strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))
	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil || smtpPort <= 0 {
		smtpPort = 587
	}
	return Config{
		Addr: addr, Environment: environment, AppURL: appURL, DatabaseURL: os.Getenv("DATABASE_URL"),
		GoogleClientID: os.Getenv("GOOGLE_CLIENT_ID"), GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"), GoogleRedirectURL: os.Getenv("GOOGLE_REDIRECT_URL"),
		JWTPrivateKeyFile: os.Getenv("JWT_PRIVATE_KEY_FILE"), JWTPublicKeyFile: os.Getenv("JWT_PUBLIC_KEY_FILE"), JWTIssuer: envOr("JWT_ISSUER", "dinar-api"), JWTAudience: envOr("JWT_AUDIENCE", "dinar-web"),
		RefreshCookieName: envOr("REFRESH_TOKEN_COOKIE_NAME", "dinar_refresh"), CSRFCookieName: envOr("CSRF_COOKIE_NAME", "dinar_csrf"), CSRFHeaderName: envOr("CSRF_HEADER_NAME", "X-CSRF-Token"), CookieSecure: secure,
		JWTAccessTTL: accessTTL, JWTRefreshTTL: refreshTTL,
		AllowedOrigins: envOr("ALLOWED_ORIGINS", "http://127.0.0.1:8180,http://localhost:8180,http://127.0.0.1:3143,http://localhost:3143"),
		IPHashSalt:     os.Getenv("IP_HASH_SALT"),
		MinIOEndpoint:  os.Getenv("MINIO_ENDPOINT"), MinIOAccessKey: os.Getenv("MINIO_ACCESS_KEY"), MinIOSecretKey: os.Getenv("MINIO_SECRET_KEY"),
		MinIOBucketPrivate: envOr("MINIO_BUCKET_PRIVATE", "dinar-private"), MinIOBucketPublic: envOr("MINIO_BUCKET_PUBLIC", "dinar-public"), PublicMediaURL: envOr("PUBLIC_MEDIA_URL", "http://localhost:9100/dinar-public"), MinIORegion: envOr("MINIO_REGION", "us-east-1"), MinIOUseSSL: minioSSL,
		SMTPHost: os.Getenv("SMTP_HOST"), SMTPPort: smtpPort, SMTPUser: os.Getenv("SMTP_USER"), SMTPPass: os.Getenv("SMTP_PASS"), MailFrom: envOr("MAIL_FROM", "Dinar DZ <no-reply@dinar.dz>"),
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func durationEnv(key string, fallback int) time.Duration {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil || value <= 0 {
		value = fallback
	}
	return time.Duration(value) * time.Second
}
