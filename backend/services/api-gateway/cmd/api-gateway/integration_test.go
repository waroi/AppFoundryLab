package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/handlers"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/metrics"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/middleware"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/worker"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/worker/workerpb"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
)

type tokenResponse struct {
	AccessToken string `json:"accessToken"`
}

type fakeWorkerServer struct {
	workerpb.UnimplementedWorkerServiceServer
}

func (s *fakeWorkerServer) Health(context.Context, *workerpb.HealthRequest) (*workerpb.HealthResponse, error) {
	return &workerpb.HealthResponse{Status: "ok", Service: "fake-worker"}, nil
}

func (s *fakeWorkerServer) ComputeFibonacci(context.Context, *workerpb.ComputeFibonacciRequest) (*workerpb.ComputeFibonacciResponse, error) {
	return &workerpb.ComputeFibonacciResponse{N: 1, Value: 1}, nil
}

func (s *fakeWorkerServer) ComputeHash(_ context.Context, req *workerpb.ComputeHashRequest) (*workerpb.ComputeHashResponse, error) {
	return &workerpb.ComputeHashResponse{
		Algorithm: "test-hash",
		Hash:      "hashed:" + req.Input,
	}, nil
}

func TestIntegrationAuthProtectedWorkerLoggerMetrics(t *testing.T) {
	t.Setenv("JWT_SECRET", "integration-test-secret")
	t.Setenv("JWT_ISSUER", "appfoundrylab")
	t.Setenv("JWT_AUDIENCE", "appfoundrylab-clients")
	t.Setenv("JWT_LEEWAY_SECONDS", "15")
	t.Setenv("BOOTSTRAP_USER", "developer")
	t.Setenv("BOOTSTRAP_USER_PASSWORD", "developer_dev_password")
	t.Setenv("LOGGER_SHARED_SECRET", "logger-secret")

	logger := newFakeLoggerServer()
	defer logger.Close()
	t.Setenv("LOGGER_ENDPOINT", logger.URL+"/ingest")

	grpcAddr, stopGRPC := startFakeWorkerGRPC(t)
	defer stopGRPC()
	t.Setenv("WORKER_GRPC_ENDPOINT", grpcAddr)
	t.Setenv("WORKER_GRPC_TLS_MODE", "insecure")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	workerClient, err := worker.NewClient(ctx)
	if err != nil {
		t.Fatalf("worker client init failed: %v", err)
	}
	defer func() { _ = workerClient.Close() }()

	apiServer := startHTTPServer(t, buildIntegrationRouter(workerClient))
	defer apiServer.Close()

	accessToken := issueAccessToken(t, apiServer.URL)
	callProtectedComputeHash(t, apiServer.URL, accessToken)

	if !waitForIngest(logger, 2*time.Second) {
		t.Fatalf("logger ingest was not observed within timeout")
	}

	resp, err := http.Get(logger.URL + "/metrics")
	if err != nil {
		t.Fatalf("logger metrics request failed: %v", err)
	}
	defer resp.Body.Close()

	var metricsPayload map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&metricsPayload); err != nil {
		t.Fatalf("decode logger metrics failed: %v", err)
	}
	if metricsPayload["ingestCount"] < 1 {
		t.Fatalf("expected ingestCount >= 1, got %d", metricsPayload["ingestCount"])
	}
}

func TestLegacyAPIIncludesDeprecationHeaders(t *testing.T) {
	t.Setenv("JWT_SECRET", "integration-test-secret")
	t.Setenv("JWT_ISSUER", "appfoundrylab")
	t.Setenv("JWT_AUDIENCE", "appfoundrylab-clients")
	t.Setenv("JWT_LEEWAY_SECONDS", "15")
	t.Setenv("BOOTSTRAP_USER", "developer")
	t.Setenv("BOOTSTRAP_USER_PASSWORD", "developer_dev_password")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	workerClient, err := worker.NewClient(ctx)
	if err != nil {
		// legacy token endpoint test does not depend on worker
		workerClient = nil
	}
	if workerClient != nil {
		defer func() { _ = workerClient.Close() }()
	}

	apiServer := startHTTPServer(t, buildIntegrationRouter(workerClient))
	defer apiServer.Close()

	body := []byte(`{"username":"developer","password":"developer_dev_password"}`)
	resp, err := http.Post(apiServer.URL+"/api/auth/token", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("legacy token request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected legacy token endpoint 200, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Deprecation") == "" {
		t.Fatal("expected Deprecation header on legacy endpoint")
	}
	if resp.Header.Get("Sunset") == "" {
		t.Fatal("expected Sunset header on legacy endpoint")
	}
}

func TestReadyInvalidateRequiresAdminRole(t *testing.T) {
	t.Setenv("JWT_SECRET", "integration-test-secret")
	t.Setenv("JWT_ISSUER", "appfoundrylab")
	t.Setenv("JWT_AUDIENCE", "appfoundrylab-clients")
	t.Setenv("JWT_LEEWAY_SECONDS", "15")
	t.Setenv("BOOTSTRAP_ADMIN_USER", "admin")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "admin_dev_password")
	t.Setenv("BOOTSTRAP_USER", "developer")
	t.Setenv("BOOTSTRAP_USER_PASSWORD", "developer_dev_password")

	apiServer := startHTTPServer(t, buildIntegrationRouter(nil))
	defer apiServer.Close()

	noAuthReq, _ := http.NewRequest(http.MethodPost, apiServer.URL+"/health/ready/invalidate", nil)
	noAuthResp, err := http.DefaultClient.Do(noAuthReq)
	if err != nil {
		t.Fatalf("invalidate request without auth failed: %v", err)
	}
	defer noAuthResp.Body.Close()
	if noAuthResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without auth, got %d", noAuthResp.StatusCode)
	}

	userToken := issueAccessTokenWithCredentials(t, apiServer.URL, "developer", "developer_dev_password")
	userReq, _ := http.NewRequest(http.MethodPost, apiServer.URL+"/health/ready/invalidate", nil)
	userReq.Header.Set("Authorization", "Bearer "+userToken)
	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil {
		t.Fatalf("invalidate request with user token failed: %v", err)
	}
	defer userResp.Body.Close()
	if userResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for user role, got %d", userResp.StatusCode)
	}

	adminToken := issueAccessTokenWithCredentials(t, apiServer.URL, "admin", "admin_dev_password")
	adminReq, _ := http.NewRequest(http.MethodPost, apiServer.URL+"/health/ready/invalidate", nil)
	adminReq.Header.Set("Authorization", "Bearer "+adminToken)
	adminResp, err := http.DefaultClient.Do(adminReq)
	if err != nil {
		t.Fatalf("invalidate request with admin token failed: %v", err)
	}
	defer adminResp.Body.Close()
	if adminResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for admin role, got %d", adminResp.StatusCode)
	}
}

