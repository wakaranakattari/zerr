// Command stacktrace demonstrates the rendering of wrapped chains with
// %v and %+v, plus the difference the herr_trace build tag makes.
//
// The trace build must be used to see per-node stacks:
//
//	$ go run ./examples/stacktrace
//	$ go run -tags herr_trace ./examples/stacktrace
package main

import (
	"fmt"

	"github.com/wakaranakattari/zerr"
)

func main() {
	err := zerr.WithCode(
		zerr.Wrapf(zerr.New("no such row"), "select", "from users where id = %d", 42),
		zerr.Code("not_found"), "load user",
		"id", 42, "table", "users",
	)

	fmt.Println("compact: ", err)
	fmt.Println()
	fmt.Println("verbose:")
	fmt.Printf("%+v\n", err)
}
