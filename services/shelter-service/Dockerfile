FROM golang:1.25-alpine AS build

WORKDIR /src
COPY services/shelter-service/go.mod ./
RUN go mod download
COPY services/shelter-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/shelter-service ./cmd/server

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/shelter-service /usr/local/bin/shelter-service

USER appuser
EXPOSE 8093
ENTRYPOINT ["shelter-service"]
