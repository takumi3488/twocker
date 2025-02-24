# Twocker

Simple http client extension with `net/http` and `json`, `github.com/PuerkitoBio/goquery`.

## Usage

```go
import "github.com/takumi3488/twocker"

t := twocker.NewTwockerClient().WithCookieJar(cookiestore.NewRedisCookieStore(&cookiestore.NewRedisCookieStoreOption{
        Addr:     "localhost:6379",
}, prefix))
resp, _ := t.Get("https://github.com/takumi3488/twocker", nil)
selection, _ := resp.Select("meta[name='description']")
description, _ := selection.Attr("content")
println(description)
```

## Features

- You can get a `TwockerResponse` with a simple `t.GET` or `t.POST` statement.
- The `TwockerResponse` has a `Select` method to easily extract elements from HTML.
- `TwockerJson` function maps JSON response from `TwockerResponse` to a structure.
- Some options for `CookieJar`
  - `InMemoryCookieStore`: destroyed at program exit
  - `RedisCookieStore`: stored in Redis (see Usage)
