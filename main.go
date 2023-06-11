package main

import (
	"context"
	"fmt"

	"github.com/jedrp/go-core/cqs"
)

type ARequest struct {
}
type AResponse struct {
	*cqs.BaseResponse
}
type AHandler struct {
}

func (r *ARequest) HandlerID() string {
	return "1"
}

func (h *AHandler) Handle(ctx context.Context, request *ARequest) (*AResponse, error) {
	fmt.Println("handling")
	return &AResponse{}, nil
}

func main() {

	err := cqs.RegisterHandler[*ARequest, *AResponse](context.Background(), &AHandler{})
	if err != nil {
		panic(err)
	}
	cqs.Send[*ARequest, *AResponse](context.Background(), &ARequest{})

}
