package cqs

import (
	"context"
	"testing"
)

type testCommand struct{}

func (c *testCommand) HandlerID() string {
	return "testHandler"
}

type testCommandResponse struct {
	Value int
}

type testHandler struct{}

func (*testHandler) Handle(ctx context.Context, command *testCommand) (*testCommandResponse, error) {
	return &testCommandResponse{Value: 1}, nil
}

func TestDitpatcherSuccess(t *testing.T) {
	ResetDispatcherSetting()
	ctx := context.Background()
	RegisterHandler[*testCommand, *testCommandResponse](ctx, &testHandler{})

	r, err := Send[*testCommand, *testCommandResponse](ctx, &testCommand{})

	if err != nil {
		t.Error("should not return error")
	}
	if r.Value != 1 {
		t.Errorf("expected 1 but got %v", r.Value)
	}

}

func TestHandlerFactorySuccess(t *testing.T) {
	ResetDispatcherSetting()
	ctx := context.Background()
	RegisterRequestHandlerFactory[*testCommand, *testCommandResponse](ctx, func() Handler[*testCommand, *testCommandResponse] {
		return &testHandler{}
	})

	r, err := Send[*testCommand, *testCommandResponse](ctx, &testCommand{})

	if err != nil {
		t.Error("should not return error")
	}
	if r.Value != 1 {
		t.Errorf("expected 1 but got %v", r.Value)
	}

}

func TestHandlerDuplicated(t *testing.T) {
	ResetDispatcherSetting()
	ctx := context.Background()
	err := RegisterHandler[*testCommand, *testCommandResponse](ctx, &testHandler{})
	if err != nil {
		t.Error("should not return error")
	}
	err = RegisterHandler[*testCommand, *testCommandResponse](ctx, &testHandler{})

	if err == nil {
		t.Errorf("expected duplicated error")
	}
}
