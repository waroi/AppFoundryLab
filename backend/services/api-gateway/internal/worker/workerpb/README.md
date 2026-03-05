This package is generated from `backend/proto/worker.proto`.

Regenerate files with:

```bash
./scripts/gen-worker-stubs.sh
```

Do not edit `worker.pb.go` or `worker_grpc.pb.go` manually.

Generation modes:
- Modern (`protoc-gen-go` + `protoc-gen-go-grpc`): `worker.pb.go` + `worker_grpc.pb.go`
- Legacy fallback (`plugins=grpc`): only `worker.pb.go`
