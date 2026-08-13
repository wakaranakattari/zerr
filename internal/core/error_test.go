package core_test

import (
	"errors"
	"fmt"
	"testing"

	zerr "github.com/wakaranakattari/zerr/internal/core"
	"github.com/wakaranakattari/zerr/join"
)

func joinAppend(err, next error) error { return join.Append(err, next) }

func TestWrapChain(t *testing.T) {
	root := zerr.New("boom")
	err := zerr.Wrap(root, "load")
	err = zerr.Wrap(err, "handle")

	if got, want := err.Error(), "handle: load: boom"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
	if !errors.Is(err, root) {
		t.Fatal("errors.Is did not reach the root through the chain")
	}
	var node *zerr.Error
	if !errors.As(err, &node) {
		t.Fatal("expected the chain to contain a *zerr.Error")
	}
	if zerr.CodeOf(err) != "" {
		t.Fatal("chain without codes must report an empty code")
	}
}

func TestWrapCauseIsPreserved(t *testing.T) {
	root := zerr.New("boom")
	err := zerr.Wrap(root, "load")
	unwrapped := errors.Unwrap(err)
	if !errors.Is(unwrapped, root) {
		t.Fatal("Unwrap did not return the wrapped cause")
	}
}

func TestWrapNilCause(t *testing.T) {
	if err := zerr.Wrap(nil, "op"); err != nil {
		t.Fatalf("Wrap(nil, ...) = %v, want nil", err)
	}
	if err := zerr.Wrapf(nil, "op", "msg"); err != nil {
		t.Fatalf("Wrapf(nil, ...) = %v, want nil", err)
	}
	if err := zerr.WithCode(nil, zerr.Code("x"), "op"); err != nil {
		t.Fatalf("WithCode(nil, ...) = %v, want nil", err)
	}
}

