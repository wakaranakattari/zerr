package httperr_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wakaranakattari/zerr"
	"github.com/wakaranakattari/zerr/httperr"
)

func TestStatusByCode(t *testing.T) {
	cases := []struct {
		code zerr.Code
		want int
	}{
		{"invalid_argument", http.StatusBadRequest},
		{"unauthenticated", http.StatusUnauthorized},
		{"forbidden", http.StatusForbidden},
		{"not_found", http.StatusNotFound},
		{"conflict", http.StatusConflict},
		{"too_many_requests", http.StatusTooManyRequests},
		{"unavailable", http.StatusServiceUnavailable},
		{"unknown_code", http.StatusInternalServerError},
		{"", http.StatusInternalServerError},
	}
	for _, tc := range cases {
		err := zerr.WithCode(zerr.New("x"), tc.code, "op")
		if got := httperr.Status(err); got != tc.want {
			t.Errorf("Status(%q) = %d, want %d", tc.code, got, tc.want)
		}
	}
	if got := httperr.Status(nil); got != http.StatusOK {
		t.Errorf("Status(nil) = %d, want 200", got)
	}
}

func TestStatusDeepChain(t *testing.T) {
	err := zerr.Wrap(zerr.Wrap(zerr.WithCode(zerr.New("x"), zerr.Code("not_found"), "db"), "load"), "api")
	if got := httperr.Status(err); got != http.StatusNotFound {
		t.Errorf("Status through a deep chain = %d, want 404", got)
	}
}

func TestCustomMapping(t *testing.T) {
	httperr.Map("voucher_expired", http.StatusGone)
	defer delete(httperr.Default, "voucher_expired")
	err := zerr.WithCode(zerr.New("x"), zerr.Code("voucher_expired"), "op")
	if got := httperr.Status(err); got != http.StatusGone {
		t.Errorf("Status(custom code) = %d, want 410", got)
	}
}

func TestWritePublicMessage(t *testing.T) {
	err := zerr.WithCode(
		zerr.Public(zerr.Wrap(zerr.New("no rows"), "db", "table", "users"), "account unavailable"),
		zerr.Code("not_found"), "load",
	)
	rec := httptest.NewRecorder()
	httperr.Write(rec, err)

	if got, want := rec.Code, http.StatusNotFound; got != want {
		t.Errorf("status = %d, want %d", got, want)
	}
	if got, want := rec.Header().Get("X-Error-Code"), "not_found"; got != want {
		t.Errorf("X-Error-Code = %q, want %q", got, want)
	}
	if got, want := rec.Body.String(), "account unavailable"; got != want {
		t.Errorf("body = %q, want the public message %q", got, want)
	}
}

func TestWriteFallbackToChain(t *testing.T) {
	err := zerr.Wrap(zerr.New("no rows"), "db")
	rec := httptest.NewRecorder()
	httperr.Write(rec, err)

	if got, want := rec.Code, http.StatusInternalServerError; got != want {
		t.Errorf("status = %d, want %d", got, want)
	}
	if got, want := rec.Body.String(), "db: no rows"; got != want {
		t.Errorf("body = %q, want the plain chain %q", got, want)
	}
	if rec.Header().Get("X-Error-Code") != "" {
		t.Error("a chain without codes must not set X-Error-Code")
	}
}

func TestWriteLeaksNoInternalDetail(t *testing.T) {
	err := zerr.Wrap(zerr.Wrap(zerr.New("no rows"), "db", "dsn", "postgres://secret@internal:5432/prod"), "load",
		zerr.Sec("password", "hunter2"))
	rec := httptest.NewRecorder()
	httperr.Write(rec, err)

	body := rec.Body.String()
	for _, secret := range []string{"hunter2", "secret@internal", "dsn"} {
		if contains(body, secret) {
			t.Errorf("response %q leaks %q", body, secret)
		}
	}
}

func TestWriteNil(t *testing.T) {
	rec := httptest.NewRecorder()
	httperr.Write(rec, nil)
	if rec.Code != http.StatusOK {
		t.Errorf("Write(nil) = %d, want 200", rec.Code)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
