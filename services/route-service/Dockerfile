FROM golang:1.25-alpine AS build

WORKDIR /src
COPY services/route-service/go.mod ./
RUN go mod download
COPY services/route-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/route-service ./cmd/server

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/route-service /usr/local/bin/route-service

USER appuser
EXPOSE 8096
ENTRYPOINT ["route-service"]
