package model

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/takumi3488/twocker/model/cookiestore"
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
	cookieStores := []http.CookieJar{
		cookiestore.NewInMemoryCookieStore(),
	}
	for _, cookieStore := range cookieStores {
		cookieTestHelper(t, cookieStore)
	}
}

func cookieTestHelper(t *testing.T, cookieStore http.CookieJar) {
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
		if cookie.Name == "ENC_KAISUU" && cookie.Value == "q6GqJCZnhTFfSxZTBfBtaA%3d%3d" {
			return
		}
	}
	t.Errorf("Expected cookie ENC_KAISUU to be set")
}
