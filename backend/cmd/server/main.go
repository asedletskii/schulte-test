package main

import (
	"context"
	"errors"
	"log/slog"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpadapter "schulte/internal/adapter/http"
	"schulte/internal/adapter/repository/memory"
	"schulte/internal/adapter/repository/postgres"
	"schulte/internal/domain"
	"schulte/internal/infrastructure/config"
	"schulte/internal/infrastructure/platform"
	"schulte/internal/infrastructure/token"
	"schulte/internal/usecase"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-healthcheck" {
		os.Exit(healthcheck())
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("сервер остановлен с ошибкой", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var (
		subjects usecase.SubjectRepository
		sessions usecase.SessionRepository
		settings usecase.SettingsRepository
	)
	switch cfg.Storage {
	case "memory":
		st := memory.NewStore()
		subjects, sessions, settings = st.Subjects(), st.Sessions(), st.Settings()
		log.Warn("используется хранилище в памяти: данные будут потеряны при остановке")
	default:
		st, err := postgres.Connect(ctx, cfg.DatabaseURL)
		if err != nil {
			return err
		}
		defer st.Close()
		subjects, sessions, settings = st.Subjects(), st.Sessions(), st.Settings()
	}

	ids, clock := platform.UUIDGenerator{}, platform.SystemClock{}
	tokens := token.New(cfg.TokenSecret, cfg.TokenTTL)
	settingsSvc := usecase.NewSettingsService(settings)
	tables := domain.NewTableGenerator(uint64(time.Now().UnixNano()), uint64(os.Getpid()))

	handler := httpadapter.NewHandler(httpadapter.Deps{
		Auth:     usecase.NewAuthService(cfg.AdminPassword, subjects, tokens),
		Subjects: usecase.NewSubjectService(subjects, ids, clock),
		Tests:    usecase.NewTestService(subjects, sessions, settingsSvc, tables, ids, clock),
		Results:  usecase.NewResultService(subjects, sessions),
		Settings: settingsSvc,
		Log:      log,
	})

	srv := &nethttp.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpadapter.NewRouter(handler, log, cfg.CORSOrigin),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("сервер запущен", "addr", cfg.HTTPAddr, "storage", cfg.Storage)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, nethttp.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
	return nil
}

func healthcheck() int {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	if addr[0] == ':' {
		addr = "127.0.0.1" + addr
	}
	client := nethttp.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://" + addr + "/api/health")
	if err != nil {
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != nethttp.StatusOK {
		return 1
	}
	return 0
}
