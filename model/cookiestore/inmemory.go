package cookiestore

import (
	"net/http"
	"net/url"
)

type InMemoryCookieStore struct {
	cookies map[string][]*http.Cookie
}

func (s *InMemoryCookieStore) SetCookies(url *url.URL, cookies []*http.Cookie) {
	s.cookies[url.Hostname()] = cookies
}

func (s *InMemoryCookieStore) Cookies(url *url.URL) []*http.Cookie {
	return s.cookies[url.Hostname()]
}

func NewInMemoryCookieStore() *InMemoryCookieStore {
	return &InMemoryCookieStore{
		cookies: make(map[string][]*http.Cookie),
	}
}
