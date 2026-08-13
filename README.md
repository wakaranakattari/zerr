# zerr

**Errors that don't suck.**

One allocation per wrap, machine-readable codes, structured logs,
and secrets that stay secret.

<p align="center">
<a href="https://go.dev"><img alt="Go Version" src="https://img.shields.io/github/go-mod/go-version/wakaranakattari/zerr?color=00ADD8&label=go&logo=go&logoColor=white"></a>
<a href="https://pkg.go.dev/github.com/wakaranakattari/zerr"><img alt="Go Reference" src="https://pkg.go.dev/badge/github.com/wakaranakattari/zerr.svg"></a>
<a href="https://github.com/wakaranakattari/zerr/actions"><img alt="Build" src="https://img.shields.io/github/actions/workflow/status/wakaranakattari/zerr/ci.yml?branch=main&label=build"></a>
<a href="LICENSE"><img alt="License" src="https://img.shields.io/github/license/wakaranakattari/zerr?color=blue"></a>
<a href="go.mod"><img alt="Dependencies" src="https://img.shields.io/badge/dependencies-none-brightgreen"></a>
<a href="https://goreportcard.com/report/github.com/wakaranakattari/zerr"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/wakaranakattari/zerr"></a>
<a href="https://github.com/wakaranakattari/zerr/stargazers"><img alt="Stars" src="https://img.shields.io/github/stars/wakaranakattari/zerr?color=gold&logo=star&logoColor=gold"></a>
</p>

```go
var ErrNotFound = zerr.Code("not_found")

func load(id int) (string, error) {
    v, err := db.Get(id)
    if err != nil {
        return "", zerr.WithCode(err, ErrNotFound, "load", "id", id)
    }
    return v, nil
}
```

<br>

<div align="center">

| | |
|---|---|
| **Performance** | `1 alloc` per wrap - faster than `fmt.Errorf` |
| **Dependencies** | none. stdlib only (`go 1.21+`) |
| **Interop** | `errors.Is` / `errors.As` / `errors.Join` out of the box |
| **Logging** | native `log/slog` via `LogValuer`, plus a `zap` submodule |
| **Debugging** | per-node stack traces with `-tags herr_trace` |
| **Safety** | private attributes can **never** leak into formatted output |
| **Boundaries** | `Public()` + `httperr`: clients see only what you choose |

</div>

<br>

## Why zerr?

Go's error handling is fine - until you need to know *what* failed and
*why*. Then you get three bad options:

- **`fmt.Errorf`** - context locked inside a string. Fast, but
  machines can't read it, and every wrap costs allocations.
- **Custom error types** - one type per error, a thousand `Is`/`As`
  dances, and the chain is only as good as the author's discipline.
- **`%w` everywhere** - works, but nothing carriers *structure*: no
  codes, no fields, no stacks, no privacy.

zerr keeps the idiomatic `if err != nil` shape of Go and gives errors
the plumbing they deserve:

- **One allocation per node** - like `errors.New`. Up to 2 attributes
  and 8 stack frames ride inside that same allocation.
- **Codes as sentinels** - `Code` *is* an `error`, so classification
  reads like stdlib: `errors.Is(err, ErrNotFound)`.
- **Attributes** - your context, machine-readable, delivered straight
  into structured logs.
- **Privacy** - `zerr.Sec(...)` marks secrets that only `Fields` can
  see. No more passwords in `%+v` dumps.
- **Exception-style flow** (`zerr/must`) - panic with the *real error*
  down a deep call stack, recover it whole at the boundary.
- **Stacks on demand** - production costs nothing; `-tags herr_trace`
  gives every node its own stack, still 1 alloc.

## Quick tour

```go
import "github.com/wakaranakattari/zerr"

err := zerr.Wrap(zerr.New("no rows"), "db", "table", "users")
err = zerr.Wrap(err, "load", "id", 42)

fmt.Println(err)          // load: db: no rows
fmt.Printf("%+v\n", err)  // load
                          //   id: 42
                          // db
                          //   table: users
                          // no rows
                          // (with herr_trace:   at store.go:132 loadUser)
```

