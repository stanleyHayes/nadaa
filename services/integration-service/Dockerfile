FROM golang:1.25-alpine AS build

WORKDIR /src
COPY services/integration-service/go.mod ./
RUN go mod download
COPY services/integration-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/integration-service .

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/integration-service /usr/local/bin/integration-service

USER appuser
EXPOSE 8088
ENTRYPOINT ["integration-service"]
