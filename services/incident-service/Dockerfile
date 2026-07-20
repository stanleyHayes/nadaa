FROM golang:1.25-alpine AS build

WORKDIR /src
COPY services/incident-service/go.mod ./
RUN go mod download
COPY services/incident-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/incident-service ./cmd/server

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/incident-service /usr/local/bin/incident-service
# Media upload directory (NADAA_INCIDENT_MEDIA_STORAGE_PATH=/app/uploads/media
# in render.yaml) must be writable by appuser or uploads fail on MkdirAll.
RUN mkdir -p /app/uploads/media && chown -R appuser /app/uploads

USER appuser
EXPOSE 8084
ENTRYPOINT ["incident-service"]
