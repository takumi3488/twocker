package model

import (
	"bytes"
	"encoding/json"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

type TwockerResponse struct {
	StatusCode int
	body       []byte
	url        *url.URL
}

func NewTwockerResponse(statusCode int, body []byte, url *url.URL) *TwockerResponse {
	return &TwockerResponse{
		StatusCode: statusCode,
		body:       body,
		url:        url,
	}
}

func TwockerJson[T any](r *TwockerResponse) (*T, error) {
	var v T
	err := json.Unmarshal(r.body, &v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *TwockerResponse) Select(selector string) (*goquery.Selection, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(r.body))
	if err != nil {
		return nil, err
	}
	return doc.Find(selector), nil
}

func (r *TwockerResponse) URL() *url.URL {
	return r.url
}

func (r *TwockerResponse) Body() []byte {
	return r.body
}

func (r *TwockerResponse) Text() string {
	return string(r.body)
}
