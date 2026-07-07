FROM golang:1.25-alpine AS build

WORKDIR /src
COPY services/notification-service/go.mod ./
RUN go mod download
COPY services/notification-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/notification-service ./cmd/server

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/notification-service /usr/local/bin/notification-service

USER appuser
EXPOSE 8090
ENTRYPOINT ["notification-service"]
