package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/database"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/handlers"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/incidents"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/metrics"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/middleware"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/repository"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/worker"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type appDependencies struct {
	pool         *pgxpool.Pool
	redisClient  *redis.Client
	workerClient *worker.Client
}

func initDependencies(ctx context.Context, cfg runtimeConfig) (appDependencies, func(), error) {
	deps := appDependencies{}
	cleanup := func() {
		if deps.pool != nil {
			deps.pool.Close()
		}
		if deps.redisClient != nil {
			_ = deps.redisClient.Close()
		}
		if deps.workerClient != nil {
			_ = deps.workerClient.Close()
		}
	}

	pool, err := initPostgres(ctx, cfg)
	if err != nil {
		cleanup()
		return deps, func() {}, err
	}
	deps.pool = pool

	redisClient, err := initRedis(ctx, cfg)
	if err != nil {
		cleanup()
		return deps, func() {}, err
	}
	deps.redisClient = redisClient

	workerClient, err := initWorker(ctx, cfg)
	if err != nil {
		cleanup()
		return deps, func() {}, err
	}
	deps.workerClient = workerClient

	return deps, cleanup, nil
}

func initPostgres(ctx context.Context, cfg runtimeConfig) (*pgxpool.Pool, error) {
	pool, err := database.PostgresPool(ctx)
	if err != nil {
		return nil, handleDependencyInitError(cfg.StrictDependencies, "postgres init", err)
	}

	if !cfg.AutoMigrate {
		return pool, nil
	}

	migrateCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := database.ApplyMigrations(migrateCtx); err != nil {
		return pool, handleDependencyInitError(cfg.StrictDependencies, "postgres migration", err)
	}

	return pool, nil
}

func initRedis(ctx context.Context, cfg runtimeConfig) (*redis.Client, error) {
	redisClient, err := database.RedisClient(ctx)
	if err != nil {
		return nil, handleDependencyInitError(cfg.StrictDependencies, "redis init", err)
	}
	return redisClient, nil
}

func initWorker(ctx context.Context, cfg runtimeConfig) (*worker.Client, error) {
	workerClient, err := worker.NewClient(ctx)
	if err != nil {
		if warnErr := handleDependencyInitError(cfg.StrictDependencies, "worker grpc init", err); warnErr != nil {
			return nil, warnErr
		}
		return nil, nil
	}
	return workerClient, nil
}

func handleDependencyInitError(strict bool, stage string, err error) error {
	if err == nil {
		return nil
	}
	if strict {
		return fmt.Errorf("%s failed: %w", stage, err)
	}
	log.Printf("%s warning (strict=false): %v", stage, err)
	return nil
}

func buildHTTPServer(cfg runtimeConfig, deps appDependencies) (*http.Server, func()) {
	r, cleanup := buildRouter(cfg, deps)
	addr := ":" + cfg.APIGatewayPort
	return &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}, cleanup
}

