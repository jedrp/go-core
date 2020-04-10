package result

import "testing"

func TestResult(t *testing.T) {
	tt := []struct {
		code               ErrorCode
		expectedFailResult bool
	}{
		{code: InvalidArgument},
		{code: Unauthenticated},
		{code: FailedPrecondition},
		{code: PermissionDenied},
	}
	for _, tc := range tt {
		r := Fail(tc.code, "fail")
		if r.IsSuccess() {
			t.Error("expect fail result but got success")
		}
		if !r.IsFailure() {
			t.Error("expect fail result but got success")
		}
		if r.Error.Code != tc.code {
			t.Errorf("expect fail %v result but got %v", tc.code, r.Error.Code)
		}
	}
}
