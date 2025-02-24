package model

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/takumi3488/twocker/cookiestore"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNewTwockerClientGet(t *testing.T) {
	c := NewTwockerClient()
	resp, err := c.Get("https://jsonplaceholder.typicode.com/todos")
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
		cookieTestHelper(t, cookieStore, "q6GqJCZnhTFfSxZTBfBtaA%3d%3d")
		cookieTestHelper(t, cookieStore, "%2bGjy7/mfbY%2bDeVzg0tAzcQ%3d%3d")
		cookieTestHelper(t, cookieStore, "lpRr6W5HRzkMzaNxjjlYHA%3d%3d")
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
	tohoho_url := "https://www.tohoho-web.com/cgi/wwwcook.cgi"
	c := NewTwockerClient().WithCookieJar(cookieStore)
	if c.client.Jar == nil {
		t.Errorf("Expected cookie jar to be set")
	}
	resp, err := c.Get(tohoho_url)
	if err != nil {
		t.Errorf("Error making GET request: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
	tohoho_url_parsed, err := url.Parse(tohoho_url)
	if err != nil {
		t.Errorf("Error parsing URL: %v", err)
	}
	cookies := c.client.Jar.Cookies(tohoho_url_parsed)
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
