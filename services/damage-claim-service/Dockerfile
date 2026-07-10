FROM golang:1.25-alpine AS build

WORKDIR /src
COPY services/damage-claim-service/go.mod ./
RUN go mod download
COPY services/damage-claim-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/damage-claim-service ./cmd/server

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/damage-claim-service /usr/local/bin/damage-claim-service

USER appuser
EXPOSE 8098
ENTRYPOINT ["damage-claim-service"]
