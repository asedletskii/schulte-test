package http

import (
	"context"
	"log/slog"
	nethttp "net/http"
	"strings"
	"time"

	"schulte/internal/domain"
	"schulte/internal/usecase"
)

type ctxKey struct{}

// principalFrom извлекает пользователя, помещённого middleware аутентификации.
func principalFrom(ctx context.Context) usecase.Principal {
	p, _ := ctx.Value(ctxKey{}).(usecase.Principal)
	return p
}

// authenticated оборачивает обработчик, требующий входа в систему.
func (h *Handler) authenticated(next nethttp.HandlerFunc) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || token == "" {
			h.writeError(w, domain.ErrUnauthorized)
			return
		}
		p, err := h.auth.Authenticate(token)
		if err != nil {
			h.writeError(w, err)
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, p)))
	}
}

// cors разрешает запросы с указанного источника (режим разработки фронтенда).
func cors(origin string, next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == nethttp.MethodOptions {
			w.WriteHeader(nethttp.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	nethttp.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) { s.status = code; s.ResponseWriter.WriteHeader(code) }

// logging пишет в журнал метод, путь, код ответа и длительность запроса.
func logging(log *slog.Logger, next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: nethttp.StatusOK}
		next.ServeHTTP(rec, r)
		log.Info("http", "method", r.Method, "path", r.URL.Path, "status", rec.status, "dur", time.Since(start).String())
	})
}
