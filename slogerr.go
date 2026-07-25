package slogerr

import (
	"errors"
	"log/slog"
)

type AnnotatedError struct {
	err   error
	attrs []slog.Attr
}

func New(err error, attrs ...slog.Attr) *AnnotatedError {
	return &AnnotatedError{err: err, attrs: attrs}
}

func (e *AnnotatedError) Error() string {
	return e.err.Error()
}

func (e *AnnotatedError) Unwrap() error {
	return e.err
}

func (e *AnnotatedError) Attrs() []slog.Attr {
	return e.allAttrs()
}

func (e *AnnotatedError) LogValue() slog.Value {
	return slog.GroupValue(e.allAttrs()...)
}

// allAttrs flattens attrs from the full chain of AnnotatedErrors.
func (e *AnnotatedError) allAttrs() []slog.Attr {
	var all []slog.Attr

	var current error = e

	for current != nil {
		if ae, ok := errors.AsType[*AnnotatedError](current); ok {
			all = append(all, ae.attrs...)
			current = ae.err
		} else {
			break
		}
	}

	return all
}
