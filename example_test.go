//go:build !herr_trace

package zerr_test

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/wakaranakattari/zerr"
	"github.com/wakaranakattari/zerr/join"
	"github.com/wakaranakattari/zerr/must"
)

func ExampleWrap() {
	root := zerr.New("no rows")
	err := zerr.Wrap(root, "db", "table", "users")
	err = zerr.Wrap(err, "load", "id", 42)

	fmt.Println(err)
	fmt.Printf("%+v\n", err)
	// Output:
	// load: db: no rows
	// load
	//   id: 42
	// db
	//   table: users
	// no rows
}

func Example_code() {
	var ErrNotFound = zerr.Code("not_found")

	err := zerr.WithCode(zerr.New("gone"), ErrNotFound, "load")
	if zerr.Is(err, ErrNotFound) {
		fmt.Println("classified as not_found")
	}

	err = zerr.Wrap(err, "handler")
	if errors.Is(err, ErrNotFound) {
		fmt.Println("still classified through the chain")
	}
	// Output:
	// classified as not_found
	// still classified through the chain
}

func ExampleSec() {
	err := zerr.Wrap(zerr.New("rejected"), "login",
		"user", "alice",
		zerr.Sec("token", "secret"),
	)

	fmt.Printf("%+v\n", err)
	fmt.Println(zerr.Fields(err)[1])
	// Output:
	// login
	//   user: alice
	// rejected
	// {token secret true}
}

func ExampleJoin() {
	a := zerr.New("validation failed")
	b := zerr.New("commit failed")
	err := join.Join(a, b)

	fmt.Println(err)
	// Output:
	// validation failed; commit failed
}

func ExampleCatch() {
	const codeParse = zerr.Code("parse")
	parse := func(s string) (int, error) {
		var n int
		_, e := fmt.Sscanf(s, "%d", &n)
		if e != nil {
			return 0, zerr.WithCode(e, codeParse, "parse", "input", s)
		}
		return n, nil
	}

	load := func() (err error) {
		defer must.Catch(&err)
		n := must.Must(parse("abc")) // panics with the wrapped error
		_ = n
		return nil
	}

	err := load()
	fmt.Println(zerr.Is(err, codeParse))
	fmt.Println(err)
	// Output:
	// true
	// parse: expected integer
}

func Example_fieldsAsLogAttrs() {
	err := zerr.Wrap(zerr.New("boom"), "load", "id", 7)
	attrs := make([]slog.Attr, 0, 4)
	for _, f := range zerr.Fields(err) {
		attrs = append(attrs, slog.Any(f.Key, f.Value))
	}
	_ = attrs // delivered to the log transport of your choice
	fmt.Println("fields:", len(attrs))
	// Output:
	// fields: 1
}
