FROM golang:1.25-alpine AS build

WORKDIR /src
COPY services/donation-service/go.mod ./
RUN go mod download
COPY services/donation-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/donation-service .

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/donation-service /usr/local/bin/donation-service

USER appuser
EXPOSE 8100
ENTRYPOINT ["donation-service"]
