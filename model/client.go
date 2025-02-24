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

func (c *TwockerClient) Get(url string) (*TwockerResponse, error) {
	resp, err := c.client.Get(url)
	if err != nil {
		return &TwockerResponse{}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &TwockerResponse{}, nil
	}

	return NewTwockerResponse(resp.StatusCode, body), nil
}

func (c *TwockerClient) Post(url string, contentType string, body io.Reader) (*TwockerResponse, error) {
	return command(c, http.MethodPost, url, contentType, body)
}

func (c *TwockerClient) Patch(url string, contentType string, body io.Reader) (*TwockerResponse, error) {
	return command(c, http.MethodPatch, url, contentType, body)
}

func (c *TwockerClient) Delete(url string, contentType string, body io.Reader) (*TwockerResponse, error) {
	return command(c, http.MethodDelete, url, contentType, body)
}

func (c *TwockerClient) Put(url string, contentType string, body io.Reader) (*TwockerResponse, error) {
	return command(c, http.MethodPut, url, contentType, body)
}

func command(c *TwockerClient, method string, url string, contentType string, body io.Reader) (*TwockerResponse, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return &TwockerResponse{}, nil
	}
	req.Header.Set("Content-Type", contentType)

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