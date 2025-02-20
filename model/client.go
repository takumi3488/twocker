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
