// Package httperr maps zerr error codes and public messages to HTTP
// responses.
//
// The idea: classification happens with zerr.WithCode deep in the
// service, the boundary stays a one-liner:
//
//	func handle(w http.ResponseWriter, r *http.Request) {
//	    v, err := load(r)
//	    if err != nil {
//	        httperr.Write(w, err)
//	        return
//	    }
//	    ...
//	}
//
// Write picks a status from the error's code (see Map for the default
// table), responds with the user-facing message (zerr.Public) or the
// plain chain text as a fallback, and stamps the code into the
// X-Error-Code header. Internal details -- operations, attributes,
// private fields -- never leave the server.
package httperr

import (
	"errors"
	"net/http"

	"github.com/wakaranakattari/zerr"
)

// Default code-to-status mapping. Codes outside this table fall back
// to StatusInternalServerError.
var Default = map[zerr.Code]int{
	"invalid_argument":  http.StatusBadRequest,
	"unauthenticated":   http.StatusUnauthorized,
	"forbidden":         http.StatusForbidden,
	"not_found":         http.StatusNotFound,
	"conflict":          http.StatusConflict,
	"rate_limited":      http.StatusTooManyRequests,
	"too_many_requests": http.StatusTooManyRequests,
	"unavailable":       http.StatusServiceUnavailable,
	"internal":          http.StatusInternalServerError,
}

// Map registers an extra code-to-status entry, mirroring Default.
// Registering a code twice overwrites the previous status.
func Map(code zerr.Code, status int) {
	Default[code] = status
}

// Status returns the HTTP status for err: the first registered code in
// the chain wins, StatusInternalServerError otherwise.
func Status(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if status, ok := Default[zerr.CodeOf(err)]; ok {
		return status
	}
	return http.StatusInternalServerError
}

// Write responds to a failed request. The body carries only what a
// client may see: the public message when present, otherwise the plain
// error chain. The code is echoed in the X-Error-Code header whenever
// the chain carries one.
func Write(w http.ResponseWriter, err error) {
	if err == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if code := zerr.CodeOf(err); code != "" {
		w.Header().Set("X-Error-Code", string(code))
	}
	w.WriteHeader(Status(err))
	if msg, ok := publicOf(err); ok {
		_, _ = w.Write([]byte(msg))
		return
	}
	_, _ = w.Write([]byte(err.Error()))
}

// publicOf extracts the user-facing message from err, reporting it as
// absent when the chain carries none.
func publicOf(err error) (string, bool) {
	var node *zerr.Error
	if !errors.As(err, &node) {
		return "", false
	}
	if msg := node.Public(); msg != "" {
		return msg, true
	}
	return "", false
}
