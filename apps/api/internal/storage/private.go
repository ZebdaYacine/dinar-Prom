package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

// PrivateStore is the narrow object-storage contract used by merchant identity.
// Implementations must never expose a public bucket URL.
type PrivateStore interface {
	ObjectStore
	PresignGet(context.Context, string, time.Duration) (string, error)
}

// ObjectStore is shared by private identity documents and public offer media.
// Buckets and URL visibility are selected when the implementation is created.
type ObjectStore interface {
	Put(context.Context, string, string, []byte) error
	Get(context.Context, string) ([]byte, error)
	Delete(context.Context, string) error
}

type MemoryStore struct {
	mu      sync.RWMutex
	objects map[string][]byte
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{objects: map[string][]byte{}} }

func (s *MemoryStore) Put(_ context.Context, key, _ string, data []byte) error {
	s.mu.Lock()
	s.objects[key] = append([]byte(nil), data...)
	s.mu.Unlock()
	return nil
}

func (s *MemoryStore) Get(_ context.Context, key string) ([]byte, error) {
	s.mu.RLock()
	data, ok := s.objects[key]
	s.mu.RUnlock()
	if !ok {
		return nil, errors.New("object not found")
	}
	return append([]byte(nil), data...), nil
}

func (s *MemoryStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	delete(s.objects, key)
	s.mu.Unlock()
	return nil
}

func (s *MemoryStore) PresignGet(_ context.Context, key string, ttl time.Duration) (string, error) {
	s.mu.RLock()
	_, ok := s.objects[key]
	s.mu.RUnlock()
	if !ok {
		return "", errors.New("object not found")
	}
	return fmt.Sprintf("memory://private/%s?expires=%d", url.PathEscape(key), time.Now().Add(ttl).Unix()), nil
}

// S3Store implements the subset of the S3 API required by the private MinIO
// bucket. It uses path-style requests and AWS Signature Version 4.
type S3Store struct {
	endpoint, accessKey, secretKey, bucket, region string
	client                                         *http.Client
	now                                            func() time.Time
}

type S3Config struct {
	Endpoint, AccessKey, SecretKey, Bucket, Region string
	UseSSL                                         bool
	Client                                         *http.Client
}

func NewS3Store(cfg S3Config) (*S3Store, error) {
	if cfg.Endpoint == "" || cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.Bucket == "" {
		return nil, errors.New("private object storage is not fully configured")
	}
	scheme := "http"
	if cfg.UseSSL {
		scheme = "https"
	}
	endpoint := strings.TrimRight(cfg.Endpoint, "/")
	if !strings.Contains(endpoint, "://") {
		endpoint = scheme + "://" + endpoint
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" {
		return nil, errors.New("invalid object storage endpoint")
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	if cfg.Client == nil {
		cfg.Client = &http.Client{Timeout: 30 * time.Second}
	}
	return &S3Store{endpoint: strings.TrimRight(u.String(), "/"), accessKey: cfg.AccessKey, secretKey: cfg.SecretKey, bucket: cfg.Bucket, region: cfg.Region, client: cfg.Client, now: time.Now}, nil
}

func (s *S3Store) objectURL(key string) string {
	parts := strings.Split(strings.TrimLeft(key, "/"), "/")
	for index := range parts {
		parts[index] = url.PathEscape(parts[index])
	}
	return s.endpoint + "/" + url.PathEscape(s.bucket) + "/" + strings.Join(parts, "/")
}

func (s *S3Store) Put(ctx context.Context, key, contentType string, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, s.objectURL(key), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	return s.do(req, data)
}

func (s *S3Store) Delete(ctx context.Context, key string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, s.objectURL(key), nil)
	if err != nil {
		return err
	}
	return s.do(req, nil)
}

func (s *S3Store) Get(ctx context.Context, key string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.objectURL(key), nil)
	if err != nil {
		return nil, err
	}
	if err := s.sign(req, nil); err != nil {
		return nil, err
	}
	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		message, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return nil, fmt.Errorf("object storage returned %s: %s", res.Status, strings.TrimSpace(string(message)))
	}
	return io.ReadAll(io.LimitReader(res.Body, 32<<20))
}

