package cutline

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/jedi-knights/go-platform/apperrors"
)

// ErrInvalidArgument is wrapped by every error returned for a bad argument
// (an empty sport or division, a negative page or limit). The SDK rejects
// these before making any request. Test for it with errors.Is.
var ErrInvalidArgument error = apperrors.New(apperrors.ErrCodeBadRequest, "cutline: invalid argument")

// ErrInvalidResponse is wrapped by the error returned when The Cutline answers
// 200 but the body is not the JSON the SDK expected, for example an HTML error
// page from a misconfigured proxy. It is not retried.
var ErrInvalidResponse error = apperrors.New(apperrors.ErrCodeUnavailable, "cutline: invalid response")

// APIError is returned when The Cutline answers with a non-2xx status.
// Inspect it with errors.As:
//
//	var apiErr *cutline.APIError
//	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusTooManyRequests { ... }
//
// The helper predicates (IsNotFound, IsOutOfScope, IsPaymentRequired,
// IsRateLimited) cover the common branches.
type APIError struct {
	// StatusCode is the HTTP status of the final response.
	StatusCode int
	// Message is the status text, for example "Not Found".
	Message string
	// Code is the Cutline-side error code ("BAD_REQUEST", "OUT_OF_SCOPE",
	// "RATE_LIMITED", …) decoded from the JSON body, when present.
	Code string
	// Detail is the Cutline-side error message decoded from the JSON body.
	Detail string
	// RequestID identifies the request for support correspondence.
	RequestID string
	// Method and URL describe the request that failed.
	Method string
	URL    string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	msg := e.Message
	if e.Detail != "" {
		msg = e.Detail
	}
	return fmt.Sprintf("cutline: %s %s: %d %s [request_id=%s]", e.Method, e.URL, e.StatusCode, msg, e.RequestID)
}

// Unwrap exposes a go-platform *apperrors.AppError whose code is derived from
// the HTTP status (404 NOT_FOUND, 429 RATE_LIMIT, 401 UNAUTHORIZED, 402
// PAYMENT_REQUIRED, 403 FORBIDDEN, 409 CONFLICT, other 4xx BAD_REQUEST,
// 502/503/504 UNAVAILABLE, other 5xx INTERNAL). It lets code built on
// go-platform use apperrors.IsNotFound(err) on SDK errors directly. The
// *APIError itself stays reachable with errors.As.
func (e *APIError) Unwrap() error {
	return apperrors.New(codeForStatus(e.StatusCode), e.Message)
}

// decorateFromBody parses the Cutline error envelope ({error, code}) from
// body, if present, and populates Code and Detail. Unparseable bodies are
// ignored — Message and StatusCode still describe the failure.
func (e *APIError) decorateFromBody(body []byte) {
	if len(body) == 0 {
		return
	}
	var env struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return
	}
	e.Code = env.Code
	e.Detail = env.Error
}

// IsNotFound reports whether err is an APIError with HTTP 404.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errorsAsAPI(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// IsPaymentRequired reports whether err is an APIError with HTTP 402. The
// caller's plan has lapsed or the account has no active subscription.
func IsPaymentRequired(err error) bool {
	var apiErr *APIError
	return errorsAsAPI(err, &apiErr) && apiErr.StatusCode == http.StatusPaymentRequired
}

// IsOutOfScope reports whether err is an APIError with HTTP 403. The caller's
// plan does not cover the requested sport, division, endpoint, or feature.
func IsOutOfScope(err error) bool {
	var apiErr *APIError
	return errorsAsAPI(err, &apiErr) && apiErr.StatusCode == http.StatusForbidden
}

// IsRateLimited reports whether err is an APIError with HTTP 429. The caller
// has exceeded either the per-minute rate limit or (on Free) the monthly
// quota. Inspect Retry-After via subsequent requests.
func IsRateLimited(err error) bool {
	var apiErr *APIError
	return errorsAsAPI(err, &apiErr) && apiErr.StatusCode == http.StatusTooManyRequests
}

// codeForStatus maps HTTP statuses to apperrors codes. The uncovered buckets
// are BAD_REQUEST for other 4xx and INTERNAL for other 5xx.
func codeForStatus(status int) apperrors.ErrorCode {
	switch status {
	case http.StatusNotFound:
		return apperrors.ErrCodeNotFound
	case http.StatusTooManyRequests:
		return apperrors.ErrCodeRateLimit
	case http.StatusUnauthorized:
		return apperrors.ErrCodeUnauthorized
	case http.StatusForbidden:
		return apperrors.ErrCodeForbidden
	case http.StatusConflict:
		return apperrors.ErrCodeConflict
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return apperrors.ErrCodeUnavailable
	}
	if status >= 400 && status < 500 {
		return apperrors.ErrCodeBadRequest
	}
	return apperrors.ErrCodeInternal
}

// invalid builds an ErrInvalidArgument-wrapping error with a formatted message.
func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{ErrInvalidArgument}, args...)...)
}

// invalidResponse builds an ErrInvalidResponse-wrapping error.
func invalidResponse(format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{ErrInvalidResponse}, args...)...)
}

// errorsAsAPI is a thin wrapper over errors.As for *APIError that keeps the
// predicate callers above tidy.
func errorsAsAPI(err error, target **APIError) bool {
	return errors.As(err, target)
}
