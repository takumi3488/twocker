package cookiestore

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

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

type CookieJson struct {
	Name       string        `json:"Name"`
	Value      string        `json:"Value"`
	Path       string        `json:"Path"`
	Domain     string        `json:"Domain"`
	Expires    string        `json:"Expires"`
	RawExpires string        `json:"RawExpires"`
	MaxAge     int           `json:"MaxAge"`
	Secure     bool          `json:"Secure"`
	HttpOnly   bool          `json:"HttpOnly"`
	SameSite   http.SameSite `json:"SameSite"`
	Raw        string        `json:"Raw"`
}

func cookiesToJson(c []*http.Cookie) string {
	if len(c) == 0 {
		return "[]"
	}
	s := "["
	for _, cookie := range c {
		s += "{"
		s += "\"Name\": \"" + cookie.Name + "\","
		s += "\"Value\": \"" + cookie.Value + "\","
		s += "\"Path\": \"" + cookie.Path + "\","
		s += "\"Domain\": \"" + cookie.Domain + "\","
		s += "\"Expires\": \"" + cookie.Expires.Format(time.RFC3339) + "\","
		s += "\"RawExpires\": \"" + cookie.RawExpires + "\","
		s += "\"MaxAge\": " + fmt.Sprint(cookie.MaxAge) + ","
		s += "\"Secure\": " + fmt.Sprint(cookie.Secure) + ","
		s += "\"HttpOnly\": " + fmt.Sprint(cookie.HttpOnly) + ","
		s += "\"SameSite\": " + fmt.Sprint(cookie.SameSite) + ","
		s += "\"Raw\": \"" + cookie.Raw + "\""
		s += "},"
	}
	s = s[:len(s)-1] + "]"
	return s
}

func jsonToCookies(s string) []*http.Cookie {
	var c []CookieJson
	err := json.Unmarshal([]byte(s), &c)
	if err != nil {
		panic(err)
	}
	res := make([]*http.Cookie, len(c))
	for i, cookie := range c {
		expire, err := time.Parse(time.RFC3339, cookie.Expires)
		if err != nil {
			panic(err)
		}
		res[i] = &http.Cookie{
			Name:       cookie.Name,
			Value:      cookie.Value,
			Path:       cookie.Path,
			Domain:     cookie.Domain,
			Expires:    expire,
			RawExpires: cookie.RawExpires,
			MaxAge:     cookie.MaxAge,
			Secure:     cookie.Secure,
			HttpOnly:   cookie.HttpOnly,
			SameSite:   cookie.SameSite,
			Raw:        cookie.Raw,
		}
	}
	return res
}
