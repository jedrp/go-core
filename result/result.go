package result

import (
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime"
)

type Error struct {
	Code    ErrorCode `json:"code,omitempty"`
	Message string    `json:"message,omitempty"`
}

type Result struct {
	Value interface{}
	Error *Error
}

func OK(v interface{}) *Result {
	return &Result{
		v,
		nil,
	}
}

func (r *Result) IsSuccess() bool {
	return r.Error == nil
}

func (r *Result) IsFailure() bool {
	return r.Error != nil
}

func Failf(code ErrorCode, f string, o ...interface{}) *Result {
	return &Result{
		nil,
		&Error{
			Code:    code,
			Message: fmt.Sprintf(f, o...),
		},
	}
}
func Fail(code ErrorCode, m string) *Result {
	return &Result{
		nil,
		&Error{
			Code:    code,
			Message: m,
		},
	}
}

// Implement Responder interface (Responder is an interface for types to implement, when they want to be considered for writing HTTP responses)
func (r *Result) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
	if r.Error == nil {
		rw.WriteHeader(200)
		if err := producer.Produce(rw, r.Value); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	} else {
		code := r.Error.Code
		switch code {
		case InvalidArgument:
			rw.WriteHeader(400)
		case NotFound:
			rw.WriteHeader(404)
		case Aborted:
			rw.WriteHeader(412)
		case Unauthenticated:
			rw.WriteHeader(401)
		case PermissionDenied:
			rw.WriteHeader(403)
		default:
			rw.WriteHeader(500)
		}
		if err := producer.Produce(rw, r.Error); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
