package twocker

import (
	"github.com/takumi3488/twocker/model"
)

type TwockerClient = model.TwockerClient
type TwockerResponse = model.TwockerResponse

func NewTwockerClient() *model.TwockerClient {
	return model.NewTwockerClient()
}
