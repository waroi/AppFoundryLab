FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api-gateway ./services/api-gateway/cmd/api-gateway

FROM alpine:3.20
RUN adduser -D appuser
WORKDIR /app
COPY --from=builder /out/api-gateway /app/api-gateway
COPY infrastructure/certs /app/certs
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/api-gateway"]
