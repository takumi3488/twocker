package cookiestore

import (
	"net/http"
	"net/url"
)

type InMemoryCookieStore struct {
	cookies map[string]map[string]*http.Cookie
}

func (s *InMemoryCookieStore) SetCookies(url *url.URL, cookies []*http.Cookie) {
	if s.cookies[url.Hostname()] == nil {
		s.cookies[url.Hostname()] = make(map[string]*http.Cookie)
	}
	for _, cookie := range cookies {
		s.cookies[url.Hostname()][cookie.Name] = cookie
	}
}

func (s *InMemoryCookieStore) Cookies(url *url.URL) []*http.Cookie {
	cookies := make([]*http.Cookie, 0)
	if s.cookies[url.Hostname()] == nil {
		return cookies
	}
	for _, cookie := range s.cookies[url.Hostname()] {
		cookies = append(cookies, cookie)
	}
	return cookies
}

func NewInMemoryCookieStore() *InMemoryCookieStore {
	return &InMemoryCookieStore{
		cookies: make(map[string]map[string]*http.Cookie),
	}
}
