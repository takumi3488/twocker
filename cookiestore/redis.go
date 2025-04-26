package cookiestore

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/redis/go-redis/v9"
)

type RedisCookieStore struct {
	redisClient *redis.Client
	prefix      string
}

type NewRedisCookieStoreOption = redis.Options

func NewRedisCookieStore(option *NewRedisCookieStoreOption, prefix *string) *RedisCookieStore {
	if prefix == nil {
		prefix = new(string)
		*prefix = "twocker"
	}
	return &RedisCookieStore{
		redisClient: redis.NewClient(option),
		prefix:      *prefix,
	}
}

func (s *RedisCookieStore) SetCookies(url *url.URL, cookies []*http.Cookie) {
	ctx := context.Background()
	for _, cookie := range s.Cookies(url) {
		flg := false
		for _, newCookie := range cookies {
			if cookie.Name == newCookie.Name {
				flg = true
				break
			}
		}
		if !flg {
			cookies = append(cookies, cookie)
		}
	}
	err := s.redisClient.Set(
		ctx,
		s.prefix+":"+url.Hostname(),
		cookiesToJson(cookies),
		0,
	).Err()
	if err != nil {
		panic(err)
	}
}

func (s *RedisCookieStore) Cookies(url *url.URL) []*http.Cookie {
	ctx := context.Background()
	res, err := s.redisClient.Get(ctx, s.prefix+":"+url.Hostname()).Result()
	if err != nil {
		return nil
	}
	return jsonToCookies(res)
}

func cookiesToJson(c []*http.Cookie) string {
	if len(c) == 0 {
		return "[]"
	}

	cookiesJSON, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(cookiesJSON)
}

func jsonToCookies(s string) []*http.Cookie {
	var cookies []*http.Cookie
	err := json.Unmarshal([]byte(s), &cookies)
	if err != nil {
		panic(err)
	}
	return cookies
}
