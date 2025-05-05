package loghandler

import (
	"context"
	"log/slog"
	"os"
)

type CustomHandler struct {
	slog.Handler
	logFile *os.File
}

func (h *CustomHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *CustomHandler) Handle(_ context.Context, r slog.Record) error {
	_, err := h.logFile.WriteString(r.Message + "\n")
	return err
}

func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *CustomHandler) WithGroup(name string) slog.Handler {
	return h
}
