package mock

import (
	"context"
	"log/slog"
)

type noopHandler struct{}

func (n *noopHandler) Enabled(context.Context, slog.Level) bool {
	return false
}

func (*noopHandler) Handle(context.Context, slog.Record) error {
	return nil
}

func (n *noopHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return n
}

func (n *noopHandler) WithGroup(name string) slog.Handler {
	return n
}

func NewMockLogger() *slog.Logger {
	h := noopHandler{}
	return slog.New(&h)
}
