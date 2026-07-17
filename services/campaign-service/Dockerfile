FROM golang:1.25-alpine AS build

WORKDIR /src
COPY services/campaign-service/go.mod ./
RUN go mod download
COPY services/campaign-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/campaign-service ./cmd/server

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/campaign-service /usr/local/bin/campaign-service

USER appuser
EXPOSE 8103
ENTRYPOINT ["campaign-service"]
