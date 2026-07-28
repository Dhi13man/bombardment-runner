FROM node:26-alpine@sha256:e88a35be04478413b7c71c455cd9865de9b9360e1f43456be5951032d7ac1a66 AS frontend
WORKDIR /build
COPY app/package.json app/package-lock.json ./
RUN npm ci
COPY app/scripts/ ./scripts/
COPY app/tailwind.config.js app/tsconfig.json ./
COPY app/src/app/ui/static/src/ ./src/app/ui/static/src/
COPY app/src/app/ui/static/icons/ ./src/app/ui/static/icons/
RUN npm run build

FROM golang:1.25.12-alpine@sha256:56961d79ea8129efddcc0b8643fd8a5416b4e6228cfd477e3fd61deb2672c587 AS builder
WORKDIR /build
COPY app/go.mod app/go.sum ./
RUN go mod download
COPY app/ .
COPY --from=frontend /build/src/app/ui/static/css/ ./src/app/ui/static/css/
COPY --from=frontend /build/src/app/ui/static/js/ ./src/app/ui/static/js/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bombardment .

FROM alpine:3.23@sha256:fd791d74b68913cbb027c6546007b3f0d3bc45125f797758156952bc2d6daf40
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