func (s *S3Store) do(req *http.Request, body []byte) error {
	if err := s.sign(req, body); err != nil {
		return err
	}
	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		message, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return fmt.Errorf("object storage returned %s: %s", res.Status, strings.TrimSpace(string(message)))
	}
	return nil
}

func (s *S3Store) sign(req *http.Request, body []byte) error {
	payloadHash := sha256Hex(body)
	now := s.now().UTC()
	req.Header.Set("X-Amz-Date", now.Format("20060102T150405Z"))
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"
	canonicalHeaders := "host:" + req.URL.Host + "\n" + "x-amz-content-sha256:" + payloadHash + "\n" + "x-amz-date:" + req.Header.Get("X-Amz-Date") + "\n"
	canonical := req.Method + "\n" + req.URL.EscapedPath() + "\n" + req.URL.Query().Encode() + "\n" + canonicalHeaders + "\n" + signedHeaders + "\n" + payloadHash
	scope := now.Format("20060102") + "/" + s.region + "/s3/aws4_request"
	stringToSign := "AWS4-HMAC-SHA256\n" + req.Header.Get("X-Amz-Date") + "\n" + scope + "\n" + sha256Hex([]byte(canonical))
	signature := hex.EncodeToString(hmacSHA256(signingKey(s.secretKey, now, s.region), stringToSign))
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+s.accessKey+"/"+scope+", SignedHeaders="+signedHeaders+", Signature="+signature)
	return nil
}

func (s *S3Store) PresignGet(_ context.Context, key string, ttl time.Duration) (string, error) {
	if ttl <= 0 || ttl > 5*time.Minute {
		return "", errors.New("signed document URLs must expire within five minutes")
	}
	now := s.now().UTC()
	u, err := url.Parse(s.objectURL(key))
	if err != nil {
		return "", err
	}
	scope := now.Format("20060102") + "/" + s.region + "/s3/aws4_request"
	q := u.Query()
	q.Set("X-Amz-Algorithm", "AWS4-HMAC-SHA256")
	q.Set("X-Amz-Credential", s.accessKey+"/"+scope)
	q.Set("X-Amz-Date", now.Format("20060102T150405Z"))
	q.Set("X-Amz-Expires", fmt.Sprintf("%d", int(ttl.Seconds())))
	q.Set("X-Amz-SignedHeaders", "host")
	u.RawQuery = canonicalQuery(q)
	canonical := http.MethodGet + "\n" + u.EscapedPath() + "\n" + u.RawQuery + "\n" + "host:" + u.Host + "\n\n" + "host\nUNSIGNED-PAYLOAD"
	stringToSign := "AWS4-HMAC-SHA256\n" + now.Format("20060102T150405Z") + "\n" + scope + "\n" + sha256Hex([]byte(canonical))
	q.Set("X-Amz-Signature", hex.EncodeToString(hmacSHA256(signingKey(s.secretKey, now, s.region), stringToSign)))
	u.RawQuery = canonicalQuery(q)
	return u.String(), nil
}

func canonicalQuery(values url.Values) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var parts []string
	for _, key := range keys {
		items := append([]string(nil), values[key]...)
		sort.Strings(items)
		for _, value := range items {
			parts = append(parts, awsEscape(key)+"="+awsEscape(value))
		}
	}
	return strings.Join(parts, "&")
}

func awsEscape(value string) string { return strings.ReplaceAll(url.QueryEscape(value), "+", "%20") }

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key []byte, value string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}

func signingKey(secret string, now time.Time, region string) []byte {
	date := hmacSHA256([]byte("AWS4"+secret), now.Format("20060102"))
	regionKey := hmacSHA256(date, region)
	serviceKey := hmacSHA256(regionKey, "s3")
	return hmacSHA256(serviceKey, "aws4_request")
}
