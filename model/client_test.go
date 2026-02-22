package model

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/takumi3488/twocker/cookiestore"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNewTwockerClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewTwockerClient()
	resp, err := c.Get(server.URL, nil)
	if err != nil {
		t.Errorf("Error making GET request: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
}

func TestNewTwockerClientWithCookieJar(t *testing.T) {
	ctx := context.Background()
	cookieStores := []http.CookieJar{
		cookiestore.NewInMemoryCookieStore(),
		createRedisCookieStore(ctx),
	}
	for _, cookieStore := range cookieStores {
		cookieTestHelper(t, cookieStore, "q6GqJCZnhTFfSxZTBfBtaA==")
		cookieTestHelper(t, cookieStore, "+Gjy7/mfbY+DeVzg0tAzcQ==")
		cookieTestHelper(t, cookieStore, "lpRr6W5HRzkMzaNxjjlYHA==")
	}
}

func createRedisCookieStore(ctx context.Context) *cookiestore.RedisCookieStore {
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}
	ip, err := redisContainer.Host(ctx)
	if err != nil {
		panic(err)
	}
	port, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		panic(err)
	}
	redisOptions := &cookiestore.NewRedisCookieStoreOption{
		Addr: ip + ":" + port.Port(),
	}
	return cookiestore.NewRedisCookieStore(redisOptions, nil)
}

func cookieTestHelper(t *testing.T, cookieStore http.CookieJar, kaisuu string) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "ENC_KAISUU",
			Value: kaisuu,
		})
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewTwockerClient().WithCookieJar(cookieStore)
	if c.Client.Jar == nil {
		t.Errorf("Expected cookie jar to be set")
	}
	resp, err := c.Get(server.URL, nil)
	if err != nil {
		t.Errorf("Error making GET request: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Errorf("Error parsing URL: %v", err)
	}
	cookies := c.Client.Jar.Cookies(serverURL)
	if len(cookies) == 0 {
		t.Errorf("Expected cookies to be set")
	}
	for _, cookie := range cookies {
		if cookie.Name == "ENC_KAISUU" && cookie.Value == kaisuu {
			return
		}
	}
	t.Errorf("Expected cookie ENC_KAISUU to be set")
}
