package model

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