### Classify with codes

Codes implement `error`, so membership works everywhere the stdlib
walks chains - including through joins:

```go
var (
    ErrNotFound  = zerr.Code("not_found")
    ErrForbidden = zerr.Code("forbidden")
)

err := zerr.Wrap(handle(), "api")

switch zerr.CodeOf(err) {        // the first code in the chain
case ErrNotFound:
    return 404
case ErrForbidden:
    return 403
}
// or match anywhere, deep or shallow:
if zerr.Is(err, ErrNotFound) { /* … */ }
```

### Attach attributes, keep secrets secret

```go
err := zerr.Wrap(err, "login",
    "user", user.ID,
    zerr.Sec("token", accessToken),  // private: invisible to fmt
)

fmt.Printf("%+v\n", err)   // user only
for _, f := range zerr.Fields(err) {  // the only way to see token
    slog.Any(f.Key, f.Value)
}
```

### Log it - structured, for free

```go
slog.Error("request failed", "err", err)
// err.op=load err.id=42 err.cause.op=db err.cause.table=users err.cause.cause.msg="no rows"
```

Any zerr error is a `slog.LogValuer`; foreign causes and multi-error
groups are nested with bounded depth.

### Zap too (`zerr/zap`)

The `zerr/zap` submodule encodes the same shape for uber-go/zap, with
the same three-level cause bound:

```go
import zerrzap "github.com/wakaranakattari/zerr/zap"

logger.Error("request failed", zerrzap.Field(err))
// "err":{"op":"load","id":42,"code":"not_found","cause":{"op":"db","table":"users"}}
```

### Hard boundaries: panic in, no clutter

```go
import "github.com/wakaranakattari/zerr/must"

func handle() (err error) {
    defer must.Catch(&err)          // recover the panic as a real error
    user := must.Must(fetchUser(uid)) // panics with the full chain
    _ = user
    return nil
}
```

### API boundaries: what the client sees

Failures have two faces. `Wrap`/`Wrapf` build the *internal* face:
operations, dsn strings, SQL text - everything the on-call engineer
wants. `Public` stamps the *client* face on top of exactly the same
chain:

```go
func store(id int) (*user, error) {
    internal := zerr.Wrapf(zerr.New("no such row"), "select", "id %d", id)
    return nil, zerr.Public(internal, "user not found")
}
```

- `(*Error).Public()` walks the chain and returns the outermost
  public message - the only text an API may return.
- Private attributes (`Sec`) and internals stay out of the message,
  the logs, and the wire.
- `httperr` turns the boundary into one line: codes pick the status,
  the public message is the body, the code rides in the header.

```go
import "github.com/wakaranakattari/zerr/httperr"

var ErrNotFound = zerr.Code("not_found")
// ... deep in the service:
return nil, zerr.WithCode(err, ErrNotFound, "load user", "id", id)
// ... at the handler:
if err != nil {
    httperr.Write(w, err)   // 404 · "user not found" · X-Error-Code: not_found
    return
}
```

Default statuses: `invalid_argument`→400, `unauthenticated`→401,
`forbidden`→403, `not_found`→404, `conflict`→409,
`too_many_requests`→429, `unavailable`→503, fallback 500. Extend
with `httperr.Map(code, status)`.

`%+v` renders the public message on its own line, marked for humans:
`public: user not found`. Logs never see it: the public message is
for clients, not for `slog`/`zap`.

### Resumé of independent failures

```go
import "github.com/wakaranakattari/zerr/join"

var err error
for _, f := range files {
    err = join.Append(err, read(f)) // nil-safe per call
}
if err != nil { /* errors.Is matches every member */ }
```

### Stack traces: free in prod, deep in debug

```sh
go build                    # zero overhead: stacks aren't even collected
go build -tags herr_trace   # every node records its own call site, 1 alloc
```

## Benchmarks

