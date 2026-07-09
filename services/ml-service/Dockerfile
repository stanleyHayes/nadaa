FROM golang:1.25-alpine AS build

WORKDIR /src
COPY services/ml-service/go.mod ./
RUN go mod download
COPY services/ml-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/ml-service ./cmd/server

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/ml-service /usr/local/bin/ml-service
COPY data/flood-risk/models /app/data/flood-risk/models
COPY data/flood-risk/generated /app/data/flood-risk/generated

USER appuser
EXPOSE 8094
ENV NADAA_ML_MODEL_DIR=/app/data/flood-risk/models
ENTRYPOINT ["ml-service"]
