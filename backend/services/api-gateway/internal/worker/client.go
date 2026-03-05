package worker

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/example/appfoundrylab/backend/pkg/env"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/worker/workerpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client workerpb.WorkerServiceClient
}

func NewClient(_ context.Context) (*Client, error) {
	// grpc.NewClient is non-blocking; connection is established lazily on first RPC call.
	// The context parameter is retained for API compatibility but is not used for dial.
	endpoint := os.Getenv("WORKER_GRPC_ENDPOINT")
	if endpoint == "" {
		endpoint = "calculator:7070"
	}

	tlsMode := env.GetWithDefault("WORKER_GRPC_TLS_MODE", "mtls")

	var dialOpt grpc.DialOption
	if tlsMode == "insecure" {
		dialOpt = grpc.WithTransportCredentials(insecure.NewCredentials())
	} else {
		tlsConfig, err := loadClientTLSConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to configure grpc tls: %w", err)
		}
		dialOpt = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	}

	conn, err := grpc.NewClient(endpoint, dialOpt)
	if err != nil {
		return nil, fmt.Errorf("grpc client create failed: %w", err)
	}

	return &Client{conn: conn, client: workerpb.NewWorkerServiceClient(conn)}, nil
}

func (c *Client) Health(ctx context.Context) (*workerpb.HealthResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
	defer cancel()
	return c.client.Health(ctx, &workerpb.HealthRequest{})
}

func (c *Client) ComputeFibonacci(ctx context.Context, n uint32) (*workerpb.ComputeFibonacciResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.client.ComputeFibonacci(ctx, &workerpb.ComputeFibonacciRequest{N: n})
}

func (c *Client) ComputeHash(ctx context.Context, input string) (*workerpb.ComputeHashResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.client.ComputeHash(ctx, &workerpb.ComputeHashRequest{Input: input})
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func loadClientTLSConfig() (*tls.Config, error) {
	caCertPath := env.GetWithDefault("WORKER_GRPC_CA_CERT_PATH", "backend/infrastructure/certs/dev/ca.crt")
	caCertPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("read ca cert: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCertPEM) {
		return nil, fmt.Errorf("append ca cert failed")
	}

	clientCertPath := env.GetWithDefault("WORKER_GRPC_CLIENT_CERT_PATH", "backend/infrastructure/certs/dev/client.crt")
	clientKeyPath := env.GetWithDefault("WORKER_GRPC_CLIENT_KEY_PATH", "backend/infrastructure/certs/dev/client.key")
	clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load client cert/key: %w", err)
	}

	serverName := env.GetWithDefault("WORKER_GRPC_SERVER_NAME", "calculator")
	return &tls.Config{
		RootCAs:      caPool,
		Certificates: []tls.Certificate{clientCert},
		ServerName:   serverName,
		MinVersion:   tls.VersionTLS13,
	}, nil
}
