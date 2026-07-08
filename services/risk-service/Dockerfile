FROM golang:1.25-alpine AS build

WORKDIR /src
COPY services/risk-service/go.mod ./
RUN go mod download
COPY services/risk-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/risk-service ./cmd/server

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/risk-service /usr/local/bin/risk-service

USER appuser
EXPOSE 8081
ENTRYPOINT ["risk-service"]
