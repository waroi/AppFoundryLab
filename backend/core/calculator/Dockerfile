FROM rust:1.85-alpine AS builder
WORKDIR /app
RUN apk add --no-cache musl-dev build-base perl
COPY core/calculator/Cargo.toml core/calculator/build.rs /app/
COPY core/calculator/src /app/src
COPY proto /proto
RUN cargo build --release

FROM alpine:3.23
RUN adduser -D appuser
WORKDIR /app
COPY --from=builder /app/target/release/calculator /app/calculator
COPY infrastructure/certs /app/certs
USER appuser
EXPOSE 7070
ENTRYPOINT ["/app/calculator"]
