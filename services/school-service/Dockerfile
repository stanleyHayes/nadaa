FROM golang:1.25-alpine AS build

WORKDIR /src
COPY services/school-service/go.mod ./
RUN go mod download
COPY services/school-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/school-service ./cmd/server

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/school-service /usr/local/bin/school-service

USER appuser
EXPOSE 8097
ENTRYPOINT ["school-service"]
