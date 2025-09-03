# syntax=docker/dockerfile:1
FROM golang:1.22 as build

WORKDIR /app
COPY go.mod .
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod go mod tidy && go build -o /out/server ./cmd/server

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=build /out/server /app/server
# app expects configs/private-key.pem mounted at runtime
COPY .env.example /app/.env.example
EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
