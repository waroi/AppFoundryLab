use sha2::{Digest, Sha256};
use std::{env, fs, net::SocketAddr};
use tokio::task;
use tonic::{transport::Server, Request, Response, Status};
use tracing::{error, info};

pub mod worker {
    tonic::include_proto!("worker");
}

use worker::worker_service_server::{WorkerService, WorkerServiceServer};
use worker::{
    ComputeFibonacciRequest, ComputeFibonacciResponse, ComputeHashRequest, ComputeHashResponse,
    HealthRequest, HealthResponse,
};

#[derive(Default)]
struct WorkerServiceImpl;

#[tonic::async_trait]
impl WorkerService for WorkerServiceImpl {
    async fn health(
        &self,
        _request: Request<HealthRequest>,
    ) -> Result<Response<HealthResponse>, Status> {
        Ok(Response::new(HealthResponse {
            status: "ok".to_string(),
            service: "calculator-grpc".to_string(),
        }))
    }

    async fn compute_fibonacci(
        &self,
        request: Request<ComputeFibonacciRequest>,
    ) -> Result<Response<ComputeFibonacciResponse>, Status> {
        let n = request.into_inner().n;
        if n > 93 {
            return Err(Status::invalid_argument("n must be between 0 and 93"));
        }

        let value = task::spawn_blocking(move || fib(n))
            .await
            .map_err(|_| Status::internal("fibonacci computation failed"))?;

        Ok(Response::new(ComputeFibonacciResponse { n, value }))
    }

    async fn compute_hash(
        &self,
        request: Request<ComputeHashRequest>,
    ) -> Result<Response<ComputeHashResponse>, Status> {
        let input = request.into_inner().input;
        if input.is_empty() {
            return Err(Status::invalid_argument("input cannot be empty"));
        }

        let mut hasher = Sha256::new();
        hasher.update(input.as_bytes());
        let hash = format!("{:x}", hasher.finalize());

        Ok(Response::new(ComputeHashResponse {
            algorithm: "sha256".to_string(),
            hash,
        }))
    }
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(tracing_subscriber::EnvFilter::from_default_env())
        .init();

    let port = env::var("CALCULATOR_PORT").unwrap_or_else(|_| "7070".to_string());
    let addr: SocketAddr = match format!("0.0.0.0:{port}").parse() {
        Ok(v) => v,
        Err(err) => {
            error!("invalid CALCULATOR_PORT: {err}");
            return;
        }
    };

    info!("calculator gRPC listening on {}", addr);

    let service = WorkerServiceImpl;
    let mut builder = Server::builder();

    if env::var("WORKER_GRPC_TLS_MODE").unwrap_or_else(|_| "mtls".to_string()) != "insecure" {
        let cert_path =
            env::var("WORKER_GRPC_SERVER_CERT_PATH").unwrap_or_else(|_| "/app/certs/dev/server.crt".to_string());
        let key_path =
            env::var("WORKER_GRPC_SERVER_KEY_PATH").unwrap_or_else(|_| "/app/certs/dev/server.key".to_string());
        let client_ca_path = env::var("WORKER_GRPC_CLIENT_CA_CERT_PATH")
            .unwrap_or_else(|_| "/app/certs/dev/ca.crt".to_string());

        let cert = match fs::read(&cert_path) {
            Ok(bytes) => bytes,
            Err(err) => {
                error!("failed to read server cert {}: {}", cert_path, err);
                return;
            }
        };

        let key = match fs::read(&key_path) {
            Ok(bytes) => bytes,
            Err(err) => {
                error!("failed to read server key {}: {}", key_path, err);
                return;
            }
        };

        let client_ca = match fs::read(&client_ca_path) {
            Ok(bytes) => bytes,
            Err(err) => {
                error!("failed to read client ca {}: {}", client_ca_path, err);
                return;
            }
        };

        let identity = tonic::transport::Identity::from_pem(cert, key);
        let client_ca_cert = tonic::transport::Certificate::from_pem(client_ca);
        let tls_config = tonic::transport::ServerTlsConfig::new()
            .identity(identity)
            .client_ca_root(client_ca_cert);

        match builder.tls_config(tls_config) {
            Ok(configured) => {
                builder = configured;
            }
            Err(err) => {
                error!("failed to apply tls config: {}", err);
                return;
            }
        }
    }

    if let Err(err) = builder
        .add_service(WorkerServiceServer::new(service))
        .serve(addr)
        .await
    {
        error!("server error: {}", err);
    }
}

fn fib(n: u32) -> u64 {
    if n <= 1 {
        return n as u64;
    }

    let mut a: u64 = 0;
    let mut b: u64 = 1;

    for _ in 2..=n {
        let c = a.saturating_add(b);
        a = b;
        b = c;
    }

    b
}

#[cfg(test)]
mod tests {
    use super::fib;

    #[test]
    fn fib_returns_expected_values() {
        assert_eq!(fib(0), 0);
        assert_eq!(fib(1), 1);
        assert_eq!(fib(10), 55);
        assert_eq!(fib(20), 6765);
    }

    #[test]
    fn fib_upper_bound_value_is_stable() {
        assert_eq!(fib(93), 12200160415121876738);
    }
}
