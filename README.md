# slogerr

Attach `slog.Attr` to an error and have them flatten across the wrap chain.

```go
import "github.com/flamingoosesoftwareinc/slogerr"

err := slogerr.New(err, slog.String("resource", "visitor"))
err = slogerr.New(err, slog.String("id", "abc123"))

slog.Error("op failed", "err", err)
// level=ERROR msg="op failed" err.resource=visitor err.id=abc123
```

`AnnotatedError` implements `error`, `Unwrap`, and `slog.LogValuer`, so a handler
renders the accumulated attributes as a group without any call-site plumbing.
Attributes are collected from every `AnnotatedError` in the chain, outermost first.

```go
func (e *AnnotatedError) Error() string
func (e *AnnotatedError) Unwrap() error
func (e *AnnotatedError) Attrs() []slog.Attr
func (e *AnnotatedError) LogValue() slog.Value
```

## License

MPL-2.0
