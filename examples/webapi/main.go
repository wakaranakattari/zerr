// Command webapi shows zerr working end to end in HTTP handlers:
// codes drive status-line classification, attributes become structured
// log fields, private attributes never leave the server, and only the
// public message reaches the client.
//
// Run with:
//
//	$ go run ./examples/webapi     # then: curl localhost:8080/users/42
//
//	# with per-node stack traces in the log:
//	$ go run -tags herr_trace ./examples/webapi
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/wakaranakattari/zerr"
	"github.com/wakaranakattari/zerr/httperr"
)

var (
	ErrNotFound = zerr.Code("not_found")
	ErrInvalid  = zerr.Code("invalid_argument")
)

type user struct {
	ID   int
	Name string
}

func main() {
	http.HandleFunc("/users/", handleUser)
	slog.Info("listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func handleUser(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Path[len("/users/"):]
	id, err := strconv.Atoi(raw)
	if err != nil {
		writeError(w, zerr.WithCode(err, ErrInvalid, "parse path", "path", raw))
		return
	}

	u, err := api(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	slog.Info("user served", "id", u.ID, "name", u.Name)
}

// api sits below the boundary: its contract is expressed in codes, so
// the handler can classify failures without inspecting internals.
func api(_ context.Context, id int) (*user, error) {
	u, err := store(id)
	if err != nil {
		return nil, zerr.WithCode(err, ErrNotFound, "load user", "id", id)
	}
	return u, nil
}

// store is the bottom of the chain, close to an actual DB call. The
// query text stays internal; clients never see it.
func store(id int) (*user, error) {
	if id != 42 {
		internal := zerr.Wrapf(zerr.New("no such row"), "select", "id %d", id)
		return nil, zerr.Public(internal, "user not found")
	}
	return &user{ID: id, Name: "zerr"}, nil
}

// writeError keeps the server side of the failure in the logs -- code,
// attributes, private fields, full chain -- and hands the client only
// the status line, the X-Error-Code header and the public message.
func writeError(w http.ResponseWriter, err error) {
	attrs := make([]slog.Attr, 0, 4)
	for _, f := range zerr.Fields(err) {
		attrs = append(attrs, slog.Any(f.Key, f.Value))
	}
	slog.Error("request failed",
		slog.String("code", string(zerr.CodeOf(err))),
		slog.Any("attrs", attrs),
	)

	httperr.Write(w, err)
}
