package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"noytech-ga-optimizer/internal/handler"
	"noytech-ga-optimizer/internal/services/importer"
	"noytech-ga-optimizer/internal/services/optimizer"
	storages "noytech-ga-optimizer/internal/storages"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		start := time.Now()

		requestID := time.Now().Format("20060102150405") + "-" + r.RemoteAddr
		ctx = context.WithValue(ctx, RequestIDKey, requestID)

		log := logger.With(
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("url", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
		)
		log.Info("Request started")

		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapper, r.WithContext(ctx))

		duration := time.Since(start)
		log.Info("Request finished",
			slog.Int("status", wrapper.statusCode),
			slog.Duration("duration", duration),
		)
	})
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		logger.Warn("Failed to load .env file", "error", err)
	}

	dbURL := os.Getenv("PG_DSN")
	if dbURL == "" {
		logger.Error("PG_DSN environment variable is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	logger.Info("Connected to PostgreSQL")

	store := storages.NewPostgresStorage(pool)
	importerSvc := importer.New(store, logger)
	optimizerSvc := optimizer.New(store, logger)

	uploadHandler := handler.NewUploadHandler(importerSvc, logger)
	optimizeHandler := handler.NewOptimizeHandler(optimizerSvc, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /upload", uploadHandler.HandleUpload)
	mux.HandleFunc("POST /optimize", optimizeHandler.HandleOptimize)

	finalHandler := loggingMiddleware(mux, logger)

	server := &http.Server{
		Addr:    ":8080",
		Handler: finalHandler,
	}

	logger.Info("Starting HTTP server", "addr", server.Addr)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server stopped gracefully")
}
