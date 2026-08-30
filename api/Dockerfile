FROM golang:1.26-bookworm AS builder
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=~/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build go build -o /app/bin/api ./src

FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/bin/api /app/api
EXPOSE 8080
CMD ["/app/api"]
