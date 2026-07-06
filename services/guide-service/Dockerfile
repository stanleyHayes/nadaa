FROM golang:1.25-alpine AS build

WORKDIR /src
COPY services/guide-service/go.mod ./
RUN go mod download
COPY services/guide-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/guide-service .

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/guide-service /usr/local/bin/guide-service

USER appuser
EXPOSE 8086
ENTRYPOINT ["guide-service"]
