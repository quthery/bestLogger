package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/dusted-go/logging/prettylog"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	// Создаем логгер
	logger := NewLogger(slog.LevelInfo, true)

	// Добавляем middleware для логирования
	r.Use(NewLoggingMiddleware(logger))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	http.ListenAndServe(":8080", r)
}

// NewLogger создает новый логгер с красивым выводом
func NewLogger(level slog.Level, addSource bool) *slog.Logger {
	prettyHandler := prettylog.NewHandler(&slog.HandlerOptions{
		Level:       level,
		AddSource:   addSource,
		ReplaceAttr: nil,
	})
	logger := slog.New(prettyHandler)
	return logger
}

// NewLoggingMiddleware возвращает middleware для логирования запросов и ответов
func NewLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Создаем обертку для записи статуса
			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Обработка запроса
			next.ServeHTTP(ww, r)

			// Логируем информацию о запросе
			logger.Info("HTTP Request",
				slog.String("method", r.Method),
				slog.String("uri", r.RequestURI),
				slog.Int("status", ww.statusCode),
				slog.String("duration",
					fmt.Sprint(float64(float64(time.Since(start))/float64(time.Microsecond)))+"µs"),
			)
		})
	}
}

// responseWriter обертка для ResponseWriter, чтобы захватывать статус
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader переопределяет метод WriteHeader для захвата статуса
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
