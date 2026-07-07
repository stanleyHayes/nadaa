FROM golang:1.25-alpine AS build

WORKDIR /src
COPY services/road-closure-service/go.mod ./
RUN go mod download
COPY services/road-closure-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/road-closure-service .

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/road-closure-service /usr/local/bin/road-closure-service

USER appuser
EXPOSE 8095
ENTRYPOINT ["road-closure-service"]
