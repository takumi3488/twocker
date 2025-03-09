package twocker

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/takumi3488/twocker/model"
)

type TwockerClient = model.TwockerClient
type TwockerResponse = model.TwockerResponse
type Selection = goquery.Selection

func NewTwockerClient() *model.TwockerClient {
	return model.NewTwockerClient()
}

func TwockerJson[T any](r *TwockerResponse) (*T, error) {
	return model.TwockerJson[T](r)
}
