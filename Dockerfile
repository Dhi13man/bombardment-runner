FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY app/go.mod app/go.sum ./
RUN go mod download
COPY app/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bombardment .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates && \
    addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /build/bombardment .
COPY app/src/app/ui ./src/app/ui
RUN mkdir -p /app/data /app/responses && chown -R appuser:appgroup /app
USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --retries=3 --start-period=5s \
  CMD wget -qO- http://localhost:8080/v1/ping || exit 1
ENTRYPOINT ["./bombardment"]
CMD ["server", "--bind-addr", "0.0.0.0"]
