package storage

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestS3StoreMinIOSmoke(t *testing.T) {
	endpoint := os.Getenv("TEST_MINIO_ENDPOINT")
	if endpoint == "" {
		t.Skip("TEST_MINIO_ENDPOINT is not configured")
	}
	store, err := NewS3Store(S3Config{
		Endpoint:  endpoint,
		AccessKey: os.Getenv("TEST_MINIO_ACCESS_KEY"),
		SecretKey: os.Getenv("TEST_MINIO_SECRET_KEY"),
		Bucket:    os.Getenv("TEST_MINIO_BUCKET"),
		Region:    "us-east-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	key := "integration/epic02-" + strconv.FormatInt(time.Now().UnixNano(), 10) + ".pdf"
	defer store.Delete(context.Background(), key)
	if err := store.Put(context.Background(), key, "application/pdf", []byte("private-smoke")); err != nil {
		t.Fatal(err)
	}
	signedURL, err := store.PresignGet(context.Background(), key, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.Get(signedURL)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	if response.StatusCode != http.StatusOK || string(body) != "private-smoke" {
		t.Fatalf("unexpected signed download: status=%d body=%q", response.StatusCode, body)
	}
}

func TestS3StoreSignsPrivateObjectOperations(t *testing.T) {
	var methods []string
	client := &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		methods = append(methods, r.Method)
		if !strings.HasPrefix(r.Header.Get("Authorization"), "AWS4-HMAC-SHA256") {
			t.Error("request was not signed")
		}
		if r.Method == http.MethodPut {
			body, _ := io.ReadAll(r.Body)
			if string(body) != "private" || r.Header.Get("Content-Type") != "application/pdf" {
				t.Errorf("unexpected upload: %q %q", body, r.Header.Get("Content-Type"))
			}
		}
		return &http.Response{StatusCode: http.StatusNoContent, Status: "204 No Content", Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
	})}
	store, err := NewS3Store(S3Config{Endpoint: "http://minio:9000", AccessKey: "access", SecretKey: "secret", Bucket: "private", Region: "us-east-1", Client: client})
	if err != nil {
		t.Fatal(err)
	}
	store.now = func() time.Time { return time.Date(2026, 10, 11, 10, 0, 0, 0, time.UTC) }
	if err := store.Put(context.Background(), "merchants/user/id/file.pdf", "application/pdf", []byte("private")); err != nil {
		t.Fatal(err)
	}
	if err := store.Delete(context.Background(), "merchants/user/id/file.pdf"); err != nil {
		t.Fatal(err)
	}
	if len(methods) != 2 || methods[0] != http.MethodPut || methods[1] != http.MethodDelete {
		t.Fatalf("unexpected methods: %v", methods)
	}
	signed, err := store.PresignGet(context.Background(), "merchants/user/id/file.pdf", 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := url.Parse(signed)
	if parsed.Query().Get("X-Amz-Expires") != "300" || parsed.Query().Get("X-Amz-Signature") == "" {
		t.Fatalf("unexpected signed URL: %s", signed)
	}
	if _, err := store.PresignGet(context.Background(), "merchants/user/id/file.pdf", 6*time.Minute); err == nil {
		t.Fatal("expected TTL above five minutes to fail")
	}
}
