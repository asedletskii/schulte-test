# syntax=docker/dockerfile:1
# Сборка backend (контекст — корень репозитория).
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY backend/ ./
# go.sum формируется при первой сборке
RUN go mod tidy && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN adduser -D -u 10001 app
USER app
WORKDIR /app
COPY --from=build /out/server /app/server
EXPOSE 8080
ENTRYPOINT ["/app/server"]
