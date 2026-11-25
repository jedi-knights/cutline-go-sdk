package cutline

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/jedi-knights/go-platform/apperrors"
)

func apiErr(status int) *APIError {
	return &APIError{
		StatusCode: status,
		Message:    http.StatusText(status),
		Method:     http.MethodGet,
		URL:        "https://example.test/v1/foo",
	}
}

func TestIsNotFound(t *testing.T) {
	t.Parallel()
	if !IsNotFound(apiErr(http.StatusNotFound)) {
		t.Error("IsNotFound(404 APIError) = false")
	}
	if IsNotFound(apiErr(http.StatusOK)) {
		t.Error("IsNotFound(200 APIError) = true")
	}
	if IsNotFound(errors.New("plain error")) {
		t.Error("IsNotFound(plain error) = true")
	}
	if IsNotFound(nil) {
		t.Error("IsNotFound(nil) = true")
	}
	// Wrapping should still work through errors.As.
	wrapped := fmt.Errorf("while getting team: %w", apiErr(http.StatusNotFound))
	if !IsNotFound(wrapped) {
		t.Error("IsNotFound(wrapped 404) = false")
	}
}

func TestIsPaymentRequired(t *testing.T) {
	t.Parallel()
	if !IsPaymentRequired(apiErr(http.StatusPaymentRequired)) {
		t.Error("IsPaymentRequired(402) = false")
	}
	if IsPaymentRequired(apiErr(http.StatusForbidden)) {
		t.Error("IsPaymentRequired(403) = true")
	}
}

func TestIsOutOfScope(t *testing.T) {
	t.Parallel()
	if !IsOutOfScope(apiErr(http.StatusForbidden)) {
		t.Error("IsOutOfScope(403) = false")
	}
	if IsOutOfScope(apiErr(http.StatusPaymentRequired)) {
		t.Error("IsOutOfScope(402) = true")
	}
}

func TestIsRateLimited(t *testing.T) {
	t.Parallel()
	if !IsRateLimited(apiErr(http.StatusTooManyRequests)) {
		t.Error("IsRateLimited(429) = false")
	}
	if IsRateLimited(apiErr(http.StatusServiceUnavailable)) {
		t.Error("IsRateLimited(503) = true")
	}
}

func TestErrInvalidArgument_IsAndAs(t *testing.T) {
	t.Parallel()
	err := invalid("sport is required")
	if !errors.Is(err, ErrInvalidArgument) {
		t.Errorf("invalid() does not errors.Is ErrInvalidArgument: %v", err)
	}
	if !strings.Contains(err.Error(), "sport is required") {
		t.Errorf("invalid() message = %q, want contains format arg", err.Error())
	}
}

func TestErrInvalidResponse_Is(t *testing.T) {
	t.Parallel()
	err := invalidResponse("bad json: %v", errors.New("eof"))
	if !errors.Is(err, ErrInvalidResponse) {
		t.Errorf("invalidResponse() does not errors.Is ErrInvalidResponse: %v", err)
	}
	if !strings.Contains(err.Error(), "bad json") {
		t.Errorf("invalidResponse() message = %q", err.Error())
	}
}

func TestCodeForStatus(t *testing.T) {
	t.Parallel()
	cases := map[int]apperrors.ErrorCode{
		http.StatusNotFound:            apperrors.ErrCodeNotFound,
		http.StatusTooManyRequests:     apperrors.ErrCodeRateLimit,
		http.StatusUnauthorized:        apperrors.ErrCodeUnauthorized,
		http.StatusForbidden:           apperrors.ErrCodeForbidden,
		http.StatusConflict:            apperrors.ErrCodeConflict,
		http.StatusBadGateway:          apperrors.ErrCodeUnavailable,
		http.StatusServiceUnavailable:  apperrors.ErrCodeUnavailable,
		http.StatusGatewayTimeout:      apperrors.ErrCodeUnavailable,
		http.StatusBadRequest:          apperrors.ErrCodeBadRequest,
		http.StatusPaymentRequired:     apperrors.ErrCodeBadRequest, // uncovered 4xx bucket
		http.StatusInternalServerError: apperrors.ErrCodeInternal,
		599:                            apperrors.ErrCodeInternal, // uncovered 5xx bucket
	}
	for status, want := range cases {
		if got := codeForStatus(status); got != want {
			t.Errorf("codeForStatus(%d) = %q, want %q", status, got, want)
		}
	}
}

func TestAPIError_UnwrapsToAppError(t *testing.T) {
	t.Parallel()
	e := apiErr(http.StatusNotFound)
	if !apperrors.IsNotFound(e) {
		t.Errorf("apperrors.IsNotFound(APIError{404}) = false")
	}
	e = apiErr(http.StatusTooManyRequests)
	if !apperrors.IsRateLimit(e) {
		t.Errorf("apperrors.IsRateLimit(APIError{429}) = false")
	}
	e = apiErr(http.StatusForbidden)
	if !apperrors.IsForbidden(e) {
		t.Errorf("apperrors.IsForbidden(APIError{403}) = false")
	}
	e = apiErr(http.StatusServiceUnavailable)
	if !apperrors.IsUnavailable(e) {
		t.Errorf("apperrors.IsUnavailable(APIError{503}) = false")
	}
}

func TestAPIError_Error_UsesDetailWhenPresent(t *testing.T) {
	t.Parallel()
	e := &APIError{
		StatusCode: http.StatusBadRequest,
		Message:    "Bad Request",
		Detail:     `sport "basketball-u" is not recognised`,
		Method:     http.MethodGet,
		URL:        "https://example.test/v1/standings",
	}
	if !strings.Contains(e.Error(), "basketball-u") {
		t.Errorf("Error() = %q, want contains detail", e.Error())
	}
}

func TestAPIError_DecorateFromBody_Populates(t *testing.T) {
	t.Parallel()
	e := apiErr(http.StatusBadRequest)
	e.decorateFromBody([]byte(`{"error":"division is required","code":"BAD_REQUEST"}`))
	if e.Code != "BAD_REQUEST" {
		t.Errorf("Code = %q, want BAD_REQUEST", e.Code)
	}
	if e.Detail != "division is required" {
		t.Errorf("Detail = %q, want 'division is required'", e.Detail)
	}
}

func TestAPIError_DecorateFromBody_UnparseableIgnored(t *testing.T) {
	t.Parallel()
	e := apiErr(http.StatusBadRequest)
	e.decorateFromBody([]byte(`<html>Nginx 400</html>`))
	if e.Code != "" || e.Detail != "" {
		t.Errorf("Code=%q Detail=%q; both want empty on unparseable body", e.Code, e.Detail)
	}
}

func TestAPIError_DecorateFromBody_EmptyBodyIgnored(t *testing.T) {
	t.Parallel()
	e := apiErr(http.StatusBadRequest)
	e.decorateFromBody(nil)
	if e.Code != "" || e.Detail != "" {
		t.Errorf("Code=%q Detail=%q; both want empty on nil body", e.Code, e.Detail)
	}
}
