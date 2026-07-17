FROM golang:1.25-alpine AS build

WORKDIR /src
COPY services/imagery-service/go.mod ./
RUN go mod download
COPY services/imagery-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/imagery-service ./cmd/server

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/imagery-service /usr/local/bin/imagery-service
# Uploads directory (IMAGERY_STORAGE_PATH=/app/uploads) must be writable by
# appuser or the service crash-loops on MkdirAll at boot.
RUN mkdir -p /app/uploads && chown appuser /app/uploads

USER appuser
EXPOSE 8099
ENTRYPOINT ["imagery-service"]
