package slogerr_test

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	"github.com/flamingoosesoftwareinc/slogerr"
	"github.com/sebdah/goldie/v2"
)

func TestErrorChain(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("sentinel")
	inner := slogerr.New(sentinel, slog.String("a", "1"))
	outer := slogerr.New(inner, slog.String("b", "2"))

	if !errors.Is(outer, sentinel) {
		t.Error("errors.Is should find sentinel through chain")
	}
	if !errors.Is(outer, inner) {
		t.Error("errors.Is should find inner through chain")
	}
	if outer.Unwrap() != inner {
		t.Error("Unwrap should return immediate wrapped error")
	}
	if outer.Error() != "sentinel" {
		t.Errorf("Error() = %q, want %q", outer.Error(), "sentinel")
	}
}

func TestLogOutput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  *slogerr.AnnotatedError
	}{
		{
			name: "single_wrap",
			err: slogerr.New(errors.New("not found"),
				slog.String("resource", "visitor"),
				slog.String("id", "abc123"),
			),
		},
		{
			name: "stacked",
			err: slogerr.New(
				slogerr.New(errors.New("db error"),
					slog.String("table", "visitors"),
					slog.String("op", "SELECT"),
				),
				slog.String("tenant", "acme"),
			),
		},
		{
			name: "deep_chain",
			err: slogerr.New(
				slogerr.New(
					slogerr.New(errors.New("timeout"),
						slog.String("host", "db.local"),
					),
					slog.String("service", "auth"),
				),
				slog.String("region", "us-east-1"),
			),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := goldie.New(t)

			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
				ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
					if a.Key == slog.TimeKey {
						return slog.Attr{}
					}
					return a
				},
			}))

			logger.Error("op failed", "err", tc.err)
			g.Assert(t, tc.name, buf.Bytes())
		})
	}
}
