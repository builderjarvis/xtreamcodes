package xtreamcodes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

// ErrResponseTooLarge indicates that a response exceeded the configured byte limit.
var ErrResponseTooLarge = errors.New("response exceeds configured byte limit")

// safeCause retains useful sentinel identities without retaining transport URLs,
// provider payloads or arbitrary error text that may contain credentials.
func safeCause(err error) error {
	for _, sentinel := range []error{context.Canceled, context.DeadlineExceeded, ErrResponseTooLarge, io.ErrUnexpectedEOF, io.EOF} {
		if errors.Is(err, sentinel) {
			return sentinel
		}
	}
	var networkError net.Error
	if errors.As(err, &networkError) && networkError.Timeout() {
		return context.DeadlineExceeded
	}
	return errors.New("request or response processing failed (details omitted to protect credentials)")
}

// HTTPError represents an error returned by the Xtream Codes API
type HTTPError struct {
	URL        string
	StatusCode int
	// Body is retained for source compatibility. Client methods never populate it
	// because server error bodies can echo credentials or other private data.
	Body *string
}

// Error implements the error interface
func (e *HTTPError) Error() string {
	return fmt.Sprintf(
		"HTTP request to %s failed with status code: %d %s",
		redactURL(e.URL),
		e.StatusCode,
		http.StatusText(e.StatusCode),
	)
}

// DecoderError represents an error that occurs while decoding API responses.
type DecoderError struct {
	URL                    string
	UnderlyingDecoderError error
}

// Error implements the error interface
func (e *DecoderError) Error() string {
	return fmt.Sprintf("failed to decode response from %s: %v", redactURL(e.URL), safeCause(e.UnderlyingDecoderError))
}

// Unwrap implements the Wrapper interface
func (e *DecoderError) Unwrap() error {
	return e.UnderlyingDecoderError
}
