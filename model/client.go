package model

import (
	"io"
	"net/http"
)

type TwockerClient struct {
	client *http.Client
}

func NewTwockerClient() *TwockerClient {
	return &TwockerClient{
		client: &http.Client{},
	}
}

func (c *TwockerClient) WithCookieJar(jar http.CookieJar) *TwockerClient {
	c.client.Jar = jar
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

	resp, err := c.client.Do(req)

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