func buildRouter(cfg runtimeConfig, deps appDependencies) (http.Handler, func()) {
	userRepo := repository.NewUserRepository(deps.pool)
	usersHandler := handlers.NewUsersHandler(userRepo)
	computeHandler := handlers.NewComputeHandler(deps.workerClient)
	readyEndpoints := handlers.NewReadyEndpoints(deps.workerClient)
	metricsStore := metrics.NewStore()
	runtimeConfigSummary := diagnosticsSummary(cfg)
	runtimeConfigHandler := handlers.RuntimeConfigHandler(runtimeConfigSummary)
	runtimeMetricsOptions := handlers.RuntimeMetricsOptions{
		ReadyEndpoints:               readyEndpoints,
		LoggerEndpoint:               cfg.LoggerEndpoint,
		RequestLoggerStatsProvider:   middleware.CurrentAsyncLogSenderStats,
		IncidentEmitterStatsProvider: incidents.CurrentEmitterStats,
	}
	runtimeDiagnosticsCache := handlers.NewRuntimeDiagnosticsCache(
		runtimeConfigSummary,
		metricsStore,
		runtimeMetricsOptions,
		cfg.RuntimeDiagnosticsCacheTTL,
	)
	runtimeMetricsHandler := handlers.RuntimeMetricsHandlerWithCache(runtimeDiagnosticsCache)
	runtimeReportHandler := handlers.RuntimeReportHandlerWithCache(runtimeDiagnosticsCache)
	runtimeIncidentReportHandler := handlers.RuntimeIncidentReportHandlerWithCache(runtimeDiagnosticsCache)
	runtimeIncidentEventsHandler := handlers.RuntimeIncidentEventsHandler(cfg.LoggerEndpoint, runtimeMetricsOptions.LoggerHTTPClient)
	runtimeRequestLogsHandler := handlers.RuntimeRequestLogsHandler(cfg.LoggerEndpoint, runtimeMetricsOptions.LoggerHTTPClient)
	reportBuilder := func() handlers.RuntimeReportSummary {
		return runtimeDiagnosticsCache.Report()
	}
	incidentMonitor := incidents.NewMonitor(
		reportBuilder,
		cfg.LoggerEndpoint,
		cfg.IncidentEventWebhookURL,
		envOrEmpty("LOGGER_SHARED_SECRET"),
		envOrEmpty("INCIDENT_EVENT_WEBHOOK_HMAC_SECRET"),
		cfg.IncidentEventSink,
		cfg.IncidentEventWebhookAllowedHosts,
		cfg.IncidentEventInterval,
		cfg.IncidentEventDedupeWindow,
		nil,
	)
	if incidentMonitor != nil {
		incidentMonitor.Start(context.Background())
	}

	r := chi.NewRouter()
	r.Use(middleware.HTTPMetrics(metricsStore))
	r.Use(middleware.TraceID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(5 * time.Second))
	r.Use(middleware.CORS)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.RequestBodyLimit(1 << 20))
	r.Use(middleware.LoadShedding(metricsStore, cfg.MaxInFlightRequests, cfg.LoadShedExemptPrefixes))
	r.Use(middleware.AsyncRequestLogger)

	authRateLimiter, apiRateLimiter := buildRateLimiters(cfg, deps.redisClient)

	r.Get("/health/live", handlers.LiveHandler)
	r.Get("/health/ready", readyEndpoints.Ready)
	r.With(authRateLimiter, middleware.AuthenticateJWT, middleware.RequireRoles("admin")).Post("/health/ready/invalidate", readyEndpoints.Invalidate)
	r.Get("/health", readyEndpoints.Ready)
	r.Get("/metrics", handlers.MetricsHandler(metricsStore))
	r.With(authRateLimiter).Post("/api/v1/auth/token", handlers.IssueToken)

	r.Route("/api/v1", func(api chi.Router) {
		api.Use(apiRateLimiter)
		api.Use(middleware.AuthenticateJWT)
		registerProtectedRoutes(api, usersHandler, computeHandler, runtimeConfigHandler, runtimeMetricsHandler, runtimeReportHandler, runtimeIncidentReportHandler, runtimeIncidentEventsHandler, runtimeRequestLogsHandler)
	})

	if cfg.LegacyAPIEnabled {
		legacyDeprecation := middleware.DeprecatedAPIVersion("/api/v1", cfg.LegacyDeprecationDate, cfg.LegacySunsetDate)
		r.With(authRateLimiter, legacyDeprecation).Post("/api/auth/token", handlers.IssueToken)
		r.Route("/api", func(api chi.Router) {
			api.Use(legacyDeprecation)
			api.Use(apiRateLimiter)
			api.Use(middleware.AuthenticateJWT)
			registerProtectedRoutes(api, usersHandler, computeHandler, runtimeConfigHandler, runtimeMetricsHandler, runtimeReportHandler, runtimeIncidentReportHandler, runtimeIncidentEventsHandler, runtimeRequestLogsHandler)
		})
	}

	cleanup := func() {
		if incidentMonitor != nil {
			incidentMonitor.Close()
		}
	}
	return r, cleanup
}

func buildRateLimiters(cfg runtimeConfig, redisClient *redis.Client) (func(http.Handler) http.Handler, func(http.Handler) http.Handler) {
	authRateLimiter := middleware.RateLimitByIP(cfg.AuthRateLimitPerMinute, time.Minute)
	apiRateLimiter := middleware.RateLimitByIP(cfg.APIRateLimitPerMinute, time.Minute)

	if cfg.RateLimitStore == "redis" && redisClient != nil {
		authRateLimiter = middleware.RateLimitByIPDistributedWithFailureMode(redisClient, "auth", cfg.AuthRateLimitPerMinute, time.Minute, cfg.RedisFailureMode)
		apiRateLimiter = middleware.RateLimitByIPDistributedWithFailureMode(redisClient, "api", cfg.APIRateLimitPerMinute, time.Minute, cfg.RedisFailureMode)
	} else if cfg.RateLimitStore == "redis" {
		log.Printf("rate limiter store requested=redis but redis client unavailable, falling back to memory")
	}

	return authRateLimiter, apiRateLimiter
}

func registerProtectedRoutes(
	api chi.Router,
	usersHandler *handlers.UsersHandler,
	computeHandler *handlers.ComputeHandler,
	runtimeConfigHandler http.HandlerFunc,
	runtimeMetricsHandler http.HandlerFunc,
	runtimeReportHandler http.HandlerFunc,
	runtimeIncidentReportHandler http.HandlerFunc,
	runtimeIncidentEventsHandler http.HandlerFunc,
	runtimeRequestLogsHandler http.HandlerFunc,
) {
	api.With(middleware.RequireRoles("user", "admin")).Get("/users", usersHandler.List)
	api.With(middleware.RequireRoles("user", "admin")).Post("/compute/fibonacci", computeHandler.Fibonacci)
	api.With(middleware.RequireRoles("user", "admin")).Post("/compute/hash", computeHandler.Hash)
	api.With(middleware.RequireRoles("admin")).Get("/admin/ping", handlers.AdminPing)
	api.With(middleware.RequireRoles("admin")).Get("/admin/runtime-config", runtimeConfigHandler)
	api.With(middleware.RequireRoles("admin")).Get("/admin/runtime-metrics", runtimeMetricsHandler)
	api.With(middleware.RequireRoles("admin")).Get("/admin/runtime-report", runtimeReportHandler)
	api.With(middleware.RequireRoles("admin")).Get("/admin/runtime-incident-report", runtimeIncidentReportHandler)
	api.With(middleware.RequireRoles("admin")).Get("/admin/incident-events", runtimeIncidentEventsHandler)
	api.With(middleware.RequireRoles("admin")).Get("/admin/request-logs", runtimeRequestLogsHandler)
}

func envOrEmpty(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}
