FROM golang:1.25-alpine AS build

WORKDIR /src
COPY services/open-data-service/go.mod ./
RUN go mod download
COPY services/open-data-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/open-data-service ./cmd/server

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/open-data-service /usr/local/bin/open-data-service

USER appuser
EXPOSE 8102
ENTRYPOINT ["open-data-service"]
