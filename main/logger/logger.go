package logger

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"time"
)

// L is the global structured logger. Available after Init is called.
var L *slog.Logger

// Init sets up the logger. If logToFile is true, logs are written to both
// stdout and logs/gosanime-YYYY-MM-DD.log. The returned function closes the
// log file and must be called before the process exits (use defer).
func Init(logToFile bool) (func(), error) {
	writers := []io.Writer{os.Stdout}
	var fileCloser func()

	if logToFile {
		if err := os.MkdirAll("logs", 0o755); err != nil {
			return func() {}, fmt.Errorf("creating logs dir: %w", err)
		}
		filename := fmt.Sprintf("logs/gosanime-%s.log", time.Now().Format("2006-01-02"))
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return func() {}, fmt.Errorf("opening log file %s: %w", filename, err)
		}
		writers = append(writers, f)
		fileCloser = func() { f.Close() }
	}

	w := io.MultiWriter(writers...)

	handler := slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	L = slog.New(handler)
	slog.SetDefault(L)

	// Redirect the standard log package to the same writer so existing
	// log.Fatal / log.Printf calls in third-party code also reach the file.
	log.SetOutput(w)
	log.SetFlags(0)

	if logToFile {
		L.Info("file logging enabled", "path", fmt.Sprintf("logs/gosanime-%s.log", time.Now().Format("2006-01-02")))
	}

	return func() {
		if fileCloser != nil {
			fileCloser()
		}
	}, nil
}

// Fatal logs an error message and exits with code 1.
func Fatal(msg string, args ...any) {
	L.Error(msg, args...)
	os.Exit(1)
}
