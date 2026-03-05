FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/logger ./services/logger/cmd/logger

FROM alpine:3.20
RUN adduser -D appuser
WORKDIR /app
COPY --from=builder /out/logger /app/logger
USER appuser
EXPOSE 8090
ENTRYPOINT ["/app/logger"]
