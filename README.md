# slogerr

Attach `slog.Attr` to an error and have them flatten across the wrap chain.

```go
import (
    "errors"
    "log/slog"

    "github.com/flamingoosesoftwareinc/slogerr"
)

err := slogerr.New(errors.New("not found"), slog.String("resource", "visitor"))
err = slogerr.New(err, slog.String("id", "abc123"))

slog.Error("op failed", "err", err)
// level=ERROR msg="op failed" err.message="not found" err.id=abc123 err.resource=visitor
```

`AnnotatedError` implements `error`, `Unwrap`, and `slog.LogValuer`, so a handler
renders the error message and accumulated attributes as a group without any call-site plumbing.
Attributes are collected from every `AnnotatedError` in the chain, outermost first.

The group's `message` field comes from `Error()`. Top-level annotations named
`message` are omitted from log output; `Attrs()` still returns them unchanged.
Use a named group for contextual message attributes, such as `context.message`.

```go
func (e *AnnotatedError) Error() string
func (e *AnnotatedError) Unwrap() error
func (e *AnnotatedError) Attrs() []slog.Attr
func (e *AnnotatedError) LogValue() slog.Value
```

## License

MPL-2.0