`go1.26 · linux/amd64 · production build · this machine`

| operation | zerr | stdlib |
|---|---|---|
| wrap an error | **~125 ns/op · 1 alloc** | `fmt.Errorf("%w")` ~163 ns/op · 2 allocs |
| wrap + 4 attributes | ~241 ns/op · 2 allocs | - |
| create root error | ~125 ns/op · 1 alloc | `errors.New` ~25 ns/op · 1 alloc |
| `errors.Is` over depth-10 chain | ~77 ns/op · 0 alloc | ~49 ns/op · 0 alloc |
| join two errors | **~82 ns/op · 1 alloc** | `errors.Join` ~44 ns/op · 1 alloc |

with `-tags herr_trace`: wrap ~500 ns/op, still **1 alloc** (8 frames inline).

Reproduce: `make bench` and `make bench-trace`.

### The competition, on the same wall clock

Head-to-head against the field, one machine, one run
(`make bench-cmp`). zerr's production build captures no stacks;
`oops` and `cockroachdb/errors` capture them at every node by
design, which is most of their cost:

| operation | zerr | `samber/oops` | `cockroachdb/errors` |
|---|---|---|---|
| create root error | **129 ns · 1 alloc** | 11.3 µs · 37 allocs | 834 ns · 5 allocs |
| wrap + 2 attributes | **166 ns · 1 alloc** | 11.3 µs · 38 allocs | 1.08 µs · 7 allocs |
| classify (10-deep chain) | **77 ns · 0 alloc**¹ | 249 ns · 1 alloc² | 118 ns · 0 alloc³ |

¹ `errors.Is(err, code)` matching a code at the bottom of a 10-node
chain. ² `oops.AsOops` - the only sound path, since `errors.Is` on an
`OopsError` panics on its embedded map in v0.19.x. ³ `errors.Is` on
the wrapped root value.

> The honest fine print: a zerr node is a 288-byte struct, so root
> creation is pricier than `errors.New` - you pay once per failure for
> context that would otherwise cost you two to three allocations and a
> hundred lines of plumbing. `Wrap` beats `fmt.Errorf` outright;
> `Join` trades ~2× of `errors.Join`'s time for a structured node with
> attributes intact; and the `Is` walk is ~1.6× of a plain
> `fmt.Errorf` chain. When you want stacks, produce them the way zerr
> does - gated behind a build tag - instead of paying for them by
> default.

## What it is not

- **Not a Result monad.** Go's `(T, error)` is fine; zerr makes the
  `error` half better instead of hiding it.
- **Not a replacement for `errors.Join`.** It *is* `errors.Join`
  semantics with the children preserved and attributes intact.
- **Not clever reflection magic.** Zero `reflect`, zero generated
  code, zero dependencies.
- **Not `oops` with a different name.** `oops` captures stack traces
  and spreads attributes across many allocations at every wrap.
  zerr's production build is one allocation per node and captures
  stacks only when you build with `-tags herr_trace`.

## Repository layout

```
├── doc.go / zerr.go        public facade: API + aliases
├── internal/core/          implementation: node, codes, attrs, format,
│   │                       slog, trace (herr_trace build tag)
│   └── *_test.go           unit · alloc · trace · bench tests
├── internal/multierror/    grouped-error node (behind join)
├── join/                   join.Join, join.Append
├── must/                   must.Must, must.MustErr, must.Catch
├── httperr/                HTTP boundaries: codes → statuses, public
│                           messages → bodies (stdlib only)
├── zap/                    submodule: zap.Field(err) for go.uber.org/zap
├── bench/                  submodule: head-to-head benchmarks vs
│                           samber/oops and cockroachdb/errors
├── examples/webapi/        HTTP handler: httperr + Public + slog
├── examples/stacktrace/    %v vs %+v with and without herr_trace
└── Makefile                make test / race / trace / bench / check
```
---

<p align="center"><b>Errors that don't suck.</b></p>
<p align="center"><a href="#zerr">Star it - one allocation, no drama.</a></p>