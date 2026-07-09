FROM golang:1.25-alpine AS build

WORKDIR /src
COPY services/missing-person-service/go.mod ./
RUN go mod download
COPY services/missing-person-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/missing-person-service ./cmd/server

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/missing-person-service /usr/local/bin/missing-person-service

USER appuser
EXPOSE 8101
ENTRYPOINT ["missing-person-service"]
