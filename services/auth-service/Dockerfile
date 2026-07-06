FROM golang:1.25-alpine AS build

WORKDIR /src
COPY services/auth-service/go.mod ./
RUN go mod download
COPY services/auth-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/auth-service .

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/auth-service /usr/local/bin/auth-service

USER appuser
EXPOSE 8082
ENTRYPOINT ["auth-service"]
