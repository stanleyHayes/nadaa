FROM golang:1.25-alpine AS build

WORKDIR /src
COPY services/alert-service/go.mod ./
RUN go mod download
COPY services/alert-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/alert-service .

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/alert-service /usr/local/bin/alert-service

USER appuser
EXPOSE 8089
ENTRYPOINT ["alert-service"]
