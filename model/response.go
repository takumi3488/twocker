package model

import "encoding/json"

type TwockerResponse struct {
	StatusCode int
	body       []byte
}

func NewTwockerResponse(statusCode int, body []byte) *TwockerResponse {
	return &TwockerResponse{
		StatusCode: statusCode,
		body:       body,
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
