package slogerr_test

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/flamingoosesoftwareinc/slogerr"
	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/require"
)

func TestLogValue(t *testing.T) {
	root := errors.New("disk unavailable")
	inner := slogerr.New(root, slog.String("host", "db"))
	wrapped := fmt.Errorf("query failed: %w", inner)
	cases := map[string]struct {
		err     *slogerr.AnnotatedError
		message string
		attrs   []slog.Attr
		unwrap  error
	}{
		"empty": {slogerr.New(root), "disk unavailable", nil, root},
		"single": {
			slogerr.New(root, slog.String("host", "db")),
			"disk unavailable",
			[]slog.Attr{slog.String("host", "db")},
			root,
		},
		"nested": {
			slogerr.New(inner, slog.String("operation", "read")),
			"disk unavailable",
			[]slog.Attr{slog.String("operation", "read"), slog.String("host", "db")},
			inner,
		},
		"wrapped": {
			slogerr.New(wrapped, slog.String("operation", "read")),
			"query failed: disk unavailable",
			[]slog.Attr{slog.String("operation", "read"), slog.String("host", "db")},
			wrapped,
		},
		"collision": {
			slogerr.New(
				root,
				slog.String("message", "spoof"),
				slog.Group("context", slog.String("message", "detail")),
			),
			"disk unavailable",
			[]slog.Attr{
				slog.String("message", "spoof"),
				slog.Group("context", slog.String("message", "detail")),
			},
			root,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.message, tc.err.Error())
			require.Equal(t, tc.attrs, tc.err.Attrs())
			require.Same(t, tc.unwrap, tc.err.Unwrap())
			require.ErrorIs(t, tc.err, root)
			var annotated *slogerr.AnnotatedError
			require.ErrorAs(t, tc.err, &annotated)
			require.Same(t, tc.err, annotated)
			for _, format := range []string{"json", "text"} {
				t.Run(format, func(t *testing.T) {
					var output bytes.Buffer
					options := &slog.HandlerOptions{
						ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
							if attr.Key == slog.TimeKey {
								return slog.Attr{}
							}
							return attr
						},
					}
					var handler slog.Handler
					if format == "json" {
						handler = slog.NewJSONHandler(&output, options)
					} else {
						handler = slog.NewTextHandler(&output, options)
					}
					slog.New(handler).Error("request failed", "error", tc.err)
					if name == "collision" && format == "json" {
						require.JSONEq(
							t,
							`{"level":"ERROR","msg":"request failed","error":{"message":"disk unavailable","context":{"message":"detail"}}}`,
							output.String(),
						)
					}
					if name == "collision" && format == "text" {
						require.Equal(
							t,
							"level=ERROR msg=\"request failed\" error.message=\"disk unavailable\" error.context.message=detail\n",
							output.String(),
						)
					}
					t.Logf("actual %s handler output: %s", format, output.String())
					goldie.New(t).Assert(t, "logvalue_"+name+"_"+format, output.Bytes())
				})
			}
		})
	}
}