func TestWrapfMessage(t *testing.T) {
	err := zerr.Wrapf(zerr.New("root"), "load", "could not load item %d", 42)
	if got, want := err.Error(), "load: could not load item 42: root"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestNewfNoArgsDoesNotFormat(t *testing.T) {
	if err := zerr.Newf("plain"); err.Error() != "plain" {
		t.Fatalf("Newf without args = %q", err.Error())
	}
}

func TestNewDistinctNodes(t *testing.T) {
	a, b := zerr.New("same"), zerr.New("same")
	if a == b {
		t.Fatal("New must create a distinct node per call")
	}
}

func TestWrapVarargs(t *testing.T) {
	err := zerr.Wrap(zerr.New("root"), "op", "id", 1, "user", "alice")
	if got, want := fmt.Sprintf("%v", err), "op: root"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	fields := zerr.Fields(err)
	if len(fields) != 2 || fields[0].Key != "id" || fields[1].Key != "user" {
		t.Fatalf("Fields = %+v, want id and user", fields)
	}
}

func TestWrapOddKVPanics(t *testing.T) {
	var got any
	func() {
		defer func() { got = recover() }()
		zerr.Wrap(zerr.New("x"), "op", "key")
	}()
	if got == nil {
		t.Fatal("expected panic on an odd number of kv arguments")
	}
}

func TestWrapBadKeyPanics(t *testing.T) {
	var got any
	func() {
		defer func() { got = recover() }()
		zerr.Wrap(zerr.New("x"), "op", 42, "value")
	}()
	if got == nil {
		t.Fatal("expected panic on a non-string attribute key")
	}
}

func TestWithCodeAndMatching(t *testing.T) {
	const code = zerr.Code("not_found")
	err := zerr.WithCode(zerr.New("gone"), code, "load")
	if !errors.Is(err, code) {
		t.Fatal("errors.Is(err, code) = false, want true")
	}
	if !zerr.Is(err, code) {
		t.Fatal("zerr.Is(err, code) = false, want true")
	}
	if zerr.Is(err, zerr.Code("other")) {
		t.Fatal("zerr.Is with a different code must be false")
	}
	if got := zerr.CodeOf(err); got != code {
		t.Fatalf("CodeOf = %q, want %q", got, code)
	}
}

func TestCodeMatchingThroughChain(t *testing.T) {
	const code = zerr.Code("not_found")
	root := zerr.WithCode(zerr.New("gone"), code, "store")
	err := zerr.Wrap(root, "api")
	err = zerr.Wrap(err, "handler")
	if !errors.Is(err, code) {
		t.Fatal("errors.Is must match a code at the bottom of a deep chain")
	}
	if got := zerr.CodeOf(err); got != code {
		t.Fatalf("CodeOf = %q, want %q", got, code)
	}
}

func TestCodeHierarchy(t *testing.T) {
	const (
		ioCode        = zerr.Code("io")
		ioTimeoutCode = zerr.Code("io.timeout")
		ioDeadline    = zerr.Code("io.deadline")
		otherIO       = zerr.Code("ioops") // sibling char-wise, different family
	)
	err := zerr.WithCode(zerr.New("slow"), ioTimeoutCode, "net")
	err = zerr.Wrap(err, "api")

	if !errors.Is(err, ioCode) {
		t.Fatal("a child code must match its family sentinel through the chain")
	}
	if !errors.Is(err, ioTimeoutCode) {
		t.Fatal("exact code matching must keep working")
	}
	if errors.Is(err, otherIO) {
		t.Fatal("Code(\"ioops\") must not match the io family: prefix is dot-bounded")
	}
	if got := zerr.CodeOf(err); got != ioTimeoutCode {
		t.Fatalf("CodeOf = %q, want the exact child code %q", got, ioTimeoutCode)
	}

	parent := zerr.WithCode(zerr.New("slow"), ioCode, "net")
	if errors.Is(parent, ioTimeoutCode) {
		t.Fatal("a parent code must not match its child sentinel")
	}

	childOnTop := zerr.WithCode(zerr.Wrap(zerr.New("slow"), "net"), ioTimeoutCode, "api")
	if !errors.Is(childOnTop, ioCode) {
		t.Fatal("the outermost node's code must also match the family")
	}
}

func TestPublicMessage(t *testing.T) {
	// End-to-end: an exotic chain with two public layers; the outermost
	// must win, internal text must stay out of Public().
	root := zerr.Wrap(zerr.New("no rows"), "db", "table", "users")
	internal := zerr.Wrapf(root, "list", "filter %q rejected", "x'")
	withPub := zerr.Public(internal, "the account could not be loaded")

	if got := internal.(*zerr.Error).Public(); got != "" {
		t.Fatalf("a chain without a public message must report %q, got %q", "", got)
	}
	if got, want := withPub.(*zerr.Error).Public(), "the account could not be loaded"; got != want {
		t.Fatalf("Public() = %q, want %q", got, want)
	}
	if got := zerr.Public(zerr.Public(root, "outer"), "inner").(*zerr.Error).Public(); got != "inner" {
		t.Fatalf("Public() must return the outermost message, got %q", got)
	}
	if zerr.Public(nil, "msg") != nil {
		t.Fatal("Public(nil, ...) must return nil")
	}
	if zerr.Public(withPub, "") == nil {
		t.Fatal("Public with an empty message must still wrap")
	}
}

func TestNodes(t *testing.T) {
	err := zerr.WithCode(
		zerr.Public(zerr.Wrap(zerr.New("boom"), "db", "table", "users"), "ph"),
		zerr.Code("not_found"), "load", "id", 7,
	)
	nodes := zerr.Nodes(err)
	if got, want := len(nodes), 4; got != want {
		t.Fatalf("Nodes() = %d entries, want %d", got, want)
	}
	outer := nodes[0]
	if outer.Op != "load" || outer.Code != zerr.Code("not_found") || outer.Pub != "" {
		t.Fatalf("outer node = %+v, want op=load code=not_found pub empty", outer)
	}
	if len(outer.Attrs) != 1 || outer.Attrs[0].Key != "id" || outer.Attrs[0].Value != 7 {
		t.Fatalf("outer attrs = %+v, want [id:7]", outer.Attrs)
	}
	if nodes[1].Op != "" || nodes[1].Pub != "ph" {
		t.Fatalf("middle node = %+v, want pub=ph", nodes[1])
	}
	if nodes[2].Op != "db" || nodes[2].Msg != "" || len(nodes[2].Attrs) != 1 {
		t.Fatalf("wrap node = %+v, want op=db with attrs", nodes[2])
	}
	if nodes[3].Msg != "boom" {
		t.Fatalf("root node = %+v, want msg=boom", nodes[3])
	}
}

func TestNodesForeignAndGroup(t *testing.T) {
	foreign := fmt.Errorf("plain %w", errors.New("cause"))
	nodes := zerr.Nodes(foreign)
	if len(nodes) != 1 || nodes[0].Msg != "plain cause" {
		t.Fatalf("foreign chain = %+v, want one message-only node", nodes)
	}
	if zerr.Nodes(nil) != nil {
		t.Fatal("Nodes(nil) must be nil")
	}

	group := joinAppend(joinAppend(nil, zerr.New("a")), zerr.New("b"))
	if g := zerr.Nodes(group); len(g) != 1 {
		t.Fatalf("group Nodes = %+v, want one collapsed node", g)
	}
}

func TestStdlibAsFindsNode(t *testing.T) {
	err := zerr.Wrap(zerr.New("boom"), "load")
	var node *zerr.Error
	if !errors.As(err, &node) {
		t.Fatal("errors.As did not find *zerr.Error")
	}
	if node.Error() != "load: boom" {
		t.Fatalf("errors.As returned the wrong node: %q", node.Error())
	}
}