func buildIntegrationRouter(workerClient *worker.Client) http.Handler {
	computeHandler := handlers.NewComputeHandler(workerClient)
	readyEndpoints := handlers.NewReadyEndpoints(workerClient)
	metricsStore := metrics.NewStore()
	legacyDeprecation := middleware.DeprecatedAPIVersion("/api/v1", "Fri, 27 Feb 2026 00:00:00 GMT", "Tue, 30 Jun 2026 23:59:59 GMT")

	r := chi.NewRouter()
	r.Use(middleware.HTTPMetrics(metricsStore))
	r.Use(middleware.TraceID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(5 * time.Second))
	r.Use(middleware.AsyncRequestLogger)

	r.Post("/api/v1/auth/token", handlers.IssueToken)
	r.With(legacyDeprecation).Post("/api/auth/token", handlers.IssueToken)
	r.With(middleware.AuthenticateJWT, middleware.RequireRoles("admin")).Post("/health/ready/invalidate", readyEndpoints.Invalidate)
	r.Get("/metrics", handlers.MetricsHandler(metricsStore))
	r.Route("/api/v1", func(api chi.Router) {
		api.Use(middleware.AuthenticateJWT)
		api.With(middleware.RequireRoles("user", "admin")).Post("/compute/hash", computeHandler.Hash)
	})
	r.Route("/api", func(api chi.Router) {
		api.Use(legacyDeprecation)
		api.Use(middleware.AuthenticateJWT)
		api.With(middleware.RequireRoles("user", "admin")).Post("/compute/hash", computeHandler.Hash)
	})
	return r
}

func issueAccessToken(t *testing.T, baseURL string) string {
	t.Helper()
	return issueAccessTokenWithCredentials(t, baseURL, "developer", "developer_dev_password")
}

func issueAccessTokenWithCredentials(t *testing.T, baseURL, username, password string) string {
	t.Helper()
	body := []byte(`{"username":"` + username + `","password":"` + password + `"}`)
	resp, err := http.Post(baseURL+"/api/v1/auth/token", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("token request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected token endpoint 200, got %d", resp.StatusCode)
	}

	var token tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		t.Fatalf("decode token response failed: %v", err)
	}
	if token.AccessToken == "" {
		t.Fatal("access token is empty")
	}
	return token.AccessToken
}

func callProtectedComputeHash(t *testing.T, baseURL, accessToken string) {
	t.Helper()
	body := []byte(`{"input":"abc"}`)
	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/compute/hash", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("build compute request failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("compute request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected compute endpoint 200, got %d", resp.StatusCode)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode compute response failed: %v", err)
	}
	if payload["algorithm"] != "test-hash" {
		t.Fatalf("unexpected algorithm: %v", payload["algorithm"])
	}
}

type fakeLogger struct {
	URL    string
	Close  func()
	count  int
	countM sync.Mutex
}

func newFakeLoggerServer() *fakeLogger {
	f := &fakeLogger{}
	mux := http.NewServeMux()
	mux.HandleFunc("/ingest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		f.countM.Lock()
		f.count++
		f.countM.Unlock()
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status":"queued"}`))
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		f.countM.Lock()
		count := f.count
		f.countM.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int{"ingestCount": count})
	})

	server := startHTTPServer(nil, mux)
	f.URL = server.URL
	f.Close = server.Close
	return f
}

func waitForIngest(logger *fakeLogger, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		logger.countM.Lock()
		count := logger.count
		logger.countM.Unlock()
		if count > 0 {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

func startFakeWorkerGRPC(t *testing.T) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen fake worker failed: %v", err)
	}

	grpcServer := grpc.NewServer()
	workerpb.RegisterWorkerServiceServer(grpcServer, &fakeWorkerServer{})

	go func() {
		_ = grpcServer.Serve(lis)
	}()

	stop := func() {
		grpcServer.Stop()
		_ = lis.Close()
	}
	return lis.Addr().String(), stop
}

type httpTestServer struct {
	URL   string
	Close func()
}

func startHTTPServer(t *testing.T, handler http.Handler) *httpTestServer {
	if handler == nil {
		handler = http.NewServeMux()
	}
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if t != nil {
			t.Fatalf("http listen failed: %v", err)
		}
		panic(err)
	}

	server := &http.Server{Handler: handler}
	go func() {
		_ = server.Serve(lis)
	}()

	return &httpTestServer{
		URL: "http://" + lis.Addr().String(),
		Close: func() {
			_ = server.Close()
			_ = lis.Close()
		},
	}
}
