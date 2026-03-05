cd backend && go test -bench=BenchmarkDispatchEvent -benchmem ./services/api-gateway/internal/incidents/... | tail -n 5
