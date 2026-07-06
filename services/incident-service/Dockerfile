FROM golang:1.25-alpine AS build

WORKDIR /src
COPY services/incident-service/go.mod ./
RUN go mod download
COPY services/incident-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/incident-service .

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/incident-service /usr/local/bin/incident-service

USER appuser
EXPOSE 8084
ENTRYPOINT ["incident-service"]
