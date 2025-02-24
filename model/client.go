package model

import (
	"io"
	"net/http"
	"net/url"
)

type TwockerClient struct {
	Client *http.Client
}

func NewTwockerClient() *TwockerClient {
	return &TwockerClient{
		Client: &http.Client{},
	}
}

func (c *TwockerClient) WithCookieJar(jar http.CookieJar) *TwockerClient {
	c.Client.Jar = jar
	return c
}

func (c *TwockerClient) Get(url string, headers [][2]string) (*TwockerResponse, error) {
	return command(c, http.MethodGet, url, nil, headers)
}

func (c *TwockerClient) Post(url string, body io.Reader, headers [][2]string) (*TwockerResponse, error) {
	return command(c, http.MethodPost, url, body, headers)
}

func (c *TwockerClient) Patch(url string, body io.Reader, headers [][2]string) (*TwockerResponse, error) {
	return command(c, http.MethodPatch, url, body, headers)
}

func (c *TwockerClient) Delete(url string, body io.Reader, headers [][2]string) (*TwockerResponse, error) {
	return command(c, http.MethodDelete, url, body, headers)
}

func (c *TwockerClient) Put(url string, contentType string, body io.Reader, headers [][2]string) (*TwockerResponse, error) {
	return command(c, http.MethodPut, url, body, headers)
}

func command(c *TwockerClient, method string, url string, body io.Reader, headers [][2]string) (*TwockerResponse, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return &TwockerResponse{}, nil
	}
	for _, header := range headers {
		req.Header.Add(header[0], header[1])
	}

	resp, err := c.Client.Do(req)

	if err != nil {
		return &TwockerResponse{}, nil
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return &TwockerResponse{}, nil
	}

	return NewTwockerResponse(resp.StatusCode, b), nil
}

func (c *TwockerClient) Cookies(url *url.URL) []*http.Cookie {
	return c.Client.Jar.Cookies(url)
}

func (c *TwockerClient) SetCookie(url *url.URL, cookie *http.Cookie) {
	c.Client.Jar.SetCookies(url, []*http.Cookie{cookie})
}
