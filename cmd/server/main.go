package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	
	"gobilling/internal/customer"
	"gobilling/internal/event"
	"gobilling/internal/invoice"
	"gobilling/internal/ledger"
	"gobilling/internal/payment"
	"gobilling/internal/platform/cache"
	"gobilling/internal/platform/config"
	"gobilling/internal/platform/database"
	"gobilling/internal/platform/middleware"
	"gobilling/internal/product"
	"gobilling/internal/subscription"
	"gobilling/internal/webhook"
	"gobilling/internal/worker"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger := setupLogger(cfg)
	slog.SetDefault(logger)

	slog.Info("starting gobilling", "env", cfg.Env, "port", cfg.Port)

	db, err := database.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	slog.Info("connected to database")

	redisClient, err := cache.NewRedisClient(ctx, cfg.Redis)
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}
	defer redisClient.Close()
	slog.Info("connected to redis")

	clk := &RealClock{}

	customerRepo := customer.NewPostgresRepository(db)
	customerService := customer.NewService(customerRepo, db, clk)
	customerHandler := customer.NewHandler(customerService)

	productRepo := product.NewPostgresProductRepository(db)
	planRepo := product.NewPostgresPlanRepository(db)
	productService := product.NewService(productRepo, planRepo, db, clk)
	productHandler := product.NewHandler(productService)

	invoiceRepo := invoice.NewPostgresRepository(db)
	invoiceService := invoice.NewService(invoiceRepo, clk)
	invoiceHandler := invoice.NewHandler(invoiceService)

	ledgerRepo := ledger.NewPostgresRepository(db)
	ledgerService := ledger.NewService(ledgerRepo)

	paymentProvider := payment.NewStubProvider(cfg.Payment.SuccessRate)
	paymentRepo := payment.NewPostgresRepository(db)
	refundRepo := payment.NewPostgresRefundRepository(db)
	paymentService := payment.NewService(paymentRepo, refundRepo, invoiceRepo, paymentProvider, ledgerService, db, clk)
	paymentHandler := payment.NewHandler(paymentService)

	subscriptionRepo := subscription.NewPostgresRepository(db)
	subscriptionService := subscription.NewService(subscriptionRepo, customerRepo, planRepo, invoiceRepo, paymentProvider, db, clk)
	subscriptionHandler := subscription.NewHandler(subscriptionService)

	webhookEndpointRepo := webhook.NewPostgresEndpointRepository(db)
	webhookDeliveryRepo := webhook.NewPostgresDeliveryRepository(db)
	webhookService := webhook.NewService(webhookEndpointRepo, webhookDeliveryRepo)
	webhookHandler := webhook.NewHandler(webhookService)

	eventRepo := event.NewPostgresRepository(db)
	outboxWorker := worker.NewOutboxWorker(eventRepo, db)
	renewalWorker := worker.NewRenewalWorker(subscriptionRepo, planRepo, invoiceRepo, paymentRepo, paymentProvider, db)
	paymentRetryWorker := worker.NewPaymentRetryWorker(invoiceRepo, paymentRepo, paymentProvider, db)
	webhookDeliveryWorker := worker.NewWebhookDeliveryWorker(webhookDeliveryRepo, webhookEndpointRepo, eventRepo, db)

	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()

	go func() {
		if err := outboxWorker.Start(workerCtx); err != nil && err != context.Canceled {
			slog.Error("outbox worker error", "error", err)
		}
	}()

	go func() {
		if err := renewalWorker.Start(workerCtx); err != nil && err != context.Canceled {
			slog.Error("renewal worker error", "error", err)
		}
	}()

	go func() {
		if err := paymentRetryWorker.Start(workerCtx); err != nil && err != context.Canceled {
			slog.Error("payment retry worker error", "error", err)
		}
	}()

	go func() {
		if err := webhookDeliveryWorker.Start(workerCtx); err != nil && err != context.Canceled {
			slog.Error("webhook delivery worker error", "error", err)
		}
	}()

	r := chi.NewRouter()

	validationMw := middleware.NewValidationMiddleware()
	idempotencyMw := middleware.NewIdempotencyMiddleware(redisClient, db)

	r.Use(chi_middleware.RequestID)
	r.Use(chi_middleware.RealIP)
	r.Use(chi_middleware.Logger)
	r.Use(chi_middleware.Recoverer)
	r.Use(chi_middleware.Timeout(60 * time.Second))
	r.Use(validationMw.ValidateRequest)
	r.Use(idempotencyMw.Handle)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := db.Health(ctx); err != nil {
			http.Error(w, "database unhealthy", http.StatusServiceUnavailable)
			return
		}

		if err := redisClient.Health(ctx); err != nil {
			http.Error(w, "redis unhealthy", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Route("/v1", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("GoBilling API v1"))
		})
		
		customerHandler.RegisterRoutes(r)
		productHandler.RegisterRoutes(r)
		subscriptionHandler.RegisterRoutes(r)
		invoiceHandler.RegisterRoutes(r)
		paymentHandler.RegisterRoutes(r)
		webhookHandler.RegisterRoutes(r)
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("starting http server", "addr", srv.Addr)
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		slog.Info("shutdown signal received", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			srv.Close()
			return fmt.Errorf("could not gracefully shutdown server: %w", err)
		}

		slog.Info("server stopped gracefully")
	}

	return nil
}

func setupLogger(cfg *config.Config) *slog.Logger {
	var level slog.Level
	switch cfg.Logging.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if cfg.Logging.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now().UTC()
}
