package server

import (
	"log/slog"
	"net/http"

	"github.com/densestvoid/krogerrecipeshopper/templates"
)

type CommitResponseWriter struct {
	http.ResponseWriter

	statusCode int
	bs         []byte
	header     http.Header
}

func (w *CommitResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *CommitResponseWriter) Write(bs []byte) (int, error) {
	w.bs = append(w.bs, bs...)
	return len(bs), nil
}

func (w *CommitResponseWriter) Header() http.Header {
	return w.header
}

func (w *CommitResponseWriter) Abort() {
	w.statusCode = 0
	w.bs = nil
	w.header = nil
}

func (w *CommitResponseWriter) Commit() {
	// Must come before writing the status code
	writerHeader := w.ResponseWriter.Header()
	for key, values := range w.header {
		for _, value := range values {
			writerHeader.Add(key, value)
		}
	}
	// Must come before writing the body
	if w.statusCode != 0 {
		w.ResponseWriter.WriteHeader(w.statusCode)
	}
	// Write the body
	if _, err := w.ResponseWriter.Write(w.bs); err != nil {
		slog.Error("committing response writer", slog.String("error", err.Error()))
		http.Error(w.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
}

func NewSlogMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writer := CommitResponseWriter{ResponseWriter: w, header: http.Header{}}
			next.ServeHTTP(&writer, r)
			statusCode := writer.statusCode

			var logFunc func(string, ...any)
			switch {
			case 400 <= statusCode && statusCode <= 499:
				logFunc = slog.Warn
			case 500 <= statusCode && statusCode <= 599:
				logFunc = slog.Error
			default:
				logFunc = slog.Info
			}
			logFunc("HTTP request",
				slog.Int("status-code", statusCode),
				slog.String("method", r.Method),
				slog.String("url", r.URL.RequestURI()),
				slog.String("user-agent", r.UserAgent()),
			)

			writer.Commit()
		})
	}
}

func NewErrorStatusMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writer := CommitResponseWriter{ResponseWriter: w, header: http.Header{}}
			next.ServeHTTP(&writer, r)
			statusCode := writer.statusCode

			switch {
			case 400 <= statusCode && statusCode <= 599:
				defer writer.Abort()
				w.WriteHeader(statusCode)
				if r.Header.Get("HX-Request") != "" {
					slog.Error("sending htmx response", slog.String("error", string(writer.bs)))
					if err := templates.Alert("Failed", "alert-danger").Render(w); err != nil {
						slog.Error("writing error alert", slog.String("error", err.Error()))
						http.Error(w, err.Error(), statusCode)
						return
					}
				} else {
					slog.Error("sending page response", slog.String("error", string(writer.bs)))
					if err := templates.ErrorPage(statusCode, http.StatusText(statusCode)).Render(w); err != nil {
						slog.Error("writing error page", slog.String("error", err.Error()))
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}
				return
			default:
				writer.Commit()
				if r.Header.Get("HX-Request") != "" && len(writer.bs) == 0 {
					if err := templates.Alert("Success", "alert-success").Render(w); err != nil {
						slog.Error("writing error alert", slog.String("error", err.Error()))
						http.Error(w, err.Error(), statusCode)
						return
					}
				}
			}
		})
	}
}
